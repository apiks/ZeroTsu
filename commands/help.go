package commands

import (
	"fmt"
	"sort"

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
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	guildWaifuModule := misc.GuildMap[m.GuildID].GuildConfig.WaifuModule
	guildReactsModule := misc.GuildMap[m.GuildID].GuildConfig.ReactsModule
	misc.MapMutex.Unlock()

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
		embedFooter.Text = fmt.Sprintf("Usage: Pick a category with %vhcategory", guildPrefix)
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
			if !commandMap[commands[i]].elevated {
				if commandMap[commands[i]].category == "waifus" {
					if !guildWaifuModule {
						continue
					}
				}
				if config.Website == "" {
					if commandMap[commands[i]].trigger == "verify" ||
						commandMap[commands[i]].trigger == "unverify" {
						continue
					}
				}
				userCommands.Value += fmt.Sprintf("`%v` - %v\n", commands[i], commandMap[commands[i]].desc)
			}
		}
		misc.MapMutex.Unlock()

		// Sets footer field
		embedFooter.Text = fmt.Sprintf("Tip: Type %vcommand to see a detailed description.", guildPrefix)
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

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return err
	}
	return err
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

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return err
	}
	return err
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

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return err
	}
	return err
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
		if commandMap[commands[i]].category == "misc" {
			commandsField.Value += fmt.Sprintf("`%v` - %v\n", commands[i], commandMap[commands[i]].desc)
		}
	}
	misc.MapMutex.Unlock()

	// Adds the field to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &commandsField)

	// Adds everything together
	embedMess.Fields = embed

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return err
	}
	return err
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

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return err
	}
	return err
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

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return err
	}
	return err
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

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return err
	}
	return err
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

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return err
	}
	return err
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

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return err
	}
	return err
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

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return err
	}
	return err
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

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return err
	}
	return err
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

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return err
	}
	return err
}

func init() {
	add(&command{
		execute:  helpEmbedCommand,
		trigger:  "help",
		aliases:  []string{"h"},
		desc:     "Print all available commands in embed form.",
		category: "normal",
	})
	//add(&command{
	//	execute:  helpPlaintextCommand,
	//	trigger:  "helpplain",
	//	desc:     "Prints all available commands in plain text.",
	//	category: "normal",
	//})
	add(&command{
		execute:  helpChannelCommand,
		trigger:  "hchannel",
		aliases:  []string{"h[channel]", "hchannels", "h[channels]"},
		desc:     "Print all channel related commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpFiltersCommand,
		trigger:  "hfilters",
		aliases:  []string{"h[filters]", "hfilter", "h[filters]"},
		desc:     "Print all commands related to filters.",
		elevated: true,
	})
	add(&command{
		execute:  helpMiscCommand,
		trigger:  "hmisc",
		aliases:  []string{"h[misc]", "hmiscellaneous", "h[miscellaneous]"},
		desc:     "Print all miscellaneous mod commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpNormalCommand,
		trigger:  "hnormal",
		aliases:  []string{"h[normal]"},
		desc:     "Print all normal user commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpPunishmentCommand,
		trigger:  "hpunishment",
		aliases:  []string{"h[punishment]", "hpunishments", "h[punishments]"},
		desc:     "Print all mod pusnihment commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpReactsCommand,
		trigger:  "helpreacts",
		aliases:  []string{"hreact", "helpreacts", "hreacts"},
		desc:     "Print all react mod commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpRssCommand,
		trigger:  "hrss",
		aliases:  []string{"h[rss]"},
		desc:     "Print all RSS feed from sub commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpStatsCommand,
		trigger:  "hstats",
		aliases:  []string{"h[stats]", "hstat", "h[stat]"},
		desc:     "Print all channel and emoji stats commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpRaffleCommand,
		trigger:  "hraffle",
		aliases:  []string{"h[raffle]", "hraffles", "h[raffles]"},
		desc:     "Print all raffle commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpWaifuCommand,
		trigger:  "hwaifu",
		aliases:  []string{"h[waifu]", "hwaifus", "h[waifus]"},
		desc:     "Print all waifu commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpGuildSettingsCommand,
		trigger:  "hsettings",
		aliases:  []string{"h[set]", "hsetting", "h[setting]", "h[settings]", "hset", "hsets", "hsetts", "hsett"},
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
	categoriesMap["Rss"] = "RSS feed from sub commands."
	categoriesMap["Stats"] = "Channel and emoji stats."
	categoriesMap["Raffles"] = "Raffle commands."
	categoriesMap["Waifus"] = "Waifu commands."
	categoriesMap["Settings"] = "Server setting commands."
	misc.MapMutex.Unlock()
}
