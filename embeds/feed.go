package embeds

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mmcdole/gofeed"
	"github.com/r-anime/ZeroTsu/entities"
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
		author := feed.GetAuthor()
		if len(author) > 2035 {
			author = author[:2035]
		}
		footerText += fmt.Sprintf(" - u/%s", author)
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
		embedImage.ProxyURL = imageLink
	}

	title := item.Title
	if len(title) > 245 {
		title = title[:245]
	}
	description := item.Description
	if len(description) > 4085 {
		description = description[:4085]
	}
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     item.Link,
			Name:    title,
			IconURL: "https://images-eu.ssl-images-amazon.com/images/I/418PuxYS63L.png",
		},
		Description: description,
		Timestamp:   item.Published,
		Color:       lightPinkColor,
		Footer: &discordgo.MessageEmbedFooter{
			Text: footerText,
		},
		Image: embedImage,
	}

	return s.ChannelMessageSendComplex(feed.GetChannelID(), &discordgo.MessageSend{Content: item.Link, Embed: embed})
}

func FeedEmbed(feed *entities.Feed, item *gofeed.Item) *discordgo.MessageEmbed {
	var (
		embedImage = &discordgo.MessageEmbedImage{}
		imageLink  = "https://"
		footerText = fmt.Sprintf("r/%s - %s", feed.GetSubreddit(), feed.GetPostType())
	)

	// Append author to footer if he exists in the feed
	if feed.GetAuthor() != "" {
		author := feed.GetAuthor()
		if len(author) > 2035 {
			author = author[:2035]
		}
		footerText += fmt.Sprintf(" - u/%s", author)
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
		embedImage.ProxyURL = imageLink
	}

	title := item.Title
	if len(title) > 245 {
		title = title[:245]
	}
	description := item.Description
	if len(description) > 4085 {
		description = description[:4085]
	}
	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     item.Link,
			Name:    title,
			IconURL: "https://images-eu.ssl-images-amazon.com/images/I/418PuxYS63L.png",
		},
		Description: description,
		Timestamp:   item.Published,
		Color:       lightPinkColor,
		Footer: &discordgo.MessageEmbedFooter{
			Text: footerText,
		},
		Image: embedImage,
	}
}
