package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"

	"github.com/bwmarrin/discordgo"
)

const jokeURL = "https://v2.jokeapi.dev/joke/Any?blacklistFlags=nsfw,religious,political,racist,sexist,explicit"

var myClient = &http.Client{Timeout: 10 * time.Second}

type Joke struct {
	ID       int    `json:"id"`
	JokeType string `json:"type"`
	Joke     string `json:"joke"`
	Setup    string `json:"setup"`
	Delivery string `json:"delivery"`
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

// jokeCommand prints a random joke
func jokeCommand() string {
	joke := new(Joke)
	err := getJson(jokeURL, joke)
	if err != nil {
		return "Error: Joke website is not working properly. Please notify apiks about it."
	}

	jokeStr := ""
	if joke.JokeType == "single" {
		jokeStr = joke.Joke
	} else if joke.JokeType == "twopart" {
		jokeStr = fmt.Sprintf("%s\n\n%s", joke.Setup, joke.Delivery)
	}

	return jokeStr
}

// jokeCommandHandler prints a random joke
func jokeCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	joke := new(Joke)
	err := getJson(jokeURL, joke)
	if err != nil {
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: Joke website is not working properly. Please notify apiks about it.")
		if err != nil {
			if m.GuildID != "" {
				guildSettings := db.GetGuildSettings(m.GuildID)
				common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
				return
			}
			return
		}
		return
	}

	jokeStr := ""
	if joke.JokeType == "single" {
		jokeStr = joke.Joke
	} else if joke.JokeType == "twopart" {
		jokeStr = fmt.Sprintf("%s\n\n%s", joke.Setup, joke.Delivery)
	}

	_, err = s.ChannelMessageSend(m.ChannelID, jokeStr)
	if err != nil {
		if m.GuildID != "" {
			guildSettings := db.GetGuildSettings(m.GuildID)
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		}
	}
}

func init() {
	Add(&Command{
		Execute: jokeCommandHandler,
		Name:    "joke",
		Desc:    "Prints a joke",
		Module:  "normal",
		DMAble:  true,
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			jokeStr := jokeCommand()
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &jokeStr,
			})
		},
	})
}
