package commands

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

const jokeURL = "https://official-joke-api.herokuapp.com/random_joke"
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
	err := getJson(jokeURL, joke)
	if err != nil {
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: Joke website is not working properly. Please notify Apiks about it.")
		if err != nil {
			misc.MapMutex.Lock()
			guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
			misc.MapMutex.Unlock()
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, joke.Setup + "\n\n" + joke.Punchline)
	if err != nil {
		misc.MapMutex.Lock()
		guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
		misc.MapMutex.Unlock()
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
}

func init() {
	add(&command{
		execute: jokeCommand,
		trigger: "joke",
		desc:    "Print a joke",
		category: "normal",
	})
}