package embeds

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/entities"
)

// Subscription sends an embed anime subscription message
func Subscription(s *discordgo.Session, show *entities.ShowAirTime, channelID, pingableRoleId string) error {
	var pingableRoleStr string

	description := fmt.Sprintf("__**%s**__ raw is out!", show.GetEpisode())
	if show.GetSubbed() {
		description = fmt.Sprintf("__**%s**__ subbed is out!", show.GetEpisode())
	}
	if pingableRoleId != "" {
		pingableRoleStr = fmt.Sprintf("<@&%s>", pingableRoleId)
	}

	_, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: pingableRoleStr,
		Embed: &discordgo.MessageEmbed{
			URL:         fmt.Sprintf("https://animeschedule.net/anime/%s", show.GetKey()),
			Title:       show.GetName(),
			Description: description,
			Timestamp:   time.Now().Format(time.RFC3339),
			Color:       purpleColor,
			Image: &discordgo.MessageEmbedImage{
				Width:  30,
				Height: 60,
				URL:    show.GetImageUrl(),
			},
			Author: &discordgo.MessageEmbedAuthor{
				URL:          "https://AnimeSchedule.net",
				Name:         "AnimeSchedule.net",
				IconURL:      "https://cdn.animeschedule.net/production/assets/public/img/logos/as-logo-855bacd96c.png",
				ProxyIconURL: "",
			},
		},
	})
	return err
}

func SubscriptionEmbed(show *entities.ShowAirTime) *discordgo.MessageEmbed {
	description := fmt.Sprintf("__**%s**__ raw is out!", show.GetEpisode())
	if show.GetSubbed() {
		description = fmt.Sprintf("__**%s**__ subbed is out!", show.GetEpisode())
	}

	return &discordgo.MessageEmbed{
		URL:         fmt.Sprintf("https://animeschedule.net/anime/%s", show.GetKey()),
		Title:       show.GetName(),
		Description: description,
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       purpleColor,
		Image: &discordgo.MessageEmbedImage{
			Width:  30,
			Height: 60,
			URL:    show.GetImageUrl(),
		},
		Author: &discordgo.MessageEmbedAuthor{
			URL:          "https://AnimeSchedule.net",
			Name:         "AnimeSchedule.net",
			IconURL:      "https://cdn.animeschedule.net/production/assets/public/img/logos/as-logo-855bacd96c.png",
			ProxyIconURL: "",
		},
	}
}
