package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
)

// Shows todays airing anime times, fetched from AnimeSchedule.net
func scheduleCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		currentDay   = int(time.Now().Weekday())
		day          = -1
		printMessage string
	)

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		// Get the current day's schedule in print format
		printMessage = getDaySchedule(currentDay)
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

		printMessage = getDaySchedule(day)
	}

	printMessage += "\n\n**Full Week:** <https://AnimeSchedule.net>"

	// Print the daily schedule
	_, _ = s.ChannelMessageSend(m.ChannelID, printMessage)
}

// Gets a target weekday's anime schedule
func getDaySchedule(weekday int) string {
	var (
		printMessage = fmt.Sprintf("**__%s:__**\n\n", time.Weekday(weekday).String())
		DST          = isTimeDST(time.Now())
		JST          *time.Location
		BST          *time.Location
		PDT          *time.Location
		PST          *time.Location
	)

	// Set timezones based on DST
	JST = time.FixedZone("JST", +9*3600)
	if DST {
		BST = time.FixedZone("BST", +1*3600)
		PDT = time.FixedZone("PDT", -7*3600)
	} else {
		PST = time.FixedZone("PST", -8*3600)
	}

	entities.Mutex.Lock()
	for dayInt, showSlice := range entities.AnimeSchedule {
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

			// Parses the time in a proper time object
			t, err := time.Parse("3:04 PM", show.GetAirTime())
			if err != nil {
				log.Println(err)
				continue
			}

			// Format print message for show's air times and timezones
			jstTimezoneString, _ := t.In(JST).Zone()
			if DST {
				ukTimezoneString, _ := t.In(BST).Zone()
				westAmericanTimezoneString, _ := t.In(PDT).Zone()

				if show.GetDelayed() == "" {
					printMessage += fmt.Sprintf("**%s %s** - %s %s **|** %s %s **|** %s %s\n\n", show.GetName(), show.GetEpisode(), t.UTC().In(BST).Format("15:04"), ukTimezoneString,
						t.UTC().In(PDT).Format("15:04"), westAmericanTimezoneString,
						t.UTC().In(JST).Format("15:04"), jstTimezoneString)
				} else {
					printMessage += fmt.Sprintf("**%s %s** __%s__ - %s %s **|** %s %s **|** %s %s\n\n", show.GetName(), show.GetEpisode(), show.GetDelayed(), t.UTC().In(BST).Format("15:04"), ukTimezoneString,
						t.UTC().In(PDT).Format("15:04"), westAmericanTimezoneString,
						t.UTC().In(JST).Format("15:04"), jstTimezoneString)
				}
			} else {
				westAmericanTimezoneString, _ := t.In(PST).Zone()

				if show.GetDelayed() == "" {
					printMessage += fmt.Sprintf("**%s %s** - %s GMT **|** %s %s **|** %s %s\n\n", show.GetName(), show.GetEpisode(), t.UTC().Format("15:04"),
						t.UTC().In(PST).Format("15:04"), westAmericanTimezoneString,
						t.UTC().In(JST).Format("15:04"), jstTimezoneString)
				} else {
					printMessage += fmt.Sprintf("**%s %s** __%s__ - %s GMT **|** %s %s **|** %s %s\n\n", show.GetName(), show.GetEpisode(), show.GetDelayed(), t.UTC().Format("15:04"),
						t.UTC().In(PST).Format("15:04"), westAmericanTimezoneString,
						t.UTC().In(JST).Format("15:04"), jstTimezoneString)
				}
			}
		}
		break
	}
	entities.Mutex.Unlock()

	return printMessage
}

// Updates the anime schedule map
func processEachShow(_ int, element *goquery.Selection) {
	var (
		day  int
		show entities.ShowAirTime
	)

	switch strings.ToLower(element.SiblingsFiltered(".timetable-column-day").Text()) {
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

	show.SetName(element.Find(".show-title-bar").Text())
	show.SetEpisode(element.Find(".show-episode").Text())
	show.SetEpisode(strings.Replace(show.GetEpisode(), "\n", "", -1))
	show.SetAirTime(element.Find(".show-air-time").Text())
	show.SetAirTime(strings.Replace(show.GetAirTime(), "\n", "", -1))
	show.SetDelayed(strings.TrimPrefix(element.Find(".show-delay-bar").Text(), " "))
	show.SetDelayed(strings.Trim(show.GetDelayed(), "\n"))
	key, exists := element.Find(".show-link").Attr("href")
	if exists == true {
		show.SetKey(key)
		show.SetKey(strings.ToLower(strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(show.GetKey(), "/shows/"), "shows/"), "/shows")))
	}
	imageUrl, exists := element.Find(".show-poster").Attr("data-src")
	if exists == true {
		show.SetImageUrl(imageUrl)
	} else {
		imageUrl, exists = element.Find(".show-poster").Attr("src")
		if exists {
			show.SetImageUrl(imageUrl)
		}
	}

	entities.AnimeSchedule[day] = append(entities.AnimeSchedule[day], &show)
}

// Scrapes https://AnimeSchedule.net for air times subbed
func UpdateAnimeSchedule() {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
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
	entities.Mutex.Lock()
	defer entities.Mutex.Unlock()
	for dayInt := range entities.AnimeSchedule {
		for len(entities.AnimeSchedule[dayInt]) > 0 {
			for i := range entities.AnimeSchedule[dayInt] {
				copy(entities.AnimeSchedule[dayInt][i:], entities.AnimeSchedule[dayInt][i+1:])
				entities.AnimeSchedule[dayInt][len(entities.AnimeSchedule[dayInt])-1] = nil
				entities.AnimeSchedule[dayInt] = entities.AnimeSchedule[dayInt][:len(entities.AnimeSchedule[dayInt])-1]
				break
			}
		}
		delete(entities.AnimeSchedule, dayInt)
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

// Posts the schedule in a target channel if a guild has enabled it
func DailySchedule(s *discordgo.Session, guildID string) {
	dailyschedule := db.GetGuildAutopost(guildID, "dailyschedule")
	if dailyschedule == (entities.Cha{}) || dailyschedule.GetID() == "" {
		return
	}

	var (
		message discordgo.Message
		author  discordgo.User
	)

	guildSettings := db.GetGuildSettings(guildID)

	author.ID = s.State.User.ID
	message.GuildID = guildID
	message.Author = &author
	message.Content = fmt.Sprintf("%sschedule", guildSettings.GetPrefix())
	message.ChannelID = dailyschedule.GetID()

	scheduleCommand(s, &message)
}

func ScheduleTimer(_ *discordgo.Session, _ *discordgo.Ready) {
	for range time.NewTicker(30 * time.Minute).C {
		// Update anime schedule
		UpdateAnimeSchedule()
	}
}

func init() {
	Add(&Command{
		Execute: scheduleCommand,
		Trigger: "schedule",
		Aliases: []string{"schedul", "schedu", "schedle", "schdule", "animeschedule", "anischedule"},
		Desc:    "Print Anime Air Times (subbed where possible.) Add a day to specify a day",
		Module:  "normal",
		DMAble:  true,
	})
}
