package commands

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

// Tracks emoji usage of server emojis
func OnMessageEmoji(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in OnMessageEmoji")
		}
	}()

	if m.GuildID == "" {
		return
	}
	if m.Author.ID == s.State.User.ID {
		return
	}

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	// Pulls the entire guild structure so we can check guild emojis from it later
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m.Message, err, guildBotLog)
	}

	// If a message contains a server emoji it tracks it
	for _, emoji := range guild.Emojis {
		if strings.Contains(m.Content, "<:"+emoji.APIName()+">") {
			var emojiStatsVar misc.Emoji

			// Counts emoji usages in a message
			emojiCount := strings.Count(m.Content, "<:"+emoji.APIName()+">")

			// If Emoji stat doesn't exist create it
			misc.MapMutex.Lock()
			if misc.GuildMap[m.GuildID].EmojiStats[emoji.ID] == nil {
				emojiStatsVar.ID = emoji.ID
				emojiStatsVar.Name = emoji.Name
				misc.GuildMap[m.GuildID].EmojiStats[emoji.ID] = &emojiStatsVar
			}
			if misc.GuildMap[m.GuildID].EmojiStats[emoji.ID].ID == "" {
				emojiStatsVar = *misc.GuildMap[m.GuildID].EmojiStats[emoji.ID]
				emojiStatsVar.ID = emoji.ID
				misc.GuildMap[m.GuildID].EmojiStats[emoji.ID] = &emojiStatsVar
			}
			misc.GuildMap[m.GuildID].EmojiStats[emoji.ID].MessageUsage += emojiCount
			misc.GuildMap[m.GuildID].EmojiStats[emoji.ID].UniqueMessageUsage += 1
			misc.MapMutex.Unlock()
		}
	}
}

// Tracks emoji react usage of server emojis
func OnMessageEmojiReact(s *discordgo.Session, r *discordgo.MessageReactionAdd) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in OnMessageEmojiReact")
		}
	}()

	if r.GuildID == "" {
		return
	}

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[r.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	// Pulls the entire guild structure so we can check guild emojis from it later
	guild, err := s.Guild(r.GuildID)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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
			if misc.GuildMap[r.GuildID].EmojiStats[emoji.ID] == nil {
				emojiStatsVar.ID = emoji.ID
				emojiStatsVar.Name = emoji.Name
				misc.GuildMap[r.GuildID].EmojiStats[emoji.ID] = &emojiStatsVar
			}
			if misc.GuildMap[r.GuildID].EmojiStats[emoji.ID].ID == "" {
				emojiStatsVar = *misc.GuildMap[r.GuildID].EmojiStats[emoji.ID]
				emojiStatsVar.ID = emoji.ID
				misc.GuildMap[r.GuildID].EmojiStats[emoji.ID] = &emojiStatsVar
			}
			misc.GuildMap[r.GuildID].EmojiStats[emoji.ID].Reactions += 1
			misc.MapMutex.Unlock()
		}
	}
}

// Tracks emoji unreact usage of server emojis
func OnMessageEmojiUnreact(s *discordgo.Session, r *discordgo.MessageReactionRemove) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in OnMessageEmojiUnreact")
		}
	}()

	if r.GuildID == "" {
		return
	}

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[r.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	// Pulls the entire guild structure so we can check guild emojis from it later
	guild, err := s.Guild(r.GuildID)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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
			if misc.GuildMap[r.GuildID].EmojiStats[emoji.ID] == nil {
				emojiStatsVar.ID = emoji.ID
				emojiStatsVar.Name = emoji.Name
				misc.GuildMap[r.GuildID].EmojiStats[emoji.ID] = &emojiStatsVar
			}
			if misc.GuildMap[r.GuildID].EmojiStats[emoji.ID].ID == "" {
				emojiStatsVar = *misc.GuildMap[r.GuildID].EmojiStats[emoji.ID]
				emojiStatsVar.ID = emoji.ID
				misc.GuildMap[r.GuildID].EmojiStats[emoji.ID] = &emojiStatsVar
			}
			misc.GuildMap[r.GuildID].EmojiStats[emoji.ID].Reactions -= 1
			misc.MapMutex.Unlock()
		}
	}
}

// Display emoji stats
func showEmojiStats(s *discordgo.Session, m *discordgo.Message) {

	var (
		msgs          []string
		printEmojiMap = make(map[string]*misc.Emoji)
		guildFlag     bool
	)

	// Merges duplicates and returns that as a map
	misc.MapMutex.Lock()
	printEmojiMap = mergeDuplicates(m.GuildID)

	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID

	// Sorts emojis by their message use from the above map
	emojis := make([]*misc.Emoji, len(printEmojiMap))
	for i := 0; i < len(printEmojiMap); i++ {
		for _, emoji := range printEmojiMap {
			emojis[i] = emoji
			i++
		}
	}
	sort.Sort(byEmojiFrequency(emojis))
	misc.MapMutex.Unlock()

	guild, err := s.Guild(m.GuildID)
	if err != nil {
		misc.ErrorLocation(err)
		return
	}

	// Add every emoji and its stats to message and format it
	message := "```CSS\nName:                         ([Message Usage] | [Unique Usage] | [Reactions]) \n\n"
	misc.MapMutex.Lock()
	for _, emoji := range emojis {

		// Checks if an emoji with that name exists on the server before adding to print
		for _, guildEmoji := range guild.Emojis {
			if guildEmoji.Name == emoji.Name {
				guildFlag = true
				break
			}
		}
		if !guildFlag {
			continue
		}

		if emoji.Name != "" {
			message += lineSpaceFormatEmoji(emoji.Name, printEmojiMap)
			msgs, message = splitStatMessages(msgs, message)
		}

		// Resets guildFlag
		guildFlag = false
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
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
	}
}

// Formats the line space length for the above to keep level spacing
func lineSpaceFormatEmoji(name string, printEmojiMap map[string]*misc.Emoji) string {
	line := fmt.Sprintf("%v", name)
	spacesRequired := 30 - len(name)
	for i := 0; i < spacesRequired; i++ {
		line += " "
	}
	line += fmt.Sprintf("([%d])", printEmojiMap[name].MessageUsage)
	spacesRequired = 47 - len(line)
	for i := 0; i < spacesRequired; i++ {
		line += " "
	}
	line += fmt.Sprintf("| ([%d])", printEmojiMap[name].UniqueMessageUsage)
	spacesRequired = 64 - len(line)
	for i := 0; i < spacesRequired; i++ {
		line += " "
	}
	line += fmt.Sprintf("| ([%d])\n", printEmojiMap[name].Reactions)

	return line
}

// Merges duplicate emotes in EmojiStats
func mergeDuplicates(guildID string) map[string]*misc.Emoji {

	var (
		duplicateMap  = make(map[string]string)
		uniqueTotal   int
		reactTotal    int
		msgTotal      int
		printEmojiMap = make(map[string]*misc.Emoji)
	)

	// Fetches the IDs of all of the emojis that have at least one duplicate in duplicateMap
	for _, emoji := range misc.GuildMap[guildID].EmojiStats {
		for _, emojiTwo := range misc.GuildMap[guildID].EmojiStats {
			if emoji.ID == emojiTwo.ID {
				continue
			}
			if emoji.Name != emojiTwo.Name {
				continue
			}
			if _, ok := duplicateMap[emojiTwo.ID]; ok {
				continue
			}
			if _, ok := duplicateMap[emoji.ID]; !ok {
				duplicateMap[emoji.ID] = emoji.Name
			}

			duplicateMap[emojiTwo.ID] = emojiTwo.Name
		}
	}

	// Merges their values and leaves only one of them in a new map for printing purposes
	for duplicateOneID, duplicateOneName := range duplicateMap {
		// Emoji var here so it resets every iteration
		var emoji misc.Emoji

		if _, ok := misc.GuildMap[guildID].EmojiStats[duplicateOneID]; !ok {
			continue
		}

		// Fetch current iteration values
		uniqueTotal = misc.GuildMap[guildID].EmojiStats[duplicateOneID].UniqueMessageUsage
		reactTotal = misc.GuildMap[guildID].EmojiStats[duplicateOneID].Reactions
		msgTotal = misc.GuildMap[guildID].EmojiStats[duplicateOneID].MessageUsage

		for duplicateTwoID, duplicateTwoName := range duplicateMap {
			if duplicateOneID == duplicateTwoID {
				continue
			}
			if duplicateOneID == "" || duplicateTwoID == "" {
				continue
			}
			if strings.ToLower(duplicateOneName) == strings.ToLower(duplicateTwoName) {
				if _, ok := misc.GuildMap[guildID].EmojiStats[duplicateTwoID]; !ok {
					continue
				}
				uniqueTotal += misc.GuildMap[guildID].EmojiStats[duplicateTwoID].UniqueMessageUsage
				reactTotal += misc.GuildMap[guildID].EmojiStats[duplicateTwoID].Reactions
				msgTotal += misc.GuildMap[guildID].EmojiStats[duplicateTwoID].MessageUsage
				continue
			}
		}

		if _, ok := printEmojiMap[duplicateOneName]; !ok {
			emoji.Name = duplicateOneName
			emoji.ID = duplicateOneID
			emoji.MessageUsage = msgTotal
			emoji.Reactions = reactTotal
			emoji.UniqueMessageUsage = uniqueTotal
			printEmojiMap[duplicateOneName] = &emoji
		}

		// Reset values
		uniqueTotal = 0
		reactTotal = 0
		msgTotal = 0
	}

	// Adds non-duplicate values to the print map
	for _, statEmoji := range misc.GuildMap[guildID].EmojiStats {
		if _, ok := printEmojiMap[statEmoji.Name]; !ok {
			printEmojiMap[statEmoji.Name] = statEmoji
		}
	}

	return printEmojiMap
}

// Sort functions for emoji use by message use
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
