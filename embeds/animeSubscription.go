package embeds

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/entities"
	"time"
)

// Subscription sends an embed anime subscription message
func Subscription(s *discordgo.Session, show *entities.ShowAirTime, channelID string) error {
	var embed = &discordgo.MessageEmbed{
		URL:         fmt.Sprintf("https://animeschedule.net/shows/%s", show.GetKey()),
		Title:       show.GetName(),
		Description: fmt.Sprintf("__**%s**__ is out!", show.GetEpisode()),
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       purpleColor,
		Image: &discordgo.MessageEmbedImage{
			Width:  30,
			Height: 60,
			URL:    show.GetImageUrl(),
		},
	}

	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	return err
}
