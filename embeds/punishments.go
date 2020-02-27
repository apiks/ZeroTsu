package embeds

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/entities"
	"time"
)

// PunishmentAddition sends an embed message about punishment addition to the target channel
func PunishmentAddition(s *discordgo.Session, m *discordgo.Message, mem *discordgo.Member, punishmentType, punishmentVerb, reason, channelID string, date *time.Time, flag ...bool) error {
	var (
		embed           discordgo.MessageEmbed
		usernameDiscrim = fmt.Sprintf("%s#%s", mem.User.Username, mem.User.Discriminator)
	)

	embed.Color = redColor
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: mem.User.AvatarURL("128")}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "User ID:", Value: mem.User.ID})
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Reason:", Value: reason})

	if (punishmentType == "mute" || punishmentType == "ban") && date != nil {
		embed.Footer = &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Un%s Date", punishmentType)}
		embed.Timestamp = date.Format(time.RFC3339)
	} else {
		embed.Timestamp = time.Now().Format(time.RFC3339)
	}

	if len(flag) != 0 && flag[0] {
		if punishmentType == "warning" {
			embed.Title = fmt.Sprintf("Added %s to %s", punishmentType, usernameDiscrim)
		} else {
			embed.Title = fmt.Sprintf("%s was perma%s by %s#%s", usernameDiscrim, punishmentVerb, m.Author.Username, m.Author.Discriminator)
		}
	} else {
		embed.Title = fmt.Sprintf("%s was %s by %s#%s", usernameDiscrim, punishmentVerb, m.Author.Username, m.Author.Discriminator)
	}

	_, err := s.ChannelMessageSendEmbed(channelID, &embed)
	return err
}

// PunishmentRemoval sends an embed message about punishment removal to the target channel
func PunishmentRemoval(s *discordgo.Session, m *discordgo.Message, punishmentType string, punishment string) error {
	var embed discordgo.MessageEmbed

	embed.Color = redColor
	embed.Title = fmt.Sprintf("Successfuly removed %s: _%s_", punishmentType, punishment)

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embed)
	return err
}

// AutoPunishmentRemoval sends an embed notification message about auto punishment removal to the guild log
func AutoPunishmentRemoval(s *discordgo.Session, user entities.UserInfo, logId, punishmentVerb string, punisher ...*discordgo.User) error {
	var embed discordgo.MessageEmbed

	embed.Timestamp = time.Now().Format(time.RFC3339)
	embed.Color = redColor

	if len(punisher) != 0 {
		embed.Title = fmt.Sprintf("%s#%s has been %s by %s#%s.", user.GetUsername(), user.GetDiscrim(), punishmentVerb, punisher[0].Username, punisher[0].Discriminator)
	} else {
		embed.Title = fmt.Sprintf("%s#%s has been %s.", user.GetUsername(), user.GetDiscrim(), punishmentVerb)
	}

	_, err := s.ChannelMessageSendEmbed(logId, &embed)
	return err
}
