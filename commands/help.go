package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/functionality"
)

// Command categories in sorted form and map form(map for descriptions)
var (
	categoriesSorted = [...]string{"Autopost", "Channel", "Filters", "Misc", "Normal", "Moderation", "Reacts", "Reddit", "Stats", "Raffles", "Waifus", "Settings"}
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

	guildSettings := db.GetGuildSettings(m.GuildID)

	// Checks for mod perms and handles accordingly
	if functionality.HasElevatedPermissions(s, m.Author.ID, m.GuildID) {
		elevated = true
	}

	// Check perms
	mem, err := s.State.Member(m.GuildID, m.Author.ID)
	if err != nil {
		mem, err = s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			return
		}
	}
	admin, err = functionality.MemberIsAdmin(s, m.GuildID, mem, discordgo.PermissionAdministrator)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
	}

	err = helpEmbed(s, m, elevated, admin)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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

		guildSettings = entities.GuildSettings{Prefix: "."}
	)

	if m.GuildID != "" {
		guildSettings = db.GetGuildSettings(m.GuildID)
	}

	// Set embed color
	embedMess.Color = 16758465

	// Sets user field
	user.Name = "Username:"
	user.Value = m.Author.Mention()

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

	// Sets usage field if elevated
	if elevated {
		// Sets footer field
		embedFooter.Text = fmt.Sprintf("Usage: %vh-category | Example: %vh-settings", guildSettings.GetPrefix(), guildSettings.GetPrefix())
		embedMess.Footer = &embedFooter
	}

	if !elevated {
		// Sets commands field
		userCommands.Name = "**Command:**"
		userCommands.Inline = true

		// Iterates through non-mod commands and adds them to the embed sorted
		for command := range CommandMap {
			commands = append(commands, command)
		}
		sort.Strings(commands)
		for i := 0; i < len(commands); i++ {
			if m.GuildID == "" {
				if !CommandMap[commands[i]].DMAble {
					continue
				}
			}
			if CommandMap[commands[i]].Permission == functionality.User {
				if CommandMap[commands[i]].Module == "waifus" {
					if !guildSettings.GetWaifuModule() {
						continue
					}
				}
				if CommandMap[commands[i]].Trigger == "startvote" {
					if !guildSettings.GetVoteModule() {
						continue
					}
				}
				userCommands.Value += fmt.Sprintf("`%v` - %v\n", commands[i], CommandMap[commands[i]].Desc)
			}
		}

		// Sets footer field
		embedFooter.Text = fmt.Sprintf("Tip: Type %v<command> to see a detailed description", guildSettings.GetPrefix())
		embedMess.Footer = &embedFooter
	} else {
		// Sets elevated commands field
		adminCategories.Name = "Categories:"
		adminCategories.Inline = true

		// Iterates through categories and their descriptions and adds them to the embed. Special behavior for waifus and reacts and settings based on settings
		for i := 0; i < len(categoriesSorted); i++ {
			if categoriesSorted[i] == "Waifus" {
				if !guildSettings.GetWaifuModule() {
					continue
				}
			}
			if categoriesSorted[i] == "Reacts" {
				if !guildSettings.GetReactsModule() {
					continue
				}
			}
			if categoriesSorted[i] == "Settings" {
				if !admin && m.Author.ID != config.OwnerID {
					continue
				}
			}
			adminCategories.Value += fmt.Sprintf("**%v** - %v\n", categoriesSorted[i], categoriesMap[categoriesSorted[i]])
		}
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
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return err
		}
	}
	return nil
}

// Mod command help page
func helpChannelCommand(s *discordgo.Session, m *discordgo.Message) {
	err := helpChannelEmbed(s, m)
	if err != nil {
		guildSettings := db.GetGuildSettings(m.GuildID)
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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

	guildSettings := db.GetGuildSettings(m.GuildID)

	// Set embed color
	embedMess.Color = 16758465

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %vcommand to see a detailed description", guildSettings.GetPrefix())
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the filter category
	for command := range CommandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if CommandMap[commands[i]].Module == "channel" {
			commandsField.Value += fmt.Sprintf("`%v` - %v\n", commands[i], CommandMap[commands[i]].Desc)
		}
	}

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return err
		}
	}
	return nil
}

// Mod command help page
func helpFiltersCommand(s *discordgo.Session, m *discordgo.Message) {
	err := helpFiltersEmbed(s, m)
	if err != nil {
		guildSettings := db.GetGuildSettings(m.GuildID)
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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

	guildSettings := db.GetGuildSettings(m.GuildID)

	// Set embed color
	embedMess.Color = 16758465

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %scommand to see a detailed description", guildSettings.GetPrefix())
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the filter category
	for command := range CommandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if CommandMap[commands[i]].Module == "filters" {
			commandsField.Value += fmt.Sprintf("`%s` - %s\n", commands[i], CommandMap[commands[i]].Desc)
		}
	}

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return err
		}
	}
	return nil
}

// Mod command help page
func helpMiscCommand(s *discordgo.Session, m *discordgo.Message) {
	err := helpMiscEmbed(s, m)
	if err != nil {
		guildSettings := db.GetGuildSettings(m.GuildID)
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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

	guildSettings := db.GetGuildSettings(m.GuildID)

	// Set embed color
	embedMess.Color = 16758465

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %scommand to see a detailed description", guildSettings.GetPrefix())
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the filter category
	for command := range CommandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if CommandMap[commands[i]].Module == "misc" {
			commandsField.Value += fmt.Sprintf("`%s` - %s\n", commands[i], CommandMap[commands[i]].Desc)
		}
	}

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return err
		}
	}
	return nil
}

// Mod command help page
func helpNormalCommand(s *discordgo.Session, m *discordgo.Message) {
	err := helpNormalEmbed(s, m)
	if err != nil {
		guildSettings := db.GetGuildSettings(m.GuildID)
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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

	guildSettings := db.GetGuildSettings(m.GuildID)

	// Set embed color
	embedMess.Color = 16758465

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %scommand to see a detailed description", guildSettings.GetPrefix())
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the filter category
	for command := range CommandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if CommandMap[commands[i]].Module == "normal" {
			commandsField.Value += fmt.Sprintf("`%s` - %s\n", commands[i], CommandMap[commands[i]].Desc)
		}
	}

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return err
		}
	}

	return nil
}

// Mod command help page
func helpModerationCommand(s *discordgo.Session, m *discordgo.Message) {
	err := helpModerationEmbed(s, m)
	if err != nil {
		guildSettings := db.GetGuildSettings(m.GuildID)
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// Mod command help page embed
func helpModerationEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	guildSettings := db.GetGuildSettings(m.GuildID)

	// Set embed color
	embedMess.Color = 16758465

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %scommand to see a detailed description", guildSettings.GetPrefix())
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the filter category
	for command := range CommandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if CommandMap[commands[i]].Module == "moderation" {
			commandsField.Value += fmt.Sprintf("`%s` - %s\n", commands[i], CommandMap[commands[i]].Desc)
		}
	}

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return err
		}
	}
	return nil
}

// Prints pretty help
func helpReactsCommand(s *discordgo.Session, m *discordgo.Message) {

	guildSettings := db.GetGuildSettings(m.GuildID)

	// Checks if reacts are disabled in the guild
	if !guildSettings.GetReactsModule() {
		return
	}

	err := helpReactsEmbed(s, m)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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

	guildSettings := db.GetGuildSettings(m.GuildID)

	// Set embed color
	embedMess.Color = 16758465

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %scommand to see a detailed description", guildSettings.GetPrefix())
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the filter category
	for command := range CommandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if CommandMap[commands[i]].Module == "reacts" {
			commandsField.Value += fmt.Sprintf("`%s` - %s\n", commands[i], CommandMap[commands[i]].Desc)
		}
	}

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return err
		}
	}
	return nil
}

// Mod command help page
func helpRedditCommand(s *discordgo.Session, m *discordgo.Message) {
	err := helpRedditEmbed(s, m)
	if err != nil {
		guildSettings := db.GetGuildSettings(m.GuildID)
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// Mod command help page embed
func helpRedditEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	guildSettings := db.GetGuildSettings(m.GuildID)

	// Set embed color
	embedMess.Color = 16758465

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %scommand to see a detailed description", guildSettings.GetPrefix())
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the filter category
	for command := range CommandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if CommandMap[commands[i]].Module == "reddit" {
			commandsField.Value += fmt.Sprintf("`%s` - %s\n", commands[i], CommandMap[commands[i]].Desc)
		}
	}

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return err
		}
	}
	return nil
}

// Mod command help page
func helpStatsCommand(s *discordgo.Session, m *discordgo.Message) {
	err := helpStatsEmbed(s, m)
	if err != nil {
		guildSettings := db.GetGuildSettings(m.GuildID)
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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

	guildSettings := db.GetGuildSettings(m.GuildID)

	// Set embed color
	embedMess.Color = 16758465

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %scommand to see a detailed description", guildSettings.GetPrefix())
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the filter category
	for command := range CommandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if CommandMap[commands[i]].Module == "stats" {
			commandsField.Value += fmt.Sprintf("`%s` - %s\n", commands[i], CommandMap[commands[i]].Desc)
		}
	}

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return err
		}
	}
	return nil
}

// Mod command help page
func helpRaffleCommand(s *discordgo.Session, m *discordgo.Message) {
	err := helpRaffleEmbed(s, m)
	if err != nil {
		guildSettings := db.GetGuildSettings(m.GuildID)
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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

	guildSettings := db.GetGuildSettings(m.GuildID)

	// Set embed color
	embedMess.Color = 16758465

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %scommand to see a detailed description", guildSettings.GetPrefix())
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the filter category
	for command := range CommandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if CommandMap[commands[i]].Module == "raffles" {
			commandsField.Value += fmt.Sprintf("`%s` - %s\n", commands[i], CommandMap[commands[i]].Desc)
		}
	}

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return err
		}
	}
	return nil
}

// Mod command help page
func helpWaifuCommand(s *discordgo.Session, m *discordgo.Message) {

	guildSettings := db.GetGuildSettings(m.GuildID)

	// Checks if waifus are disabled in the guild
	if !guildSettings.GetWaifuModule() {
		return
	}

	err := helpWaifuEmbed(s, m)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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

	guildSettings := db.GetGuildSettings(m.GuildID)

	// Set embed color
	embedMess.Color = 16758465

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %scommand to see a detailed description", guildSettings.GetPrefix())
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the waifus category
	for command := range CommandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if CommandMap[commands[i]].Module == "waifus" {
			commandsField.Value += fmt.Sprintf("`%s` - %s\n", commands[i], CommandMap[commands[i]].Desc)
		}
	}

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return err
		}
	}
	return nil
}

// Mod command help page
func helpAutopostCommand(s *discordgo.Session, m *discordgo.Message) {
	err := helpAutopostEmbed(s, m)
	if err != nil {
		guildSettings := db.GetGuildSettings(m.GuildID)
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// Mod command help page embed
func helpAutopostEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	guildSettings := db.GetGuildSettings(m.GuildID)

	// Set embed color
	embedMess.Color = 16758465

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %scommand to see a detailed description", guildSettings.GetPrefix())
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the waifus category
	for command := range CommandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if CommandMap[commands[i]].Module == "autopost" {
			commandsField.Value += fmt.Sprintf("`%s` - %s\n", commands[i], CommandMap[commands[i]].Desc)
		}
	}

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return err
		}
	}
	return nil
}

// Mod command help page
func helpGuildSettingsCommand(s *discordgo.Session, m *discordgo.Message) {
	err := helpGuildSettingsEmbed(s, m)
	if err != nil {
		guildSettings := db.GetGuildSettings(m.GuildID)
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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

	guildSettings := db.GetGuildSettings(m.GuildID)

	// Set embed color
	embedMess.Color = 16758465

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %scommand to see a detailed description", guildSettings.GetPrefix())
	embedMess.Footer = &embedFooter

	// Sets command field
	commandsField.Name = "Command:"
	commandsField.Inline = true

	// Iterates through commands in the waifus category
	for command := range CommandMap {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for i := 0; i < len(commands); i++ {
		if CommandMap[commands[i]].Module == "settings" {
			commandsField.Value += fmt.Sprintf("`%s` - %s\n", commands[i], CommandMap[commands[i]].Desc)
		}
	}

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed
	embeds := splitHelpEmbedField(&embedMess, true)

	for _, splitEmbed := range embeds {
		// Sends embed in channel
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, splitEmbed)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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
		newFooter     *discordgo.MessageEmbedFooter
		targetIndex   int
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
		for i := len(commands) / 2; i < len(commands); i++ {
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
			// Username
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
	Add(&Command{
		Execute: helpEmbedCommand,
		Trigger: "help",
		Aliases: []string{"h"},
		Desc:    "Print all commands available to you",
		Module:  "normal",
		DMAble:  true,
	})
	Add(&Command{
		Execute:    helpChannelCommand,
		Trigger:    "h-channel",
		Aliases:    []string{"h[channel]", "hchannels", "h[channels]", "h-chanel", "help-channel", "help-chanel", "hchannel", "h-channels", "help-channels", "channel"},
		Desc:       "Print all channel related commands",
		Permission: functionality.Mod,
	})
	Add(&Command{
		Execute:    helpFiltersCommand,
		Trigger:    "h-filters",
		Aliases:    []string{"h[filters]", "hfilter", "h[filters]", "h-filter", "help-filters", "help-filter", "hfilters"},
		Desc:       "Print all commands related to filters",
		Permission: functionality.Mod,
	})
	Add(&Command{
		Execute:    helpMiscCommand,
		Trigger:    "h-misc",
		Aliases:    []string{"h[misc]", "hmiscellaneous", "h[miscellaneous]", "help-misc", "hmisc", "misc"},
		Desc:       "Print all miscellaneous mod commands",
		Permission: functionality.Mod,
	})
	Add(&Command{
		Execute:    helpNormalCommand,
		Trigger:    "h-normal",
		Aliases:    []string{"h[normal]", "h-norma", "h-norm", "help-normal", "hnormal", "normal"},
		Desc:       "Print all normal user commands",
		Permission: functionality.Mod,
	})
	Add(&Command{
		Execute:    helpModerationCommand,
		Trigger:    "h-moderation",
		Aliases:    []string{"h[moderation]", "hmoderation", "h-mod", "h-mode", "help-moderation", "moderation"},
		Desc:       "Print all mod moderation commands",
		Permission: functionality.Mod,
	})
	Add(&Command{
		Execute:    helpReactsCommand,
		Trigger:    "h-reacts",
		Aliases:    []string{"helpreacts", "helpreacts", "hreact", "h-react", "help-reacts", "help-react", "hreacts"},
		Desc:       "Print all react mod commands",
		Permission: functionality.Mod,
	})
	Add(&Command{
		Execute:    helpRedditCommand,
		Trigger:    "h-reddit",
		Aliases:    []string{"h[reddit]", "help-reddit", "hreddit", "reddit"},
		Desc:       "Print all Reddit feed commands",
		Permission: functionality.Mod,
	})
	Add(&Command{
		Execute:    helpStatsCommand,
		Trigger:    "h-stats",
		Aliases:    []string{"h[stats]", "hstat", "h[stat]", "help-stats", "hstats", "h-stats", "help-stats"},
		Desc:       "Print all channel & emoji stat commands",
		Permission: functionality.Mod,
	})
	Add(&Command{
		Execute:    helpRaffleCommand,
		Trigger:    "h-raffles",
		Aliases:    []string{"h[raffle]", "hraffles", "h[raffles]", "help-raffle", "help-raffles", "h-raffle", "hraffle", "raffle"},
		Desc:       "Print all raffle commands",
		Permission: functionality.Mod,
	})
	Add(&Command{
		Execute:    helpWaifuCommand,
		Trigger:    "h-waifu",
		Aliases:    []string{"h[waifu]", "hwaifus", "h[waifus]", "help-waifu", "help-waifus", "h-waifus", "hwaifu"},
		Desc:       "Print all waifu commands",
		Permission: functionality.Mod,
	})
	Add(&Command{
		Execute:    helpAutopostCommand,
		Trigger:    "h-autopost",
		Aliases:    []string{"h[autopost]", "hautopost", "h[auto]", "h[autoposts]", "hautopost", "hautoposts", "hautos", "hauto", "h-autopost", "help-autopost", "help-auto", "h-autos", "autopost"},
		Desc:       "Print all autopost commands",
		Permission: functionality.Mod,
	})
	Add(&Command{
		Execute:    helpGuildSettingsCommand,
		Trigger:    "h-settings",
		Aliases:    []string{"h[set]", "hsetting", "h[setting]", "h[settings]", "hset", "hsets", "hsetts", "hsett", "h-set", "help-settings", "help-set", "hsettings", "settings"},
		Desc:       "Print all server setting commands",
		Permission: functionality.Admin,
	})

	categoriesMap["Channel"] = "Mod channel-related commands"
	categoriesMap["Filters"] = "Phrase, extension and emoji filters"
	categoriesMap["Misc"] = "Miscellaneous Mod commands"
	categoriesMap["Normal"] = "Normal Username commands"
	categoriesMap["Moderation"] = "Moderation commands"
	categoriesMap["Reacts"] = "React Autorole commands"
	categoriesMap["Reddit"] = "Reddit Feed commands"
	categoriesMap["Stats"] = "Channel & Emoji Stats commands"
	categoriesMap["Raffles"] = "Raffle commands"
	categoriesMap["Waifus"] = "Waifu commands"
	categoriesMap["Autopost"] = "Autopost commands"
	categoriesMap["Settings"] = "Server setting commands"
}
