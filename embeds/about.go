package embeds

import "github.com/bwmarrin/discordgo"

// About sends an embed about message
func About(s *discordgo.Session, m *discordgo.Message) error {
	var embed = &discordgo.MessageEmbed{
		URL:         "https://discordbots.org/bot/614495694769618944",
		Title:       s.State.User.Username,
		Description: "Written in **Go** by _Apiks#8969_",
		Color:       purpleColor,
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: s.State.User.AvatarURL("256")},
		Fields: []*discordgo.MessageEmbedField{{
			Name: "**Features:**",
			Value: "**-** Autopost Anime Episodes, Anime Schedule (_subbed_)\n" +
				"**-** Autopost Reddit Feed\n**-** React-based Auto Role\n" +
				"**-** Raffles" +
				"\n[Invite Link](https://discord.com/oauth2/authorize?client_id=614495694769618944&scope=bot&permissions=335899736)"}},
	}

	if s.State.User.ID != "432579417974374400" {
		var (
			animeTimesEmbedField = &discordgo.MessageEmbedField{
				Name:  "**Anime Times:**",
				Value: "The Anime features derive their data from [AnimeSchedule.net](https://animeschedule.net), a site dedicated to showing you when and what anime are airing this week.",
			}

			supporterPerksEmbedField = &discordgo.MessageEmbedField{
				Name:  "**Supporter Perks:**",
				Value: "Consider becoming a [Patron](https://patreon.com/animeschedule) if you want to support me and get: \n**-** Increased database limits for you and a server of your choice\n**-** Development Updates\n**-** Be in the BOT Status",
			}
		)

		embed.Fields = append(embed.Fields, animeTimesEmbedField)
		embed.Fields = append(embed.Fields, supporterPerksEmbedField)
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}
