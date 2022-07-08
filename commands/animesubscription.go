package commands

import (
	"fmt"
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
	"golang.org/x/sync/errgroup"

	"github.com/bwmarrin/discordgo"
)

var animeSubFeedBlock events.Block

// subscribeCommand subscribes to notifications for anime episode releases SUBBED
func subscribeCommand(title, authorID string) string {
	var (
		now           = time.Now().UTC()
		showExists    bool
		hasAiredToday bool
	)

	// Iterates over all of the anime shows saved from AnimeSchedule and checks if it finds one
	entities.AnimeSchedule.RLock()
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

				// Iterate over existing anime subscription users to see if he's already subbed to this show
				for userID, subscriptions := range entities.SharedInfo.GetAnimeSubsMap() {
					if subscriptions == nil {
						continue
					}

					// Skip users that are not the author
					if userID != authorID {
						continue
					}

					// Check if user is already subscribed to that show and throw an error if so
					for _, userShow := range subscriptions {
						if userShow == nil {
							continue
						}

						if strings.ToLower(userShow.GetShow()) == strings.ToLower(show.GetName()) {
							entities.AnimeSchedule.RUnlock()
							return fmt.Sprintf("Error: You are already subscribed to `%s`", show.GetName())
						}
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
					scheduleDate := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), now.Second(), now.Nanosecond(), now.Location())

					// Calculates whether the show has already aired Today
					difference := now.Sub(scheduleDate)
					if difference > 0 {
						hasAiredToday = true
					}
				}

				// Add that show to the user anime subs list and break out of loops
				entities.SharedInfo.Lock()
				if hasAiredToday {
					entities.SharedInfo.AnimeSubs[authorID] = append(entities.SharedInfo.AnimeSubs[authorID], entities.NewShowSub(show.GetName(), true, false))
				} else {
					entities.SharedInfo.AnimeSubs[authorID] = append(entities.SharedInfo.AnimeSubs[authorID], entities.NewShowSub(show.GetName(), false, false))
				}
				entities.SharedInfo.Unlock()
				break Loop
			}
		}
	}
	entities.AnimeSchedule.RUnlock()

	if !showExists {
		return "Error: That is not a valid airing show name. It has to be airing. Make sure you're using the exact romaji anime title from `/schedule` or AnimeSchedule.net."
	}

	// Write to shared AnimeSubs DB
	err := entities.AnimeSubsWrite(entities.SharedInfo.GetAnimeSubsMap())
	if err != nil {
		return err.Error()
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
			return
		}
		return
	}

	now := time.Now().UTC()

	// Iterates over all of the anime shows saved from AnimeSchedule and checks if it finds one
	entities.AnimeSchedule.RLock()
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

				// Iterate over existing anime subscription users to see if he's already subbed to this show
				for userID, subscriptions := range entities.SharedInfo.GetAnimeSubsMap() {
					if subscriptions == nil {
						continue
					}

					// Skip users that are not this user for performance
					if userID != m.Author.ID {
						continue
					}

					// Check if user is already subscribed to that show and throw an error if so
					for _, userShows := range subscriptions {
						if userShows == nil {
							continue
						}

						if strings.ToLower(userShows.GetShow()) == strings.ToLower(show.GetName()) {
							name := show.GetName()
							entities.AnimeSchedule.RUnlock()
							_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: You are already subscribed to `%s`", name))
							if err != nil {
								common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
								return
							}
							return
						}
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
					scheduleDate := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), now.Second(), now.Nanosecond(), now.Location())

					// Calculates whether the show has already aired Today
					difference := now.Sub(scheduleDate)
					if difference > 0 {
						hasAiredToday = true
					}
				}

				// Add that show to the user anime subs list and break out of loops
				entities.SharedInfo.Lock()
				if hasAiredToday {
					entities.SharedInfo.AnimeSubs[m.Author.ID] = append(entities.SharedInfo.AnimeSubs[m.Author.ID], entities.NewShowSub(show.GetName(), true, false))
				} else {
					entities.SharedInfo.AnimeSubs[m.Author.ID] = append(entities.SharedInfo.AnimeSubs[m.Author.ID], entities.NewShowSub(show.GetName(), false, false))
				}
				entities.SharedInfo.Unlock()
				break Loop
			}
		}
	}
	entities.AnimeSchedule.RUnlock()

	if showName == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: That is not a valid airing show name. It has to be airing. Make sure you're using the exact show name from `"+guildSettings.GetPrefix()+"schedule`")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Write to shared AnimeSubs DB
	err = entities.AnimeSubsWrite(entities.SharedInfo.GetAnimeSubsMap())
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! You have subscribed to notifications for `%s`", showName))
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// unsubscribeCommand removes a subscription for notifications for anime episode releases SUBBED
func unsubscribeCommand(title, authorID string) string {
	var isDeleted bool

	// Iterate over all of the user's subscriptions and remove the target one if it finds it
LoopShowRemoval:
	for userID, userSubs := range entities.SharedInfo.GetAnimeSubsMap() {

		// Skip users that are not the message author so they don't delete everyone's subscriptions
		if userID != authorID {
			continue
		}

		for i, show := range userSubs {
			if show == nil {
				continue
			}

			if strings.ToLower(show.GetShow()) == strings.ToLower(title) {

				// Delete either the entire object or remove just one item from it
				entities.SharedInfo.Lock()
				if len(userSubs) == 1 {
					delete(entities.SharedInfo.AnimeSubs, userID)
				} else {
					if i < len(entities.SharedInfo.AnimeSubs[userID])-1 {
						copy(entities.SharedInfo.AnimeSubs[userID][i:], entities.SharedInfo.AnimeSubs[userID][i+1:])
					}
					entities.SharedInfo.AnimeSubs[userID][len(entities.SharedInfo.AnimeSubs[userID])-1] = nil
					entities.SharedInfo.AnimeSubs[userID] = entities.SharedInfo.AnimeSubs[userID][:len(entities.SharedInfo.AnimeSubs[userID])-1]
				}
				entities.SharedInfo.Unlock()

				isDeleted = true
				break LoopShowRemoval
			}
		}
	}

	// Send an error if the target show is not one the user is subscribed to
	if !isDeleted {
		return fmt.Sprintf("Error: You are not subscribed to `%s`", title)
	}

	// Write to shared AnimeSubs DB
	err := entities.AnimeSubsWrite(entities.SharedInfo.GetAnimeSubsMap())
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("Success! You have unsubscribed from `%s`", title)
}

// unsubscribeCommandHandler removes a subscription for notifications for anime episode releases SUBBED
func unsubscribeCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	var (
		isDeleted bool

		err           error
		guildSettings = entities.GuildSettings{Prefix: "."}
	)

	if m.GuildID != "" {
		guildSettings = db.GetGuildSettings(m.GuildID)
	}

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vunsub [anime]`\n\nAnime is the anime name from <https://AnimeSchedule.net> or the schedule command", guildSettings.GetPrefix()))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Iterate over all of the user's subscriptions and remove the target one if it finds it
LoopShowRemoval:
	for userID, userSubs := range entities.SharedInfo.GetAnimeSubsMap() {

		// Skip users that are not the message author so they don't delete everyone's subscriptions
		if userID != m.Author.ID {
			continue
		}

		for i, show := range userSubs {
			if show == nil {
				continue
			}

			if strings.ToLower(show.GetShow()) == strings.ToLower(commandStrings[1]) {

				// Delete either the entire object or remove just one item from it
				entities.SharedInfo.Lock()
				if len(userSubs) == 1 {
					delete(entities.SharedInfo.AnimeSubs, userID)
				} else {
					if i < len(entities.SharedInfo.AnimeSubs[userID])-1 {
						copy(entities.SharedInfo.AnimeSubs[userID][i:], entities.SharedInfo.AnimeSubs[userID][i+1:])
					}
					entities.SharedInfo.AnimeSubs[userID][len(entities.SharedInfo.AnimeSubs[userID])-1] = nil
					entities.SharedInfo.AnimeSubs[userID] = entities.SharedInfo.AnimeSubs[userID][:len(entities.SharedInfo.AnimeSubs[userID])-1]
				}
				entities.SharedInfo.Unlock()

				isDeleted = true
				break LoopShowRemoval
			}
		}
	}

	// Send an error if the target show is not one the user is subscribed to
	if !isDeleted {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: You are not subscribed to `%s`", commandStrings[1]))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	err = entities.AnimeSubsWrite(entities.SharedInfo.GetAnimeSubsMap())
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! You have unsubscribed from `%s`", commandStrings[1]))
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// viewSubscriptions prints out all the anime the user is subscribed to
func viewSubscriptions(authorID string) []string {
	var (
		message  string
		messages []string
	)

	// Iterates over all of a user's subscribed shows and adds them to the message string
	for userID, shows := range entities.SharedInfo.GetAnimeSubsMap() {
		if shows == nil {
			continue
		}

		if userID != authorID {
			continue
		}

		for i := 0; i < len(shows); i++ {
			message += fmt.Sprintf("**%d.** %s\n", i+1, shows[i].GetShow())
		}
	}

	if len(message) == 0 {
		return []string{"Error: You have no active anime subscriptions."}
	}

	// Splits the message if it's too big into multiple ones
	if len(message) > 1900 {
		messages = common.SplitLongMessage(message)
	}

	if messages == nil {
		return []string{message}
	}

	return messages
}

// viewSubscriptionsHandler prints out all the anime the user is subscribed to
func viewSubscriptionsHandler(s *discordgo.Session, m *discordgo.Message) {
	var (
		message  string
		messages []string

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
			return
		}
		return
	}

	// Iterates over all of a user's subscribed shows and adds them to the message string
	for userID, shows := range entities.SharedInfo.GetAnimeSubsMap() {
		if shows == nil {
			continue
		}

		if userID != m.Author.ID {
			continue
		}

		for i := 0; i < len(shows); i++ {
			message += fmt.Sprintf("**%d.** %s\n", i+1, shows[i].GetShow())
		}
	}

	if len(message) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: You have no active show subscriptions.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Splits the message if it's too big into multiple ones
	if len(message) > 1900 {
		messages = common.SplitLongMessage(message)
	}

	if messages == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	for i := 0; i < len(messages); i++ {
		_, err := s.ChannelMessageSend(m.ChannelID, messages[i])
		if err != nil {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Cannot send anime notification subscriptions message.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
	}
}

// animeSubsHandler handles sending notifications to users when it's time
func animeSubsHandler() {
	var (
		now        = time.Now()
		todayShows []*entities.ShowAirTime
		eg         errgroup.Group
	)

	Today.RLock()
	if int(Today.Time.Weekday()) != int(now.Weekday()) {
		Today.RUnlock()
		return
	}
	Today.RUnlock()

	animeSubFeedBlock.Lock()
	if animeSubFeedBlock.Block {
		animeSubFeedBlock.Unlock()
		return
	}
	animeSubFeedBlock.Block = true
	animeSubFeedBlock.Unlock()

	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		return
	}
	now = now.In(location)

	// Fetches Today's shows
	entities.AnimeSchedule.RLock()
	for _, show := range entities.AnimeSchedule.AnimeSchedule[int(now.Weekday())] {
		todayShows = append(todayShows, show)
	}
	entities.AnimeSchedule.RUnlock()

	// Iterates over all users and their shows and sends notifications if need be
	for userID, subscriptions := range entities.SharedInfo.GetAnimeSubsMap() {
		if subscriptions == nil {
			continue
		}

		var (
			session       *discordgo.Session
			guildSettings entities.GuildSettings
			isGuild       bool
		)
		if len(subscriptions) >= 1 && subscriptions[0].GetGuild() {
			isGuild = true
		}
		if isGuild {
			guildIDInt, err := strconv.ParseInt(userID, 10, 64)
			if err != nil {
				continue
			}
			session = config.Mgr.SessionForGuild(guildIDInt)
			guildSettings = db.GetGuildSettings(userID)
		} else {
			session = config.Mgr.SessionForDM()
		}

		for subKey, userShow := range subscriptions {
			if userShow == nil {
				continue
			}
			if userShow.GetNotified() {
				continue
			}

			for _, scheduleShow := range todayShows {
				if scheduleShow == nil {
					continue
				}
				if scheduleShow.GetDelayed() != "" {
					continue
				}
				if strings.ToLower(userShow.GetShow()) != strings.ToLower(scheduleShow.GetName()) {
					continue
				}
				if userShow.GetNotified() {
					continue
				}
				if userShow.GetGuild() {
					if !guildSettings.GetDonghua() && scheduleShow.GetDonghua() {
						continue
					}
				}

				// Parse the air hour and minute
				t, err := time.Parse("3:04 PM", scheduleShow.GetAirTime())
				if err != nil {
					log.Println(err)
					continue
				}

				// Form the air date for Today
				scheduleDate := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), now.Second(), now.Nanosecond(), now.Location())

				// Calculates whether the show has already aired today
				difference := now.Sub(scheduleDate)
				if difference <= 0 {
					continue
				}

				// Wait some milliseconds so it doesn't hit the rate limit easily
				time.Sleep(time.Millisecond * 50)

				uid := userID
				us := userShow
				ss := scheduleShow
				sk := subKey
				s := session
				eg.Go(func() error {
					// Sends notification to user DMs if possible, or to guild autopost channel
					if us.GetGuild() {
						newepisodes := db.GetGuildAutopost(uid, "newepisodes")
						if newepisodes == (entities.Cha{}) {
							return nil
						}

						// Sends embed in Guild
						err = embeds.Subscription(s, ss, newepisodes.GetID())
						if err != nil {
							return err
						}

						// Sets the show as notified for that guild
						entities.SharedInfo.Lock()
						entities.SharedInfo.AnimeSubs[uid][sk].SetNotified(true)
						entities.SharedInfo.Unlock()

						return nil
					}

					// Sends embed in DMs
					dm, err := s.UserChannelCreate(uid)
					if err != nil {
						return err
					}
					err = embeds.Subscription(s, ss, dm.ID)
					if err != nil {
						return err
					}

					// Sets the show as notified for that user
					entities.SharedInfo.Lock()
					entities.SharedInfo.AnimeSubs[uid][sk].SetNotified(true)
					entities.SharedInfo.Unlock()

					return nil
				})

				// Wait some milliseconds so it doesn't hit the rate limit easily
				time.Sleep(time.Millisecond * 50)
			}
		}
	}

	err = eg.Wait()
	if err != nil {
		log.Println(err)
	}

	// Write to shared AnimeSubs DB
	_ = entities.AnimeSubsWrite(entities.SharedInfo.GetAnimeSubsMap())

	animeSubFeedBlock.Lock()
	animeSubFeedBlock.Block = false
	animeSubFeedBlock.Unlock()
}

func AnimeSubsTimer(_ *discordgo.Session, _ *discordgo.Ready) {
	for range time.NewTicker(1 * time.Minute).C {
		animeSubFeedBlock.RLock()
		if animeSubFeedBlock.Block {
			animeSubFeedBlock.RUnlock()
			return
		}
		animeSubFeedBlock.RUnlock()

		// Anime Episodes subscription
		animeSubsHandler()
	}
}

// ResetSubscriptions Resets anime sub notifications status
func ResetSubscriptions() {
	var todayShows []*entities.ShowAirTime
	now := time.Now()

	// Fetches Today's shows
	entities.AnimeSchedule.RLock()
	for dayInt, scheduleShows := range entities.AnimeSchedule.AnimeSchedule {
		// Checks if the target schedule day is Today or not
		if int(now.Weekday()) != dayInt {
			continue
		}

		// Saves Today's shows
		todayShows = scheduleShows
		break
	}
	entities.AnimeSchedule.RUnlock()

	nowUTC := now.UTC()

	for userID, subscriptions := range entities.SharedInfo.GetAnimeSubsMap() {
		if subscriptions == nil {
			continue
		}

		for subKey, userShow := range subscriptions {
			if userShow == nil {
				continue
			}

			for _, scheduleShow := range todayShows {
				if scheduleShow == nil {
					continue
				}

				// Checks if the target show matches
				if strings.ToLower(userShow.GetShow()) != strings.ToLower(scheduleShow.GetName()) {
					continue
				}

				// Parse the air hour and minute
				t, err := time.Parse("3:04 PM", scheduleShow.GetAirTime())
				if err != nil {
					log.Println(err)
					continue
				}

				// Form the air date for Today
				scheduleDate := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), nowUTC.Second(), nowUTC.Nanosecond(), nowUTC.Location())

				// Calculates whether the show has already aired today
				difference := now.Sub(scheduleDate)
				entities.SharedInfo.Lock()
				if difference >= 0 {
					entities.SharedInfo.AnimeSubs[userID][subKey].SetNotified(true)
				} else {
					entities.SharedInfo.AnimeSubs[userID][subKey].SetNotified(false)
				}
				entities.SharedInfo.Unlock()
			}
		}
	}

	// Write to shared AnimeSubs DB
	_ = entities.AnimeSubsWrite(entities.SharedInfo.GetAnimeSubsMap())
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
			if i.ApplicationCommandData().Options != nil {
				for _, option := range i.ApplicationCommandData().Options {
					if option.Name == "anime" {
						anime = option.StringValue()
					}
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: subscribeCommand(anime, i.Member.User.ID),
				},
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
			if i.ApplicationCommandData().Options != nil {
				for _, option := range i.ApplicationCommandData().Options {
					if option.Name == "anime" {
						anime = option.StringValue()
					}
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: unsubscribeCommand(anime, i.Member.User.ID),
				},
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
			messages := viewSubscriptions(i.Member.User.ID)
			if messages == nil {
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: messages[0],
				},
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
