package embeds

import "github.com/bwmarrin/discordgo"

const inviteLink = "https://discord.com/oauth2/authorize?client_id=614495694769618944&scope=bot&permissions=335899736"

// Invite sends an embed invite message
func Invite(s *discordgo.Session, m *discordgo.Message) error {
	embed := &discordgo.MessageEmbed{
		URL:         inviteLink,
		Title:       "Invite Link",
		Description: "Be sure to assign command roles using `.addcommandrole` after inviting me if you want my moderation features to work with non-administrator moderators!",
		Color:       purpleColor,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: s.State.User.AvatarURL("256")},
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}
