package functionality

import (
	"bytes"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
)

// Send number of servers via post request
func SendServers(guildCountStr string, s *discordgo.Session) {
	if s.State.User.ID != "614495694769618944" {
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Discord Bots
	discordBotsGuildCount(client, guildCountStr)

	// Discord Boats
	discordBoatsGuildCount(client, guildCountStr)
}

// Sends guild count to discordbots.org
func discordBotsGuildCount(client *http.Client, guildCount string) {

	data := url.Values{
		"server_count": {guildCount},
	}
	req, err := http.NewRequest("POST", "https://discordbots.org/api/bots/614495694769618944/stats", bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Println("discordBots Err")
		log.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	req.Header.Add("Authorization", config.DiscordBotsSecret)
	response, err := client.Do(req)
	if err != nil {
		log.Println("discordBots Err")
		log.Println(err)
		return
	}
	response.Body.Close()
}

// Sends guild count to discord.boats
func discordBoatsGuildCount(client *http.Client, guildCount string) {
	data := url.Values{
		"server_count": {guildCount},
	}
	req, err := http.NewRequest("POST", "https://discord.boats/api/bot/614495694769618944", bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Println("discordBoats Err")
		log.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	req.Header.Add("Authorization", config.DiscordBoatsSecret)
	response, err := client.Do(req)
	if err != nil {
		log.Println("discordBoats Err")
		log.Println(err)
		return
	}
	response.Body.Close()
}
