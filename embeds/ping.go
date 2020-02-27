package embeds

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/entities"
	"time"
)

// Ping sends an embed ping message
func Ping(s *discordgo.Session, m *discordgo.Message, settings *entities.GuildSettings) error {
	embed := &discordgo.MessageEmbed{
		Title:       ":ping_pong:",
		Description: settings.GetPingMessage(),
		Color:       purpleColor,
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: s.State.User.AvatarURL("256")},
	}

	// Parses and edits the ping message with how long it took to send message
	now := time.Now()
	embedMsg, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		return err
	}
	delay := time.Now()
	embed.Title += fmt.Sprintf(" %s", delay.Sub(now).Truncate(time.Millisecond).String())

	_, err = s.ChannelMessageEditEmbed(embedMsg.ChannelID, embedMsg.ID, embed)
	return err
}
