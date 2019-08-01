package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

// Returns a message on "uptime" for BOT uptime
func uptimeCommand(s *discordgo.Session, m *discordgo.Message) {
	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("I've been online for %s.", misc.Uptime()))
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

func init() {
	add(&command{
		execute:  uptimeCommand,
		trigger:  "uptime",
		desc:     "Print how long I've been on for.",
		category: "normal",
	})
}
