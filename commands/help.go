package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"sort"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Command categories in sorted form and map form(map for descriptions)
var (
	categoriesSorted = [8]string{"Channel", "Filters", "Misc", "Normal", "Punishment", "Reacts", "Rss", "Stats"}
	categoriesMap = make(map[string]string)
)

// Prints pretty help command
func helpEmbedCommand(s *discordgo.Session, m *discordgo.Message) {

	var admin bool

	// Checks if it's within the config server
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

	// Pulls info on message author
	mem, err := s.State.Member(config.ServerID, m.Author.ID)
	if err != nil {
		mem, err = s.GuildMember(config.ServerID, m.Author.ID)
		if err != nil {
			return
		}
	}
	// Checks for mod perms and handles accordingly
	s.State.RWMutex.RLock()
	if misc.HasPermissions(mem) {
		admin = true
	}
	s.State.RWMutex.RUnlock()

	err = helpEmbed(s, m, admin)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}
}

// Embed message for general all-purpose help message
func helpEmbed(s *discordgo.Session, m *discordgo.Message, admin bool) error {

	var (
		embedMess          discordgo.MessageEmbed
		embedFooter	   	   discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed              []*discordgo.MessageEmbedField
		user               discordgo.MessageEmbedField
		permission         discordgo.MessageEmbedField
		userCommands       discordgo.MessageEmbedField
		adminCategories	   discordgo.MessageEmbedField

		// Slice for sorting
		commands		   []string
	)

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets user field
	user.Name = "User:"
	user.Value = m.Author.Mention()
	user.Inline = true

	// Sets permission field
	permission.Name = "Permission Level:"
	if admin {
		permission.Value = "_Admin_"
	} else {
		permission.Value = "_User_"
	}
	permission.Inline = true

	// Sets usage field if admin
	if admin {
		// Sets footer field
		embedFooter.Text = fmt.Sprintf("Usage: Pick a category with %vh[category]", config.BotPrefix)
		embedMess.Footer = &embedFooter
	}

	if !admin {
		// Sets user commands field
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
				userCommands.Value += fmt.Sprintf("`%v` - %v\n", commands[i], commandMap[commands[i]].desc)
			}
		}
		misc.MapMutex.Unlock()

		// Sets footer field
		embedFooter.Text = fmt.Sprintf("Tip: Type %v[command] to see a detailed description.", config.BotPrefix)
		embedMess.Footer = &embedFooter
	} else {
		// Sets user commands field
		adminCategories.Name = "Categories:"
		adminCategories.Inline = true

		// Iterates through categories and their descriptions and adds them to the embed
		misc.MapMutex.Lock()
		for i := 0; i < len(categoriesSorted); i++ {
			adminCategories.Value += fmt.Sprintf("%v - %v\n", categoriesSorted[i], categoriesMap[categoriesSorted[i]])
		}
		misc.MapMutex.Unlock()
	}

	// Adds the fields to embed slice (because embedMess.Fields requires slice input)
	embed = append(embed, &user)
	embed = append(embed, &permission)
	if admin {
		embed = append(embed, &adminCategories)
	} else {
		embed = append(embed, &userCommands)
	}

	// Adds everything together
	embedMess.Fields = embed

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return err
		}
		return err
	}
	return err
}

// Mod command help page
func helpChannelCommand(s *discordgo.Session, m *discordgo.Message) {
	// Checks if it's within the config server
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

	err = helpChannelEmbed(s, m)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}
}

// Mod command help page embed
func helpChannelEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess          discordgo.MessageEmbed
		embedFooter	   	   discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed    		   []*discordgo.MessageEmbedField
		commandsField  	   discordgo.MessageEmbedField

		// Slice for sorting
		commands		   []string
	)

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %v[command] to see a detailed description.", config.BotPrefix)
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
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return err
		}
		return err
	}
	return err
}

// Mod command help page
func helpFiltersCommand(s *discordgo.Session, m *discordgo.Message) {
	// Checks if it's within the config server
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

	err = helpFiltersEmbed(s, m)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}
}

// Mod command help page embed
func helpFiltersEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess          discordgo.MessageEmbed
		embedFooter	   	   discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed    		   []*discordgo.MessageEmbedField
		commandsField  	   discordgo.MessageEmbedField

		// Slice for sorting
		commands		   []string
	)

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %v[command] to see a detailed description.", config.BotPrefix)
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
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return err
		}
		return err
	}
	return err
}

// Mod command help page
func helpMiscCommand(s *discordgo.Session, m *discordgo.Message) {
	// Checks if it's within the config server
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

	err = helpMiscEmbed(s, m)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}
}

// Mod command help page embed
func helpMiscEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess          discordgo.MessageEmbed
		embedFooter	   	   discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed    		   []*discordgo.MessageEmbedField
		commandsField  	   discordgo.MessageEmbedField

		// Slice for sorting
		commands		   []string
	)

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %v[command] to see a detailed description.", config.BotPrefix)
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
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return err
		}
		return err
	}
	return err
}

// Mod command help page
func helpNormalCommand(s *discordgo.Session, m *discordgo.Message) {
	// Checks if it's within the config server
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

	err = helpNormalEmbed(s, m)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}
}

// Mod command help page embed
func helpNormalEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess          discordgo.MessageEmbed
		embedFooter	   	   discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed    		   []*discordgo.MessageEmbedField
		commandsField  	   discordgo.MessageEmbedField

		// Slice for sorting
		commands		   []string
	)

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %v[command] to see a detailed description.", config.BotPrefix)
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
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return err
		}
		return err
	}
	return err
}

// Mod command help page
func helpPunishmentCommand(s *discordgo.Session, m *discordgo.Message) {
	// Checks if it's within the config server
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

	err = helpPunishmentEmbed(s, m)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}
}

// Mod command help page embed
func helpPunishmentEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess          discordgo.MessageEmbed
		embedFooter	   	   discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed    		   []*discordgo.MessageEmbedField
		commandsField  	   discordgo.MessageEmbedField

		// Slice for sorting
		commands		   []string
	)

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %v[command] to see a detailed description.", config.BotPrefix)
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
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return err
		}
		return err
	}
	return err
}

// Mod command help page
func helpReactsCommand(s *discordgo.Session, m *discordgo.Message) {
	// Checks if it's within the config server
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

	err = helpReactsEmbed(s, m)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}
}

// Mod command help page embed
func helpReactsEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess          discordgo.MessageEmbed
		embedFooter	   	   discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed    		   []*discordgo.MessageEmbedField
		commandsField  	   discordgo.MessageEmbedField

		// Slice for sorting
		commands		   []string
	)

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %v[command] to see a detailed description.", config.BotPrefix)
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
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return err
		}
		return err
	}
	return err
}

// Mod command help page
func helpRssCommand(s *discordgo.Session, m *discordgo.Message) {
	// Checks if it's within the config server
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

	err = helpRssEmbed(s, m)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}
}

// Mod command help page embed
func helpRssEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess          discordgo.MessageEmbed
		embedFooter	   	   discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed    		   []*discordgo.MessageEmbedField
		commandsField  	   discordgo.MessageEmbedField

		// Slice for sorting
		commands		   []string
	)

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %v[command] to see a detailed description.", config.BotPrefix)
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
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return err
		}
		return err
	}
	return err
}

// Mod command help page
func helpStatsCommand(s *discordgo.Session, m *discordgo.Message) {
	// Checks if it's within the config server
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

	err = helpStatsEmbed(s, m)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}
}

// Mod command help page embed
func helpStatsEmbed(s *discordgo.Session, m *discordgo.Message) error {

	var (
		embedMess          discordgo.MessageEmbed
		embedFooter	   	   discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embed    		   []*discordgo.MessageEmbedField
		commandsField  	   discordgo.MessageEmbedField

		// Slice for sorting
		commands		   []string
	)

	// Set embed color
	embedMess.Color = 0x00ff00

	// Sets footer field
	embedFooter.Text = fmt.Sprintf("Tip: Type %v[command] to see a detailed description.", config.BotPrefix)
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
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return err
		}
		return err
	}
	return err
}

// Prints two versions of help depending on whether the user is a mod or not in plain text
func helpPlaintextCommand(s *discordgo.Session, m *discordgo.Message) {

	// Pulls message author
	mem, err := s.State.Member(config.ServerID, m.Author.ID)
	if err != nil {
		mem, err = s.GuildMember(config.ServerID, m.Author.ID)
		if err != nil {
			return
		}
	}

	// Checks for mod perms
	s.State.RWMutex.RLock()
	if misc.HasPermissions(mem) {

		// Help message 1 if user is a mod
		successMod := "`" + config.BotPrefix + "about` | Shows information about me. \n " +
			"`" + config.BotPrefix + "filters` | Shows all current filters. \n " +
			"`" + config.BotPrefix + "addfilter [filter]` | Adds a normal or regex word to the filter. \n " +
			"`" + config.BotPrefix + "removefilter [filter]` | Removes a word from the filter. \n " +
			"`" + config.BotPrefix + "avatar [@mention or user ID]` | Returns user avatar URL and image embed. \n " +
			"`" + config.BotPrefix + "create [name] [airing, general or temp; defaults to opt-in] [category ID] [description; must have at least one other non-name parameter]` | Creates a channel and role of the same name. \n " +
			"`" + config.BotPrefix + "emoji` | Shows emoji web. \n " +
			"`" + config.BotPrefix + "web` | Shows web. \n " +
			"`" + config.BotPrefix + "help` | Lists commands and their usage. \n " +
			"`" + config.BotPrefix + "join [channel name]` | Joins an opt-in channel. `" + config.BotPrefix + "joinchannel` works too. \n " +
			"`" + config.BotPrefix + "joke` | Prints a random joke. \n" +
			"`" + config.BotPrefix + "leave [channel name]` | Leaves an opt-in channel. `" + config.BotPrefix + "leavechannel` works too. \n " +
			"`" + config.BotPrefix + "lock` | Locks a non-mod channel. Takes a few seconds only if the channel has no custom mod permissions set. \n " +
			"`" + config.BotPrefix + "unlock` | Unlocks a non-mod channel. \n " +
			"`" + config.BotPrefix + "ping` | Returns Pong message. \n " +
			"`" + config.BotPrefix + "say OPTIONAL[channelID] [message]` | Sends a message from the bot. \n "

		_, err = s.ChannelMessageSend(m.ChannelID, successMod)
		if err != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				s.State.RWMutex.RUnlock()
				return
			}
			s.State.RWMutex.RUnlock()
			return
		}

		// Help message 2 if user is a mod
		successMod = "`" + config.BotPrefix + "edit [channelID] [messageID] [message]` | Edits a bot message with the command's set message, replacing it entirely. \n" +
			"`" + config.BotPrefix + "setreactjoin [messageID] [emote] [role]` | Sets a specific message's emote to give those reacted a role. \n " +
			"`" + config.BotPrefix + "setreactjoin [messageID] [emote] [role]` | Sets a specific message's emote to give those reacted a role. \n " +
			"`" + config.BotPrefix + "removereactjoin [messageID] OPTIONAL[emote]` | Removes the set react emote join from an entire message or only a specific emote of that message. \n " +
			"`" + config.BotPrefix + "viewreacts` | Prints out all currently set message react emote joins. \n " +
			"`" + config.BotPrefix + "viewrss` | Prints out all currently set rss thread post. \n " +
			"`" + config.BotPrefix + "setrss OPTIONAL[/u/author] [thread name]` | Set a thread name which it'll look for in /new by the author (default /u/AutoLovepon) and then post that thread in the channel this command was executed in. \n " +
			"`" + config.BotPrefix + "removerss OPTIONAL[/u/author] [thread name]` | Remove a thread name from a previously set rss command. \n " +
			"`" + config.BotPrefix + "sortcategory [category name or ID]` | Sorts all channels within given category alphabetically. \n " +
			"`" + config.BotPrefix + "sortroles` | Sorts spoiler roles created with the create command between opt-in dummy roles alphabetically. Freezes server for a few seconds. Use preferably with large batches.\n" +
			"`" + config.BotPrefix + "startvote OPTIONAL[required votes] [name] OPTIONAL[type] OPTIONAL[categoryID] + OPTIONAL[description]` | Starts a reaction vote in the channel the command is in. " +
			"Creates and sorts the channel if successful. Required votes are how many non-bot reacts are needed for channel creation(default 7). Types are airing, general, temp and optin(default)." +
			"CategoryID is what category to put the channel in and sort alphabetically. Description is the channel description but NEEDS a categoryID or type to work.\n"

		_, err = s.ChannelMessageSend(m.ChannelID, successMod)
		if err != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				s.State.RWMutex.RUnlock()
				return
			}
			s.State.RWMutex.RUnlock()
			return
		}
	} else {

		// Help message if user is not a mod
		successUser := "`" + config.BotPrefix + "about` | Shows information about me. \n " +
			"`" + config.BotPrefix + "avatar [@mention or user ID]` | Returns user avatar URL and image embed. \n " +
			"`" + config.BotPrefix + "help` | Lists commands and their usage. \n " +
			"`" + config.BotPrefix + "join [channel name]` | Joins an opt-in channel. `" + config.BotPrefix + "joinchannel` works too. \n " +
			"`" + config.BotPrefix + "joke` | Prints a random joke." +
			"`" + config.BotPrefix + "leave [channel name]` | Leaves an opt-in channel. `" + config.BotPrefix + "leavechannel` works too. \n " +
			"`" + config.BotPrefix + "startvote [channel name]` | Starts a 3-person vote for the creation of a temp spoilers channel that will be removed 3 hours after last message. \n "

		_, err = s.ChannelMessageSend(m.ChannelID, successUser)
		if err != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				s.State.RWMutex.RUnlock()
				return
			}
			s.State.RWMutex.RUnlock()
			return
		}
	}
	s.State.RWMutex.RUnlock()
}

func init() {
	add(&command{
		execute:  helpEmbedCommand,
		trigger:  "help",
		aliases:  []string{"h"},
		desc:     "Print all available commands in embed form.",
		category: "normal",
	})
	add(&command{
		execute:  helpPlaintextCommand,
		trigger:  "helpplain",
		desc:     "Prints all available commands in plain text.",
		category: "normal",
	})
	add(&command{
		execute:  helpChannelCommand,
		trigger:  "hchannel",
		aliases:  []string{"h[channel]"},
		desc:     "Print all channel related commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpFiltersCommand,
		trigger:  "hfilters",
		aliases:  []string{"h[filters]"},
		desc:     "Print all commands related to filters.",
		elevated: true,
	})
	add(&command{
		execute:  helpMiscCommand,
		trigger:  "hmisc",
		aliases:  []string{"h[misc]"},
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
		aliases:  []string{"h[punishment]"},
		desc:     "Print all mod pusnihment commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpReactsCommand,
		trigger:  "hreacts",
		aliases:  []string{"h[reacts]"},
		desc:     "Print all channel join via react commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpRssCommand,
		trigger:  "hrss",
		desc:     "Print all RSS feed from sub commands.",
		elevated: true,
	})
	add(&command{
		execute:  helpStatsCommand,
		trigger:  "hstats",
		aliases:  []string{"h[stats]"},
		desc:     "Print all channel and emoji stats commands.",
		elevated: true,
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
	misc.MapMutex.Unlock()
}