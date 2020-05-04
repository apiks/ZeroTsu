package events

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/embeds"
	"github.com/r-anime/ZeroTsu/entities"
	"github.com/r-anime/ZeroTsu/functionality"
	"log"
	"math/rand"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mmcdole/gofeed"

	"github.com/r-anime/ZeroTsu/config"
)

const feedCheckLifespanHours = 2160

var redditFeedBlock bool

// Write Events
func WriteEvents(s *discordgo.Session, _ *discordgo.Ready) {
	var (
		t time.Time
		randomPlayingMsg string

		guild *discordgo.Guild
		roles []*discordgo.Role

		guildIds []string
		guildID string

		guildChannelStats map[string]entities.Channel
		stat entities.Channel
		roleUserAmount int
		emptyStatMap = make(map[string]int)

		emojiStats map[string]entities.Emoji
		userChangeStats map[string]int
		verifiedStats map[string]int

		err error
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
			_ = s.UpdateStatus(0, randomPlayingMsg)
		}

		for _, guild = range s.State.Guilds {
			if guild == nil {
				continue
			}
			guildIds = append(guildIds, guild.ID)
		}

		for _, guildID = range guildIds {
			entities.HandleNewGuild(guildID)

			// Updates BOT nickname
			DynamicNicknameChange(s, guildID)

			// Clears up spoilerRoles.json
			err = cleanSpoilerRoles(s, guildID)
			if err != nil {
				log.Println(err)
			}

			// Fetches all server roles
			roles, err = s.GuildRoles(guildID)
			if err != nil {
				log.Println(err)
			}

			// Updates optin role stat
			guildChannelStats = db.GetGuildChannelStats(guildID)
			if roles != nil {
				guild, err = s.Guild(guildID)
				if err == nil {
					for _, stat = range guildChannelStats {
						if stat.GetRoleCountMap() == nil {
							stat.SetRoleCountMap(emptyStatMap)
						}
						if stat.GetOptin() {
							roleUserAmount = common.GetRoleUserAmount(guild, roles, stat.GetName())
							stat.SetRoleCount(t.Format(common.ShortDateFormat), roleUserAmount)
						}
					}
				}
			}
			// Writes channel stats to disk
			db.SetGuildChannelStats(guildID, guildChannelStats)

			// Writes emoji stats to disk
			emojiStats = db.GetGuildEmojiStats(guildID)
			db.SetGuildEmojiStats(guildID, emojiStats)

			// Writes user gain stats to disk
			userChangeStats = db.GetGuildUserChangeStats(guildID)
			db.SetGuildUserChangeStats(guildID, userChangeStats)

			// Writes verified stats to disk
			if config.Website != "" {
				verifiedStats = db.GetGuildVerifiedStats(guildID)
				db.SetGuildVerifiedStats(guildID, verifiedStats)
			}
		}

		// Sends server count to bot list sites if it's the public ZeroTsu
		functionality.SendServers(s)

		guildIds = []string{}
	}
}

// Common Timer Events
func CommonEvents(s *discordgo.Session, _ *discordgo.Ready) {
	var (
		guildIds []string
		guildID string
		guild *discordgo.Guild

		memberInfo map[string]entities.UserInfo
	)

	for range time.NewTicker(1 * time.Minute).C {
		for _, guild = range s.State.Guilds {
			guildIds = append(guildIds, guild.ID)
		}

		for _, guildID = range guildIds {
			memberInfo = entities.Guilds.DB[guildID].GetMemberInfoMap()
			db.SetGuildMemberInfo(guildID, memberInfo)

			// Handles Unbans and Unmutes
			punishmentHandler(s, guildID)

			// Handles RemindMes
			remindMeHandler(s, guildID)
		}

		// Handles Reddit Feeds
		feedHandler(s, guildIds)

		guildIds = []string{}
	}
}

func punishmentHandler(s *discordgo.Session, guildID string) {
	// Fetches all server bans so it can check if the memberInfo Username is banned there (whether he's been manually unbanned for example)
	bans, err := s.GuildBans(guildID)
	if err != nil {
		return
	}

	t := time.Now()

	// Checks if there are punishedUsers in this guild
	punishedUsers := db.GetGuildPunishedUsers(guildID)
	if punishedUsers == nil || len(punishedUsers) == 0 {
		return
	}

	// Unbans/Unmutes users
	for _, user := range punishedUsers {
		fieldRemoved := unbanHandler(s, guildID, user, bans, &t)
		if fieldRemoved {
			continue
		}
		unmuteHandler(s, guildID, user, &t)
	}
}

func unbanHandler(s *discordgo.Session, guildID string, user entities.PunishedUsers, bans []*discordgo.GuildBan, t *time.Time) bool {
	if user.GetUnbanDate() == (time.Time{}) {
		return false
	}
	banDifference := t.Sub(user.GetUnbanDate())
	if banDifference <= 0 {
		return false
	}

	// Set flag for whether user is a banned user in guild and db
	banFlag := false
	for _, ban := range bans {
		if ban.User.ID == user.GetID() {
			banFlag = true
			break
		}
	}

	// Unbans Username if possible and needed
	if banFlag {
		_ = s.GuildBanDelete(guildID, user.GetID())
	}

	// Removes unban date entirely
	mem := db.GetGuildMember(guildID, user.GetID())
	if mem.GetID() != "" {
		mem.SetUnbanDate("")
	}

	// Removes the unbanDate or the entire object
	if user.GetUnmuteDate() != (time.Time{}) {
		_ = db.SetGuildPunishedUser(guildID, entities.NewPunishedUsers(user.GetID(), user.GetUsername(), time.Time{}, user.GetUnmuteDate()))
	} else {
		_ = db.SetGuildPunishedUser(guildID, entities.NewPunishedUsers(user.GetID(), user.GetUsername(), user.GetUnbanDate(), user.GetUnmuteDate()), true)
	}

	// Sends an embed message to bot-log
	if banFlag && mem.GetID() != "" {
		guildSettings := db.GetGuildSettings(guildID)
		if guildSettings.BotLog != (entities.Cha{}) && guildSettings.BotLog.GetID() != "" {
			_ = embeds.AutoPunishmentRemoval(s, mem, guildSettings.BotLog.GetID(), "unbanned")
		}
	}

	if mem.GetID() != "" {
		db.SetGuildMember(guildID, mem)
	}

	return true
}

func unmuteHandler(s *discordgo.Session, guildID string, user entities.PunishedUsers, t *time.Time) {
	if user.GetUnmuteDate() == (time.Time{}) {
		return
	}
	muteDifference := t.Sub(user.GetUnmuteDate())
	if muteDifference <= 0 {
		return
	}

	// Set flag for whether user has the muted role and unmutes them
	muteFlag := false
	guildMember, err := s.State.Member(guildID, user.GetID())
	if err != nil {
		guildMember, _ = s.GuildMember(guildID, user.GetID())
	}
	if guildMember != nil {
		guildSettings := db.GetGuildSettings(guildID)

		for _, roleID := range guildMember.Roles {
			if guildSettings.GetMutedRole() != (entities.Role{}) {
				if roleID == guildSettings.GetMutedRole().GetID() {
					err := s.GuildMemberRoleRemove(guildID, user.GetID(), roleID)
					if err == nil {
						muteFlag = true
					}
					break
				}
			} else {
				deb, _ := s.GuildRoles(guildID)
				for _, role := range deb {
					if strings.ToLower(role.Name) == "muted" || strings.ToLower(role.Name) == "t-mute" {
						err := s.GuildMemberRoleRemove(guildID, user.GetID(), role.ID)
						if err == nil {
							muteFlag = true
						}
						break
					}
				}
			}
			if muteFlag {
				break
			}
		}
	}

	// Removes unmute date entirely
	mem := db.GetGuildMember(guildID, user.GetID())
	if mem.GetID() != "" {
		mem.SetUnmuteDate("")
	}

	// Removes the unmuteDate or the entire object
	if user.GetUnbanDate() != (time.Time{}) {
		_ = db.SetGuildPunishedUser(guildID, entities.NewPunishedUsers(user.GetID(), user.GetUsername(), user.GetUnbanDate(), time.Time{}))
	} else {
		_ = db.SetGuildPunishedUser(guildID, entities.NewPunishedUsers(user.GetID(), user.GetUsername(), user.GetUnbanDate(), user.GetUnmuteDate()), true)
	}

	// Sends an embed message to bot-log
	if muteFlag && mem.GetID() != "" {
		guildSettings := db.GetGuildSettings(guildID)
		if guildSettings.BotLog != (entities.Cha{}) && guildSettings.BotLog.GetID() != "" {
			_ = embeds.AutoPunishmentRemoval(s, mem, guildSettings.BotLog.GetID(), "unmuted")
		}
	}

	if mem.GetID() != "" {
		db.SetGuildMember(guildID, mem)
	}
}

// remindMeHandler handles sending remindMe messages when called if it's time.
// Sends either a DM, or, failing that, a ping in the channel the remindMe was set.
func remindMeHandler(s *discordgo.Session, guildID string) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("%v\n%s", rec, string(debug.Stack()))
		}
	}()

	var writeFlag bool
	var wg sync.WaitGroup
	t := time.Now()

	entities.Mutex.Lock()
	if entities.SharedInfo == nil || entities.SharedInfo.GetRemindMesMap() == nil || len(entities.SharedInfo.GetRemindMesMap()) == 0 {
		entities.Mutex.Unlock()
		return
	}

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
			if t.Sub(remindMe.GetDate()) <= 0 {
				remindMeSlice.GetRemindMeSlice()[i] = remindMe
				i++
				continue
			}

			msgDM := fmt.Sprintf("RemindMe: %s", remindMe.GetMessage())
			msgChannel := fmt.Sprintf("<@%s> Remindme: %s", userID, remindMe.GetMessage())
			cmdChannel := remindMe.GetCommandChannel()

			wg.Add(1)
			go func(userID, msgDM, msgChannel, cmdChannel, guildID string) {
				defer wg.Done()

				dm, err := s.UserChannelCreate(userID)
				if err == nil {
					_, err = s.ChannelMessageSend(dm.ID, msgDM)
				}
				if err != nil && guildID != "" {
					// Checks if the user is in the server and then pings him if true
					_, err := s.GuildMember(guildID, userID)
					if err == nil {
						_, _ = s.ChannelMessageSend(cmdChannel, msgChannel)
					}
				}
			}(userID, msgDM, msgChannel, cmdChannel, guildID)

			writeFlag = true
		}
		remindMeSlice.SetRemindMeSlice(remindMeSlice.GetRemindMeSlice()[:i])
	}

	if !writeFlag {
		entities.Mutex.Unlock()
		return
	}

	err := entities.RemindMeWrite(entities.SharedInfo.GetRemindMesMap())
	if err != nil && guildID != "" {
		entities.Mutex.Unlock()
		guildSettings := db.GetGuildSettings(guildID)
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
	entities.Mutex.Unlock()

	wg.Wait()
}

// Fetches reddit feeds and returns the feeds that need to posted for all guilds
func feedHandler(s *discordgo.Session, guildIds []string) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in feedHandler")
			log.Println("stacktrace from panic: \n" + string(debug.Stack()))
		}
	}()

	// Blocks handling of new feeds if there are some currently being sent
	entities.Mutex.Lock()
	if redditFeedBlock {
		entities.Mutex.Unlock()
		return
	}
	redditFeedBlock = true
	entities.Mutex.Unlock()

	// Handles feeds
	var guildPostsMap = make(map[string]map[entities.Feed][]*gofeed.Item)
	for _, guildID := range guildIds {
		feedsToPost, err := guildFeedsHandler(guildID)
		if err != nil || feedsToPost == nil || len(feedsToPost) == 0 {
			continue
		}
		guildPostsMap[guildID] = feedsToPost
	}

	// Posts feeds
	feedPoster(s, guildPostsMap, guildIds)

	entities.Mutex.Lock()
	redditFeedBlock = false
	entities.Mutex.Unlock()
}

// Handles guild feeds and returns a map of the feed items to send in discord
func guildFeedsHandler(guildID string) (map[entities.Feed][]*gofeed.Item, error) {
	// Fetches all feeds and feed checks for this guild
	guildFeeds := db.GetGuildFeeds(guildID)
	guildFeedChecks := db.GetGuildFeedChecks(guildID)

	var (
		removedCheck bool
		subMap       = make(map[string]*gofeed.Feed)
		feedsToPost  = make(map[entities.Feed][]*gofeed.Item)
	)

	// Store current time
	t := time.Now()

	// Removes a check if more than its allowed lifespan hours have passed
	for _, feedCheck := range guildFeedChecks {
		dateRemoval := feedCheck.GetDate().Add(feedCheckLifespanHours)
		if t.Sub(dateRemoval) > 0 {
			continue
		}

		db.SetGuildFeedCheck(guildID, feedCheck, true)
		removedCheck = true
	}
	if removedCheck {
		guildFeedChecks = db.GetGuildFeedChecks(guildID)
	}

	// Sets up the reddit feed parser
	fp := gofeed.NewParser()
	fp.Client = &http.Client{Transport: &common.UserAgentTransport{RoundTripper: http.DefaultTransport}, Timeout: time.Second * 10}

	// Save all feed subreddits and their feedParsers early as an optimization
	for _, feed := range guildFeeds {
		if _, ok := subMap[fmt.Sprintf("%s:%s", feed.GetSubreddit(), feed.GetPostType())]; ok {
			continue
		}
		// Parse the feed
		feedParser, err := fp.ParseURL(fmt.Sprintf("https://www.reddit.com/r/%s/%s/.rss", feed.GetSubreddit(), feed.GetPostType()))
		if err != nil {
			if _, ok := err.(*gofeed.HTTPError); ok {
				if err.(*gofeed.HTTPError).StatusCode == 429 {
					return nil, nil
				}
			}
			continue
		}
		subMap[fmt.Sprintf("%s:%s", feed.GetSubreddit(), feed.GetPostType())] = feedParser
	}

	// Store latest current time
	t = time.Now()

	for _, feed := range guildFeeds {
		// Get the necessary feed parser from the subMap
		feedParser, ok := subMap[fmt.Sprintf("%s:%s", feed.GetSubreddit(), feed.GetPostType())]
		if !ok {
			continue
		}

		// Iterates through each feed parser item to see if it finds something that should be posted
		for _, item := range feedParser.Items {
			var skip bool

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

			// Adds the feed to the send map
			feedsToPost[feed] = append(feedsToPost[feed], item)
		}
	}

	return feedsToPost, nil
}

// Sends feed posts in all guilds
func feedPoster(s *discordgo.Session, feedsToPost map[string]map[entities.Feed][]*gofeed.Item, guildIds []string) {
	var wg sync.WaitGroup

	// Stores current time
	t := time.Now()

	wg.Add(len(guildIds))
	for _, guildID := range guildIds {
		go func(guildID string, t time.Time) {
			defer wg.Done()
			if _, ok := feedsToPost[guildID]; !ok || feedsToPost[guildID] == nil || len(feedsToPost[guildID]) == 0 {
				return
			}
			feedPostHandler(s, guildID, feedsToPost[guildID], t)
		}(guildID, t)
	}
	wg.Wait()
}

// Sends feed posts in a guild
func feedPostHandler(s *discordgo.Session, guildID string, feedsToPost map[entities.Feed][]*gofeed.Item, t time.Time) {
	var wg sync.WaitGroup

	wg.Add(len(feedsToPost))
	for feed, items := range feedsToPost {
		go func(feed entities.Feed, items []*gofeed.Item) {
			defer wg.Done()
			postFeedItems(s, feed, items, t, guildID)
		}(feed, items)
	}
	wg.Wait()

	db.SetGuildFeedChecks(guildID, entities.Guilds.DB[guildID].GetFeedChecks())
}

func postFeedItems(s *discordgo.Session, feed entities.Feed, items []*gofeed.Item, t time.Time, guildID string) {
	var pinnedItems = make(map[*gofeed.Item]bool)

	for _, item := range items {

		var ok bool

		// Wait five seconds so it doesn't hit the rate limit easily
		time.Sleep(time.Second * 5)

		// Stops the iteration if the feed doesn't exist anymore
		guildFeeds := db.GetGuildFeeds(guildID)
		for _, guildFeed := range guildFeeds {
			if guildFeed.GetSubreddit() == feed.GetSubreddit() &&
				guildFeed.GetChannelID() == feed.GetChannelID() {
				ok = true
				break
			}
		}
		if !ok {
			break
		}

		// Checks if the item has already been posted
		feedChecks := db.GetGuildFeedChecks(guildID)
		for _, feedCheck := range feedChecks {
			if feedCheck.GetGUID() == item.GUID {
				ok = true
				break
			}
		}
		if !ok {
			continue
		}

		// Sends the feed item
		message, err := embeds.Feed(s, &feed, item)
		if err != nil {
			continue
		}

		// Adds that the feed has been posted
		db.AddGuildFeedCheck(guildID, entities.NewFeedCheck(feed, t, item.GUID))

		// Pins/unpins the feed items if necessary
		if !feed.GetPin() {
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

		// Pins
		_ = s.ChannelMessagePin(message.ChannelID, message.ID)
		pinnedItems[item] = true
	}
}
