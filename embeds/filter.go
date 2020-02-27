package embeds

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"strings"
	"time"
)

// Filter sends an embed filter message
func Filter(s *discordgo.Session, m *discordgo.Message, removals []string, channelID string) error {
	var (
		embed   discordgo.MessageEmbed
		content string
	)

	guildSettings := db.GetGuildSettings(m.GuildID)

	// Checks if the message contains a mention and replaces it with the actual username instead of ID
	content = common.MentionParser(s, m.Content, m.GuildID)

	// Trims the message if it is too big. Removal location sensitive
	if len(content) > 951 {
		var flag bool

		for _, removal := range removals {
			if strings.Contains(content[:900], removal) {
				content = fmt.Sprintf("%s...", content[:900])
				flag = true
				break
			}
		}

		if !flag {
			for _, removal := range removals {
				if strings.Contains(content[1153:], removal) {
					content = fmt.Sprintf("...%s", content[:1150])
					break
				}
			}
		}

		content = fmt.Sprintf("%s", content)
	}

	embed.Timestamp = time.Now().Format(time.RFC3339)
	embed.Color = redColor
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: m.Author.AvatarURL("128")}
	embed.Title = "User:"
	embed.Description = m.Author.Mention()

	// Fields
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Filtered:", Value: fmt.Sprintf("**%s**", strings.Join(removals, ", ")), Inline: true})
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Channel:", Value: common.ChMentionID(channelID), Inline: true})
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Message:", Value: fmt.Sprintf("%s", content)})

	// Sends embed in log channel channel
	if guildSettings.BotLog != (entities.Cha{}) && guildSettings.BotLog.GetID() != "" {
		_, err := s.ChannelMessageSendEmbed(guildSettings.BotLog.GetID(), &embed)
		if err != nil {
			return err
		}
	}

	return nil
}
