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

const jokeURL = "https://official-joke-api.herokuapp.com/random_joke"

var myClient = &http.Client{Timeout: 10 * time.Second}

type Joke struct {
	ID        int    `json:"id"`
	JokeType  string `json:"type"`
	Setup     string `json:"setup"`
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

// jokeCommand prints a random joke
func jokeCommand() string {
	joke := new(Joke)
	err := getJson(jokeURL, joke)
	if err != nil {
		return "Error: Joke website is not working properly. Please notify Apiks#8969 about it."
	}

	return fmt.Sprintf("%s\n\n%s", joke.Setup, joke.Punchline)
}

// jokeCommandHandler prints a random joke
func jokeCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	joke := new(Joke)
	err := getJson(jokeURL, joke)
	if err != nil {
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: Joke website is not working properly. Please notify Apiks#8969 about it.")
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

	_, err = s.ChannelMessageSend(m.ChannelID, joke.Setup+"\n\n"+joke.Punchline)
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
		Desc:    "Prints a (bad) joke",
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
