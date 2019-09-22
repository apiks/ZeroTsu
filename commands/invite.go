package commands

import (
	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

const inviteLink = "https://discordapp.com/api/oauth2/authorize?client_id=614495694769618944&permissions=401960278&scope=bot"

// Prints Public ZeroTsu's invite link
func inviteCommand(s *discordgo.Session, m *discordgo.Message) {
	err := inviteEmbed(s, m)
	if err != nil {
		if m.GuildID != "" {
			misc.MapMutex.Lock()
			guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
			misc.MapMutex.Unlock()
			misc.CommandErrorHandler(s, m, err, guildBotLog)
		}
	}
}

func inviteEmbed(s *discordgo.Session, m *discordgo.Message) error {
	embed := &discordgo.MessageEmbed{
		URL:         inviteLink,
		Title:       "Invite Link",
		Description: "Be sure to assign command roles after inviting it if you want it to work with non-administrator permission moderators!",
		Color:       16758465,
		Thumbnail: &discordgo.MessageEmbedThumbnail {
			URL:s.State.User.AvatarURL("256"),
		},
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	add(&command{
		execute:  inviteCommand,
		trigger:  "invite",
		aliases:  []string{"inv", "invit"},
		desc:     "Display my server invite link",
		category: "normal",
		DMAble: true,
	})
}
