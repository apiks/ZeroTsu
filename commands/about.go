package commands

import (
	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

// Prints information about the BOT
func aboutCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		guildPrefix = "."
		guildBotLog string
	)

	if m.GuildID != "" {
		misc.MapMutex.Lock()
		guildPrefix = misc.GuildMap[m.GuildID].GuildConfig.Prefix
		guildBotLog = misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
		misc.MapMutex.Unlock()
	}

	err := aboutEmbed(s, m, guildPrefix)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
	}
}

func aboutEmbed(s *discordgo.Session, m *discordgo.Message, prefix string) error {
	var embed *discordgo.MessageEmbed

	if s.State.User.ID == "614495694769618944" {
		embed = &discordgo.MessageEmbed {
			URL:         "https://discordbots.org/bot/614495694769618944",
			Title:       s.State.User.Username,
			Description: "Written in **Go** by _Apiks#8969_ with a focus on Moderation",
			Fields: []*discordgo.MessageEmbedField {
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
			Color:       16758465,
			Thumbnail: &discordgo.MessageEmbedThumbnail {
				URL:s.State.User.AvatarURL("256"),
			},
		}
	} else {
		embed = &discordgo.MessageEmbed {
			URL:         "https://github.com/r-anime/ZeroTsu",
			Title:       s.State.User.Username,
			Description: "Written in **Go** by _Apiks#8969_ with a focus on Moderation",
			Fields: []*discordgo.MessageEmbedField {
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
			Color:       16758465,
			Thumbnail: &discordgo.MessageEmbedThumbnail {
				URL:s.State.User.AvatarURL("256"),
			},
		}
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	add(&command{
		execute:  aboutCommand,
		trigger:  "about",
		desc:     "Get info about me.",
		category: "normal",
		DMAble: true,
	})
}
