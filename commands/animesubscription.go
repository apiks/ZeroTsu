package commands

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"ZeroTsu/config"
	"ZeroTsu/functionality"
)

// Add Notifications for anime episode releases SUBBED
func subscribeCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		showName      string
		hasAiredToday bool

		guildSettings = &functionality.GuildSettings{
			Prefix: ".",
			BotLog: nil,
		}
	)

	if m.GuildID != "" {
		functionality.Mutex.RLock()
		guildSettings = functionality.GuildMap[m.GuildID].GetGuildSettings()
		functionality.Mutex.RUnlock()
	}

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {

		if config.ServerID == "267799767843602452" {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%ssub [anime]`\n\nAnime is the anime name from the schedule command", guildSettings.Prefix))
			if err != nil {
				functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
				return
			}
			return
		}

		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%ssub [anime]`\n\nAnime is the anime name from <https://AnimeSchedule.net> or the schedule command", guildSettings.Prefix))
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	now := time.Now()
	now = now.UTC()

	// Iterates over all of the anime shows saved from AnimeSchedule and checks if it finds one
	functionality.Mutex.RLock()
	animeSchedule := functionality.AnimeSchedule
	animeSubs := functionality.SharedInfo.AnimeSubs
	functionality.Mutex.RUnlock()
Loop:
	for dayInt, dailyShows := range animeSchedule {
		for _, show := range dailyShows {
			if strings.ToLower(show.Name) == commandStrings[1] {

				// Iterate over existing anime subscription users to see if he's already subbed to this show
				for userID, subscriptions := range animeSubs {

					// Skip users that are not this user for performance
					if userID != m.Author.ID {
						continue
					}

					// Check if user is already subscribed to that show and throw an error if so
					for _, userShows := range subscriptions {
						if strings.ToLower(userShows.Show) == strings.ToLower(show.Name) {
							_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: You are already subscribed to `%s`", show.Name))
							if err != nil {
								functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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
					scheduleTime := strings.Split(show.AirTime, ":")
					scheduleHour, err := strconv.Atoi(scheduleTime[0])
					if err != nil {
						continue
					}
					scheduleMinute, err := strconv.Atoi(scheduleTime[1])
					if err != nil {
						continue
					}

					// Form the air date for Today
					scheduleDate := time.Date(now.Year(), now.Month(), now.Day(), scheduleHour, scheduleMinute, now.Second(), now.Nanosecond(), now.Location())

					// Calculates whether the show has already aired Today
					difference := now.Sub(scheduleDate)
					if difference > 0 {
						hasAiredToday = true
					}
				}

				// Add that show to the user anime subs list and break out of loops
				functionality.Mutex.Lock()
				if hasAiredToday {
					functionality.SharedInfo.AnimeSubs[m.Author.ID] = append(functionality.SharedInfo.AnimeSubs[m.Author.ID], &functionality.ShowSub{Show: show.Name, Notified: true})
				} else {
					functionality.SharedInfo.AnimeSubs[m.Author.ID] = append(functionality.SharedInfo.AnimeSubs[m.Author.ID], &functionality.ShowSub{Show: show.Name, Notified: false})
				}
				functionality.Mutex.Unlock()
				showName = show.Name
				break Loop
			}
		}
	}

	if showName == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: That is not a valid airing show name. It has to be airing. Make sure you're using the exact show name from `"+guildSettings.Prefix+"schedule`")
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Write to shared AnimeSubs DB
	functionality.Mutex.Lock()
	err := functionality.AnimeSubsWrite(functionality.SharedInfo.AnimeSubs)
	if err != nil {
		functionality.Mutex.Unlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	functionality.Mutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! You have subscribed to notifications for `%s`", showName))
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// Remove Notifications for anime episode releases SUBBED
func unsubscribeCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		isDeleted bool

		guildSettings = &functionality.GuildSettings{
			Prefix: ".",
			BotLog: nil,
		}
	)

	if m.GuildID != "" {
		functionality.Mutex.RLock()
		guildSettings = functionality.GuildMap[m.GuildID].GetGuildSettings()
		functionality.Mutex.RUnlock()
	}

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		if config.ServerID == "267799767843602452" {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vunsub [anime]`\n\nAnime is the anime name from the schedule command", guildSettings.Prefix))
			if err != nil {
				functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
				return
			}
			return
		}

		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vunsub [anime]`\n\nAnime is the anime name from <https://AnimeSchedule.net> or the schedule command", guildSettings.Prefix))
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Iterate over all of the seasonal anime and see if it's a valid one
	functionality.Mutex.RLock()
	animeSubs := functionality.SharedInfo.AnimeSubs
	functionality.Mutex.RUnlock()

	// Iterate over all of the user's subscriptions and remove the target one if it finds it
LoopShowRemoval:
	for userID, userSubs := range animeSubs {

		// Skip users that are not the message author so they don't delete everyone's subscriptions
		if userID != m.Author.ID {
			continue
		}

		for i, show := range userSubs {
			if strings.ToLower(show.Show) == strings.ToLower(commandStrings[1]) {

				// Delete either the entire object or remove just one item from it
				functionality.Mutex.Lock()
				if len(userSubs) == 1 {
					delete(functionality.SharedInfo.AnimeSubs, userID)
				} else {

					if i < len(functionality.SharedInfo.AnimeSubs[userID])-1 {
						copy(functionality.SharedInfo.AnimeSubs[userID][i:], functionality.SharedInfo.AnimeSubs[userID][i+1:])
					}
					functionality.SharedInfo.AnimeSubs[userID][len(functionality.SharedInfo.AnimeSubs[userID])-1] = nil
					functionality.SharedInfo.AnimeSubs[userID] = functionality.SharedInfo.AnimeSubs[userID][:len(functionality.SharedInfo.AnimeSubs[userID])-1]

				}
				functionality.Mutex.Unlock()

				isDeleted = true
				break LoopShowRemoval
			}
		}
	}

	// Send an error if the target show is not one the user is subscribed to
	if !isDeleted {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: You are not subscribed to `%s`", commandStrings[1]))
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	functionality.Mutex.Lock()
	err := functionality.AnimeSubsWrite(functionality.SharedInfo.AnimeSubs)
	if err != nil {
		functionality.Mutex.Unlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	functionality.Mutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! You have unsubscribed from `%v`", commandStrings[1]))
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
	}
}

// Print all shows the user is subscribed to
func viewSubscriptions(s *discordgo.Session, m *discordgo.Message) {

	var (
		message  string
		messages []string

		guildSettings = &functionality.GuildSettings{
			Prefix: ".",
			BotLog: nil,
		}
	)

	if m.GuildID != "" {
		functionality.Mutex.RLock()
		guildSettings = functionality.GuildMap[m.GuildID].GetGuildSettings()
		functionality.Mutex.RUnlock()
	}

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vsubs`", guildSettings.Prefix))
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Iterates over all of a user's subscribed shows and adds them to the message string
	functionality.Mutex.RLock()
	animeSubs := functionality.SharedInfo.AnimeSubs
	functionality.Mutex.RUnlock()
	for userID, shows := range animeSubs {
		if userID != m.Author.ID {
			continue
		}

		for i := 0; i < len(shows); i++ {
			message += fmt.Sprintf("**%d.** %v\n", i+1, shows[i].Show)
		}
	}

	if len(message) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: You have no active show subscriptions.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Splits the message if it's too big into multiple ones
	if len(message) > 1900 {
		messages = functionality.SplitLongMessage(message)
	}

	if messages == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	for i := 0; i < len(messages); i++ {
		_, err := s.ChannelMessageSend(m.ChannelID, messages[i])
		if err != nil {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Cannot send anime notification subscriptions message.")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
	}
}

// Handles sending notifications to users when it's time
func animeSubsHandler(s *discordgo.Session) {
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in animeSubsHandler")
		}
	}()

	var todayShows []*functionality.ShowAirTime

	now := time.Now()

	functionality.Mutex.RLock()
	if int(Today.Weekday()) != int(now.Weekday()) {
		functionality.Mutex.RUnlock()
		return
	}

	animeSchedule := functionality.AnimeSchedule
	animeSubs := functionality.SharedInfo.AnimeSubs
	functionality.Mutex.RUnlock()

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
			if userShow.Notified {
				continue
			}

			for _, scheduleShow := range todayShows {

				if scheduleShow.Delayed != "" {
					continue
				}

				// Checks if the target show matches
				if strings.ToLower(userShow.Show) != strings.ToLower(scheduleShow.Name) {
					continue
				}

				// Parse the air hour and minute
				scheduleTime := strings.Split(scheduleShow.AirTime, ":")
				scheduleHour, err := strconv.Atoi(scheduleTime[0])
				if err != nil {
					continue
				}
				scheduleMinute, err := strconv.Atoi(scheduleTime[1])
				if err != nil {
					continue
				}

				// Form the air date for Today
				scheduleDate := time.Date(now.Year(), now.Month(), now.Day(), scheduleHour, scheduleMinute, now.Second(), now.Nanosecond(), now.Location())

				// Calculates whether the show has already aired today
				difference := now.Sub(scheduleDate)
				if difference <= 0 {
					continue
				}

				// Sends notification to user DMs if possible, or to guild autopost channel
				if userShow.Guild {

					functionality.Mutex.RLock()
					newepisodes, ok := functionality.GuildMap[userID].Autoposts["newepisodes"]
					if !ok || newepisodes == nil {
						functionality.Mutex.RUnlock()
						continue
					}
					functionality.Mutex.RUnlock()

					// Sends embed
					err := functionality.SubEmbed(s, scheduleShow, newepisodes.ID)
					if err != nil {
						continue
					}

					// Sets the show as notified for that guild
					functionality.Mutex.Lock()
					functionality.SharedInfo.AnimeSubs[userID][subKey].Notified = true
					functionality.Mutex.Unlock()
					continue
				}

				dm, _ := s.UserChannelCreate(userID)
				if config.ServerID != "267799767843602452" {
					_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("**%s __%s__** is out!\nSource: <https://animeschedule.net/shows/%s>", scheduleShow.Name, scheduleShow.Episode, scheduleShow.Key))
				} else {
					_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("**%s __%s__** is out!", scheduleShow.Name, scheduleShow.Episode))
				}

				// Sets the show as notified for that user
				functionality.Mutex.Lock()
				functionality.SharedInfo.AnimeSubs[userID][subKey].Notified = true
				functionality.Mutex.Unlock()
			}
		}
	}

	// Write to shared AnimeSubs DB
	functionality.Mutex.Lock()
	_ = functionality.AnimeSubsWrite(functionality.SharedInfo.AnimeSubs)
	functionality.Mutex.Unlock()
}

func AnimeSubsTimer(s *discordgo.Session, e *discordgo.Ready) {
	for range time.NewTicker(1 * time.Minute).C {
		// Anime Episodes subscription
		animeSubsHandler(s)
	}
}

// Set up all notified bools for anime subscriptions
func resetSubNotified() {
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in resetSubNotified")
		}
	}()

	var todayShows []*functionality.ShowAirTime

	now := time.Now()

	functionality.Mutex.RLock()
	if int(now.Weekday()) == int(Today.Weekday()) {
		return
	}

	animeSchedule := functionality.AnimeSchedule
	animeSubs := functionality.SharedInfo.AnimeSubs
	functionality.Mutex.RUnlock()

	// Fetches Today's shows
	for dayInt, scheduleShows := range animeSchedule {
		// Checks if the target schedule day is Today or not
		if int(now.UTC().Weekday()) != dayInt {
			continue
		}

		// Saves Today's shows
		todayShows = scheduleShows
		break
	}

	// Iterates over all users and their shows and resets notified status
	for userID, subscriptions := range animeSubs {
		for subKey, userShow := range subscriptions {
			for _, scheduleShow := range todayShows {
				if strings.ToLower(scheduleShow.Name) == strings.ToLower(userShow.Show) {
					functionality.Mutex.Lock()
					functionality.SharedInfo.AnimeSubs[userID][subKey].Notified = false
					functionality.Mutex.Unlock()
				}
			}
		}
	}

	// Write to shared AnimeSubs DB
	functionality.Mutex.Lock()
	_ = functionality.AnimeSubsWrite(functionality.SharedInfo.AnimeSubs)
	functionality.Mutex.Unlock()
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

	var todayShows []*functionality.ShowAirTime

	now := time.Now()

	functionality.Mutex.RLock()
	animeSchedule := functionality.AnimeSchedule
	animeSubs := functionality.SharedInfo.AnimeSubs
	functionality.Mutex.RUnlock()

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
		for subKey, userShow := range subscriptions {
			for _, scheduleShow := range todayShows {

				// Checks if the target show matches
				if strings.ToLower(userShow.Show) != strings.ToLower(scheduleShow.Name) {
					continue
				}

				// Parse the air hour and minute
				scheduleTime := strings.Split(scheduleShow.AirTime, ":")
				scheduleHour, err := strconv.Atoi(scheduleTime[0])
				if err != nil {
					continue
				}
				scheduleMinute, err := strconv.Atoi(scheduleTime[1])
				if err != nil {
					continue
				}

				// Form the air date for Today
				scheduleDate := time.Date(now.Year(), now.Month(), now.Day(), scheduleHour, scheduleMinute, nowUTC.Second(), nowUTC.Nanosecond(), nowUTC.Location())

				// Calculates whether the show has already aired today
				difference := now.Sub(scheduleDate)
				functionality.Mutex.Lock()
				if difference >= 0 {
					functionality.SharedInfo.AnimeSubs[userID][subKey].Notified = true
				} else {
					functionality.SharedInfo.AnimeSubs[userID][subKey].Notified = false
				}
				functionality.Mutex.Unlock()
			}
		}
	}

	// Write to shared AnimeSubs DB
	functionality.Mutex.Lock()
	_ = functionality.AnimeSubsWrite(functionality.SharedInfo.AnimeSubs)
	functionality.Mutex.Unlock()
}

func init() {
	functionality.Add(&functionality.Command{
		Execute: subscribeCommand,
		Trigger: "sub",
		Aliases: []string{"subscribe", "subs", "animesub", "subanime", "addsub"},
		Desc:    "Get a message whenever an anime's new episode is released (subbed if possible). Please have your DM settings accept messages from non-friends",
		Module:  "normal",
		DMAble:  true,
	})
	functionality.Add(&functionality.Command{
		Execute: unsubscribeCommand,
		Trigger: "unsub",
		Aliases: []string{"unsubscribe", "unsubs", "unanimesub", "unsubanime", "removesub", "killsub", "stopsub"},
		Desc:    "Stop getting messages whenever an anime's new episodes are released",
		Module:  "normal",
		DMAble:  true,
	})
	functionality.Add(&functionality.Command{
		Execute: viewSubscriptions,
		Trigger: "subs",
		Aliases: []string{"subscriptions", "animesubs", "showsubs", "showsubscriptions", "viewsubs", "viewsubscriptions"},
		Desc:    "Print which shows you are getting new episode notifications for",
		Module:  "normal",
		DMAble:  true,
	})
}
