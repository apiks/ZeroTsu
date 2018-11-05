package commands

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
)

var myClient = &http.Client{Timeout: 10 * time.Second}

type Joke struct {
	ID int `json:"id"`
	JokeType string `json:"type"`
	Setup string `json:"setup"`
	Punchline string `json:"punchline"`
}

// Gets json from url body
func getJson(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

// Prints a random joke in chat
func jokeCommand(s *discordgo.Session, m *discordgo.Message) {
	joke := new(Joke)
	getJson("https://08ad1pao69.execute-api.us-east-1.amazonaws.com/dev/random_joke", joke)

	_, err := s.ChannelMessageSend(m.ChannelID, joke.Setup + "\n\n" + joke.Punchline)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

func init() {
	add(&command{
		execute: jokeCommand,
		trigger: "joke",
		desc:    "Print a joke.",
		category: "normal",
	})
}