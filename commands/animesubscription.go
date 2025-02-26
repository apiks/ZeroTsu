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

	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/embeds"
	"github.com/r-anime/ZeroTsu/entities"
	"github.com/r-anime/ZeroTsu/events"
	"github.com/sasha-s/go-deadlock"
	"golang.org/x/sync/errgroup"

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

			if strings.ToLower(show.GetName()) == strings.ToLower(title) {
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
		if strings.ToLower(show.GetShow()) == strings.ToLower(title) {
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

// viewSubscriptions prints out all the anime the user is subscribed to
func viewSubscriptions(authorID string) []string {
	var message string

	// Fetch user's anime subscriptions from MongoDB
	subscribedShows := db.GetAnimeSubs(authorID)
	if subscribedShows == nil || len(subscribedShows) == 0 {
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
	if subscribedShows == nil || len(subscribedShows) == 0 {
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

	// Fetch anime subscriptions from MongoDB
	animeSubsMap := db.GetAllAnimeSubs()

	for guildID, subs := range animeSubsMap {
		if subs == nil || len(subs) == 0 {
			continue
		}

		// Determine if the subscription is for a guild
		isGuild := subs[0].GetGuild()
		if !isGuild {
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
		err = png.Encode(out, avatar)
		if err != nil {
			continue
		}
		base64Img := base64.StdEncoding.EncodeToString(out.Bytes())
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

// animeSubsWebhookHandler handles sending notifications to users when it's time with webhooks
func animeSubsWebhookHandler() {
	var todayShows []*entities.ShowAirTime

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
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		animeSubFeedWebhookBlock.Lock()
		animeSubFeedWebhookBlock.Block = false
		animeSubFeedWebhookBlock.Unlock()
		return
	}
	now = now.In(location)

	// Fetch anime subscriptions from MongoDB
	animeSubsMap := db.GetAllAnimeSubs()

	// Iterate over guilds and send notifications
	for guildID, subscriptions := range animeSubsMap {
		if subscriptions == nil || len(subscriptions) == 0 {
			continue
		}

		// Ensure this is a guild subscription
		if !subscriptions[0].GetGuild() {
			continue
		}

		guid := guildID
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

			var updatedSubs []*entities.ShowSub
			for _, guildShow := range subs {
				if guildShow == nil {
					continue
				}

				if guildShow.GetNotified() {
					updatedSubs = append(updatedSubs, guildShow) // Preserve existing notified shows
					continue
				}

				for _, scheduleShow := range todayShows {
					if scheduleShow == nil || scheduleShow.GetDelayed() != "" {
						continue
					}
					if !strings.EqualFold(guildShow.GetShow(), scheduleShow.GetName()) {
						continue
					}
					if !guildSettings.GetDonghua() && scheduleShow.GetDonghua() {
						continue
					}

					// Parse the air hour and minute
					t, err := time.Parse("3:04 PM", scheduleShow.GetAirTime())
					if err != nil {
						log.Println(err)
						continue
					}

					// Form the air date for today
					scheduleDate := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)

					// Check whether the show has already aired today
					if now.Before(scheduleDate) {
						continue
					}

					// Use webhook to post if available
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

					// Mark as notified
					guildShow.SetNotified(true)
					break
				}

				updatedSubs = append(updatedSubs, guildShow)
			}

			// Save updated subscriptions to MongoDB
			db.SetAnimeSubs(guid, updatedSubs, true)

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

// animeSubsHandler handles sending notifications to users when it's time
func animeSubsHandler() {
	var todayShows []*entities.ShowAirTime

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
	Today.RLock()
	if int(Today.Time.Weekday()) != int(now.Weekday()) {
		Today.RUnlock()
		animeSubFeedBlock.Lock()
		animeSubFeedBlock.Block = false
		animeSubFeedBlock.Unlock()
		return
	}
	Today.RUnlock()

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

	// Fetch all anime subscriptions from the database
	animeSubsMap := db.GetAllAnimeSubs()

	// Iterate over all users and their subscriptions
	for userID, subscriptions := range animeSubsMap {
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
			guildIDInt, err := strconv.ParseInt(userID, 10, 64)
			if err != nil {
				continue
			}
			session = config.Mgr.SessionForGuild(guildIDInt)

			// Check if bot is in target guild
			_, err = session.State.Guild(userID)
			if err != nil {
				continue
			}
		} else {
			session = config.Mgr.SessionForDM()
		}

		if isGuild {
			newepisodes := db.GetGuildAutopost(userID, "newepisodes")
			if newepisodes == (entities.Cha{}) {
				continue
			}

			// Check if bot has required permissions
			perms, err := session.State.UserChannelPermissions(session.State.User.ID, newepisodes.GetID())
			if err != nil {
				continue
			}
			if perms&discordgo.PermissionManageWebhooks == discordgo.PermissionManageWebhooks ||
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

			guildSettings = db.GetGuildSettings(userID)
		}

		// Process subscriptions
		var updatedSubs []*entities.ShowSub

		for _, userShow := range subscriptions {
			if userShow == nil {
				continue
			}

			if userShow.GetNotified() {
				updatedSubs = append(updatedSubs, userShow) // Preserve existing notified shows
				continue
			}

			for _, scheduleShow := range todayShows {
				if scheduleShow == nil || scheduleShow.GetDelayed() != "" {
					continue
				}

				if !strings.EqualFold(userShow.GetShow(), scheduleShow.GetName()) {
					continue
				}

				if userShow.GetGuild() && !guildSettings.GetDonghua() && scheduleShow.GetDonghua() {
					continue
				}

				// Parse the air hour and minute
				t, err := time.Parse("3:04 PM", scheduleShow.GetAirTime())
				if err != nil {
					log.Println(err)
					continue
				}

				// Form the air date for Today
				scheduleDate := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)

				// Skip if it's not time yet
				if now.Before(scheduleDate) {
					continue
				}

				// Wait to avoid rate limit
				time.Sleep(time.Millisecond * 150)

				// Send notifications
				if userShow.GetGuild() {
					newepisodes := db.GetGuildAutopost(userID, "newepisodes")
					if newepisodes == (entities.Cha{}) {
						continue
					}

					err = embeds.Subscription(session, scheduleShow, newepisodes.GetID(), newepisodes.GetRoleID())
					if err != nil {
						continue
					}
				} else {
					dm, err := session.UserChannelCreate(userID)
					if err != nil {
						continue
					}
					err = embeds.Subscription(session, scheduleShow, dm.ID, "")
					if err != nil {
						continue
					}
				}

				// Mark as notified
				userShow.SetNotified(true)
				break
			}

			updatedSubs = append(updatedSubs, userShow)
		}

		db.SetAnimeSubs(userID, updatedSubs, isGuild)
	}

	animeSubFeedBlock.Lock()
	animeSubFeedBlock.Block = false
	animeSubFeedBlock.Unlock()
}

func AnimeSubsTimer(_ *discordgo.Session, _ *discordgo.Ready) {
	for range time.NewTicker(1 * time.Minute).C {
		animeSubsHandler()
	}
}

func AnimeSubsWebhookTimer(_ *discordgo.Session, _ *discordgo.Ready) {
	for range time.NewTicker(1 * time.Minute).C {
		animeSubsWebhookHandler()
	}
}

func AnimeSubsWebhooksMapTimer(_ *discordgo.Session, _ *discordgo.Ready) {
	for range time.NewTicker(15 * time.Minute).C {
		WebhooksMapHandler()
	}
}

// ResetSubscriptions resets anime sub notifications status
func ResetSubscriptions() {
	var todayShows []*entities.ShowAirTime

	now := time.Now()
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return
	}
	now = now.In(location)

	// Fetch today's shows
	entities.AnimeSchedule.RLock()
	todayShows = entities.AnimeSchedule.AnimeSchedule[int(now.Weekday())]
	entities.AnimeSchedule.RUnlock()

	// Fetch all anime subscriptions from the database
	animeSubsMap := db.GetAllAnimeSubs()

	for userID, subscriptions := range animeSubsMap {
		if len(subscriptions) == 0 {
			continue
		}

		updated := false
		var updatedSubs []*entities.ShowSub

		// Check if this is a guild (Discord server)
		isGuild := subscriptions[0].GetGuild()

		if isGuild {
			// **Guilds are always subscribed to all shows**
			updatedSubs = make([]*entities.ShowSub, len(todayShows))
			for i, scheduleShow := range todayShows {
				if scheduleShow == nil {
					continue
				}

				// Create a new ShowSub for the guild
				newGuildShow := &entities.ShowSub{}
				newGuildShow.SetShow(scheduleShow.GetName())
				newGuildShow.SetGuild(true)

				// Parse the air hour and minute
				t, err := time.Parse("3:04 PM", scheduleShow.GetAirTime())
				if err != nil {
					log.Println(err)
					continue
				}

				// Form the air date for today
				scheduleDate := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)

				// Reset notification status based on the time
				if now.Before(scheduleDate) {
					newGuildShow.SetNotified(false)
				} else {
					newGuildShow.SetNotified(true)
				}

				updatedSubs[i] = newGuildShow
			}
			updated = true
		} else {
			// **Process regular user subscriptions**
			for _, userShow := range subscriptions {
				if userShow == nil {
					continue
				}

				for _, scheduleShow := range todayShows {
					if scheduleShow == nil {
						continue
					}

					// Check if the show matches
					if !strings.EqualFold(userShow.GetShow(), scheduleShow.GetName()) {
						continue
					}

					// Parse the air hour and minute
					t, err := time.Parse("3:04 PM", scheduleShow.GetAirTime())
					if err != nil {
						log.Println(err)
						continue
					}

					// Form the air date for today
					scheduleDate := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)

					// Reset notification status based on the time
					if now.Before(scheduleDate) {
						userShow.SetNotified(false)
					} else {
						userShow.SetNotified(true)
					}

					updatedSubs = append(updatedSubs, userShow)
					updated = true
				}
			}
		}

		// Save the updated subscriptions to MongoDB
		if updated {
			db.SetAnimeSubs(userID, updatedSubs, isGuild)
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
