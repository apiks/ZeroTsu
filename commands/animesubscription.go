package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/embeds"
	"github.com/r-anime/ZeroTsu/entities"
	"github.com/r-anime/ZeroTsu/events"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var animeSubFeedBlock events.Block

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

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! You have unsubscribed from `%v`", commandStrings[1]))
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
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

// Handles sending notifications to users when it's time
func animeSubsHandler(s *discordgo.Session) {
	now := time.Now()
	var todayShows []*entities.ShowAirTime

	entities.Mutex.Lock()
	if int(Today.Weekday()) != int(now.Weekday()) {
		entities.Mutex.Unlock()
		return
	}
	entities.Mutex.Unlock()

	animeSubFeedBlock.RLock()
	if animeSubFeedBlock.Block {
		animeSubFeedBlock.RUnlock()
		return
	}
	animeSubFeedBlock.RUnlock()

	animeSubFeedBlock.Lock()
	animeSubFeedBlock.Block = true
	animeSubFeedBlock.Unlock()

	now = now.UTC()

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

	// Iterates over all users and their shows and sends notifications if need be
	for userID, subscriptions := range entities.SharedInfo.GetAnimeSubsMap() {
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
				time.Sleep(time.Millisecond * 100)

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
					entities.SharedInfo.Lock()
					entities.SharedInfo.AnimeSubs[userID][subKey].SetNotified(true)
					entities.SharedInfo.Unlock()
					continue
				}

				dm, err := s.UserChannelCreate(userID)
				if err != nil {
					continue
				}
				_, err = s.ChannelMessageSend(dm.ID, fmt.Sprintf("**%s __%s__** is out!\nSource: <https://animeschedule.net/anime/%s>", scheduleShow.GetName(), scheduleShow.GetEpisode(), scheduleShow.GetKey()))
				if err != nil {
					continue
				}

				// Sets the show as notified for that user
				entities.SharedInfo.Lock()
				entities.SharedInfo.AnimeSubs[userID][subKey].SetNotified(true)
				entities.SharedInfo.Unlock()
			}
		}
	}

	// Write to shared AnimeSubs DB
	_ = entities.AnimeSubsWrite(entities.SharedInfo.GetAnimeSubsMap())

	animeSubFeedBlock.Lock()
	animeSubFeedBlock.Block = false
	animeSubFeedBlock.Unlock()
}

func AnimeSubsTimer(s *discordgo.Session, _ *discordgo.Ready) {
	for range time.NewTicker(2 * time.Minute).C {
		animeSubFeedBlock.RLock()
		if animeSubFeedBlock.Block {
			animeSubFeedBlock.RUnlock()
			return
		}
		animeSubFeedBlock.RUnlock()

		// Anime Episodes subscription
		animeSubsHandler(s)
	}
}

// Resets anime sub notifications status on bot start
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
