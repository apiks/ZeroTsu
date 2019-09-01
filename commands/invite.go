package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

// Prints Public ZeroTsu's invite link
func inviteCommand(s *discordgo.Session, m *discordgo.Message) {

	inviteLink := "https://discordapp.com/api/oauth2/authorize?client_id=614495694769618944&permissions=401960278&scope=bot"

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Invite me to your server by using this link!\n\n<%v>", inviteLink))
	if err != nil {
		var guildBotLog string
		if m.GuildID != "" {
			misc.MapMutex.Lock()
			guildBotLog = misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
			misc.MapMutex.Unlock()
		}
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

}

func init() {
	add(&command{
		execute:  inviteCommand,
		trigger:  "invite",
		aliases:  []string{"inv", "invit"},
		desc:     "Print the BOT's invite link",
		category: "normal",
		DMAble: true,
	})
}
