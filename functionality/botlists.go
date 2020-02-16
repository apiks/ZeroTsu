package functionality

import (
	"bytes"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
)

// Send number of servers via post request
func sendServers(s *discordgo.Session) {

	if s.State.User.ID != "614495694769618944" {
		return
	}

	guildCountStr := strconv.Itoa(len(s.State.Guilds))
	client := &http.Client{Timeout: 10 * time.Second}

	// Discord Bots
	discordBotsGuildCount(client, guildCountStr)

	// Discord Boats
	discordBoatsGuildCount(client, guildCountStr)

	// Bots on Discord
	discordBotsOnDiscordGuildCount(client, guildCountStr)
}

// Sends guild count to discordbots.org
func discordBotsGuildCount(client *http.Client, guildCount string) {

	data := url.Values{
		"server_count": {guildCount},
	}
	req, err := http.NewRequest("POST", "https://discordbots.org/api/bots/614495694769618944/stats", bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	req.Header.Add("Authorization", config.DiscordBotsSecret)
	_, err = client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
}

// Sends guild count to discord.boats
func discordBoatsGuildCount(client *http.Client, guildCount string) {
	data := url.Values{
		"server_count": {guildCount},
	}
	req, err := http.NewRequest("POST", "https://discord.boats/api/bot/614495694769618944", bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	req.Header.Add("Authorization", config.DiscordBoatsSecret)
	_, err = client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
}

// Sends guild count to bots.ondiscord.xyz
func discordBotsOnDiscordGuildCount(client *http.Client, guildCount string) {
	data := url.Values{
		"guildCount": {guildCount},
	}
	req, err := http.NewRequest("POST", "https://bots.ondiscord.xyz/bot-api/bots/614495694769618944/guilds", bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	req.Header.Add("Authorization", config.BotsOnDiscordSecret)
	_, err = client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
}
