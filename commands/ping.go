package commands

import (
	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

// Returns a message on "ping" to see if bot is alive
func pingCommand(s *discordgo.Session, m *discordgo.Message) {
	misc.MapMutex.Lock()
	_, err := s.ChannelMessageSend(m.ChannelID, misc.GuildMap[m.GuildID].GuildConfig.PingMessage)
	if err != nil {
		_, err = s.ChannelMessageSend(misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()
}

func init() {
	add(&command{
		execute:  pingCommand,
		trigger:  "ping",
		aliases:  []string{"pingme"},
		desc:     "Am I alive?",
		elevated: true,
		category: "misc",
	})
}
