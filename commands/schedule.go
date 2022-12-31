package commands

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"

	"github.com/PuerkitoBio/goquery"
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

	message += "\n**Full Week:** <https://AnimeSchedule.net>"

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

	printMessage += "\n**Full Week:** <https://AnimeSchedule.net>"

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
			londonTZ, err := time.LoadLocation("Europe/London")
			if err != nil {
				realTime = time.Date(targetDay.Year(), targetDay.Month(), targetDay.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)
			} else {
				realTime = time.Date(targetDay.Year(), targetDay.Month(), targetDay.Day(), t.Hour(), t.Minute(), 0, 0, londonTZ)
			}

			if show.GetDelayed() != "" {
				printMessage += fmt.Sprintf("**%s** - %s %s - <t:%d:t>\n\n", show.GetName(), show.GetEpisode(), show.GetDelayed(), realTime.UTC().Unix())
			} else {
				printMessage += fmt.Sprintf("**%s** - %s - <t:%d:t>\n\n", show.GetName(), show.GetEpisode(), realTime.UTC().Unix())
			}

		}
		break
	}

	return printMessage
}

// Updates the anime schedule map
func processEachShow(_ int, element *goquery.Selection) {
	var (
		day  int
		show entities.ShowAirTime
	)

	date := strings.ToLower(element.SiblingsFiltered(".timetable-column-date").Find(".timetable-column-day").Text())
	if strings.Contains(date, "sunday") {
		day = 0
	} else if strings.Contains(date, "monday") {
		day = 1
	} else if strings.Contains(date, "tuesday") {
		day = 2
	} else if strings.Contains(date, "wednesday") {
		day = 3
	} else if strings.Contains(date, "thursday") {
		day = 4
	} else if strings.Contains(date, "friday") {
		day = 5
	} else if strings.Contains(date, "saturday") {
		day = 6
	}

	show.SetName(element.Find(".show-title-bar").Text())
	show.SetEpisode(element.Find(".show-episode").Text())
	show.SetEpisode(strings.Replace(show.GetEpisode(), "\n", "", -1))
	show.SetAirTime(element.Find(".show-air-time").Text())
	show.SetAirTime(strings.TrimSuffix(strings.TrimPrefix(strings.Replace(show.GetAirTime(), "\n", "", -1), " "), " "))
	show.SetDelayed(strings.TrimPrefix(element.Find(".show-delay-bar").Text(), " "))
	show.SetDelayed(strings.Trim(show.GetDelayed(), "\n"))
	key, exists := element.Find(".show-link").Attr("href")
	if exists {
		show.SetKey(key)
		show.SetKey(strings.ToLower(strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(show.GetKey(), "/anime/"), "anime/"), "/anime")))
	}
	imageUrl, exists := element.Find(".show-poster").Attr("data-src")
	if exists {
		show.SetImageUrl(imageUrl)
	} else {
		imageUrl, exists = element.Find(".show-poster").Attr("src")
		if exists {
			show.SetImageUrl(imageUrl)
		}
	}
	if element.Find(".air-type-text").Text() == "" {
		show.SetSubbed(true)
	} else if strings.ToLower(element.Find(".air-type-text").Text()) == "dub" {
		return
	} else {
		show.SetSubbed(false)
	}

	donghua, _ := element.Attr("chinese")
	if donghua == "true" {
		show.SetDonghua(true)
	}

	entities.AnimeSchedule.AnimeSchedule[day] = append(entities.AnimeSchedule.AnimeSchedule[day], &show)
}

// UpdateAnimeSchedule Scrapes https://AnimeSchedule.net for air times subbed
func UpdateAnimeSchedule() {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create and modify HTTP request before sending
	request, err := http.NewRequest("GET", "https://animeschedule.net", nil)
	if err != nil {
		log.Println(err)
		return
	}
	request.Header.Set("User-Agent", common.UserAgent)

	// Make request
	response, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return
	}
	defer response.Body.Close()

	// Create a goquery document from the HTTP response
	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Println("Error loading HTTP response body. ", err)
		return
	}

	// Find all airing shows and process them after resetting map
	entities.AnimeSchedule.Lock()
	defer entities.AnimeSchedule.Unlock()
	for dayInt := range entities.AnimeSchedule.AnimeSchedule {
		for len(entities.AnimeSchedule.AnimeSchedule[dayInt]) > 0 {
			for i := range entities.AnimeSchedule.AnimeSchedule[dayInt] {
				copy(entities.AnimeSchedule.AnimeSchedule[dayInt][i:], entities.AnimeSchedule.AnimeSchedule[dayInt][i+1:])
				entities.AnimeSchedule.AnimeSchedule[dayInt][len(entities.AnimeSchedule.AnimeSchedule[dayInt])-1] = nil
				entities.AnimeSchedule.AnimeSchedule[dayInt] = entities.AnimeSchedule.AnimeSchedule[dayInt][:len(entities.AnimeSchedule.AnimeSchedule[dayInt])-1]
				break
			}
		}
		delete(entities.AnimeSchedule.AnimeSchedule, dayInt)
	}
	document.Find(".timetable-column .timetable-column-show").Each(processEachShow)
}

// isTimeDST returns true if time t occurs within daylight saving time
// for its time zone.
func isTimeDST(t time.Time) bool {
	// If the most recent (within the last year) clock change
	// was forward then assume the change was from std to dst.
	hh, mm, _ := t.UTC().Clock()
	tClock := hh*60 + mm
	for m := -1; m > -12; m-- {
		// assume dst lasts for least one month
		hh, mm, _ := t.AddDate(0, m, 0).UTC().Clock()
		clock := hh*60 + mm
		if clock != tClock {
			if clock > tClock {
				// std to dst
				return true
			}
			// dst to std
			return false
		}
	}
	// assume no dst
	return false
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
	for range time.NewTicker(30 * time.Minute).C {
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
						day = option.StringValue()
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
