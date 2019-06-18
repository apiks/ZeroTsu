package commands

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
	"log"
	"net/http"
	"time"
)

var printMessage string

// Shows todays airing anime times, fetched from AnimeSchedule.net
func scheduleCommand(s *discordgo.Session, m *discordgo.Message) {
	_, err := s.ChannelMessageSend(m.ChannelID, printMessage)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Append each show and its air time to the print message
func processEachShow(index int, element *goquery.Selection) {
	printMessage += fmt.Sprintf("__%v__ - *%v UTC*\n\n", element.Find(".show-name").Text(), element.Find(".air-time").Text())
}

// Updates the print message
func UpdatePrintMessageSchedule() {

	// Fetch current weekday
	currentDay := time.Now().Weekday()

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
	// Find all airing shows today and process them
	// Also reset message to daily default
	printMessage = fmt.Sprintf("**%v:**\n\n", currentDay)
	document.Find(fmt.Sprintf("#%v h3", currentDay)).Each(processEachShow)
}

func init() {
	add(&command{
		execute: scheduleCommand,
		trigger: "schedule",
		desc:    "Print today's airing show times",
		category:"normal",
	})
}