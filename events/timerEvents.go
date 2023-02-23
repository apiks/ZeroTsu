package events

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"image/png"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/embeds"
	"github.com/r-anime/ZeroTsu/entities"
	"github.com/r-anime/ZeroTsu/functionality"
	"github.com/sasha-s/go-deadlock"
	"golang.org/x/sync/errgroup"

	"github.com/bwmarrin/discordgo"
	"github.com/mmcdole/gofeed"

	"github.com/r-anime/ZeroTsu/config"
)

const feedCheckLifespanHours = 720

var (
	DailyScheduleWebhooksMap = &safeWebhooksMap{WebhooksMap: make(map[string]*discordgo.Webhook)}

	DailyScheduleWebhooksMapBlock Block
	redditFeedBlock               Block
	redditFeedWebhookBlock        Block
	remindMesFeedBlock            Block
)

type Block struct {
	deadlock.RWMutex
	Block bool
}

type safeWebhooksMap struct {
	deadlock.RWMutex
	WebhooksMap map[string]*discordgo.Webhook
}

func UpdateDailyScheduleWebhooks() {
	DailyScheduleWebhooksMapBlock.Lock()
	if DailyScheduleWebhooksMapBlock.Block {
		DailyScheduleWebhooksMapBlock.Unlock()
		return
	}
	DailyScheduleWebhooksMapBlock.Block = true
	DailyScheduleWebhooksMapBlock.Unlock()

	// Store all of the valid guilds' valid webhooks in a map
	tempWebhooksMap := make(map[string]*discordgo.Webhook)
	animeSubsMap := entities.SharedInfo.GetAnimeSubsMapCopy()
	for guildID, subs := range animeSubsMap {
		if subs == nil {
			continue
		}
		isGuild := false
		if len(subs) >= 1 && subs[0].GetGuild() {
			isGuild = true
		}
		if !isGuild {
			continue
		}

		guildIDInt, err := strconv.ParseInt(guildID, 10, 64)
		if err != nil {
			continue
		}
		s := config.Mgr.SessionForGuild(guildIDInt)

		// Checks if bot is in target guild
		_, err = s.State.Guild(guildID)
		if err != nil {
			continue
		}

		// Check if bot has required permissions for the daily schedule channel
		newepisodes := db.GetGuildAutopost(guildID, "dailyschedule")
		if newepisodes == (entities.Cha{}) {
			continue
		}
		perms, err := s.State.UserChannelPermissions(s.State.User.ID, newepisodes.GetID())
		if err != nil {
			continue
		}
		if perms&discordgo.PermissionManageWebhooks != discordgo.PermissionManageWebhooks {
			continue
		}
		if perms&discordgo.PermissionViewChannel != discordgo.PermissionViewChannel {
			continue
		}
		if perms&discordgo.PermissionSendMessages != discordgo.PermissionSendMessages {
			continue
		}

		// Get valid webhook
		ws, err := s.ChannelWebhooks(newepisodes.GetID())
		if err != nil {
			continue
		}
		for _, w := range ws {
			if w.User.ID != s.State.User.ID ||
				w.ChannelID != newepisodes.GetID() {
				continue
			}
			tempWebhooksMap[guildID] = w
			break
		}

		if _, ok := tempWebhooksMap[guildID]; ok {
			continue
		}

		// Create the webhook if it doesn't exist
		avatar, err := s.UserAvatarDecode(s.State.User)
		if err != nil {
			continue
		}
		out := new(bytes.Buffer)
		err = png.Encode(out, avatar)
		if err != nil {
			continue
		}
		base64Img := base64.StdEncoding.EncodeToString(out.Bytes())
		wh, err := s.WebhookCreate(newepisodes.GetID(), s.State.User.Username, fmt.Sprintf("data:image/png;base64,%s", base64Img))
		if err != nil {
			continue
		}
		tempWebhooksMap[guildID] = wh
	}

	DailyScheduleWebhooksMap.Lock()
	DailyScheduleWebhooksMap.WebhooksMap = make(map[string]*discordgo.Webhook)
	for guid, w := range tempWebhooksMap {
		DailyScheduleWebhooksMap.WebhooksMap[guid] = w
	}
	DailyScheduleWebhooksMap.Unlock()

	DailyScheduleWebhooksMapBlock.Lock()
	DailyScheduleWebhooksMapBlock.Block = false
	DailyScheduleWebhooksMapBlock.Unlock()
}

func WriteEvents(s *discordgo.Session, _ *discordgo.Ready) {
	var (
		t                time.Time
		randomPlayingMsg string
	)

	for range time.NewTicker(30 * time.Minute).C {
		t = time.Now()
		rand.Seed(t.UnixNano())

		// Updates playing status
		entities.Mutex.RLock()
		if len(config.PlayingMsg) > 1 {
			randomPlayingMsg = config.PlayingMsg[rand.Intn(len(config.PlayingMsg))]
		}
		entities.Mutex.RUnlock()
		if randomPlayingMsg != "" {
			_ = s.UpdateGameStatus(0, randomPlayingMsg)
		}

		// Sends server count to bot list sites if it's the public ZeroTsu
		functionality.SendServers(strconv.Itoa(config.Mgr.GuildCount()), s)
	}
}

func CommonEvents(_ *discordgo.Session, _ *discordgo.Ready) {
	var guildIds []string

	for range time.NewTicker(1 * time.Minute).C {
		GuildIds.RLock()
		for gID := range GuildIds.Ids {
			guildIds = append(guildIds, gID)
		}
		GuildIds.RUnlock()

		// Handles RemindMes
		remindMeHandler(config.Mgr.SessionForDM())

		// Handles Reddit Feeds
		feedWebhookHandler(guildIds)
		feedHandler(guildIds)

		guildIds = []string{}
	}
}

// remindMeHandler handles sending remindMe messages when called if it's time.
func remindMeHandler(s *discordgo.Session) {
	remindMesFeedBlock.Lock()
	if remindMesFeedBlock.Block {
		remindMesFeedBlock.Unlock()
		return
	}
	remindMesFeedBlock.Block = true
	remindMesFeedBlock.Unlock()

	if entities.SharedInfo.GetRemindMesMap() == nil || len(entities.SharedInfo.GetRemindMesMap()) == 0 {
		return
	}

	var writeFlag bool
	for userID, remindMeSlice := range entities.SharedInfo.GetRemindMesMap() {
		if remindMeSlice == nil || remindMeSlice.GetRemindMeSlice() == nil || len(remindMeSlice.GetRemindMeSlice()) == 0 {
			continue
		}

		// Filter in place if needed
		i := 0
		for _, remindMe := range remindMeSlice.GetRemindMeSlice() {
			if remindMe == nil {
				continue
			}

			// Checks if it's time to send message/ping the user
			if time.Now().Sub(remindMe.GetDate()) <= 0 {
				remindMeSlice.GetRemindMeSlice()[i] = remindMe
				i++
				continue
			}

			dm, err := s.UserChannelCreate(userID)
			if err != nil {
				break
			}
			_, err = s.ChannelMessageSend(dm.ID, fmt.Sprintf("RemindMe: %s", remindMe.GetMessage()))
			if err != nil {
				continue
			}

			writeFlag = true
		}
		remindMeSlice.SetRemindMeSlice(remindMeSlice.GetRemindMeSlice()[:i])
	}

	if !writeFlag {
		remindMesFeedBlock.Lock()
		remindMesFeedBlock.Block = false
		remindMesFeedBlock.Unlock()

		return
	}

	err := entities.RemindMeWrite(entities.SharedInfo.GetRemindMesMap())
	if err != nil {
		log.Println(err)
	}

	remindMesFeedBlock.Lock()
	remindMesFeedBlock.Block = false
	remindMesFeedBlock.Unlock()
}

// Fetches reddit feeds and returns the feeds that need to posted for all guilds with webhook
func feedWebhookHandler(guildIds []string) {
	redditFeedWebhookBlock.Lock()
	if redditFeedWebhookBlock.Block {
		redditFeedWebhookBlock.Unlock()
		return
	}
	redditFeedWebhookBlock.Block = true
	redditFeedWebhookBlock.Unlock()

	// Remove expired checks
	for _, guildID := range guildIds {
		var guildFeedChecks = db.GetGuildFeedChecks(guildID)

		// Removes a check if more than its allowed lifespan hours have passed
		for _, feedCheck := range guildFeedChecks {
			dateRemoval := feedCheck.GetDate().Add(feedCheckLifespanHours)
			if time.Since(dateRemoval) > 0 {
				continue
			}

			db.SetGuildFeedCheck(guildID, feedCheck, true)
		}
	}

	var (
		feedsMap       = make(map[string][]entities.Feed)
		parsedFeedsMap = make(map[string]*gofeed.Feed)
		webhooksMap    = make(map[string]*discordgo.Webhook)
	)

	// Get webhooks and feeds
	for _, guildID := range guildIds {
		var guildFeeds = db.GetGuildFeeds(guildID)
		guildIDInt, err := strconv.ParseInt(guildID, 10, 64)
		if err != nil {
			continue
		}
		s := config.Mgr.SessionForGuild(guildIDInt)

		for _, feed := range guildFeeds {
			perms, err := s.State.UserChannelPermissions(s.State.User.ID, feed.GetChannelID())
			if err != nil {
				continue
			}
			if perms&discordgo.PermissionManageWebhooks != discordgo.PermissionManageWebhooks {
				continue
			}
			if perms&discordgo.PermissionViewChannel != discordgo.PermissionViewChannel {
				continue
			}
			if perms&discordgo.PermissionSendMessages != discordgo.PermissionSendMessages {
				continue
			}

			feedsMap[feed.GetChannelID()] = append(feedsMap[feed.GetChannelID()], feed)
			if _, ok := webhooksMap[feed.GetChannelID()]; ok {
				continue
			}

			ws, err := s.ChannelWebhooks(feed.GetChannelID())
			if err != nil {
				continue
			}
			for _, w := range ws {
				if w.User.ID != s.State.User.ID ||
					w.ChannelID != feed.GetChannelID() {
					continue
				}
				webhooksMap[feed.GetChannelID()] = w
				break
			}

			// Create webhook if it doesn't exist and is needed
			if _, ok := webhooksMap[feed.GetChannelID()]; ok {
				continue
			}

			avatar, err := s.UserAvatarDecode(s.State.User)
			if err != nil {
				continue
			}
			out := new(bytes.Buffer)
			err = png.Encode(out, avatar)
			if err != nil {
				continue
			}
			base64Img := base64.StdEncoding.EncodeToString(out.Bytes())
			w, err := s.WebhookCreate(feed.GetChannelID(), s.State.User.Username, fmt.Sprintf("data:image/png;base64,%s", base64Img))
			if err == nil {
				webhooksMap[feed.GetChannelID()] = w
			}
		}
	}
	if len(webhooksMap) == 0 {
		return
	}

	feedParseMap := make(map[string]bool)
	for k, feeds := range feedsMap {
		if _, ok := webhooksMap[k]; !ok {
			continue
		}
		for _, feed := range feeds {
			feedParseMap[fmt.Sprintf("%s/%s", feed.GetSubreddit(), feed.GetPostType())] = true
		}
	}

	for len(feedParseMap) > 0 {
		i := 0
		if len(parsedFeedsMap) == 0 {
			for k := range feedParseMap {
				if i > 14 {
					break
				}

				// Parse the feed
				time.Sleep(time.Second * 2)
				key := strings.TrimSuffix(k, "/")
				key = strings.TrimPrefix(key, "https://www.reddit.com/r/")
				key = strings.TrimPrefix(key, "/")
				feedParser, statusCode, err := common.GetRedditRSSFeed(fmt.Sprintf("https://www.reddit.com/r/%s/.rss", key), 1)
				if err != nil {
					if statusCode == 429 {
						log.Println("HIT REDDIT RATE LIMIT feedWebhookHandler!")
						time.Sleep(10 * time.Minute)
					} else {
						delete(feedParseMap, k)
					}
					continue
				}

				parsedFeedsMap[k] = feedParser
				delete(feedParseMap, k)
				i++
			}
		}

		var (
			parsedFeedsMapCopy = make(map[string]*gofeed.Feed)

			eg            errgroup.Group
			maxGoroutines = 16
			guard         = make(chan struct{}, maxGoroutines)
		)
		for k, v := range parsedFeedsMap {
			parsedFeedsMapCopy[k] = v
		}

		i = 0
		parsedFeedsMap = make(map[string]*gofeed.Feed)
		guard <- struct{}{}
		eg.Go(func() error {
			for k := range feedParseMap {
				if i > 14 {
					break
				}

				// Parse the feed
				time.Sleep(time.Second * 2)
				key := strings.TrimSuffix(k, "/")
				key = strings.TrimPrefix(key, "https://www.reddit.com/r/")
				key = strings.TrimPrefix(key, "/")
				feedParser, statusCode, err := common.GetRedditRSSFeed(fmt.Sprintf("https://www.reddit.com/r/%s/.rss", key), 1)
				if err != nil {
					if statusCode == 429 {
						log.Println("HIT REDDIT RATE LIMIT feedWebhookHandler!")
						time.Sleep(10 * time.Minute)
					} else {
						delete(feedParseMap, k)
					}
					continue
				}

				parsedFeedsMap[k] = feedParser
				delete(feedParseMap, k)
				i++
			}

			<-guard
			return nil
		})

		for _, guildID := range guildIds {
			guid := guildID

			guard <- struct{}{}
			eg.Go(func() error {
				guildFeeds := db.GetGuildFeeds(guid)
				if len(guildFeeds) == 0 {
					<-guard
					return nil
				}

				guildIDInt, err := strconv.ParseInt(guid, 10, 64)
				if err != nil {
					<-guard
					return err
				}
				s := config.Mgr.SessionForGuild(guildIDInt)

				for k, w := range webhooksMap {
					if w.GuildID != guid {
						continue
					}
					breakFromWebhook := false

					var (
						newFeedChecks []entities.FeedCheck
						embedsSlice   []*discordgo.MessageEmbed
					)

					for _, feed := range feedsMap[k] {
						if _, ok := parsedFeedsMapCopy[fmt.Sprintf("%s/%s", feed.GetSubreddit(), feed.GetPostType())]; !ok {
							continue
						}

						// Iterates through each feed parser item to see if it finds something that should be posted
						// var pinnedItems = make(map[*gofeed.Item]bool)
						for _, item := range parsedFeedsMapCopy[fmt.Sprintf("%s/%s", feed.GetSubreddit(), feed.GetPostType())].Items {
							var skip bool

							// Checks if the item has already been posted
							guildFeedChecks := db.GetGuildFeedChecks(guid)
							for _, feedCheck := range guildFeedChecks {
								if feedCheck.GetGUID() == item.GUID &&
									feedCheck.GetFeed().GetChannelID() == feed.GetChannelID() {
									skip = true
									break
								}
							}
							if skip {
								continue
							}

							// Check if author is same and skip if not true
							if feed.GetAuthor() != "" && item.Author != nil && strings.ToLower(item.Author.Name) != fmt.Sprintf("/u/%s", feed.GetAuthor()) {
								continue
							}

							// Check if the item title starts with the set feed title
							if feed.GetTitle() != "" && !strings.HasPrefix(strings.ToLower(item.Title), feed.GetTitle()) {
								continue
							}

							// Save embed
							embedsSlice = append(embedsSlice, embeds.FeedEmbed(&feed, item))
							newFeedChecks = append(newFeedChecks, entities.NewFeedCheck(feed, time.Now(), item.GUID))

							// Use webhook to post embeds if necessary
							if len(embedsSlice) >= 10 {
								_, err := s.WebhookExecute(w.ID, w.Token, false, &discordgo.WebhookParams{
									Embeds: embedsSlice,
								})
								embedsSlice = nil
								if err != nil {
									log.Println("Failed webhookExecute in feedWebhookHandler:", err)
									breakFromWebhook = true
								}

								// Adds that the feeds have been posted
								db.AddGuildFeedChecks(guid, newFeedChecks)
								newFeedChecks = nil
							}

							if breakFromWebhook {
								break
							}
						}
					}
					if breakFromWebhook {
						continue
					}

					// Use webhook to post last embeds available
					if len(embedsSlice) > 0 && len(embedsSlice) <= 10 {
						// Use webhook to post last embeds available
						_, err := s.WebhookExecute(w.ID, w.Token, false, &discordgo.WebhookParams{
							Embeds: embedsSlice,
						})
						if err != nil {
							log.Println("Failed webhookExecute in feedWebhookHandler:", err)
							breakFromWebhook = true
						}

						// Adds that the feeds have been posted
						db.AddGuildFeedChecks(guid, newFeedChecks)
						newFeedChecks = nil
					}
					if breakFromWebhook {
						continue
					}
				}

				<-guard
				return nil
			})
		}

		err := eg.Wait()
		if err != nil {
			log.Println(err)
		}
	}

	redditFeedWebhookBlock.Lock()
	redditFeedWebhookBlock.Block = false
	redditFeedWebhookBlock.Unlock()
}

// Fetches reddit feeds and returns the feeds that need to posted for all guilds no webhook
func feedHandler(guildIds []string) {
	redditFeedBlock.Lock()
	if redditFeedBlock.Block {
		redditFeedBlock.Unlock()
		return
	}
	redditFeedBlock.Block = true
	redditFeedBlock.Unlock()

	for _, guildID := range guildIds {
		var (
			guildFeeds      = db.GetGuildFeeds(guildID)
			guildFeedChecks = db.GetGuildFeedChecks(guildID)
			fp              = gofeed.NewParser()
		)
		fp.Client = &http.Client{
			Transport: &common.UserAgentTransport{RoundTripper: &http.Transport{
				TLSNextProto: map[string]func(authority string, c *tls.Conn) http.RoundTripper{},
			}},
			Timeout: time.Second * 10}
		fp.UserAgent = common.UserAgent

		guildIDInt, err := strconv.ParseInt(guildID, 10, 64)
		if err != nil {
			continue
		}
		s := config.Mgr.SessionForGuild(guildIDInt)

		for _, feed := range guildFeeds {
			// Check if bot has required permissions for this channel
			perms, err := s.State.UserChannelPermissions(s.State.User.ID, feed.GetChannelID())
			if err != nil {
				continue
			}
			if perms&discordgo.PermissionManageWebhooks == discordgo.PermissionManageWebhooks {
				continue
			}
			if perms&discordgo.PermissionViewChannel != discordgo.PermissionViewChannel {
				continue
			}
			if perms&discordgo.PermissionSendMessages != discordgo.PermissionSendMessages {
				continue
			}

			var pinnedItems = make(map[*gofeed.Item]bool)

			// Get the RSS feed
			time.Sleep(4 * time.Second)
			feedParser, statusCode, err := common.GetRedditRSSFeed(fmt.Sprintf("https://www.reddit.com/r/%s/%s/.rss", feed.GetSubreddit(), feed.GetPostType()), 1)
			if err != nil {
				if statusCode == 429 {
					log.Println("HIT REDDIT RATE LIMIT feedWebhookHandler!")
					time.Sleep(10 * time.Minute)
				}
				continue
			}

			// Iterates through each feed parser item to see if it finds something that should be posted
			for _, item := range feedParser.Items {
				var (
					skip   bool
					exists bool
				)

				for _, feedCheck := range guildFeedChecks {
					if feedCheck.GetGUID() == item.GUID &&
						feedCheck.GetFeed().GetChannelID() == feed.GetChannelID() {
						skip = true
						break
					}
				}
				if skip {
					continue
				}

				// Check if author is same and skip if not true
				if feed.GetAuthor() != "" && item.Author != nil && strings.ToLower(item.Author.Name) != fmt.Sprintf("/u/%s", feed.GetAuthor()) {
					continue
				}

				// Check if the item title starts with the set feed title
				if feed.GetTitle() != "" && !strings.HasPrefix(strings.ToLower(item.Title), feed.GetTitle()) {
					continue
				}

				// Stops the iteration if the feed doesn't exist anymore
				guildFeeds = db.GetGuildFeeds(guildID)
				for _, guildFeed := range guildFeeds {
					if guildFeed.GetSubreddit() == feed.GetSubreddit() &&
						guildFeed.GetChannelID() == feed.GetChannelID() {
						exists = true
						break
					}
				}
				guildFeeds = nil
				if !exists {
					break
				}
				exists = false

				// Checks if the item has already been posted
				feedChecks := db.GetGuildFeedChecks(guildID)
				for _, feedCheck := range feedChecks {
					if feedCheck.GetGUID() == item.GUID &&
						feedCheck.GetFeed().GetChannelID() == feed.GetChannelID() {
						exists = true
						break
					}
				}
				feedChecks = nil
				if exists {
					continue
				}
				exists = false

				// Wait for Discord API Rate limit
				time.Sleep(time.Millisecond * 250)

				// Sends the feed item
				message, err := embeds.Feed(s, &feed, item)
				if err != nil {
					continue
				}

				// Adds that the feed has been posted
				db.AddGuildFeedCheck(guildID, entities.NewFeedCheck(feed, time.Now(), item.GUID))

				// Pins/unpins the feed items if necessary
				if !feed.GetPin() {
					continue
				}
				if perms&discordgo.PermissionManageMessages != discordgo.PermissionManageMessages {
					continue
				}
				if _, ok := pinnedItems[item]; ok {
					continue
				}

				pins, err := s.ChannelMessagesPinned(message.ChannelID)
				if err != nil {
					continue
				}

				// Unpins if necessary
				for _, pin := range pins {
					// Checks for whether the pin is one that should be unpinned
					if pin.Author.ID != s.State.User.ID {
						continue
					}
					if len(pin.Embeds) == 0 {
						continue
					}
					if pin.Embeds[0].Author == nil {
						continue
					}
					if !strings.HasPrefix(strings.ToLower(pin.Embeds[0].Author.URL), fmt.Sprintf("https://www.reddit.com/r/%s/comments/", feed.GetSubreddit())) {
						continue
					}

					_ = s.ChannelMessageUnpin(pin.ChannelID, pin.ID)
				}
				pins = nil

				// Pins
				_ = s.ChannelMessagePin(message.ChannelID, message.ID)
			}
		}
	}

	redditFeedBlock.Lock()
	redditFeedBlock.Block = false
	redditFeedBlock.Unlock()
}
