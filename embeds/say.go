package embeds

import "github.com/bwmarrin/discordgo"

// Say sends an embed message to the target channel
func Say(s *discordgo.Session, message, channelID string) error {
	embed := &discordgo.MessageEmbed{Description: message, Color: lightPinkColor}
	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	return err
}

// Edit edits an embed message
func Edit(s *discordgo.Session, channelID, messageID, message string) error {
	embed := &discordgo.MessageEmbed{Description: message, Color: lightPinkColor}
	_, err := s.ChannelMessageEditEmbed(channelID, messageID, embed)
	return err
}
