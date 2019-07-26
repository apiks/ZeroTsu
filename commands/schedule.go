package commands

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
	"log"
	"net/http"
	"strings"
	"time"
)

var AnimeSchedule = make(map[int][]ShowAirTime)

type ShowAirTime struct {
	Name 	string
	AirTime string
}

// Shows todays airing anime times, fetched from AnimeSchedule.net
func scheduleCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		currentDay = int(time.Now().Weekday())
		day = -1
		printMessage string
	)

	command := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(command, " ", 2)

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
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
			return
		}

		printMessage = getDaySchedule(day)
	}

	// Print the daily schedule
	_, err := s.ChannelMessageSend(m.ChannelID, printMessage)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Gets a target weekday's anime schedule
func getDaySchedule(weekday int) string {
	var printMessage = fmt.Sprintf("**%v:**\n\n", time.Weekday(weekday).String())

	misc.MapMutex.Lock()
	for dayInt, showSlice := range AnimeSchedule {
		if dayInt == weekday {
			for _, show := range showSlice {
				printMessage += fmt.Sprintf("__%v__ - *%v UTC*\n\n", show.Name, show.AirTime)
			}
			break
		}
	}
	misc.MapMutex.Unlock()

	return printMessage
}

// Updates the anime schedule map
func processEachShow(index int, element *goquery.Selection) {
	var (
		day int
		show ShowAirTime
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
	show.AirTime = element.Find(".air-time").Text()

	misc.MapMutex.Lock()
	AnimeSchedule[day] = append(AnimeSchedule[day], show)
	misc.MapMutex.Unlock()
}

// Scrapes https://AnimeSchedule.net for air times subbed
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
	request.Header.Set("User-Agent", misc.UserAgent)

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
	misc.MapMutex.Lock()
	for dayInt := range AnimeSchedule {
		delete(AnimeSchedule, dayInt)
	}
	misc.MapMutex.Unlock()
	document.Find(".columns h3").Each(processEachShow)
}

func init() {
	add(&command{
		execute: scheduleCommand,
		trigger: "schedule",
		desc:    "Print anime air times SUBBED. Add a day to specify a day",
		category:"normal",
	})
}