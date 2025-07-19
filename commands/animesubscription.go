package commands

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/r-anime/ZeroTsu/cache"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/embeds"
	"github.com/r-anime/ZeroTsu/entities"
	"github.com/r-anime/ZeroTsu/events"
	"github.com/sasha-s/go-deadlock"
	"golang.org/x/sync/errgroup"

	"sync"

	"github.com/bwmarrin/discordgo"
)

type safeWebhooksMap struct {
	deadlock.RWMutex
	webhooksMap map[string]*discordgo.Webhook
}

type safeThreadWebhooksMap struct {
	deadlock.RWMutex
	webhooksMap map[string]string
}

var (
	newEpisodesWebhooksMap       = &safeWebhooksMap{webhooksMap: make(map[string]*discordgo.Webhook)}
	newEpisodesThreadWebhooksMap = &safeThreadWebhooksMap{webhooksMap: make(map[string]string)}

	newEpisodeswebhooksMapBlock events.Block
	animeSubFeedWebhookBlock    events.Block
	animeSubFeedBlock           events.Block
	animeSubGlobalBlock         events.Block
)

// subscribeCommand subscribes to notifications for anime episode releases SUBBED
func subscribeCommand(title, authorID string) string {
	var (
		now           = time.Now().UTC()
		showExists    bool
		hasAiredToday bool
	)

	// Iterates over all the anime shows saved from AnimeSchedule and checks if it finds one
	entities.AnimeSchedule.RLock()
	defer entities.AnimeSchedule.RUnlock()

Loop:
	for dayInt, dailyShows := range entities.AnimeSchedule.AnimeSchedule {
		if dailyShows == nil {
			continue
		}

		for _, show := range dailyShows {
			if show == nil {
				continue
			}

			if strings.EqualFold(show.GetName(), title) {
				showExists = true

				// Check if user is already subscribed
				userSubs := db.GetAnimeSubs(authorID)
				for _, sub := range userSubs {
					if sub.GetShow() == show.GetName() {
						return fmt.Sprintf("Error: You are already subscribed to `%s`", show.GetName())
					}
				}

				// Checks if the show is from Today and whether it has already passed (to avoid notifying the user Today if it has passed)
				if int(now.Weekday()) == dayInt {
					// Reset bool
					hasAiredToday = false

					// Parse the air hour and minute
					t, err := time.Parse("3:04 PM", show.GetAirTime())
					if err != nil {
						log.Println(err)
						continue
					}

					// Form the air date for Today
					scheduleDate := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)

					// Calculates whether the show has already aired Today
					if now.After(scheduleDate) {
						hasAiredToday = true
					}
				}

				// Add the new anime subscription to MongoDB
				newSub := entities.NewShowSub(show.GetName(), hasAiredToday, false)
				db.AddAnimeSub(authorID, newSub, false)
				break Loop
			}
		}
	}

	if !showExists {
		return "Error: That is not a valid airing show name. It has to be airing. Make sure you're using the exact romaji anime title from `/schedule` or AnimeSchedule.net."
	}

	return fmt.Sprintf("Success! You have subscribed to DM notifications for `%s`", title)
}

// subscribeCommandHandler subscribes to notifications for anime episode releases SUBBED
func subscribeCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	var (
		showName      string
		hasAiredToday bool
		err           error
		guildSettings = entities.GuildSettings{Prefix: "."}
	)

	if m.GuildID != "" {
		guildSettings = db.GetGuildSettings(m.GuildID)
	}

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%ssub [anime]`\n\nAnime is the anime name from <https://AnimeSchedule.net> or the schedule command", guildSettings.GetPrefix()))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		}
		return
	}

	now := time.Now().UTC()
	entities.AnimeSchedule.RLock()
	defer entities.AnimeSchedule.RUnlock()

Loop:
	for dayInt, dailyShows := range entities.AnimeSchedule.AnimeSchedule {
		if dailyShows == nil {
			continue
		}

		for _, show := range dailyShows {
			if show == nil {
				continue
			}

			if strings.ToLower(show.GetName()) == commandStrings[1] {
				showName = show.GetName()

				// Check if user is already subscribed
				userSubs := db.GetAnimeSubs(m.Author.ID)
				for _, sub := range userSubs {
					if sub.GetShow() == show.GetName() {
						_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: You are already subscribed to `%s`", show.GetName()))
						if err != nil {
							common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
						}
						return
					}
				}

				// Check if the show aired today
				if int(now.Weekday()) == dayInt {
					// Parse the air hour and minute
					t, err := time.Parse("3:04 PM", show.GetAirTime())
					if err != nil {
						log.Println(err)
						continue
					}

					// Form the air date for today
					scheduleDate := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)

					// Determines if the show has aired today
					if now.After(scheduleDate) {
						hasAiredToday = true
					}
				}

				// Add subscription to MongoDB
				newSub := entities.NewShowSub(show.GetName(), hasAiredToday, false)
				db.AddAnimeSub(m.Author.ID, newSub, false)
				break Loop
			}
		}
	}

	if showName == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: That is not a valid airing show name. It has to be airing. Make sure you're using the exact show name from `"+guildSettings.GetPrefix()+"schedule`")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		}
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! You have subscribed to notifications for `%s`", showName))
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
	}
}

func unsubscribeCommand(title, authorID string) string {
	// Fetch user subscriptions from MongoDB
	userSubs := db.GetAnimeSubs(authorID)

	// Check if user has any subscriptions
	if len(userSubs) == 0 {
		return fmt.Sprintf("Error: You are not subscribed to `%s`", title)
	}

	// Try to remove the subscription
	var isDeleted bool
	for _, show := range userSubs {
		if strings.EqualFold(show.GetShow(), title) {
			db.RemoveAnimeSub(authorID, title)
			isDeleted = true
			break
		}
	}

	// Send an error if the target show is not one the user is subscribed to
	if !isDeleted {
		return fmt.Sprintf("Error: You are not subscribed to `%s`", title)
	}

	return fmt.Sprintf("Success! You have unsubscribed from `%s`", title)
}

// unsubscribeCommandHandler removes a subscription for notifications for anime episode releases SUBBED
func unsubscribeCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	var (
		err           error
		guildSettings = entities.GuildSettings{Prefix: "."}
	)

	// Fetch guild settings if in a server
	if m.GuildID != "" {
		guildSettings = db.GetGuildSettings(m.GuildID)
	}

	// Parse the command arguments
	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)
	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
			"Usage: `%sunsub [anime]`\n\nAnime is the anime name from <https://AnimeSchedule.net> or the schedule command",
			guildSettings.GetPrefix(),
		))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		}
		return
	}

	showName := commandStrings[1]

	// Check if user is subscribed to the anime
	subscribedShows := db.GetAnimeSubs(m.Author.ID)
	if subscribedShows == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: You are not subscribed to `%s`", showName))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		}
		return
	}

	// Remove the subscription
	db.RemoveAnimeSub(m.Author.ID, showName)

	// Confirmation message
	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! You have unsubscribed from `%s`", showName))
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
	}
}

func clearAnimeSubsCommand(userID string) string {
	db.SetAnimeSubs(userID, []*entities.ShowSub{}, false)
	return "Success! All your anime subscriptions have been cleared."
}

func clearAnimeSubsCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	var (
		guildSettings = entities.GuildSettings{Prefix: "."}
	)

	if m.GuildID != "" {
		guildSettings = db.GetGuildSettings(m.GuildID)
	}

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")
	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%sclearsubs`", guildSettings.GetPrefix()))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		}
		return
	}

	db.SetAnimeSubs(m.Author.ID, []*entities.ShowSub{}, false)

	_, err := s.ChannelMessageSend(m.ChannelID, "Success! All your anime subscriptions have been cleared.")
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
	}
}

// viewSubscriptions prints out all the anime the user is subscribed to
func viewSubscriptions(authorID string) []string {
	var message string

	// Fetch user's anime subscriptions from MongoDB
	subscribedShows := db.GetAnimeSubs(authorID)
	if len(subscribedShows) == 0 {
		return []string{"Error: You have no active anime subscriptions."}
	}

	// Format the subscription list
	for i, show := range subscribedShows {
		message += fmt.Sprintf("**%d.** %s\n", i+1, show.GetShow())
	}

	// Split long messages if necessary
	if len(message) > 1900 {
		return common.SplitLongMessage(message)
	}

	return []string{message}
}

// viewSubscriptionsHandler prints out all the anime the user is subscribed to
func viewSubscriptionsHandler(s *discordgo.Session, m *discordgo.Message) {
	var (
		message       string
		messages      []string
		guildSettings = entities.GuildSettings{Prefix: "."}
	)

	if m.GuildID != "" {
		guildSettings = db.GetGuildSettings(m.GuildID)
	}

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%ssubs`", guildSettings.GetPrefix()))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		}
		return
	}

	// Fetch user's anime subscriptions from MongoDB
	subscribedShows := db.GetAnimeSubs(m.Author.ID)
	if len(subscribedShows) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: You have no active show subscriptions.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
		}
		return
	}

	// Format subscription list
	for i, show := range subscribedShows {
		message += fmt.Sprintf("**%d.** %s\n", i+1, show.GetShow())
	}

	// Split the message if too long
	if len(message) > 1900 {
		messages = common.SplitLongMessage(message)
	} else {
		messages = []string{message}
	}

	// Send messages to the user
	for _, msg := range messages {
		_, err := s.ChannelMessageSend(m.ChannelID, msg)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
	}
}

// WebhooksMapHandler updates the anime subs guilds' webhooks map
func WebhooksMapHandler() {
	newEpisodeswebhooksMapBlock.Lock()
	if newEpisodeswebhooksMapBlock.Block {
		newEpisodeswebhooksMapBlock.Unlock()
		return
	}
	newEpisodeswebhooksMapBlock.Block = true
	newEpisodeswebhooksMapBlock.Unlock()

	// Store all valid guilds' webhooks in a temporary map
	tempWebhooksMap := make(map[string]*discordgo.Webhook)
	tempThreadWebhooksMap := make(map[string]string)

	// Fetch only guild anime subscriptions from cache
	animeSubsMap := cache.AnimeSubs.GetGuild()

	for guildID, subs := range animeSubsMap {
		if len(subs) == 0 {
			continue
		}

		guildIDInt, err := strconv.ParseInt(guildID, 10, 64)
		if err != nil {
			continue
		}
		s := config.Mgr.SessionForGuild(guildIDInt)

		// Check if the bot is in the target guild
		_, err = s.State.Guild(guildID)
		if err != nil {
			continue
		}

		// Check if bot has required permissions for the new episodes channel
		isThread := false
		newEpisodes := db.GetGuildAutopost(guildID, "newepisodes")
		if newEpisodes == (entities.Cha{}) {
			continue
		}
		perms, err := s.State.UserChannelPermissions(s.State.User.ID, newEpisodes.GetID())
		if err != nil {
			continue
		}
		if perms&discordgo.PermissionManageWebhooks != discordgo.PermissionManageWebhooks ||
			perms&discordgo.PermissionViewChannel != discordgo.PermissionViewChannel ||
			perms&discordgo.PermissionSendMessages != discordgo.PermissionSendMessages {
			continue
		}

		// Handle threads
		newEpisodesChannel, err := s.State.Channel(newEpisodes.GetID())
		if err != nil {
			newEpisodesChannel, err = s.Channel(newEpisodes.GetID())
			if err == nil && newEpisodesChannel.IsThread() {
				isThread = true
			}
		} else if newEpisodesChannel.IsThread() {
			isThread = true
		}

		if isThread && perms&discordgo.PermissionSendMessagesInThreads != discordgo.PermissionSendMessagesInThreads {
			continue
		}

		channelID := newEpisodes.GetID()
		if isThread {
			channelID = newEpisodesChannel.ParentID
		}

		// Get valid webhook
		ws, err := s.ChannelWebhooks(channelID)
		if err != nil {
			continue
		}

		for _, w := range ws {
			if w.User.ID == s.State.User.ID && w.ChannelID == channelID {
				tempWebhooksMap[guildID] = w
				if isThread {
					tempThreadWebhooksMap[guildID] = newEpisodes.GetID()
				}
				break
			}
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
		defer out.Reset()

		err = png.Encode(out, avatar)
		if err != nil {
			continue
		}
		base64Img := base64.StdEncoding.EncodeToString(out.Bytes())
		out.Reset()

		wh, err := s.WebhookCreate(channelID, s.State.User.Username, fmt.Sprintf("data:image/png;base64,%s", base64Img))
		if err != nil {
			continue
		}

		tempWebhooksMap[guildID] = wh
		if isThread {
			tempThreadWebhooksMap[guildID] = newEpisodes.GetID()
		}
	}

	// Update webhooks map with validated data
	newEpisodesWebhooksMap.Lock()
	newEpisodesWebhooksMap.webhooksMap = tempWebhooksMap
	newEpisodesWebhooksMap.Unlock()

	newEpisodesThreadWebhooksMap.Lock()
	newEpisodesThreadWebhooksMap.webhooksMap = tempThreadWebhooksMap
	newEpisodesThreadWebhooksMap.Unlock()

	// Unlock webhook block
	newEpisodeswebhooksMapBlock.Lock()
	newEpisodeswebhooksMapBlock.Block = false
	newEpisodeswebhooksMapBlock.Unlock()
}

func animeSubsWebhookHandler() {
	var todayShows []*entities.ShowAirTime

	animeSubGlobalBlock.Lock()
	if animeSubGlobalBlock.Block {
		animeSubGlobalBlock.Unlock()
		return
	}
	animeSubGlobalBlock.Block = true
	animeSubGlobalBlock.Unlock()

	defer func() {
		animeSubGlobalBlock.Lock()
		animeSubGlobalBlock.Block = false
		animeSubGlobalBlock.Unlock()
	}()

	animeSubFeedWebhookBlock.Lock()
	if animeSubFeedWebhookBlock.Block {
		animeSubFeedWebhookBlock.Unlock()
		return
	}
	animeSubFeedWebhookBlock.Block = true
	animeSubFeedWebhookBlock.Unlock()

	DailyScheduleEventsBlock.RLock()
	if DailyScheduleEventsBlock.Block {
		DailyScheduleEventsBlock.RUnlock()
		animeSubFeedWebhookBlock.Lock()
		animeSubFeedWebhookBlock.Block = false
		animeSubFeedWebhookBlock.Unlock()
		return
	}
	DailyScheduleEventsBlock.RUnlock()

	now := time.Now()
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		animeSubFeedWebhookBlock.Lock()
		animeSubFeedWebhookBlock.Block = false
		animeSubFeedWebhookBlock.Unlock()
		return
	}
	now = now.In(location)

	Today.RLock()
	if int(Today.Time.Weekday()) != int(now.Weekday()) {
		Today.RUnlock()
		animeSubFeedWebhookBlock.Lock()
		animeSubFeedWebhookBlock.Block = false
		animeSubFeedWebhookBlock.Unlock()
		return
	}
	Today.RUnlock()

	newEpisodeswebhooksMapBlock.RLock()
	if newEpisodeswebhooksMapBlock.Block {
		newEpisodeswebhooksMapBlock.RUnlock()
		animeSubFeedWebhookBlock.Lock()
		animeSubFeedWebhookBlock.Block = false
		animeSubFeedWebhookBlock.Unlock()
		return
	}
	newEpisodeswebhooksMapBlock.RUnlock()

	// Fetch today's shows
	entities.AnimeSchedule.RLock()
	todayShows = append(todayShows, entities.AnimeSchedule.AnimeSchedule[int(now.Weekday())]...)
	entities.AnimeSchedule.RUnlock()

	var (
		eg            errgroup.Group
		maxGoroutines = 32
		guard         = make(chan struct{}, maxGoroutines)
	)

	// Fetch only guild anime subscriptions from cache
	animeSubsMap := cache.AnimeSubs.GetGuild()

	for id, subscriptions := range animeSubsMap {
		if len(subscriptions) == 0 {
			continue
		}

		guid := id
		subs := subscriptions

		guard <- struct{}{}
		eg.Go(func() error {
			defer func() { <-guard }()

			guildIDInt, err := strconv.ParseInt(guid, 10, 64)
			if err != nil {
				return nil
			}
			s := config.Mgr.SessionForGuild(guildIDInt)

			// Check if bot is in target guild
			_, err = s.State.Guild(guid)
			if err != nil {
				return nil
			}
			guildSettings := db.GetGuildSettings(guid)

			// Get pingable role ID
			newEpisodes := db.GetGuildAutopost(guid, "newepisodes")
			pingableRoleID := newEpisodes.GetRoleID()

			// Get valid webhook
			newEpisodesWebhooksMap.RLock()
			w, exists := newEpisodesWebhooksMap.webhooksMap[guid]
			newEpisodesWebhooksMap.RUnlock()
			if !exists {
				return nil
			}

			threadID := ""
			newEpisodesThreadWebhooksMap.RLock()
			if ID, ok := newEpisodesThreadWebhooksMap.webhooksMap[guid]; ok {
				threadID = ID
			}
			newEpisodesThreadWebhooksMap.RUnlock()

			updated := false
			subsMap := make(map[string]*entities.ShowSub)
			processedShows := make(map[string]bool)
			for _, sub := range subs {
				if sub == nil {
					continue
				}
				subsMap[strings.ToLower(sub.GetShow())] = sub
			}
			for _, scheduleShow := range todayShows {
				if scheduleShow == nil || scheduleShow.GetDelayed() != "" {
					continue
				}

				showName := scheduleShow.GetName()
				showNameLower := strings.ToLower(showName)

				if processedShows[showNameLower] {
					continue
				}

				if sub, exists := subsMap[showNameLower]; exists {
					if !guildSettings.GetDonghua() && scheduleShow.GetDonghua() {
						continue
					}

					// Parse the air hour and minute
					t, err := time.Parse("3:04 PM", scheduleShow.GetAirTime())
					if err != nil {
						log.Println(err)
						continue
					}

					// Form the air date and time for today
					airDatetime := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)

					if now.Before(airDatetime) {
						continue
					}
					if sub.GetNotified() {
						continue
					}

					var pingableRoleStr string
					if pingableRoleID != "" {
						pingableRoleStr = fmt.Sprintf("<@&%s>", pingableRoleID)
					}
					params := &discordgo.WebhookParams{
						Content: pingableRoleStr,
						Embeds:  []*discordgo.MessageEmbed{embeds.SubscriptionEmbed(scheduleShow)},
					}

					if threadID != "" {
						_, err = s.WebhookThreadExecute(w.ID, w.Token, false, threadID, params)
					} else {
						_, err = s.WebhookExecute(w.ID, w.Token, false, params)
					}

					if err != nil {
						log.Println("Failed webhookExecute in animeSubsWebhookHandler: ", err)
						break
					}

					sub.SetNotified(true)
					updated = true

					processedShows[showNameLower] = true
				}
			}

			var updatedSubs []*entities.ShowSub
			for _, sub := range subsMap {
				updatedSubs = append(updatedSubs, sub)
			}

			if len(updatedSubs) > 0 && updated {
				db.SetAnimeSubs(guid, updatedSubs, true)
			}

			return nil
		})
	}

	err = eg.Wait()
	if err != nil {
		log.Printf("anime subs webhook handler: %s", err)
	}

	animeSubFeedWebhookBlock.Lock()
	animeSubFeedWebhookBlock.Block = false
	animeSubFeedWebhookBlock.Unlock()
}

// animeSubsHandler handles sending notifications to users or guilds when it's time
func animeSubsHandler() {
	var todayShows []*entities.ShowAirTime

	animeSubGlobalBlock.Lock()
	if animeSubGlobalBlock.Block {
		animeSubGlobalBlock.Unlock()
		return
	}
	animeSubGlobalBlock.Block = true
	animeSubGlobalBlock.Unlock()

	defer func() {
		animeSubGlobalBlock.Lock()
		animeSubGlobalBlock.Block = false
		animeSubGlobalBlock.Unlock()
	}()

	animeSubFeedBlock.Lock()
	if animeSubFeedBlock.Block {
		animeSubFeedBlock.Unlock()
		return
	}
	animeSubFeedBlock.Block = true
	animeSubFeedBlock.Unlock()

	DailyScheduleEventsBlock.RLock()
	if DailyScheduleEventsBlock.Block {
		DailyScheduleEventsBlock.RUnlock()
		animeSubFeedBlock.Lock()
		animeSubFeedBlock.Block = false
		animeSubFeedBlock.Unlock()
		return
	}
	DailyScheduleEventsBlock.RUnlock()

	now := time.Now()
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		animeSubFeedBlock.Lock()
		animeSubFeedBlock.Block = false
		animeSubFeedBlock.Unlock()
		return
	}
	now = now.In(location)

	// Fetch today's shows
	entities.AnimeSchedule.RLock()
	todayShows = append(todayShows, entities.AnimeSchedule.AnimeSchedule[int(now.Weekday())]...)
	entities.AnimeSchedule.RUnlock()

	// Fetch all anime subscriptions from cache
	animeSubsMap := cache.AnimeSubs.Get()

	for id, subscriptions := range animeSubsMap {
		if len(subscriptions) == 0 {
			continue
		}

		var (
			session       *discordgo.Session
			guildSettings entities.GuildSettings
			isGuild       bool
			isThread      bool
		)
		if subscriptions[0].GetGuild() {
			isGuild = true
		}

		if isGuild {
			guildIDInt, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				continue
			}
			session = config.Mgr.SessionForGuild(guildIDInt)

			// Check if bot is in target guild
			_, err = session.State.Guild(id)
			if err != nil {
				continue
			}
		} else {
			session = config.Mgr.SessionForDM()
		}

		if isGuild {
			newepisodes := db.GetGuildAutopost(id, "newepisodes")
			if newepisodes == (entities.Cha{}) {
				continue
			}

			// Check if bot has required permissions
			perms, err := session.State.UserChannelPermissions(session.State.User.ID, newepisodes.GetID())
			if err != nil {
				continue
			}
			if perms&discordgo.PermissionManageWebhooks != discordgo.PermissionManageWebhooks ||
				perms&discordgo.PermissionViewChannel != discordgo.PermissionViewChannel ||
				perms&discordgo.PermissionSendMessages != discordgo.PermissionSendMessages ||
				perms&discordgo.PermissionEmbedLinks != discordgo.PermissionEmbedLinks {
				continue
			}

			newEpisodesChannel, err := session.State.Channel(newepisodes.ID)
			if err != nil {
				newEpisodesChannel, err = session.Channel(newepisodes.GetID())
				if err == nil && newEpisodesChannel.IsThread() {
					isThread = true
				}
			} else if newEpisodesChannel.IsThread() {
				isThread = true
			}

			if isThread && perms&discordgo.PermissionSendMessagesInThreads != discordgo.PermissionSendMessagesInThreads {
				continue
			}

			guildSettings = db.GetGuildSettings(id)
		}

		updated := false
		var updatedSubs []*entities.ShowSub
		subsMap := make(map[string]*entities.ShowSub)
		for _, sub := range subscriptions {
			if sub == nil {
				continue
			}
			subsMap[strings.ToLower(sub.GetShow())] = sub
			updatedSubs = append(updatedSubs, sub)
		}

		for _, scheduleShow := range todayShows {
			if scheduleShow == nil || scheduleShow.GetDelayed() != "" {
				continue
			}

			showName := strings.ToLower(scheduleShow.GetName())

			// If user or guild is already subscribed, check if we need to mark it as notified
			if sub, exists := subsMap[showName]; exists {
				if isGuild && !guildSettings.GetDonghua() && scheduleShow.GetDonghua() {
					continue
				}

				// Parse the air hour and minute
				t, err := time.Parse("3:04 PM", scheduleShow.GetAirTime())
				if err != nil {
					log.Println(err)
					continue
				}

				// Form the air date and time for today
				airDatetime := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)

				// If it's time, notify & mark as notified
				if now.Before(airDatetime) {
					continue
				}
				if sub.GetNotified() {
					continue
				}

				time.Sleep(time.Millisecond * 150)

				err = nil
				if isGuild {
					newepisodes := db.GetGuildAutopost(id, "newepisodes")
					if newepisodes == (entities.Cha{}) {
						continue
					}
					err = embeds.Subscription(session, scheduleShow, newepisodes.GetID(), newepisodes.GetRoleID())
				} else {
					dm, dmErr := session.UserChannelCreate(id)
					if dmErr != nil {
						continue
					}
					err = embeds.Subscription(session, scheduleShow, dm.ID, "")
				}

				if err == nil {
					sub.SetNotified(true)
					updated = true
				}
			}
		}

		if len(updatedSubs) > 0 && updated {
			db.SetAnimeSubs(id, updatedSubs, isGuild)
		}
	}

	animeSubFeedBlock.Lock()
	animeSubFeedBlock.Block = false
	animeSubFeedBlock.Unlock()
}

func AnimeSubsTimer(_ *discordgo.Session, _ *discordgo.Ready) {
	var processingMutex sync.Mutex
	var isProcessing bool

	for range time.NewTicker(1 * time.Minute).C {
		processingMutex.Lock()
		if isProcessing {
			log.Println("Previous AnimeSubsTimer execution still running, skipping this cycle")
			processingMutex.Unlock()
			continue
		}
		isProcessing = true
		processingMutex.Unlock()

		animeSubsHandler()

		processingMutex.Lock()
		isProcessing = false
		processingMutex.Unlock()
	}
}

func AnimeSubsWebhookTimer(_ *discordgo.Session, _ *discordgo.Ready) {
	var processingMutex sync.Mutex
	var isProcessing bool

	for range time.NewTicker(1 * time.Minute).C {
		processingMutex.Lock()
		if isProcessing {
			log.Println("Previous AnimeSubsWebhookTimer execution still running, skipping this cycle")
			processingMutex.Unlock()
			continue
		}
		isProcessing = true
		processingMutex.Unlock()

		animeSubsWebhookHandler()

		processingMutex.Lock()
		isProcessing = false
		processingMutex.Unlock()
	}
}

func AnimeSubsWebhooksMapTimer(_ *discordgo.Session, _ *discordgo.Ready) {
	var processingMutex sync.Mutex
	var isProcessing bool

	for range time.NewTicker(15 * time.Minute).C {
		processingMutex.Lock()
		if isProcessing {
			log.Println("Previous AnimeSubsWebhooksMapTimer execution still running, skipping this cycle")
			processingMutex.Unlock()
			continue
		}
		isProcessing = true
		processingMutex.Unlock()

		WebhooksMapHandler()

		processingMutex.Lock()
		isProcessing = false
		processingMutex.Unlock()
	}
}

// ResetSubscriptions resets anime sub notifications status and guild subs
func ResetSubscriptions() {
	now := time.Now()
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return
	}
	now = now.In(location)

	// Fetch today's shows
	entities.AnimeSchedule.RLock()
	todayShows := entities.AnimeSchedule.AnimeSchedule[int(now.Weekday())]
	entities.AnimeSchedule.RUnlock()

	// Fetch all anime subscriptions from cache
	animeSubsMap := cache.AnimeSubs.Get()

	for id, subscriptions := range animeSubsMap {
		if len(subscriptions) == 0 {
			continue
		}

		updated := false
		var updatedSubs []*entities.ShowSub

		// Check if this is a guild
		isGuild := subscriptions[0].GetGuild()

		// If it's a guild make it subscribed to all of today's anime that it's missing
		if isGuild {
			subsMap := make(map[string]*entities.ShowSub)
			for _, sub := range subscriptions {
				if sub == nil {
					continue
				}
				subsMap[strings.ToLower(sub.GetShow())] = sub
			}
			for _, todayShow := range todayShows {
				if todayShow == nil {
					continue
				}
				if _, exists := subsMap[strings.ToLower(todayShow.GetName())]; exists {
					continue
				}
				subscriptions = append(subscriptions, &entities.ShowSub{
					Show:     todayShow.GetName(),
					Notified: false,
					Guild:    true,
				})
			}
		}

		// Process all subscriptions
		for _, show := range subscriptions {
			if show == nil {
				continue
			}

			// Check if the show airs today and needs a status update
			for _, scheduleShow := range todayShows {
				if scheduleShow == nil {
					continue
				}

				// Match show names
				if !strings.EqualFold(show.GetShow(), scheduleShow.GetName()) {
					continue
				}

				// Parse the air time
				t, err := time.Parse("3:04 PM", scheduleShow.GetAirTime())
				if err != nil {
					log.Println(err)
					continue
				}

				// Form the air date and time for today
				airDatetime := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)

				// Update notification status
				oldStatus := show.GetNotified()
				if now.Before(airDatetime) {
					show.SetNotified(false)
				} else {
					show.SetNotified(true)
				}

				// Mark as updated if the status changed
				if show.GetNotified() != oldStatus {
					updated = true
				}
			}

			// Append the processed show to updatedSubs
			updatedSubs = append(updatedSubs, show)
		}

		// Save the updated subscriptions to MongoDB (only if there are changes)
		if updated {
			db.SetAnimeSubs(id, updatedSubs, isGuild)
		}
	}
}

func AutoRemoveFinishedAnimeSubsTimer(_ *discordgo.Session, _ *discordgo.Ready) {
	RemoveFinishedAnimeUserSubs()

	var processingMutex sync.Mutex
	var isProcessing bool

	for range time.NewTicker(30 * time.Minute).C {
		processingMutex.Lock()
		if isProcessing {
			log.Println("Previous AutoRemoveFinishedAnimeSubsTimer execution still running, skipping this cycle")
			processingMutex.Unlock()
			continue
		}
		isProcessing = true
		processingMutex.Unlock()

		RemoveFinishedAnimeUserSubs()

		processingMutex.Lock()
		isProcessing = false
		processingMutex.Unlock()
	}
}

func RemoveFinishedAnimeUserSubs() {
	animeSubsMap := cache.AnimeSubs.Get()
	if animeSubsMap == nil {
		return
	}

	entities.AnimeSchedule.RLock()
	defer entities.AnimeSchedule.RUnlock()

	scheduleByTitle := make(map[string][]*entities.ShowAirTime)
	for _, shows := range entities.AnimeSchedule.AnimeSchedule {
		for _, show := range shows {
			if show == nil {
				continue
			}
			airType := strings.ToLower(show.GetAirType())
			if airType != "sub" && airType != "raw" {
				continue
			}
			key := strings.ToLower(show.GetName())
			scheduleByTitle[key] = append(scheduleByTitle[key], show)
		}
	}

	for userID, subs := range animeSubsMap {
		if len(subs) == 0 || subs[0].GetGuild() {
			continue
		}

		var newSubs []*entities.ShowSub
		updated := false

		for _, sub := range subs {
			showName := strings.ToLower(sub.GetShow())
			scheduleVariants := scheduleByTitle[showName]

			if len(scheduleVariants) == 0 {
				newSubs = append(newSubs, sub)
				continue
			}

			allFinished := true
			for _, variant := range scheduleVariants {
				ep := strings.ToLower(variant.GetEpisode())
				if !strings.HasSuffix(ep, "f") && !strings.Contains(ep, "final") {
					allFinished = false
					break
				}
			}

			// Only remove if the anime is finished and the user has been notified
			if allFinished && sub.GetNotified() {
				updated = true
			} else {
				newSubs = append(newSubs, sub)
			}
		}

		if updated {
			db.SetAnimeSubs(userID, newSubs, false)
		}
	}
}

func init() {
	Add(&Command{
		Execute: subscribeCommandHandler,
		Name:    "sub",
		Aliases: []string{"subscribe", "subs", "animesub", "subanime", "addsub"},
		Desc:    "Subscribe to receive a message when an anime's new episode is released (subbed if possible).",
		Module:  "normal",
		DMAble:  true,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "anime",
				Description: "The romaji title of an ongoing anime from AnimeSchedule.net",
				Required:    true,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.ApplicationCommandData().Options == nil {
				return
			}

			anime := ""
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			userID := ""
			if i.Member == nil {
				userID = i.User.ID
			} else {
				userID = i.Member.User.ID
			}

			if i.ApplicationCommandData().Options != nil {
				for _, option := range i.ApplicationCommandData().Options {
					if option.Name == "anime" {
						anime = option.StringValue()
					}
				}
			}

			respStr := subscribeCommand(anime, userID)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &respStr,
			})
		},
	})
	Add(&Command{
		Execute: unsubscribeCommandHandler,
		Name:    "unsub",
		Aliases: []string{"unsubscribe", "unsubs", "unanimesub", "unsubanime", "removesub", "killsub", "stopsub"},
		Desc:    "Stop getting messages whenever an anime's new episodes are released",
		Module:  "normal",
		DMAble:  true,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "anime",
				Description: "The romaji title of an ongoing anime from AnimeSchedule.net",
				Required:    true,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.ApplicationCommandData().Options == nil {
				return
			}

			anime := ""
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			userID := ""
			if i.Member == nil {
				userID = i.User.ID
			} else {
				userID = i.Member.User.ID
			}

			if i.ApplicationCommandData().Options != nil {
				for _, option := range i.ApplicationCommandData().Options {
					if option.Name == "anime" {
						anime = option.StringValue()
					}
				}
			}

			respStr := unsubscribeCommand(anime, userID)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &respStr,
			})
		},
	})
	Add(&Command{
		Execute: clearAnimeSubsCommandHandler,
		Name:    "clearsubs",
		Aliases: []string{"clearanimesubs", "subsclear", "unsuball"},
		Desc:    "Clear all your anime episode notifications.",
		Module:  "normal",
		DMAble:  true,
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			userID := ""
			if i.Member == nil {
				userID = i.User.ID
			} else {
				userID = i.Member.User.ID
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			respStr := clearAnimeSubsCommand(userID)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &respStr,
			})
		},
	})
	Add(&Command{
		Execute: viewSubscriptionsHandler,
		Name:    "subs",
		Aliases: []string{"subscriptions", "animesubs", "showsubs", "showsubscriptions", "viewsubs", "viewsubscriptions"},
		Desc:    "Print out which anime you are getting new episode notifications for.",
		Module:  "normal",
		DMAble:  true,
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			userID := ""
			if i.Member == nil {
				userID = i.User.ID
			} else {
				userID = i.Member.User.ID
			}

			messages := viewSubscriptions(userID)
			if messages == nil {
				return
			}

			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &messages[0],
			})

			if len(messages) > 1 {
				for j, message := range messages {
					if j == 0 {
						continue
					}

					s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
						Content: message,
					})
				}
			}
		},
	})
}
