package embeds

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/mmcdole/gofeed"
	"github.com/r-anime/ZeroTsu/entities"
	"strings"
)

// Feed sends a reddit feed embed message
func Feed(s *discordgo.Session, feed *entities.Feed, item *gofeed.Item) (*discordgo.Message, error) {
	var (
		embedImage = &discordgo.MessageEmbedImage{}
		imageLink  = "https://"
		footerText = fmt.Sprintf("r/%s - %s", feed.GetSubreddit(), feed.GetPostType())
	)

	// Append author to footer if he exists in the feed
	if feed.GetAuthor() != "" {
		footerText += fmt.Sprintf(" - u/%s", feed.GetAuthor())
	}

	// Parse image if it exist
	imageStrings := strings.Split(item.Content, "[link]")
	if len(imageStrings) > 1 {
		imageStrings = strings.Split(imageStrings[0], "https://")
		imageLink += strings.Split(imageStrings[len(imageStrings)-1], "\"")[0]
	}
	if strings.HasSuffix(imageLink, ".jpg") ||
		strings.HasSuffix(imageLink, ".jpeg") ||
		strings.HasSuffix(imageLink, ".png") ||
		strings.HasSuffix(imageLink, ".webp") ||
		strings.HasSuffix(imageLink, ".gifv") ||
		strings.HasSuffix(imageLink, ".gif") {
		embedImage.URL = imageLink
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:          item.Link,
			Name:         item.Title,
			IconURL:      "https://images-eu.ssl-images-amazon.com/images/I/418PuxYS63L.png",
			ProxyIconURL: "",
		},
		Description: item.Description,
		Timestamp:   item.Published,
		Color:       purpleColor,
		Footer: &discordgo.MessageEmbedFooter{
			Text: footerText,
		},
		Image: embedImage,
	}

	return s.ChannelMessageSendComplex(feed.GetChannelID(), &discordgo.MessageSend{Content: item.Link, Embed: embed})
}
