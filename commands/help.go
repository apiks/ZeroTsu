package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Command categories in sorted form and map form(map for descriptions)
var (
	categoriesSorted = [...]string{"Channel", "Filters", "Misc", "Normal", "Punishment", "Reacts", "Rss", "Stats", "Raffles", "Waifus", "Settings"}
	categoriesMap    = make(map[string]string)
)

// Prints pretty help command
func helpEmbedCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		elevated bool
		admin    bool
	)

	if m.GuildID == "" {
		_ = helpEmbed(s, m, elevated, admin)
		return
	}

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID

	// Checks for mod perms and handles accordingly
	if HasElevatedPermissions(s, m.Author.ID, m.GuildID) {
		elevated = true
	}
	misc.MapMutex.Unlock()

	// Check perms
	mem, err := s.State.Member(m.GuildID, m.Author.ID)
	if err != nil {
		mem, err = s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			return
		}
	}
	admin, err = MemberIsAdmin(s, m.GuildID, mem, discordgo.PermissionAdministrator)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
	}

	err = helpEmbed(s, m, elevated, admin)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
}

// Embed message for general all-purpose help message
func helpEmbed(s *discordgo.Session, m *discordgo.Message, elevated bool, admin bool) error {

	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed           []*discordgo.MessageEmbedField
		user            discordgo.MessageEmbedField
		permission      discordgo.MessageEmbedField
		userCommands    discordgo.MessageEmbedField
		adminCategories discordgo.MessageEmbedField

		// Slice for sorting
		commands []string

		guildPrefix = "."
		guildBotLog string
		guildWaifuModule bool
		guildReactsModule bool
	)

	if m.GuildID != "" {
		misc.MapMutex.Lock()
		guildPrefix = misc.GuildMap[m.GuildID].GuildConfig.Prefix
		guildBotLog = misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
		guildWaifuModule = misc.GuildMap[m.GuildID].GuildConfig.WaifuModule
		guildReactsModule = misc.GuildMap[m.GuildID].GuildConfig.ReactsModule
		misc.MapMutex.Unlock()
	}

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets user field
	user.Name = "User:"
	user.Value = m.Author.Mention()
	user.Inline = true

	// Sets permission field
	permission.Name = "Permission Level:"
	if m.Author.ID == config.OwnerID {
		permission.Value = "_Owner_"
	} else if admin {
		permission.Value = "_Admin_"
	} else if elevated {
		permission.Value = "_Mod_"
	} else {
		permission.Value = "_User_"
	}
	permission.Inline = true

	// Sets usage field if elevated
	if elevated {
		// Sets footer field
		embedFooter.Text = fmt.Sprintf("Usage: %vh-category | (Example: %vh-settings)", guildPrefix, guildPrefix)
		embedMess.Footer = &embedFooter
	}

	if !elevated {
		// Sets commands field
		userCommands.Name = "Command:"
		userCommands.Inline = true

		// Iterates through non-mod commands and adds them to the embed sorted
		misc.MapMutex.Lock()
		for command := range commandMap {
			commands = append(commands, command)
		}
		sort.Strings(commands)
		for i := 0; i < len(commands); i++ {
			if m.GuildID == "" {
				if !commandMap[commands[i]].DMAble {
					continue
				}
			}
			if !commandMap[commands[i]].elevated {
				if commandMap[commands[i]].category == "waifus" {
					if !guildWaifuModule {
						continue
					}
				}
				userCommands.Value += fmt.Sprintf("`%v` - %v\n", commands[i], commandMap[commands[i]].desc)
			}
		}
		misc.MapMutex.Unlock()

		// Sets footer field
		embedFooter.Text = fmt.Sprintf("Tip: Type %v<command> to see a detailed description.", guildPrefix)
		embedMess.Footer = &embedFooter
	} else {
		// Sets elevated commands field
		adminCategories.Name = "Categories:"
		adminCategories.Inline = true

		// Iterates through categories and their descriptions and adds them to the embed. Special behavior for waifus and reacts and settings based on settings
		misc.MapMutex.Lock()
		for i := 0; i < len(categoriesSorted); i++ {
			if categoriesSorted[i] == "Waifus" {
				if !guildWaifuModule {
					continue
				}
			}
			if categoriesSorted[i] == "Reacts" {
				if !guildReactsModule {
					continue
				}
			}
			if categoriesSorted[i] == "Settings" {
				if !admin && m.Author.ID != config.OwnerID {
					continue
				}
			}
			adminCategories.Value += fmt.Sprintf("%v - %v\n", categoriesSorted[i], categoriesMap[categoriesSorted[i]])
		}
		misc.MapMutex.Unlock()
	}

	// Adds the fields to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &user)
	embed = append(embed, &permission)
	if elevated {
		embed = append(embed, &adminCategories)
	} else {
		embed = append(embed, &userCommands)
	}

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, elevated)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return err
		}
	}
	return nil
}

// Mod command help page
func helpChannelCommand(s *discordgo.Session, m *discordgo.Message) {
	err := helpChannelEmbed(s, m)
	if err != nil {

		misc.MapMutex.Lock()
		guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
		misc.MapMutex.Unlock()

		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
}

// Mod command help page embed
func helpChannelEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %vcommand to see a detailed description.", guildPrefix)
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the filter category
	misc.MapMutex.Lock()
	for command := range commandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if commandMap[commands[i]].category == "channel" {
			commandsField.Value += fmt.Sprintf("`%v` - %v\n", commands[i], commandMap[commands[i]].desc)
		}
	}
	misc.MapMutex.Unlock()

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return err
		}
	}
	return nil
}

// Mod command help page
func helpFiltersCommand(s *discordgo.Session, m *discordgo.Message) {
	err := helpFiltersEmbed(s, m)
	if err != nil {

		misc.MapMutex.Lock()
		guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
		misc.MapMutex.Unlock()

		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
}

// Mod command help page embed
func helpFiltersEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %vcommand to see a detailed description.", guildPrefix)
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the filter category
	misc.MapMutex.Lock()
	for command := range commandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if commandMap[commands[i]].category == "filters" {
			commandsField.Value += fmt.Sprintf("`%v` - %v\n", commands[i], commandMap[commands[i]].desc)
		}
	}
	misc.MapMutex.Unlock()

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return err
		}
	}
	return nil
}

// Mod command help page
func helpMiscCommand(s *discordgo.Session, m *discordgo.Message) {
	err := helpMiscEmbed(s, m)
	if err != nil {

		misc.MapMutex.Lock()
		guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
		misc.MapMutex.Unlock()

		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
}

// Mod command help page embed
func helpMiscEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %vcommand to see a detailed description.", guildPrefix)
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the filter category
	misc.MapMutex.Lock()
	for command := range commandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if config.Website == "" {
			if commandMap[commands[i]].trigger == "verify" ||
				commandMap[commands[i]].trigger == "unverify" {
				continue
			}
		}
		if commandMap[commands[i]].category == "misc" {
			commandsField.Value += fmt.Sprintf("`%v` - %v\n", commands[i], commandMap[commands[i]].desc)
		}
	}
	misc.MapMutex.Unlock()

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return err
		}
	}
	return nil
}

// Mod command help page
func helpNormalCommand(s *discordgo.Session, m *discordgo.Message) {
	err := helpNormalEmbed(s, m)
	if err != nil {

		misc.MapMutex.Lock()
		guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
		misc.MapMutex.Unlock()

		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
}

// Mod command help page embed
func helpNormalEmbed(s *discordgo.Session, m *discordgo.Message) error {
	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %vcommand to see a detailed description.", guildPrefix)
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the filter category
	misc.MapMutex.Lock()
	for command := range commandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if commandMap[commands[i]].category == "normal" {
			commandsField.Value += fmt.Sprintf("`%v` - %v\n", commands[i], commandMap[commands[i]].desc)
		}
	}
	misc.MapMutex.Unlock()

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return err
		}
	}

	return nil
}

// Mod command help page
func helpPunishmentCommand(s *discordgo.Session, m *discordgo.Message) {
	err := helpPunishmentEmbed(s, m)
	if err != nil {

		misc.MapMutex.Lock()
		guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
		misc.MapMutex.Unlock()

		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
}

// Mod command help page embed
func helpPunishmentEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %vcommand to see a detailed description.", guildPrefix)
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the filter category
	misc.MapMutex.Lock()
	for command := range commandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if commandMap[commands[i]].category == "punishment" {
			commandsField.Value += fmt.Sprintf("`%v` - %v\n", commands[i], commandMap[commands[i]].desc)
		}
	}
	misc.MapMutex.Unlock()

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return err
		}
	}
	return nil
}

// Prints pretty help
func helpReactsCommand(s *discordgo.Session, m *discordgo.Message) {

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	guildReactsModule := misc.GuildMap[m.GuildID].GuildConfig.ReactsModule
	misc.MapMutex.Unlock()

	// Checks if reacts are disabled in the guild
	if !guildReactsModule {
		return
	}

	err := helpReactsEmbed(s, m)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
}

// Mod command help page embed
func helpReactsEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %vcommand to see a detailed description.", guildPrefix)
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the filter category
	misc.MapMutex.Lock()
	for command := range commandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if commandMap[commands[i]].category == "reacts" {
			commandsField.Value += fmt.Sprintf("`%v` - %v\n", commands[i], commandMap[commands[i]].desc)
		}
	}
	misc.MapMutex.Unlock()

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return err
		}
	}
	return nil
}

// Mod command help page
func helpRssCommand(s *discordgo.Session, m *discordgo.Message) {
	err := helpRssEmbed(s, m)
	if err != nil {

		misc.MapMutex.Lock()
		guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
		misc.MapMutex.Unlock()

		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
}

// Mod command help page embed
func helpRssEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %vcommand to see a detailed description.", guildPrefix)
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the filter category
	misc.MapMutex.Lock()
	for command := range commandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if commandMap[commands[i]].category == "rss" {
			commandsField.Value += fmt.Sprintf("`%v` - %v\n", commands[i], commandMap[commands[i]].desc)
		}
	}
	misc.MapMutex.Unlock()

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return err
		}
	}
	return nil
}

// Mod command help page
func helpStatsCommand(s *discordgo.Session, m *discordgo.Message) {
	err := helpStatsEmbed(s, m)
	if err != nil {

		misc.MapMutex.Lock()
		guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
		misc.MapMutex.Unlock()

		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
}

// Mod command help page embed
func helpStatsEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %vcommand to see a detailed description.", guildPrefix)
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the filter category
	misc.MapMutex.Lock()
	for command := range commandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if commandMap[commands[i]].category == "stats" {
			commandsField.Value += fmt.Sprintf("`%v` - %v\n", commands[i], commandMap[commands[i]].desc)
		}
	}
	misc.MapMutex.Unlock()

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return err
		}
	}
	return nil
}

// Mod command help page
func helpRaffleCommand(s *discordgo.Session, m *discordgo.Message) {
	err := helpRaffleEmbed(s, m)
	if err != nil {

		misc.MapMutex.Lock()
		guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
		misc.MapMutex.Unlock()

		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
}

// Mod command help page embed
func helpRaffleEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %vcommand to see a detailed description.", guildPrefix)
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the filter category
	misc.MapMutex.Lock()
	for command := range commandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if commandMap[commands[i]].category == "raffles" {
			commandsField.Value += fmt.Sprintf("`%v` - %v\n", commands[i], commandMap[commands[i]].desc)
		}
	}
	misc.MapMutex.Unlock()

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return err
		}
	}
	return nil
}

// Mod command help page
func helpWaifuCommand(s *discordgo.Session, m *discordgo.Message) {

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	guildWaifuModule := misc.GuildMap[m.GuildID].GuildConfig.WaifuModule
	misc.MapMutex.Unlock()

	// Checks if waifus are disabled in the guild
	if !guildWaifuModule {
		return
	}

	err := helpWaifuEmbed(s, m)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
}

// Mod command help page embed
func helpWaifuEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %vcommand to see a detailed description.", guildPrefix)
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the waifus category
	misc.MapMutex.Lock()
	for command := range commandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if commandMap[commands[i]].category == "waifus" {
			commandsField.Value += fmt.Sprintf("`%v` - %v\n", commands[i], commandMap[commands[i]].desc)
		}
	}
	misc.MapMutex.Unlock()

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return err
		}
	}
	return nil
}

// Mod command help page
func helpGuildSettingsCommand(s *discordgo.Session, m *discordgo.Message) {
	err := helpGuildSettingsEmbed(s, m)
	if err != nil {

		misc.MapMutex.Lock()
		guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
		misc.MapMutex.Unlock()

		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
}

// Mod command help page embed
func helpGuildSettingsEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %vcommand to see a detailed description.", guildPrefix)
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the waifus category
	misc.MapMutex.Lock()
	for command := range commandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if commandMap[commands[i]].category == "settings" {
			commandsField.Value += fmt.Sprintf("`%v` - %v\n", commands[i], commandMap[commands[i]].desc)
		}
	}
	misc.MapMutex.Unlock()

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return err
		}
	}
	return nil
}

// Split a help embed into multiple parts
func splitHelpEmbedField(embed *discordgo.MessageEmbed, elevated bool) []*discordgo.MessageEmbed {

	var (
		totalLen      int
		newEmbed      *discordgo.MessageEmbed
		newEmbeds     []*discordgo.MessageEmbed
		newEmbedField *discordgo.MessageEmbedField
		newFooter	  *discordgo.MessageEmbedFooter
		targetIndex 	int
	)

	// Set the proper field index
	if !elevated {
		targetIndex = 2
	}

	// Split off all commands and calculate their total char length
	commands := strings.Split(embed.Fields[targetIndex].Value, "\n")
	for _, command := range commands {
		totalLen += len(command)
	}

	// If they're over 1024 chars  then split them into more embeds
	for totalLen > 1024 {

		newEmbedField = nil
		newEmbedField = new(discordgo.MessageEmbedField)
		newEmbedField.Name = embed.Fields[targetIndex].Name
		for i := 0; i < len(commands)/2; i++ {
			newEmbedField.Value += fmt.Sprintf("%v\n", commands[i])
			totalLen -= len(commands[i])
		}
		newEmbedField.Inline = embed.Fields[targetIndex].Inline

		newEmbed = nil
		newEmbed = new(discordgo.MessageEmbed)
		newEmbed.Fields = append(newEmbed.Fields, newEmbedField)
		newEmbed.Color = embed.Color
		newEmbeds = append(newEmbeds, newEmbed)

		newEmbedField = nil
		newEmbedField = new(discordgo.MessageEmbedField)
		newEmbedField.Name = embed.Fields[targetIndex].Name
		for i := len(commands)/2; i < len(commands); i++ {
			newEmbedField.Value += fmt.Sprintf("%v\n", commands[i])
			totalLen -= len(commands[i])
		}
		newEmbedField.Inline = embed.Fields[targetIndex].Inline

		newEmbed = nil
		newEmbed = new(discordgo.MessageEmbed)
		newEmbed.Fields = append(newEmbed.Fields, newEmbedField)
		newEmbed.Color = embed.Color
		newEmbeds = append(newEmbeds, newEmbed)
	}

	// Set up the footer for the last embed and also the user and permission level for the first embed
	if len(newEmbeds) != 0 {
		newFooter = new(discordgo.MessageEmbedFooter)
		newFooter.Text = embed.Footer.Text
		newFooter.ProxyIconURL = embed.Footer.ProxyIconURL
		newFooter.IconURL = embed.Footer.IconURL
		newEmbeds[len(newEmbeds)-1].Footer = newFooter

		// Move the fields dynamically to add user and permission level to the first embed
		if !elevated {
			// User
			newEmbedField = nil
			newEmbedField = new(discordgo.MessageEmbedField)
			newEmbedField.Name = embed.Fields[0].Name
			newEmbedField.Value = embed.Fields[0].Value
			newEmbedField.Inline = embed.Fields[0].Inline
			newEmbeds[0].Fields = append(newEmbeds[0].Fields, newEmbedField)

			// Permission level
			newEmbedField = nil
			newEmbedField = new(discordgo.MessageEmbedField)
			newEmbedField.Name = embed.Fields[1].Name
			newEmbedField.Value = embed.Fields[1].Value
			newEmbedField.Inline = embed.Fields[1].Inline
			newEmbeds[0].Fields = append(newEmbeds[0].Fields, newEmbedField)

			// Save and remove the commands field and then readd it at the end
			newEmbedField = nil
			newEmbedField = new(discordgo.MessageEmbedField)
			newEmbedField = newEmbeds[0].Fields[0]
			newEmbeds[0].Fields = append(newEmbeds[0].Fields[:0], newEmbeds[0].Fields[0+1:]...)
			newEmbeds[0].Fields = append(newEmbeds[0].Fields, newEmbedField)
		}
	}

	if len(newEmbeds) == 0 {
		newEmbeds = append(newEmbeds, embed)
	}

	return newEmbeds
}

func init() {
	add(&command{
		execute:  helpEmbedCommand,
		trigger:  "help",
		aliases:  []string{"h"},
		desc:     "Print all available commands in embed form.",
		category: "normal",
		DMAble: true,
	})
	//add(&command{
	//	execute:  helpPlaintextCommand,
	//	trigger:  "helpplain",
	//	desc:     "Prints all available commands in plain text.",
	//	category: "normal",
	//})
	add(&command{
		execute:  helpChannelCommand,
		trigger:  "h-channel",
		aliases:  []string{"h[channel]", "hchannels", "h[channels]", "h-chanel", "help-channel", "help-chanel", "hchannel", "h-channels", "help-channels"},
		desc:     "Print all channel related commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpFiltersCommand,
		trigger:  "h-filters",
		aliases:  []string{"h[filters]", "hfilter", "h[filters]", "h-filter", "help-filters", "help-filter", "hfilters"},
		desc:     "Print all commands related to filters.",
		elevated: true,
	})
	add(&command{
		execute:  helpMiscCommand,
		trigger:  "h-misc",
		aliases:  []string{"h[misc]", "hmiscellaneous", "h[miscellaneous]", "help-misc", "hmisc"},
		desc:     "Print all miscellaneous mod commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpNormalCommand,
		trigger:  "h-normal",
		aliases:  []string{"h[normal]", "h-norma", "h-norm", "help-normal", "hnormal"},
		desc:     "Print all normal user commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpPunishmentCommand,
		trigger:  "h-punishment",
		aliases:  []string{"h[punishment]", "hpunishments", "h[punishments]", "h-punish", "h-pun", "help-punishment", "help-punishments", "h-punishments", "hpunishment"},
		desc:     "Print all mod punishment commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpReactsCommand,
		trigger:  "h-reacts",
		aliases:  []string{"helpreacts", "helpreacts", "hreact", "h-react", "help-reacts", "help-react", "hreacts"},
		desc:     "Print all react mod commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpRssCommand,
		trigger:  "h-rss",
		aliases:  []string{"h[rss]", "help-rss", "hrss"},
		desc:     "Print all RSS feed from sub commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpStatsCommand,
		trigger:  "h-stats",
		aliases:  []string{"h[stats]", "hstat", "h[stat]", "help-stats", "hstats", "h-stats", "help-stats"},
		desc:     "Print all channel and emoji stats commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpRaffleCommand,
		trigger:  "h-raffles",
		aliases:  []string{"h[raffle]", "hraffles", "h[raffles]", "help-raffle", "help-raffles", "h-raffle", "hraffle"},
		desc:     "Print all raffle commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpWaifuCommand,
		trigger:  "h-waifu",
		aliases:  []string{"h[waifu]", "hwaifus", "h[waifus]", "help-waifu", "help-waifus", "h-waifus", "hwaifu"},
		desc:     "Print all waifu commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpGuildSettingsCommand,
		trigger:  "h-settings",
		aliases:  []string{"h[set]", "hsetting", "h[setting]", "h[settings]", "hset", "hsets", "hsetts", "hsett", "h-set", "help-settings", "help-set", "hsettings"},
		desc:     "Print all server setting commands.",
		elevated: true,
		admin:    true,
	})

	misc.MapMutex.Lock()
	categoriesMap["Channel"] = "Mod channel-related commands."
	categoriesMap["Filters"] = "Word and emoji filters."
	categoriesMap["Misc"] = "Miscellaneous mod commands."
	categoriesMap["Normal"] = "Normal user commands."
	categoriesMap["Punishment"] = "Warnings, kicks and bans."
	categoriesMap["Reacts"] = "Channel join via react commands."
	categoriesMap["Rss"] = "Reddit RSS feed commands."
	categoriesMap["Stats"] = "Channel and emoji stats."
	categoriesMap["Raffles"] = "Raffle commands."
	categoriesMap["Waifus"] = "Waifu commands."
	categoriesMap["Settings"] = "Server setting commands."
	misc.MapMutex.Unlock()
}
