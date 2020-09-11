package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"log"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
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

	entities.HandleNewGuild(m.GuildID)

	var saveFlag bool

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildEmojiStatsReference := db.GetGuildEmojiStats(m.GuildID)
	guildEmojiStats := guildEmojiStatsReference

	// Pulls the entire guild structure so we can check guild emojis from it later
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		guild, err = s.Guild(m.GuildID)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
	}

	// If a message contains a server emoji it tracks it
	for _, emoji := range guild.Emojis {
		if !strings.Contains(m.Content, "<:"+emoji.APIName()+">") {
			continue
		}

		// Counts emoji usages in a message
		emojiCount := strings.Count(m.Content, "<:"+emoji.APIName()+">")

		// If Emoji stat doesn't exist create it
		if _, ok := guildEmojiStats[emoji.ID]; !ok {
			guildEmojiStats[emoji.ID] = entities.NewEmoji(emoji.ID, emoji.Name, 0, 0, 0)
		}
		// If it's missing ID or Name add it while preserving stats
		if guildEmojiStats[emoji.ID].GetID() == "" || guildEmojiStats[emoji.ID].GetName() == "" {
			guildEmojiStats[emoji.ID] = guildEmojiStats[emoji.ID].SetID(emoji.ID)
			guildEmojiStats[emoji.ID] = guildEmojiStats[emoji.ID].SetName(emoji.Name)
		}

		// Adds to that emoji usage
		guildEmojiStats[emoji.ID] = guildEmojiStats[emoji.ID].AddMessageUsage(emojiCount)
		guildEmojiStats[emoji.ID] = guildEmojiStats[emoji.ID].AddUniqueMessageUsage(1)

		saveFlag = true
	}

	if saveFlag {
		db.SetGuildEmojiStats(m.GuildID, guildEmojiStats)
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
	if r.UserID == s.State.User.ID {
		return
	}

	entities.HandleNewGuild(r.GuildID)

	var saveFlag bool

	guildSettings := db.GetGuildSettings(r.GuildID)
	guildEmojiStatsReference := db.GetGuildEmojiStats(r.GuildID)
	guildEmojiStats := guildEmojiStatsReference

	// Pulls the entire guild structure so we can check guild emojis from it later
	guild, err := s.State.Guild(r.GuildID)
	if err != nil {
		guild, err = s.Guild(r.GuildID)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
	}

	// If a message contains a server emoji it tracks it
	for _, emoji := range guild.Emojis {
		if r.Emoji.ID != emoji.ID {
			continue
		}

		// If Emoji stat doesn't exist create it
		if _, ok := guildEmojiStats[emoji.ID]; !ok {
			guildEmojiStats[emoji.ID] = entities.NewEmoji(emoji.ID, emoji.Name, 0, 0, 0)
		}
		// If it's missing ID or Name add it while preserving stats
		if guildEmojiStats[emoji.ID].GetID() == "" || guildEmojiStats[emoji.ID].GetName() == "" {
			guildEmojiStats[emoji.ID] = guildEmojiStats[emoji.ID].SetID(emoji.ID)
			guildEmojiStats[emoji.ID] = guildEmojiStats[emoji.ID].SetName(emoji.Name)
		}

		// Adds to that emoji usage
		guildEmojiStats[emoji.ID] = guildEmojiStats[emoji.ID].AddSetReactions(1)

		saveFlag = true
	}

	if saveFlag {
		db.SetGuildEmojiStats(r.GuildID, guildEmojiStats)
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
	if r.UserID == s.State.User.ID {
		return
	}

	entities.HandleNewGuild(r.GuildID)

	var saveFlag bool

	guildSettings := db.GetGuildSettings(r.GuildID)
	guildEmojiStatsReference := db.GetGuildEmojiStats(r.GuildID)
	guildEmojiStats := guildEmojiStatsReference

	// Pulls the entire guild structure so we can check guild emojis from it later
	guild, err := s.State.Guild(r.GuildID)
	if err != nil {
		guild, err = s.Guild(r.GuildID)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
	}

	// If a message contains a server emoji it tracks it
	for _, emoji := range guild.Emojis {
		if r.Emoji.ID != emoji.ID {
			continue
		}

		// If Emoji stat doesn't exist create it
		if _, ok := guildEmojiStats[emoji.ID]; !ok {
			guildEmojiStats[emoji.ID] = entities.NewEmoji(emoji.ID, emoji.Name, 0, 0, 0)
		}
		// If it's missing ID or Name add it while preserving stats
		if guildEmojiStats[emoji.ID].GetID() == "" || guildEmojiStats[emoji.ID].GetName() == "" {
			guildEmojiStats[emoji.ID] = guildEmojiStats[emoji.ID].SetID(emoji.ID)
			guildEmojiStats[emoji.ID] = guildEmojiStats[emoji.ID].SetName(emoji.Name)
		}

		// Adds to that emoji usage
		guildEmojiStats[emoji.ID] = guildEmojiStats[emoji.ID].AddSetReactions(-1)

		saveFlag = true
	}

	if saveFlag {
		db.SetGuildEmojiStats(r.GuildID, guildEmojiStats)
	}
}

// Display emoji stats
func showEmojiStats(s *discordgo.Session, m *discordgo.Message) {
	entities.HandleNewGuild(m.GuildID)

	var msgs []string

	guildSettings := db.GetGuildSettings(m.GuildID)
	printEmojiMap, err := mergeDuplicates(m.GuildID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Sorts emojis by their message use from the above map
	emojis := make([]entities.Emoji, len(printEmojiMap))
	for i := 0; i < len(printEmojiMap); i++ {
		for _, emoji := range printEmojiMap {
			emojis[i] = emoji
			i++
		}
	}
	sort.Sort(byEmojiFrequency(emojis))

	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		guild, err = s.Guild(m.GuildID)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
	}

	// Add every emoji and its stats to message and format it
	message := "```CSS\nName:                         ([Message Usage] | [Unique Usage] | [Reactions]) \n\n"
	for _, emoji := range emojis {
		var guildFlag bool

		// Checks if an emoji with that name exists on the server before adding to print
		for _, guildEmoji := range guild.Emojis {
			if guildEmoji.Name != emoji.GetName() {
				continue
			}
			guildFlag = true
			break
		}
		if !guildFlag {
			continue
		}
		if emoji.GetName() == "" {
			continue
		}

		message += lineSpaceFormatEmoji(emoji.GetName(), printEmojiMap)
		msgs, message = splitStatMessages(msgs, message)
	}

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
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
	}
}

// Formats the line space length for the above to keep level spacing
func lineSpaceFormatEmoji(name string, printEmojiMap map[string]entities.Emoji) string {
	line := fmt.Sprintf("%v", name)
	spacesRequired := 30 - len(name)
	for i := 0; i < spacesRequired; i++ {
		line += " "
	}
	line += fmt.Sprintf("([%d])", printEmojiMap[name].GetMessageUsage())
	spacesRequired = 47 - len(line)
	for i := 0; i < spacesRequired; i++ {
		line += " "
	}
	line += fmt.Sprintf("| ([%d])", printEmojiMap[name].GetUniqueMessageUsage())
	spacesRequired = 64 - len(line)
	for i := 0; i < spacesRequired; i++ {
		line += " "
	}

	return line + fmt.Sprintf("| ([%d])\n", printEmojiMap[name].GetReactions())
}

// Merges duplicate emotes in EmojiStats
func mergeDuplicates(guildID string) (map[string]entities.Emoji, error) {
	var (
		duplicateMap  = make(map[string]string)
		uniqueTotal   int
		reactTotal    int
		msgTotal      int
		printEmojiMap = make(map[string]entities.Emoji)
	)

	guildEmojiStats := db.GetGuildEmojiStats(guildID)

	// Fetches the IDs of all of the emojis that have at least one duplicate in duplicateMap
	for _, emoji := range guildEmojiStats {
		for _, emojiTwo := range guildEmojiStats {
			if emoji.GetID() == emojiTwo.GetID() {
				continue
			}
			if emoji.GetName() != emojiTwo.GetName() {
				continue
			}
			if _, ok := duplicateMap[emojiTwo.GetID()]; ok {
				continue
			}
			if _, ok := duplicateMap[emoji.GetID()]; !ok {
				duplicateMap[emoji.GetID()] = emoji.GetName()
			}

			duplicateMap[emojiTwo.GetID()] = emojiTwo.GetName()
		}
	}

	// Merges their values and leaves only one of them in a new map for printing purposes
	for duplicateOneID, duplicateOneName := range duplicateMap {
		// Emoji var here so it resets every iteration
		var emoji entities.Emoji

		emojiStatOne := db.GetGuildEmojiStat(guildID, duplicateOneID)

		// Fetch current iteration values
		msgTotal = emojiStatOne.GetMessageUsage()
		reactTotal = emojiStatOne.GetReactions()
		uniqueTotal = emojiStatOne.GetUniqueMessageUsage()

		for duplicateTwoID, duplicateTwoName := range duplicateMap {
			if duplicateOneID == duplicateTwoID {
				continue
			}
			if duplicateOneID == "" || duplicateTwoID == "" {
				continue
			}
			if strings.ToLower(duplicateOneName) == strings.ToLower(duplicateTwoName) {
				emojiStatTwo := db.GetGuildEmojiStat(guildID, duplicateTwoID)

				msgTotal += emojiStatTwo.GetMessageUsage()
				uniqueTotal += emojiStatTwo.GetUniqueMessageUsage()
				reactTotal += emojiStatTwo.GetReactions()
				continue
			}
		}

		if _, ok := printEmojiMap[duplicateOneName]; !ok {
			emoji = emoji.SetName(duplicateOneName)
			emoji = emoji.SetID(duplicateOneID)
			emoji = emoji.SetMessageUsage(msgTotal)
			emoji = emoji.SetUniqueMessageUsage(uniqueTotal)
			emoji = emoji.SetReactions(reactTotal)
			printEmojiMap[duplicateOneName] = emoji
		}

		// Reset values
		uniqueTotal = 0
		reactTotal = 0
		msgTotal = 0
	}

	// Adds non-duplicate values to the print map
	for _, statEmoji := range guildEmojiStats {
		if _, ok := printEmojiMap[statEmoji.GetName()]; !ok {
			printEmojiMap[statEmoji.GetName()] = statEmoji
		}
	}

	return printEmojiMap, nil
}

// Sort functions for emoji use by message use
type byEmojiFrequency []entities.Emoji

func (e byEmojiFrequency) Len() int {
	return len(e)
}
func (e byEmojiFrequency) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
func (e byEmojiFrequency) Less(i, j int) bool {
	return e[j].GetMessageUsage() < e[i].GetMessageUsage()
}

// Adds emoji stat command to the commandHandler
func init() {
	Add(&Command{
		Execute:    showEmojiStats,
		Trigger:    "emoji",
		Aliases:    []string{"emojistats", "emojis", "emotes", "emote"},
		Desc:       "Print server emoji usage stats",
		Permission: functionality.Mod,
		Module:     "stats",
	})
}
