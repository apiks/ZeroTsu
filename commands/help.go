package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/functionality"
)

// Command categories in sorted form and map form(map for descriptions)
var (
	categoriesSorted = [...]string{"Anime", "Misc", "Normal", "Reacts", "Reddit", "Raffles", "Settings"}
	categoriesMap    = make(map[string]string)
)

// helpEmbedCommand prints help command
func helpEmbedCommand(s *discordgo.Session, guildID string, author *discordgo.User, DM bool) []*discordgo.MessageEmbed {
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

		elevated      bool
		admin         bool
		guildSettings entities.GuildSettings
	)

	if !DM {
		guildSettings = db.GetGuildSettings(guildID)

		// Checks for mod perms and handles accordingly
		if functionality.HasElevatedPermissions(s, author.ID, guildID) {
			elevated = true
		}

		// Check perms
		mem, err := s.State.Member(guildID, author.ID)
		if err != nil {
			mem, err = s.GuildMember(guildID, author.ID)
			if err != nil {
				return nil
			}
		}
		admin, err = functionality.MemberIsAdmin(s, guildID, mem, discordgo.PermissionAdministrator)
		if err != nil {
			return nil
		}
	}

	// Set embed color
	embedMess.Color = 16758465

	// Sets user field
	user.Name = "Username:"
	user.Value = author.Mention()

	// Sets permission field
	permission.Name = "Permission Level:"
	if author.ID == config.OwnerID {
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
		embedFooter.Text = "Usage: /h-category | Example: /h-settings"
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
			if guildID == "" {
				if !CommandMap[commands[i]].DMAble {
					continue
				}
			}
			if CommandMap[commands[i]].Permission == functionality.User {
				userCommands.Value += fmt.Sprintf("`%v` - %v\n", commands[i], CommandMap[commands[i]].Desc)
			}
		}

		// Sets footer field
		embedFooter.Text = "Tip: Type /<command> to see a detailed description"
		embedMess.Footer = &embedFooter
	} else {
		// Sets elevated commands field
		adminCategories.Name = "Categories:"
		adminCategories.Inline = true

		// Iterates through categories and their descriptions and adds them to the embed. Special behavior for waifus and reacts and settings based on settings
		for i := 0; i < len(categoriesSorted); i++ {
			if categoriesSorted[i] == "Reacts" {
				if !guildSettings.GetReactsModule() {
					continue
				}
			}
			if categoriesSorted[i] == "Settings" {
				if !admin && author.ID != config.OwnerID {
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

	return splitHelpEmbedField(&embedMess, elevated)
}

// helpEmbedCommandHandler prints help command
func helpEmbedCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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

// helpEmbed sends embed message for general all-purpose help
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

// helpMiscEmbedCommand sends misc command help page
func helpMiscEmbedCommand() []*discordgo.MessageEmbed {
	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	// Set embed color
	embedMess.Color = 16758465

	// Sets footer field
	embedFooter.Text = "Tip: Type /<command> to see a detailed description"
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

	return splitHelpEmbedField(&embedMess, true)
}

// helpMiscCommandHandler sends misc command help page
func helpMiscCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	err := helpMiscEmbed(s, m)
	if err != nil {
		guildSettings := db.GetGuildSettings(m.GuildID)
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// helpMiscEmbed misc command help page embed
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
	embedFooter.Text = fmt.Sprintf("Tip: Type %s<command> to see a detailed description", guildSettings.GetPrefix())
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

// helpNormalEmbedCommand sends normal command help page
func helpNormalEmbedCommand() []*discordgo.MessageEmbed {
	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	// Set embed color
	embedMess.Color = 16758465

	// Sets footer field
	embedFooter.Text = "Tip: Type /<command> to see a detailed description"
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

	return splitHelpEmbedField(&embedMess, true)
}

// helpNormalCommandHandler sends normal command help page
func helpNormalCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	err := helpNormalEmbed(s, m)
	if err != nil {
		guildSettings := db.GetGuildSettings(m.GuildID)
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// helpNormalEmbed normal command help page embed
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
	embedFooter.Text = fmt.Sprintf("Tip: Type %s<command> to see a detailed description", guildSettings.GetPrefix())
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

// helpReactsEmbedCommand prints reacts help command
func helpReactsEmbedCommand() []*discordgo.MessageEmbed {
	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	// Set embed color
	embedMess.Color = 16758465

	// Sets footer field
	embedFooter.Text = "Tip: Type /<command> to see a detailed description"
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

	return splitHelpEmbedField(&embedMess, true)
}

// helpReactsCommandHandler sends reacts command help
func helpReactsCommandHandler(s *discordgo.Session, m *discordgo.Message) {

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

// helpReactsEmbed reacts command help page embed
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
	embedFooter.Text = fmt.Sprintf("Tip: Type %s<command> to see a detailed description", guildSettings.GetPrefix())
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

// helpRedditEmbedCommand prints reddit help command
func helpRedditEmbedCommand() []*discordgo.MessageEmbed {
	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	// Set embed color
	embedMess.Color = 16758465

	// Sets footer field
	embedFooter.Text = "Tip: Type /<command> to see a detailed description"
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
	return splitHelpEmbedField(&embedMess, true)
}

// helpRedditCommandHandler reddit command help page
func helpRedditCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	err := helpRedditEmbed(s, m)
	if err != nil {
		guildSettings := db.GetGuildSettings(m.GuildID)
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// helpRedditEmbed sends reddit command help page embed
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
	embedFooter.Text = fmt.Sprintf("Tip: Type %s<command> to see a detailed description", guildSettings.GetPrefix())
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

// helpRaffleEmbedCommand prints raffles help command
func helpRaffleEmbedCommand() []*discordgo.MessageEmbed {
	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	// Set embed color
	embedMess.Color = 16758465

	// Sets footer field
	embedFooter.Text = "Tip: Type /<command> to see a detailed description"
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

	return splitHelpEmbedField(&embedMess, true)
}

// helpRaffleCommandHandler sends raffle command help page
func helpRaffleCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	err := helpRaffleEmbed(s, m)
	if err != nil {
		guildSettings := db.GetGuildSettings(m.GuildID)
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// helpRaffleEmbed sends raffle command help page embed
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
	embedFooter.Text = fmt.Sprintf("Tip: Type %s<command> to see a detailed description", guildSettings.GetPrefix())
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

// helpAutopostEmbedCommand prints autopost help command
func helpAutopostEmbedCommand() []*discordgo.MessageEmbed {
	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	// Set embed color
	embedMess.Color = 16758465

	// Sets footer field
	embedFooter.Text = "Tip: Type /<command> to see a detailed description"
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

	return splitHelpEmbedField(&embedMess, true)
}

// helpAutopostCommandHandler sends autopost command help page
func helpAutopostCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	err := helpAutopostEmbed(s, m)
	if err != nil {
		guildSettings := db.GetGuildSettings(m.GuildID)
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// helpAutopostEmbed sends autopost command help page embed
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
	embedFooter.Text = fmt.Sprintf("Tip: Type %s<command> to see a detailed description", guildSettings.GetPrefix())
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

// helpGuildSettingsEmbedCommand prints guild settings help command
func helpGuildSettingsEmbedCommand() []*discordgo.MessageEmbed {
	var (
		embedMess   discordgo.MessageEmbed
		embedFooter discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed         []*discordgo.MessageEmbedField
		commandsField discordgo.MessageEmbedField

		// Slice for sorting
		commands []string
	)

	// Set embed color
	embedMess.Color = 16758465

	// Sets footer field
	embedFooter.Text = "Tip: Type /<command> to see a detailed description"
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

	return splitHelpEmbedField(&embedMess, true)
}

// helpGuildSettingsCommandHandler sends guild settings command help page
func helpGuildSettingsCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	err := helpGuildSettingsEmbed(s, m)
	if err != nil {
		guildSettings := db.GetGuildSettings(m.GuildID)
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// helpGuildSettingsEmbed sends guild settings command help page embed
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
	embedFooter.Text = fmt.Sprintf("Tip: Type %s<command> to see a detailed description", guildSettings.GetPrefix())
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

// splitHelpEmbedField splits a help embed into multiple sendable parts
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
		Execute: helpEmbedCommandHandler,
		Name:    "help",
		Aliases: []string{"h"},
		Desc:    "Print all commands available to you",
		Module:  "normal",
		DMAble:  true,
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var (
				user *discordgo.User
				dm   bool
			)
			emptyContent := ""
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			if i.Member == nil {
				dm = true
				user = i.User
			} else {
				user = i.Member.User
			}

			embedsResp := helpEmbedCommand(s, i.GuildID, user, dm)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &emptyContent,
				Embeds:  &embedsResp,
			})
		},
	})
	Add(&Command{
		Execute:    helpMiscCommandHandler,
		Name:       "help-misc",
		Aliases:    []string{"h-misc, h[misc]", "hmiscellaneous", "h[miscellaneous]", "hmisc", "misc"},
		Desc:       "Print all miscellaneous mod commands",
		Permission: functionality.Mod,
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			emptyContent := ""
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			err := VerifySlashCommand(s, "help-misc", i)
			if err != nil {
				errStr := err.Error()
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &errStr,
				})
				return
			}

			embedsResp := helpMiscEmbedCommand()
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &emptyContent,
				Embeds:  &embedsResp,
			})
		},
	})
	Add(&Command{
		Execute:    helpNormalCommandHandler,
		Name:       "help-normal",
		Aliases:    []string{"h-normal", "h[normal]", "h-norma", "h-norm", "hnormal", "normal"},
		Desc:       "Print all normal user commands",
		Permission: functionality.Mod,
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			emptyContent := ""
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			err := VerifySlashCommand(s, "help-normal", i)
			if err != nil {
				errStr := err.Error()
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &errStr,
				})
				return
			}

			embedsResp := helpNormalEmbedCommand()
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &emptyContent,
				Embeds:  &embedsResp,
			})
		},
	})
	Add(&Command{
		Execute:    helpReactsCommandHandler,
		Name:       "help-reacts",
		Aliases:    []string{"h-reacts", "helpreacts", "helpreacts", "hreact", "h-react", "help-react", "hreacts"},
		Desc:       "Print all react mod commands",
		Permission: functionality.Mod,
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			emptyContent := ""
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			err := VerifySlashCommand(s, "help-reacts", i)
			if err != nil {
				errStr := err.Error()
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &errStr,
				})
				return
			}

			embedsResp := helpReactsEmbedCommand()
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &emptyContent,
				Embeds:  &embedsResp,
			})
		},
	})
	Add(&Command{
		Execute:    helpRedditCommandHandler,
		Name:       "help-reddit",
		Aliases:    []string{"h-reddit", "h[reddit]", "hreddit", "reddit"},
		Desc:       "Print all Reddit feed commands",
		Permission: functionality.Mod,
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			emptyContent := ""
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			err := VerifySlashCommand(s, "help-reddit", i)
			if err != nil {
				errStr := err.Error()
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &errStr,
				})
				return
			}

			embedsResp := helpRedditEmbedCommand()
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &emptyContent,
				Embeds:  &embedsResp,
			})
		},
	})
	Add(&Command{
		Execute:    helpRaffleCommandHandler,
		Name:       "help-raffles",
		Aliases:    []string{"h-raffles", "h[raffle]", "hraffles", "h[raffles]", "help-raffle", "h-raffle", "hraffle", "raffle"},
		Desc:       "Print all raffle commands",
		Permission: functionality.Mod,
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			emptyContent := ""
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			err := VerifySlashCommand(s, "help-raffles", i)
			if err != nil {
				errStr := err.Error()
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &errStr,
				})
				return
			}

			embedsResp := helpRaffleEmbedCommand()
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &emptyContent,
				Embeds:  &embedsResp,
			})
		},
	})
	Add(&Command{
		Execute:    helpAutopostCommandHandler,
		Name:       "help-anime",
		Aliases:    []string{"h-anime", "h[anime]", "hanime", "h[animes]", "hanimes", "h-anime", "help-animes", "anime"},
		Desc:       "Print all anime commands",
		Permission: functionality.Mod,
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			emptyContent := ""
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			err := VerifySlashCommand(s, "help-anime", i)
			if err != nil {
				errStr := err.Error()
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &errStr,
				})
				return
			}

			embedsResp := helpAutopostEmbedCommand()
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &emptyContent,
				Embeds:  &embedsResp,
			})
		},
	})
	Add(&Command{
		Execute:    helpGuildSettingsCommandHandler,
		Name:       "help-settings",
		Aliases:    []string{"h-settings", "h[set]", "hsetting", "h[setting]", "h[settings]", "hset", "hsets", "hsetts", "hsett", "h-set", "help-set", "hsettings", "settings"},
		Desc:       "Print all server setting commands",
		Permission: functionality.Admin,
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			emptyContent := ""
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			err := VerifySlashCommand(s, "help-settings", i)
			if err != nil {
				errStr := err.Error()
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &errStr,
				})
				return
			}

			embedsResp := helpGuildSettingsEmbedCommand()
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &emptyContent,
				Embeds:  &embedsResp,
			})
		},
	})

	categoriesMap["Misc"] = "Miscellaneous Mod commands"
	categoriesMap["Normal"] = "Normal User commands"
	categoriesMap["Reacts"] = "React Autorole commands"
	categoriesMap["Reddit"] = "Reddit Feed commands"
	categoriesMap["Raffles"] = "Raffle commands"
	categoriesMap["Anime"] = "Anime commands"
	categoriesMap["Settings"] = "Server setting commands"
}
