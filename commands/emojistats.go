package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Tracks emoji usage of server emojis
func OnMessageEmoji(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string))
			if err != nil {
				return
			}
		}
	}()

	if m.Author.ID == s.State.User.ID {
		return
	}
	// Checks if it's within config server
	ch, err := s.State.Channel(m.ChannelID)
	if err != nil {
		ch, err = s.Channel(m.ChannelID)
		if err != nil {
			return
		}
	}
	if ch.GuildID != config.ServerID {
		return
	}
	// Pulls the entire guild structure so we can check guild emojis from it later
	guild, err := s.Guild(ch.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m.Message, err)
	}

	// If a message contains a server emoji it tracks it
	for _, emoji := range guild.Emojis {
		if strings.Contains(m.Content, "<:" + emoji.APIName() + ">") {
			var emojiStatsVar misc.Emoji

			// Counts emoji usages in a message
			emojiCount := strings.Count(m.Content, "<:" + emoji.APIName() + ">")

			// If Emoji stat doesn't exist create it
			misc.MapMutex.Lock()
			if misc.EmojiStats[emoji.ID] == nil {
				emojiStatsVar.ID = emoji.ID
				emojiStatsVar.Name = emoji.Name
				misc.EmojiStats[emoji.ID] = &emojiStatsVar
			}
			if misc.EmojiStats[emoji.ID].ID == "" {
				emojiStatsVar = *misc.EmojiStats[emoji.ID]
				emojiStatsVar.ID = emoji.ID
				misc.EmojiStats[emoji.ID] = &emojiStatsVar
			}
			misc.EmojiStats[emoji.ID].MessageUsage += emojiCount
			misc.EmojiStats[emoji.ID].UniqueMessageUsage += 1
			misc.MapMutex.Unlock()
		}
	}
}

// Tracks emoji react usage of server emojis
func OnMessageEmojiReact(s *discordgo.Session, r *discordgo.MessageReactionAdd) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string))
			if err != nil {
				return
			}
		}
	}()

	// Checks if it's within the /r/anime server
	ch, err := s.State.Channel(r.ChannelID)
	if err != nil {
		ch, err = s.Channel(r.ChannelID)
		if err != nil {
			return
		}
	}
	if ch.GuildID != config.ServerID {
		return
	}
	// Pulls the entire guild structure so we can check guild emojis from it later
	guild, err := s.Guild(ch.GuildID)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	// If a message contains a server emoji it tracks it
	for _, emoji := range guild.Emojis {
		if r.Emoji.ID == emoji.ID {
			var emojiStatsVar misc.Emoji

			// If Emoji stat doesn't exist create it
			misc.MapMutex.Lock()
			if misc.EmojiStats[emoji.ID] == nil {
				emojiStatsVar.ID = emoji.ID
				emojiStatsVar.Name = emoji.Name
				misc.EmojiStats[emoji.ID] = &emojiStatsVar
			}
			if misc.EmojiStats[emoji.ID].ID == "" {
				emojiStatsVar = *misc.EmojiStats[emoji.ID]
				emojiStatsVar.ID = emoji.ID
				misc.EmojiStats[emoji.ID] = &emojiStatsVar
			}
			misc.EmojiStats[emoji.ID].Reactions += 1
			misc.MapMutex.Unlock()
		}
	}
}

// Tracks emoji unreact usage of server emojis
func OnMessageEmojiUnreact(s *discordgo.Session, r *discordgo.MessageReactionRemove) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string))
			if err != nil {
				return
			}
		}
	}()

	// Checks if it's within the /r/anime server
	ch, err := s.State.Channel(r.ChannelID)
	if err != nil {
		ch, err = s.Channel(r.ChannelID)
		if err != nil {
			return
		}
	}
	if ch.GuildID != config.ServerID {
		return
	}
	// Pulls the entire guild structure so we can check guild emojis from it later
	guild, err := s.Guild(ch.GuildID)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	// If a message contains a server emoji it tracks it
	for _, emoji := range guild.Emojis {
		if r.Emoji.ID == emoji.ID {
			var emojiStatsVar misc.Emoji

			// If Emoji stat doesn't exist create it
			misc.MapMutex.Lock()
			if misc.EmojiStats[emoji.ID] == nil {
				emojiStatsVar.ID = emoji.ID
				emojiStatsVar.Name = emoji.Name
				misc.EmojiStats[emoji.ID] = &emojiStatsVar
			}
			if misc.EmojiStats[emoji.ID].ID == "" {
				emojiStatsVar = *misc.EmojiStats[emoji.ID]
				emojiStatsVar.ID = emoji.ID
				misc.EmojiStats[emoji.ID] = &emojiStatsVar
			}
			misc.EmojiStats[emoji.ID].Reactions -= 1
			misc.MapMutex.Unlock()
		}
	}
}

// Display emoji stats
func showEmojiStats(s *discordgo.Session, m *discordgo.Message) {

	var msgs []string

	// Sorts emojis by their message use
	misc.MapMutex.Lock()
	emojis := make([]*misc.Emoji, len(misc.EmojiStats))
	for i := 0; i < len(misc.EmojiStats); i++ {
		for _, emoji := range misc.EmojiStats {
			emojis[i] = emoji
			i++
		}
	}
	sort.Sort(byEmojiFrequency(emojis))
	misc.MapMutex.Unlock()

	// Pull guild info
	guild, err := s.State.Guild(config.ServerID)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	// Add every emoji and its stats to message and format it
	message := "```CSS\nName:                         ([Message Usage] | [Unique Usage] | [Reactions]) \n\n"
	misc.MapMutex.Lock()
	for _, emoji := range emojis {
		// Fixes emojis without ID
		if emoji.ID == "" {
			for index := range guild.Emojis {
				if guild.Emojis[index].Name == emoji.Name {
					emoji.ID = guild.Emojis[index].ID
					misc.EmojiStats[emoji.ID] = emoji
					break
				}
			}
		}

		if emoji.ID != "" {
			message += lineSpaceFormatEmoji(emoji.ID)
			msgs, message = splitStatMessages(msgs, message)
		}
	}
	misc.MapMutex.Unlock()

	msgs, message = splitStatMessages(msgs, message)
	if message != "" {
		msgs = append(msgs, message)
	}
	msgs[0] += "```"
	for i := 1; i < len(msgs); i++ {
		msgs[i] = "```CSS\n" + msgs[i] + "\n```"
	}

	for j := 0; j < len(msgs); j++ {
		_, err := s.ChannelMessageSend(m.ChannelID, msgs[j])
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
	}
}

// Formats the line space length for the above to keep level spacing
func lineSpaceFormatEmoji(id string) string {
	line := fmt.Sprintf("%v", misc.EmojiStats[id].Name)
	spacesRequired := 30 - len(misc.EmojiStats[id].Name)
	for i := 0; i < spacesRequired; i++ {
		line += " "
	}
	line += fmt.Sprintf("([%d])", misc.EmojiStats[id].MessageUsage)
	spacesRequired = 47 - len(line)
	for i := 0; i < spacesRequired; i++ {
		line += " "
	}
	line += fmt.Sprintf("| ([%d])", misc.EmojiStats[id].UniqueMessageUsage)
	spacesRequired = 64 - len(line)
	for i := 0; i < spacesRequired; i++ {
		line += " "
	}
	line += fmt.Sprintf("| ([%d])\n", misc.EmojiStats[id].Reactions)

	return line
}

// Sort functions for emoji use by message use. By Kagumi
type byEmojiFrequency []*misc.Emoji

func (e byEmojiFrequency) Len() int {
	return len(e)
}
func (e byEmojiFrequency) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
func (e byEmojiFrequency) Less(i, j int) bool {
	return e[j].MessageUsage < e[i].MessageUsage
}

// Adds emoji stat command to the commandHandler
func init() {
	add(&command{
		execute:  showEmojiStats,
		trigger:  "emoji",
		aliases:  []string{"emojistats", "emojis"},
		desc:     "Prints server emoji usage stats.",
		elevated: true,
		category: "stats",
	})
}