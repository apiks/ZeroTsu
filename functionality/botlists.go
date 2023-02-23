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

// SendServers sends number of servers via post request
func SendServers(guildCountStr string, s *discordgo.Session) {
	if s.State.User.ID != "614495694769618944" {
		return
	}

	// Discord Bots
	discordBotsGuildCount(&http.Client{Timeout: 120 * time.Second}, guildCountStr, strconv.Itoa(s.ShardCount))
}

// Sends guild count to top.gg
func discordBotsGuildCount(client *http.Client, guildCount, shardCount string) {
	data := url.Values{
		"server_count": {guildCount},
		"shard_count":  {shardCount},
	}
	req, err := http.NewRequest("POST", "https://top.gg/api/bots/614495694769618944/stats", bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Println("top.gg Err:", err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	req.Header.Add("Authorization", config.DiscordBotsSecret)
	response, err := client.Do(req)
	if err != nil {
		log.Println("top.gg Err:", err)
		return
	}
	response.Body.Close()
}
