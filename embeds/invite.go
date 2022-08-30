package embeds

import (
	"github.com/bwmarrin/discordgo"
)

const inviteLink = "https://discord.com/oauth2/authorize?client_id=614495694769618944&scope=bot%20applications.commands&permissions=335883328"

func CreateInviteEmbed(botUser *discordgo.User) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		URL:         inviteLink,
		Title:       "Invite Link",
		Description: "Be sure to assign command roles using `addcommandrole` after inviting me if you want my admin features to work with non-administrator moderators in servers!",
		Color:       lightPinkColor,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: botUser.AvatarURL("256")},
	}
}

// Invite sends an embed invite message
func Invite(s *discordgo.Session, m *discordgo.Message) error {
	embed := &discordgo.MessageEmbed{
		URL:         inviteLink,
		Title:       "Invite Link",
		Description: "Be sure to assign command roles using `.addcommandrole` after inviting me if you want my moderation features to work with non-administrator moderators in servers!",
		Color:       lightPinkColor,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: s.State.User.AvatarURL("256")},
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}
