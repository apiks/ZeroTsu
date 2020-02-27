package embeds

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"time"
)

// Verification sends an embed message about the un/verify command to the target channel
func Verification(s *discordgo.Session, m *discordgo.Message, mem *discordgo.Member, username ...string) error {
	var embed discordgo.MessageEmbed

	embed.Timestamp = time.Now().Format(time.RFC3339)
	embed.Color = greenColor

	if username == nil {
		embed.Title = fmt.Sprintf("Successfully unverified %s#%s", mem.User.Username, mem.User.Discriminator)
	} else {
		embed.Title = fmt.Sprintf("Successfully verified %s#%s with /u/%s", mem.User.Username, mem.User.Discriminator, username[0])
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embed)
	return err
}
