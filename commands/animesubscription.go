package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/embeds"
	"github.com/r-anime/ZeroTsu/entities"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Add Notifications for anime episode releases SUBBED
func subscribeCommand(s *discordgo.Session, m *discordgo.Message) {
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

	now := time.Now()
	now = now.UTC()

	// Iterates over all of the anime shows saved from AnimeSchedule and checks if it finds one
	entities.Mutex.RLock()
	animeSchedule := entities.AnimeSchedule
	animeSubs := entities.SharedInfo.GetAnimeSubsMap()
	entities.Mutex.RUnlock()
Loop:
	for dayInt, dailyShows := range animeSchedule {
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
				for userID, subscriptions := range animeSubs {
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
							_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: You are already subscribed to `%s`", show.GetName()))
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
				entities.Mutex.Lock()
				if hasAiredToday {
					entities.SharedInfo.GetAnimeSubsMap()[m.Author.ID] = append(entities.SharedInfo.GetAnimeSubsMap()[m.Author.ID], entities.NewShowSub(show.GetName(), true, false))
				} else {
					entities.SharedInfo.GetAnimeSubsMap()[m.Author.ID] = append(entities.SharedInfo.GetAnimeSubsMap()[m.Author.ID], entities.NewShowSub(show.GetName(), false, false))
				}
				entities.Mutex.Unlock()
				break Loop
			}
		}
	}

	if showName == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: That is not a valid airing show name. It has to be airing. Make sure you're using the exact show name from `"+guildSettings.GetPrefix()+"schedule`")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Write to shared AnimeSubs DB
	entities.Mutex.Lock()
	err = entities.AnimeSubsWrite(entities.SharedInfo.GetAnimeSubsMap())
	if err != nil {
		entities.Mutex.Unlock()
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	entities.Mutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! You have subscribed to notifications for `%s`", showName))
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// Remove Notifications for anime episode releases SUBBED
func unsubscribeCommand(s *discordgo.Session, m *discordgo.Message) {

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

	// Iterate over all of the seasonal anime and see if it's a valid one
	entities.Mutex.RLock()
	animeSubs := entities.SharedInfo.GetAnimeSubsMap()
	entities.Mutex.RUnlock()

	// Iterate over all of the user's subscriptions and remove the target one if it finds it
LoopShowRemoval:
	for userID, userSubs := range animeSubs {

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
				entities.Mutex.Lock()
				if len(userSubs) == 1 {
					delete(entities.SharedInfo.GetAnimeSubsMap(), userID)
				} else {
					if i < len(entities.SharedInfo.GetAnimeSubsMap()[userID])-1 {
						copy(entities.SharedInfo.GetAnimeSubsMap()[userID][i:], entities.SharedInfo.GetAnimeSubsMap()[userID][i+1:])
					}
					entities.SharedInfo.GetAnimeSubsMap()[userID][len(entities.SharedInfo.GetAnimeSubsMap()[userID])-1] = nil
					entities.SharedInfo.GetAnimeSubsMap()[userID] = entities.SharedInfo.GetAnimeSubsMap()[userID][:len(entities.SharedInfo.GetAnimeSubsMap()[userID])-1]

				}
				entities.Mutex.Unlock()

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

	entities.Mutex.Lock()
	err = entities.AnimeSubsWrite(entities.SharedInfo.GetAnimeSubsMap())
	if err != nil {
		entities.Mutex.Unlock()
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	entities.Mutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! You have unsubscribed from `%v`", commandStrings[1]))
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
	}
}

// Print all shows the user is subscribed to
func viewSubscriptions(s *discordgo.Session, m *discordgo.Message) {

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
	entities.Mutex.RLock()
	animeSubs := entities.SharedInfo.GetAnimeSubsMap()
	entities.Mutex.RUnlock()
	for userID, shows := range animeSubs {
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

// Handles sending notifications to users when it's time
func animeSubsHandler(s *discordgo.Session) {
	var todayShows []*entities.ShowAirTime
	now := time.Now()

	entities.Mutex.RLock()
	if int(Today.Weekday()) != int(now.Weekday()) {
		entities.Mutex.RUnlock()
		return
	}

	animeSchedule := entities.AnimeSchedule
	animeSubs := entities.SharedInfo.GetAnimeSubsMap()
	entities.Mutex.RUnlock()

	now = now.UTC()

	// Fetches Today's shows
	for dayInt, scheduleShows := range animeSchedule {
		// Checks if the target schedule day is Today or not
		if int(now.Weekday()) != dayInt {
			continue
		}

		// Saves Today's shows
		todayShows = scheduleShows
		break
	}

	// Iterates over all users and their shows and sends notifications if need be
	for userID, subscriptions := range animeSubs {
		if subscriptions == nil {
			continue
		}
		for subKey, userShow := range subscriptions {
			if userShow == nil {
				continue
			}

			// Checks if the user has already been notified for this show
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
				scheduleDate := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), now.Second(), now.Nanosecond(), now.Location())

				// Calculates whether the show has already aired today
				difference := now.Sub(scheduleDate)
				if difference <= 0 {
					continue
				}

				// Wait some milliseconds so it doesn't hit the rate limit easily
				time.Sleep(time.Millisecond * 200)

				// Sends notification to user DMs if possible, or to guild autopost channel
				if userShow.GetGuild() {

					newepisodes := db.GetGuildAutopost(userID, "newepisodes")
					if newepisodes == (entities.Cha{}) {
						continue
					}

					// Sends embed
					err := embeds.Subscription(s, scheduleShow, newepisodes.GetID())
					if err != nil {
						continue
					}

					// Sets the show as notified for that guild
					entities.Mutex.Lock()
					entities.SharedInfo.GetAnimeSubsMap()[userID][subKey].SetNotified(true)
					entities.Mutex.Unlock()
					continue
				}

				dm, _ := s.UserChannelCreate(userID)
				_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("**%s __%s__** is out!\nSource: <https://animeschedule.net/shows/%s>", scheduleShow.GetName(), scheduleShow.GetEpisode(), scheduleShow.GetKey()))

				// Sets the show as notified for that user
				entities.Mutex.Lock()
				entities.SharedInfo.GetAnimeSubsMap()[userID][subKey].SetNotified(true)
				entities.Mutex.Unlock()
			}
		}
	}

	// Write to shared AnimeSubs DB
	entities.Mutex.Lock()
	_ = entities.AnimeSubsWrite(entities.SharedInfo.GetAnimeSubsMap())
	entities.Mutex.Unlock()
}

func AnimeSubsTimer(s *discordgo.Session, e *discordgo.Ready) {
	for range time.NewTicker(2 * time.Minute).C {
		// Anime Episodes subscription
		animeSubsHandler(s)
	}
}

// Resets anime sub notifications status on bot start
func ResetSubscriptions() {
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in ResetSubscriptions")
		}
	}()

	var todayShows []*entities.ShowAirTime

	now := time.Now()

	entities.Mutex.RLock()
	animeSchedule := entities.AnimeSchedule
	animeSubs := entities.SharedInfo.GetAnimeSubsMap()
	entities.Mutex.RUnlock()

	// Fetches Today's shows
	for dayInt, scheduleShows := range animeSchedule {
		// Checks if the target schedule day is Today or not
		if int(now.Weekday()) != dayInt {
			continue
		}

		// Saves Today's shows
		todayShows = scheduleShows
		break
	}

	nowUTC := now.UTC()

	for userID, subscriptions := range animeSubs {
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
				entities.Mutex.Lock()
				if difference >= 0 {
					entities.SharedInfo.GetAnimeSubsMap()[userID][subKey].SetNotified(true)
				} else {
					entities.SharedInfo.GetAnimeSubsMap()[userID][subKey].SetNotified(false)
				}
				entities.Mutex.Unlock()
			}
		}
	}

	// Write to shared AnimeSubs DB
	entities.Mutex.Lock()
	_ = entities.AnimeSubsWrite(entities.SharedInfo.GetAnimeSubsMap())
	entities.Mutex.Unlock()
}

func init() {
	Add(&Command{
		Execute: subscribeCommand,
		Trigger: "sub",
		Aliases: []string{"subscribe", "subs", "animesub", "subanime", "addsub"},
		Desc:    "Get a message whenever an anime's new episode is released (subbed if possible). Please have your DM settings accept messages from non-friends",
		Module:  "normal",
		DMAble:  true,
	})
	Add(&Command{
		Execute: unsubscribeCommand,
		Trigger: "unsub",
		Aliases: []string{"unsubscribe", "unsubs", "unanimesub", "unsubanime", "removesub", "killsub", "stopsub"},
		Desc:    "Stop getting messages whenever an anime's new episodes are released",
		Module:  "normal",
		DMAble:  true,
	})
	Add(&Command{
		Execute: viewSubscriptions,
		Trigger: "subs",
		Aliases: []string{"subscriptions", "animesubs", "showsubs", "showsubscriptions", "viewsubs", "viewsubscriptions"},
		Desc:    "Print which shows you are getting new episode notifications for",
		Module:  "normal",
		DMAble:  true,
	})
}
