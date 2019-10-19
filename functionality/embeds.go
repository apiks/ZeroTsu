package functionality

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mmcdole/gofeed"
)

const inviteLink = "https://discordapp.com/api/oauth2/authorize?client_id=614495694769618944&permissions=401960278&scope=bot"

func UnbanEmbed(s *discordgo.Session, user *UserInfo, mod string, botLog string) error {

	var (
		embedMess discordgo.MessageEmbed
		embed     []*discordgo.MessageEmbedField
	)

	// Sets timestamp of unban
	t := time.Now()
	now := t.Format(time.RFC3339)
	embedMess.Timestamp = now

	// Set embed color
	embedMess.Color = 16758465

	if mod == "" {
		embedMess.Title = fmt.Sprintf("%v#%v has been unbanned.", user.Username, user.Discrim)
	} else {
		embedMess.Title = fmt.Sprintf("%v#%v has been unbanned by %v.", user.Username, user.Discrim, mod)
	}

	// Adds everything together
	embedMess.Fields = embed

	// Sends embed in bot-log
	_, err := s.ChannelMessageSendEmbed(botLog, &embedMess)
	if err != nil {
		return err
	}
	return err
}

func UnmuteEmbed(s *discordgo.Session, user *UserInfo, mod string, botLog string) error {

	var (
		embedMess discordgo.MessageEmbed
		embed     []*discordgo.MessageEmbedField
	)

	// Sets timestamp of unban
	t := time.Now()
	now := t.Format(time.RFC3339)
	embedMess.Timestamp = now

	// Set embed color
	embedMess.Color = 16758465

	if mod == "" {
		embedMess.Title = fmt.Sprintf("%v#%v has been unmuted.", user.Username, user.Discrim)
	} else {
		embedMess.Title = fmt.Sprintf("%v#%v has been unmuted by %v.", user.Username, user.Discrim, mod)
	}

	// Adds everything together
	embedMess.Fields = embed

	// Sends embed in bot-log
	_, err := s.ChannelMessageSendEmbed(botLog, &embedMess)
	if err != nil {
		return err
	}
	return err
}

func feedEmbed(s *discordgo.Session, thread *RssThread, item *gofeed.Item) (*discordgo.Message, error) {
	var (
		embedImage = &discordgo.MessageEmbedImage{}
		imageLink  = "https://"
		footerText = fmt.Sprintf("r/%v - %v", thread.Subreddit, thread.PostType)
	)

	// Append custom user author to footer if he exists in thread
	if thread.Author != "" {
		footerText += fmt.Sprintf(" - u/%v", thread.Author)
	}

	// Parse image if it exists between a preset number of allowed domains
	imageStrings := strings.Split(item.Content, "[link]")
	if len(imageStrings) > 1 {
		imageStrings = strings.Split(imageStrings[0], "https://")
		imageLink += strings.Split(imageStrings[len(imageStrings)-1], "\"")[0]
	}
	if strings.HasPrefix(imageLink, "https://i.redd.it/") ||
		strings.HasPrefix(imageLink, "https://i.imgur.com/") ||
		strings.HasPrefix(imageLink, "https://i.gyazo.com/") {
		if strings.Contains(imageLink, ".jpg") ||
			strings.Contains(imageLink, ".jpeg") ||
			strings.Contains(imageLink, ".png") ||
			strings.Contains(imageLink, ".webp") ||
			strings.Contains(imageLink, ".gifv") ||
			strings.Contains(imageLink, ".gif") {
			embedImage.URL = imageLink
		}
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
		Color:       16758465,
		Footer: &discordgo.MessageEmbedFooter{
			Text: footerText,
		},
		Image: embedImage,
	}

	// Creates the complex message to send
	data := &discordgo.MessageSend{
		Content: item.Link,
		Embed:   embed,
	}

	// Sends the message
	message, err := s.ChannelMessageSendComplex(thread.ChannelID, data)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func VerifyEmbed(s *discordgo.Session, m *discordgo.Message, mem *discordgo.Member, username string) error {

	var embedMess discordgo.MessageEmbed

	// Sets punishment embed color
	embedMess.Color = 0x00ff00
	embedMess.Title = fmt.Sprintf("Successfully verified %v#%v with /u/%v", mem.User.Username, mem.User.Discriminator, username)

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	return err
}

func UnverifyEmbed(s *discordgo.Session, m *discordgo.Message, mem string) error {

	var embedMess discordgo.MessageEmbed

	// Sets punishment embed color
	embedMess.Color = 0x00ff00
	embedMess.Title = fmt.Sprintf("Successfully unverified %v", mem)

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	return err
}

func WarningEmbed(s *discordgo.Session, m *discordgo.Message, mem *discordgo.User, reason string, channelID string, discrete bool) error {

	var (
		embedMess      discordgo.MessageEmbed
		embedThumbnail discordgo.MessageEmbedThumbnail

		// Embed slice and its fields
		embedField       []*discordgo.MessageEmbedField
		embedFieldUserID discordgo.MessageEmbedField
		embedFieldReason discordgo.MessageEmbedField
	)
	t := time.Now()

	// Sets timestamp for warning
	embedMess.Timestamp = t.Format(time.RFC3339)

	// Sets warning embed color
	embedMess.Color = 0xff0000

	// Saves user avatar as thumbnail
	embedThumbnail.URL = mem.AvatarURL("128")

	// Sets field titles
	embedFieldUserID.Name = "User ID:"
	embedFieldReason.Name = "Reason:"

	// Sets field content
	embedFieldUserID.Value = mem.ID
	embedFieldReason.Value = reason

	// Adds the two fields to embedField slice (because embedMess.Fields requires slice input)
	embedField = append(embedField, &embedFieldUserID)
	embedField = append(embedField, &embedFieldReason)

	if discrete {
		embedMess.Title = fmt.Sprintf("Added warning to %v#%v", mem.Username, mem.Discriminator)
	} else {
		embedMess.Title = mem.Username + "#" + mem.Discriminator + " was warned by " + m.Author.Username
	}

	// Adds user thumbnail and the two other fields as well
	embedMess.Thumbnail = &embedThumbnail
	embedMess.Fields = embedField

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(channelID, &embedMess)
	return err
}

func RemovePunishmentEmbed(s *discordgo.Session, m *discordgo.Message, punishment string) error {

	var embedMess discordgo.MessageEmbed

	// Sets punishment embed color
	embedMess.Color = 16758465
	embedMess.Title = fmt.Sprintf("Successfuly removed punishment: _%v_", punishment)

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	return err
}

func PingEmbed(s *discordgo.Session, m *discordgo.Message, settings *GuildSettings) error {
	embed := &discordgo.MessageEmbed{
		Title:       ":ping_pong:",
		Description: fmt.Sprintf("\n%s", settings.PingMessage),
		Color:       16758465,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: s.State.User.AvatarURL("256"),
		},
	}

	// Parses and edits message with how long it took to send message
	now := time.Now()
	embedMsg, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		return err
	}
	delay := time.Now()

	embed.Title += fmt.Sprintf(" %s", delay.Sub(now).Truncate(time.Millisecond).String())
	_, err = s.ChannelMessageEditEmbed(embedMsg.ChannelID, embedMsg.ID, embed)
	if err != nil {
		return err
	}

	return nil
}

func MuteEmbed(s *discordgo.Session, m *discordgo.Message, mem *discordgo.User, reason string, length time.Time, perma bool, channelID string) error {

	var (
		embedMess      discordgo.MessageEmbed
		embedThumbnail discordgo.MessageEmbedThumbnail
		embedFooter    discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embedField       []*discordgo.MessageEmbedField
		embedFieldUserID discordgo.MessageEmbedField
		embedFieldReason discordgo.MessageEmbedField
	)

	// Sets timestamp for unmute date and footer
	unmuteDate := length.Format(time.RFC3339)
	embedMess.Timestamp = unmuteDate
	embedFooter.Text = "Unmute Date"
	embedMess.Footer = &embedFooter

	// Sets mute embed color
	embedMess.Color = 0xff0000

	// Saves user avatar as thumbnail
	embedThumbnail.URL = mem.AvatarURL("128")

	// Sets field titles
	embedFieldUserID.Name = "User ID:"
	embedFieldReason.Name = "Reason:"

	// Sets field content
	embedFieldUserID.Value = mem.ID
	embedFieldReason.Value = reason

	// Adds the two fields to embedField slice (because embedMess.Fields requires slice input)
	embedField = append(embedField, &embedFieldUserID)
	embedField = append(embedField, &embedFieldReason)

	// Sets embed title and its description (which it uses the same way as a field)
	if !perma {
		embedMess.Title = mem.Username + "#" + mem.Discriminator + " was muted by " + m.Author.Username
	} else {
		embedMess.Title = mem.Username + "#" + mem.Discriminator + " was permamutted by " + m.Author.Username
	}

	// Adds user thumbnail and the two other fields as well
	embedMess.Thumbnail = &embedThumbnail
	embedMess.Fields = embedField

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(channelID, &embedMess)
	return err
}

func KickEmbed(s *discordgo.Session, m *discordgo.Message, mem *discordgo.User, reason string, channelID string) error {

	var (
		embedMess      discordgo.MessageEmbed
		embedThumbnail discordgo.MessageEmbedThumbnail

		// Embed slice and its fields
		embedField       []*discordgo.MessageEmbedField
		embedFieldUserID discordgo.MessageEmbedField
		embedFieldReason discordgo.MessageEmbedField
	)
	t := time.Now()

	// Sets timestamp for warning
	embedMess.Timestamp = t.Format(time.RFC3339)

	// Sets warning embed color
	embedMess.Color = 0xff0000

	// Saves user avatar as thumbnail
	embedThumbnail.URL = mem.AvatarURL("128")

	// Sets field titles
	embedFieldUserID.Name = "User ID:"
	embedFieldReason.Name = "Reason:"

	// Sets field content
	embedFieldUserID.Value = mem.ID
	embedFieldReason.Value = reason

	// Sets field inline
	embedFieldUserID.Inline = true
	embedFieldReason.Inline = true

	// Adds the two fields to embedField slice (because embedMess.Fields requires slice input)
	embedField = append(embedField, &embedFieldUserID)
	embedField = append(embedField, &embedFieldReason)

	// Sets embed title and its description (which it uses the same way as a field)
	embedMess.Title = mem.Username + "#" + mem.Discriminator + " was kicked by " + m.Author.Username

	// Adds user thumbnail and the two other fields as well
	embedMess.Thumbnail = &embedThumbnail
	embedMess.Fields = embedField

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(channelID, &embedMess)
	return err
}

func InviteEmbed(s *discordgo.Session, m *discordgo.Message) error {
	embed := &discordgo.MessageEmbed{
		URL:         inviteLink,
		Title:       "Invite Link",
		Description: "Be sure to assign command roles after inviting me if you want it to work with non-administrator permission moderators!",
		Color:       16758465,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: s.State.User.AvatarURL("256"),
		},
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		return err
	}

	return nil
}

func FilterEmbed(s *discordgo.Session, m *discordgo.Message, removals, channelID string) error {

	var (
		embedMess      discordgo.MessageEmbed
		embedThumbnail discordgo.MessageEmbedThumbnail

		// Embed slice and its fields
		embedField        []*discordgo.MessageEmbedField
		embedFieldFilter  discordgo.MessageEmbedField
		embedFieldMessage discordgo.MessageEmbedField
		embedFieldChannel discordgo.MessageEmbedField

		content string
	)

	MapMutex.Lock()
	guildSettings := GuildMap[m.GuildID].GetGuildSettings()
	MapMutex.Unlock()

	// Checks if the message contains a mention and replaces it with the actual nick instead of ID
	content = m.Content
	content = MentionParser(s, content, m.GuildID)

	// Sets timestamp for removal
	t := time.Now()
	now := t.Format(time.RFC3339)
	embedMess.Timestamp = now

	// Sets ban embed color
	embedMess.Color = 0xff0000

	// Saves user avatar as thumbnail
	embedThumbnail.URL = m.Author.AvatarURL("128")

	// Sets field titles
	embedFieldFilter.Name = "Filtered:"
	embedFieldMessage.Name = "Message:"
	embedFieldChannel.Name = "Channel:"

	// Sets field content
	embedFieldFilter.Value = fmt.Sprintf("**%v**", removals)
	embedFieldMessage.Value = fmt.Sprintf("`%v`", content)
	embedFieldChannel.Value = ChMentionID(channelID)

	// Sets field inline
	embedFieldFilter.Inline = true
	embedFieldChannel.Inline = true

	// Adds the two fields to embedField slice (because embedMess.Fields requires slice input)
	embedField = append(embedField, &embedFieldFilter)
	embedField = append(embedField, &embedFieldChannel)
	embedField = append(embedField, &embedFieldMessage)

	// Sets embed title and its description (which it uses the same way as a field)
	embedMess.Title = "User:"
	embedMess.Description = m.Author.Mention()

	// Adds user thumbnail and the two other fields as well
	embedMess.Thumbnail = &embedThumbnail
	embedMess.Fields = embedField

	// Sends embed in bot-log channel
	if guildSettings.BotLog != nil {
		if guildSettings.BotLog.ID != "" {
			_, err := s.ChannelMessageSendEmbed(guildSettings.BotLog.ID, &embedMess)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func BanEmbed(s *discordgo.Session, m *discordgo.Message, mem *discordgo.User, reason string, length time.Time, perma bool, channelID string) error {

	var (
		embedMess      discordgo.MessageEmbed
		embedThumbnail discordgo.MessageEmbedThumbnail
		embedFooter    discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embedField       []*discordgo.MessageEmbedField
		embedFieldUserID discordgo.MessageEmbedField
		embedFieldReason discordgo.MessageEmbedField
	)

	// Sets timestamp for unban date and footer
	banDate := length.Format(time.RFC3339)
	embedMess.Timestamp = banDate
	embedFooter.Text = "Unban Date"
	embedMess.Footer = &embedFooter

	// Sets ban embed color
	embedMess.Color = 0xff0000

	// Saves user avatar as thumbnail
	embedThumbnail.URL = mem.AvatarURL("128")

	// Sets field titles
	embedFieldUserID.Name = "User ID:"
	embedFieldReason.Name = "Reason:"

	// Sets field content
	embedFieldUserID.Value = mem.ID
	embedFieldReason.Value = reason

	// Adds the two fields to embedField slice (because embedMess.Fields requires slice input)
	embedField = append(embedField, &embedFieldUserID)
	embedField = append(embedField, &embedFieldReason)

	// Sets embed title and its description (which it uses the same way as a field)
	if !perma {
		embedMess.Title = mem.Username + "#" + mem.Discriminator + " was banned by " + m.Author.Username
	} else {
		embedMess.Title = mem.Username + "#" + mem.Discriminator + " was permabanned by " + m.Author.Username
	}

	// Adds user thumbnail and the two other fields as well
	embedMess.Thumbnail = &embedThumbnail
	embedMess.Fields = embedField

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(channelID, &embedMess)
	return err
}

// Embed message for subscriptions
func SubEmbed(s *discordgo.Session, show *ShowAirTime, channelID string) error {
	imageLink := fmt.Sprintf("https://animeschedule.net/img/shows/%v.webp", show.Key)
	embed := &discordgo.MessageEmbed{
		URL:         fmt.Sprintf("https://animeschedule.net/shows/%v", show.Key),
		Title:       show.Name,
		Description: fmt.Sprintf("__**%v**__ is out!", show.Episode),
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       16758465,
		Image: &discordgo.MessageEmbedImage{
			Width:  30,
			Height: 60,
			URL:    imageLink,
		},
	}

	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		return err
	}

	return nil
}

func AboutEmbed(s *discordgo.Session, m *discordgo.Message) error {
	var embed *discordgo.MessageEmbed

	if s.State.User.ID == "614495694769618944" {
		embed = &discordgo.MessageEmbed{
			URL:         "https://discordbots.org/bot/614495694769618944",
			Title:       s.State.User.Username,
			Description: "Written in **Go** by _Apiks#8969_ with a focus on Moderation",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "\n**Features:**",
					Value:  "**-** Moderation Toolset & Member System\n**-** Autopost Anime Episodes (_subbed_) & Daily Anime Schedule\n**-** Autopost Reddit Feed\n**-** React-based Auto Role\n**-** Channel & Emoji stats\n**-** Raffles\n**-** and more!\n[Invite Link](https://discordapp.com/api/oauth2/authorize?client_id=614495694769618944&permissions=401960278&scope=bot)",
					Inline: false,
				},
				{
					Name:   "**Anime Times:**",
					Value:  "The Anime features derive their data from [AnimeSchedule.net](https://animeschedule.net), a site dedicated to showing you what anime are airing this week",
					Inline: false,
				},
				{
					Name:   "**Supporter Perks:**",
					Value:  "Consider becoming a [Patron](https://patreon.com/apiks) if you want to support me\n\n**-** Increased database limits for you and a server of your choice\n**-** Development Updates\n**-** Custom BOT Ping Messages\n**-** Be in the BOT Playing Field",
					Inline: false,
				},
			},
			Color: 16758465,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: s.State.User.AvatarURL("256"),
			},
		}
	} else if s.State.User.ID == "432579417974374400" {
		embed = &discordgo.MessageEmbed{
			URL:         "https://github.com/r-anime/ZeroTsu",
			Title:       s.State.User.Username,
			Description: "Written in **Go** by _Apiks#8969_ with a focus on Moderation",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "\n**Features:**",
					Value:  "**-** Moderation Toolset & Member System\n**-** Autopost Anime Episodes (_subbed_) & Daily Anime Schedule\n**-** Autopost Reddit Feed\n**-** React-based Auto Role\n**-** Channel & Emoji stats\n**-** Raffles\n**-** and more!\n[Invite Link](https://discordapp.com/api/oauth2/authorize?client_id=614495694769618944&permissions=401960278&scope=bot)",
					Inline: false,
				},
			},
			Color: 16758465,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: s.State.User.AvatarURL("256"),
			},
		}
	} else {
		embed = &discordgo.MessageEmbed{
			URL:         "https://github.com/r-anime/ZeroTsu",
			Title:       s.State.User.Username,
			Description: "Written in **Go** by _Apiks#8969_ with a focus on Moderation",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "\n**Features:**",
					Value:  "**-** Moderation Toolset & Member System\n**-** Autopost Anime Episodes (_subbed_) & Daily Anime Schedule\n**-** Autopost Reddit Feed\n**-** React-based Auto Role\n**-** Channel & Emoji stats\n**-** Raffles\n**-** and more!\n[Invite Link](https://discordapp.com/api/oauth2/authorize?client_id=614495694769618944&permissions=401960278&scope=bot)",
					Inline: false,
				},
				{
					Name:   "**Anime Times:**",
					Value:  "The Anime features derive their data from [AnimeSchedule.net](https://animeschedule.net), a site dedicated to showing you what anime are airing this week",
					Inline: false,
				},
			},
			Color: 16758465,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: s.State.User.AvatarURL("256"),
			},
		}
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		return err
	}

	return nil
}
