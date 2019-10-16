package commands

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/functionality"
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
		if m.Author.ID == s.State.User.ID {
			functionality.MapMutex.Unlock()
		}
		printMessage = getDaySchedule(currentDay)
		if m.Author.ID == s.State.User.ID {
			functionality.MapMutex.Lock()
		}
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
					if m.Author.ID != s.State.User.ID {
						functionality.MapMutex.Lock()
					}
					guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
					if m.Author.ID != s.State.User.ID {
						functionality.MapMutex.Unlock()
					}
					functionality.LogError(s, guildSettings.BotLog, err)
				}
				return
			}
			return
		}

		printMessage = getDaySchedule(day)
	}

	// Add AnimeSchedule.net if public ZeroTsu
	if config.ServerID != "267799767843602452" {
		printMessage += "\n\nFull Week: <https://AnimeSchedule.net>"
	}

	// Print the daily schedule
	_, _ = s.ChannelMessageSend(m.ChannelID, printMessage)
}

// Gets a target weekday's anime schedule
func getDaySchedule(weekday int) string {
	var (
		printMessage = fmt.Sprintf("**__%v:__**\n\n", time.Weekday(weekday).String())
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

	functionality.MapMutex.Lock()
	for dayInt, showSlice := range functionality.AnimeSchedule {
		if dayInt == weekday {
			for _, show := range showSlice {

				// Parses the time in a proper time object
				t, err := time.Parse("15:04", show.AirTime)
				if err != nil {
					log.Println(err)
					continue
				}

				// Format print message for show's air times and timezones
				jstTimezoneString, _ := t.In(JST).Zone()
				if DST {
					ukTimezoneString, _ := t.In(BST).Zone()
					westAmericanTimezoneString, _ := t.In(PDT).Zone()

					if show.Delayed == "" {
						printMessage += fmt.Sprintf("**%v %v** - %v %v **|** %v %v **|** %v %v\n\n", show.Name, show.Episode, t.UTC().In(BST).Format("15:04"), ukTimezoneString,
							t.UTC().In(PDT).Format("15:04"), westAmericanTimezoneString,
							t.UTC().In(JST).Format("15:04"), jstTimezoneString)
					} else {
						printMessage += fmt.Sprintf("**%v %v** __%v__ - %v %v **|** %v %v **|** %v %v\n\n", show.Name, show.Episode, show.Delayed, t.UTC().In(BST).Format("15:04"), ukTimezoneString,
							t.UTC().In(PDT).Format("15:04"), westAmericanTimezoneString,
							t.UTC().In(JST).Format("15:04"), jstTimezoneString)
					}
				} else {
					westAmericanTimezoneString, _ := t.In(PST).Zone()

					if show.Delayed == "" {
						printMessage += fmt.Sprintf("**%v %v** - %v GMT **|** %v %v **|** %v %v\n\n", show.Name, show.Episode, t.UTC().Format("15:04"),
							t.UTC().In(PST).Format("15:04"), westAmericanTimezoneString,
							t.UTC().In(JST).Format("15:04"), jstTimezoneString)
					} else {
						printMessage += fmt.Sprintf("**%v %v** __%v__ - %v GMT **|** %v %v **|** %v %v\n\n", show.Name, show.Episode, show.Delayed, t.UTC().Format("15:04"),
							t.UTC().In(PST).Format("15:04"), westAmericanTimezoneString,
							t.UTC().In(JST).Format("15:04"), jstTimezoneString)
					}
				}
			}
			break
		}
	}
	functionality.MapMutex.Unlock()

	return printMessage
}

// Updates the anime schedule map
func processEachShow(index int, element *goquery.Selection) {
	var (
		day  int
		show functionality.ShowAirTime
	)

	switch strings.ToLower(element.Parent().Parent().Parent().SiblingsFiltered(".column-title").Text()) {
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

	show.Name = element.Find(".show-name").Text()
	show.Episode = "Ep " + element.Parent().Parent().Parent().Find(".episode-number").Text()
	show.Episode = strings.Replace(show.Episode, "\n", "", -1)
	show.AirTime = element.Find(".air-time").Text()
	show.Delayed = strings.TrimPrefix(element.Parent().Parent().Parent().Find(".delay").Text(), " ")
	show.Delayed = strings.Trim(show.Delayed, "\n")
	show.Key, _ = element.Parent().Parent().Parent().Find(".show-link").Attr("href")
	show.Key = strings.ToLower(strings.TrimPrefix(show.Key, "/shows/"))

	functionality.AnimeSchedule[day] = append(functionality.AnimeSchedule[day], &show)
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
	request.Header.Set("User-Agent", functionality.UserAgent)

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
	functionality.MapMutex.Lock()
	for dayInt := range functionality.AnimeSchedule {
		delete(functionality.AnimeSchedule, dayInt)
	}
	document.Find(".columns h3").Each(processEachShow)
	functionality.MapMutex.Unlock()
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
	if dailyschedule, ok := functionality.GuildMap[guildID].Autoposts["dailyschedule"]; !ok {
		return
	} else if dailyschedule == nil {
		return
	}
	if functionality.GuildMap[guildID].Autoposts["dailyschedule"].ID == "" {
		return
	}

	var (
		message discordgo.Message
		author  discordgo.User
	)

	guildSettings := functionality.GuildMap[guildID].GetGuildSettings()

	author.ID = s.State.User.ID
	message.GuildID = guildID
	message.Author = &author
	message.Content = fmt.Sprintf("%sschedule", guildSettings.Prefix)
	message.ChannelID = functionality.GuildMap[guildID].Autoposts["dailyschedule"].ID

	scheduleCommand(s, &message)
}

func ScheduleTimer(s *discordgo.Session, e *discordgo.Ready) {
	for range time.NewTicker(20 * time.Minute).C {
		// Update anime schedule
		UpdateAnimeSchedule()
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute: scheduleCommand,
		Trigger: "schedule",
		Aliases: []string{"schedul", "schedu", "schedle", "schdule", "animeschedule", "anischedule"},
		Desc:    "Print Anime Air Times (subbed where possible.) Add a day to specify a day",
		Module:  "normal",
		DMAble:  true,
	})
}
