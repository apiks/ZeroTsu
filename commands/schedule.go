package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"

	"github.com/bwmarrin/discordgo"
)

// scheduleCommand Prints out the target day's airing anime times, fetched from AnimeSchedule.net
func scheduleCommand(targetDay, guildID string) []string {
	var (
		currentDay    = int(time.Now().Weekday())
		day           = -1
		message       string
		messages      []string
		guildSettings entities.GuildSettings
		donghua       = true
	)

	// Disable donghuas if disabled in the target guild
	if guildID != "" {
		guildSettings = db.GetGuildSettings(guildID)
		if !guildSettings.GetDonghua() {
			donghua = false
		}
	}

	if targetDay == "" {
		// Get the current day's schedule in print format
		message = getDaySchedule(currentDay, donghua)
	} else {
		// Else get the target day's schedule in print format
		switch targetDay {
		case "sunday", "sundays", "sun":
			day = 0
		case "monday", "mondays", "mon":
			day = 1
		case "tuesday", "tuesdays", "tue", "tues":
			day = 2
		case "wednesday", "wednesdays", "wed":
			day = 3
		case "thursday", "thursdays", "thu", "thurs", "thur":
			day = 4
		case "friday", "fridays", "fri":
			day = 5
		case "saturday", "saturdays", "sat":
			day = 6
		}

		// Check if it's a valid int
		if day < 0 || day > 6 {
			return []string{"Error: Cannot parse that day."}
		}

		message = getDaySchedule(day, donghua)
	}

	message += "\n\n**Full Week:** <https://AnimeSchedule.net>"

	// Splits the message if it's too big into multiple ones
	if len(message) > 1900 {
		messages = common.SplitLongMessage(message)
	}

	if messages == nil {
		return []string{message}
	}

	return messages
}

// scheduleCommandHandler Prints out the target day's airing anime times, fetched from AnimeSchedule.net
func scheduleCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	var (
		currentDay   = int(time.Now().Weekday())
		day          = -1
		printMessage string
	)
	guildSettings := db.GetGuildSettings(m.GuildID)
	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		// Get the current day's schedule in print format
		printMessage = getDaySchedule(currentDay, guildSettings.GetDonghua())
	} else {
		// Else get the target day's schedule in print format
		switch commandStrings[1] {
		case "sunday", "sundays", "sun":
			day = 0
		case "monday", "mondays", "mon":
			day = 1
		case "tuesday", "tuesdays", "tue", "tues":
			day = 2
		case "wednesday", "wednesdays", "wed":
			day = 3
		case "thursday", "thursdays", "thu", "thurs", "thur":
			day = 4
		case "friday", "fridays", "fri":
			day = 5
		case "saturday", "saturdays", "sat":
			day = 6
		}

		// Check if it's a valid int
		if day < 0 || day > 6 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Cannot parse that day.")
			if err != nil {
				if m.GuildID != "" {
					guildSettings := db.GetGuildSettings(m.GuildID)
					common.LogError(s, guildSettings.BotLog, err)
				}
				return
			}
			return
		}

		printMessage = getDaySchedule(day, guildSettings.GetDonghua())
	}

	printMessage += "\n\n**Full Week:** <https://AnimeSchedule.net>"

	// Print the daily schedule
	_, _ = s.ChannelMessageSend(m.ChannelID, printMessage)
}

// Gets a target weekday's anime schedule
func getDaySchedule(weekday int, donghua bool) string {
	var (
		printMessage = fmt.Sprintf("**__%s:__**\n\n", time.Weekday(weekday).String())
		targetDay    = common.WeekStart(time.Now().ISOWeek())
	)
	for int(targetDay.Weekday()) != weekday {
		targetDay = targetDay.AddDate(0, 0, 1)
	}

	entities.AnimeSchedule.RLock()
	defer entities.AnimeSchedule.RUnlock()

	for dayInt, showSlice := range entities.AnimeSchedule.AnimeSchedule {
		if showSlice == nil {
			continue
		}

		if dayInt != weekday {
			continue
		}

		for _, show := range showSlice {
			if show == nil {
				continue
			}
			if !donghua && show.GetDonghua() {
				continue
			}

			// Parses the time in a proper time object
			var realTime time.Time
			t, err := time.Parse("3:04 PM", show.GetAirTime())
			if err != nil {
				log.Println(err)
				continue
			}
			realTime = time.Date(targetDay.Year(), targetDay.Month(), targetDay.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)

			if show.GetDelayed() != "" {
				printMessage += fmt.Sprintf("- **%s** - %s %s - <t:%d:t>\n", show.GetName(), show.GetEpisode(), show.GetDelayed(), realTime.UTC().Unix())
			} else {
				printMessage += fmt.Sprintf("- **%s** - %s - <t:%d:t>\n", show.GetName(), show.GetEpisode(), realTime.UTC().Unix())
			}

		}
		break
	}

	return printMessage
}

// UpdateAnimeSchedule fetches the AnimeSchedule.net timetable via scraping
func UpdateAnimeSchedule() {
	var timetableAnime []entities.ASAnime
	var subAnimeExists = make(map[string]bool)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 300 * time.Second,
	}

	// Create and modify HTTP request before sending
	request, err := http.NewRequest("GET", "https://animeschedule.net/api/v3/timetables", nil)
	if err != nil {
		log.Println(err)
		return
	}
	request.Header.Set("User-Agent", common.UserAgent)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.AnimeScheduleAppSecret))

	// Make request
	response, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&timetableAnime)
	if err != nil {
		log.Println(err)
		return
	}

	// Check all sub anime for later filtering
	for _, anime := range timetableAnime {
		if anime.GetAirType() != "sub" {
			continue
		}
		subAnimeExists[anime.GetRoute()] = true
	}

	// Reset map
	entities.AnimeSchedule.Lock()
	defer entities.AnimeSchedule.Unlock()
	entities.AnimeSchedule.AnimeSchedule = make(map[int][]*entities.ShowAirTime)

	// Add timetable anime
	for _, anime := range timetableAnime {
		if anime.GetAirType() == "dub" {
			continue
		}
		if anime.GetAirType() == "raw" {
			if _, ok := subAnimeExists[anime.GetRoute()]; ok {
				continue
			}
		}

		episodeStr := fmt.Sprintf("Ep %s", strconv.Itoa(anime.GetEpisodeNumber()))
		if anime.GetEpisodes() == anime.GetEpisodeNumber() {
			episodeStr = fmt.Sprintf("Ep %sF", strconv.Itoa(anime.GetEpisodeNumber()))
		}

		delayedText := ""
		if anime.GetDelayedFrom() != (time.Time{}) || anime.GetDelayedUntil() != (time.Time{}) {
			if anime.GetDelayedUntil() == (time.Time{}) {
				if !anime.GetEpisodeDate().Before(anime.GetDelayedFrom()) {
					delayedText = "Delayed"
				}
			} else if anime.GetDelayedFrom() == (time.Time{}) {
				if !anime.GetEpisodeDate().After(anime.GetDelayedUntil()) {
					delayedText = "Delayed"
				}
			} else if anime.GetEpisodeDate().After(anime.GetDelayedFrom()) && anime.GetEpisodeDate().Before(anime.GetDelayedUntil()) {
				delayedText = "Delayed"
			}
		}

		isSubbed := false
		if anime.GetAirType() == "sub" {
			isSubbed = true
		}

		isDonghua := false
		if anime.GetDonghua() {
			isDonghua = true
		}

		entities.AnimeSchedule.AnimeSchedule[int(anime.EpisodeDate.Weekday())] = append(entities.AnimeSchedule.AnimeSchedule[int(anime.EpisodeDate.Weekday())], entities.NewShowAirTime(
			anime.GetTitle(),
			anime.GetEpisodeDate().Format("03:04 PM"),
			episodeStr,
			delayedText,
			anime.GetRoute(),
			fmt.Sprintf("https://cdn.animeschedule.net/production/assets/public/img/%s", anime.GetImageVersionRoute()),
			isSubbed,
			isDonghua,
		))
	}
}

// TODO: DailyScheduleWebhook posts the schedule in a target channel if a guild has enabled it via webhook
func DailyScheduleWebhook(s *discordgo.Session, guildID string) {
	var (
		message discordgo.Message
		author  discordgo.User
	)

	dailyschedule := db.GetGuildAutopost(guildID, "dailyschedule")
	if dailyschedule == (entities.Cha{}) || dailyschedule.GetID() == "" {
		return
	}

	guildSettings := db.GetGuildSettings(guildID)

	author.ID = s.State.User.ID
	message.GuildID = guildID
	message.Author = &author
	message.Content = fmt.Sprintf("%sschedule", guildSettings.GetPrefix())
	message.ChannelID = dailyschedule.GetID()

	scheduleCommandHandler(s, &message)
}

// DailySchedule posts the schedule in a target channel if a guild has enabled it
func DailySchedule(s *discordgo.Session, guildID string) {
	var (
		message discordgo.Message
		author  discordgo.User
	)

	dailyschedule := db.GetGuildAutopost(guildID, "dailyschedule")
	if dailyschedule == (entities.Cha{}) || dailyschedule.GetID() == "" {
		return
	}

	guildSettings := db.GetGuildSettings(guildID)

	author.ID = s.State.User.ID
	message.GuildID = guildID
	message.Author = &author
	message.Content = fmt.Sprintf("%sschedule", guildSettings.GetPrefix())
	message.ChannelID = dailyschedule.GetID()

	scheduleCommandHandler(s, &message)
}

func ScheduleTimer(_ *discordgo.Session, _ *discordgo.Ready) {
	for range time.NewTicker(15 * time.Minute).C {
		UpdateAnimeSchedule()
	}
}

func init() {
	Add(&Command{
		Execute: scheduleCommandHandler,
		Name:    "schedule",
		Aliases: []string{"schedul", "schedu", "schedle", "schdule", "animeschedule", "anischedule"},
		Desc:    "Prints out all of today's anime release times (subbed where possible.)",
		Module:  "normal",
		DMAble:  true,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "day",
				Description: "The target day you want to get the schedule of.",
				Required:    false,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			day := ""
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			if i.ApplicationCommandData().Options != nil {
				for _, option := range i.ApplicationCommandData().Options {
					if option.Name == "day" {
						day = strings.ToLower(option.StringValue())
					}
				}
			}

			messages := scheduleCommand(day, i.GuildID)
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
