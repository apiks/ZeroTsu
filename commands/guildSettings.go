package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"github.com/r-anime/ZeroTsu/events"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Adds a role to the command role list
func addCommandRole(s *discordgo.Session, m *discordgo.Message) {
	var role entities.Role

	guildSettings := db.GetGuildSettings(m.GuildID)
	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"addocommandrole [Role ID]`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parse role for roleID
	roleID, roleName := common.RoleParser(s, commandStrings[1], m.GuildID)
	role = role.SetID(roleID)
	role = role.SetName(roleName)
	if role.GetID() == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role exists.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if the role already exists as a command role
	for _, commandRole := range guildSettings.GetCommandRoles() {
		if commandRole.GetID() == role.GetID() {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: That role is already a command role.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
	}

	// Adds the role to the guild command roles
	guildSettings = guildSettings.AppendToCommandRoles(role)
	db.SetGuildSettings(m.GuildID, guildSettings)

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Role `%s` is now a privileged role.", role.GetName()))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Removes a role from the command role list
func removeCommandRole(s *discordgo.Session, m *discordgo.Message) {
	var roleExists bool

	guildSettings := db.GetGuildSettings(m.GuildID)
	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"removecommandrole [Role ID]`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parse role for roleID
	roleID, roleName := common.RoleParser(s, commandStrings[1], m.GuildID)
	if roleID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role exists.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if that role is in the command role list
	for _, commandRole := range guildSettings.GetCommandRoles() {
		if commandRole.GetID() == roleID {
			roleExists = true
			break
		}
	}

	if !roleExists {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role in the command role list.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	for i, role := range guildSettings.GetCommandRoles() {
		if role.GetID() == roleID {
			guildSettings = guildSettings.RemoveFromCommandRoles(i)
			break
		}
	}

	db.SetGuildSettings(m.GuildID, guildSettings)

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Role `%v` has been removed from the command role list.", roleName))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Prints all command roles
func viewCommandRoles(s *discordgo.Session, m *discordgo.Message) {
	var (
		message      string
		splitMessage []string
	)

	guildSettings := db.GetGuildSettings(m.GuildID)
	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"commandroles`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	if len(guildSettings.GetCommandRoles()) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no privileged command roles.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	for _, role := range guildSettings.GetCommandRoles() {
		message += fmt.Sprintf("**Name:** `%s` | **ID:** `%s`\n", role.GetName(), role.GetID())
	}

	// Splits the message if it's over 1900 characters
	if len(message) > 1900 {
		splitMessage = common.SplitLongMessage(message)
	}

	// Prints split or unsplit
	if splitMessage == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	for i := 0; i < len(splitMessage); i++ {
		_, err := s.ChannelMessageSend(m.ChannelID, splitMessage[i])
		if err != nil {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot send commandroles message.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
		}
	}
}

// Handles prefix view or change
func prefixCommand(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)
	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", -1)

	// Displays current prefix
	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current prefix is: `%s` \n\n To change prefix please use `%sprefix [new prefix]`", guildSettings.GetPrefix(), guildSettings.GetPrefix()))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	} else if len(commandStrings) > 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: You cannot do multi-word prefixes due to technical reasons. Please try a single word prefix."))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Changes and writes new prefix to storage
	guildSettings = guildSettings.SetPrefix(commandStrings[1])
	db.SetGuildSettings(m.GuildID, guildSettings)

	events.DynamicNicknameChange(s, m.GuildID)

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New prefix is: `%s`", guildSettings.GetPrefix()))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Handles botlog view or change
func botLogCommand(s *discordgo.Session, m *discordgo.Message) {
	var (
		message     string
		guildBotLog entities.Cha
	)

	guildSettings := db.GetGuildSettings(m.GuildID)
	if guildSettings.BotLog != (entities.Cha{}) {
		guildBotLog = guildSettings.BotLog
	}

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Displays current botlog channel
	if len(commandStrings) == 1 {
		if guildSettings.BotLog == (entities.Cha{}) {
			message = fmt.Sprintf("Error: Bot Log is currently not set. Please use `%sbotlog [channel]`", guildSettings.GetPrefix())
		} else if guildSettings.BotLog.GetID() == "" {
			message = fmt.Sprintf("Error: Bot Log is currently not set. Please use `%sbotlog [channel]`", guildSettings.GetPrefix())
		} else {
			message = fmt.Sprintf("Current Bot Log is: `%s - %s` \n\nTo change Bot Log please use `%sbotlog [channel]`\nTo disable it please use `%sbotlog disable`", guildSettings.BotLog.GetName(), guildSettings.BotLog.GetID(), guildSettings.GetPrefix(), guildSettings.GetPrefix())
		}

		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parse and save the target channel
	if commandStrings[1] == "disable" {
		guildBotLog = entities.Cha{}
	} else {
		channelID, channelName := common.ChannelParser(s, commandStrings[1], m.GuildID)
		if channelID == "" && channelName == "" {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such channel")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		guildBotLog = entities.NewCha(channelName, channelID)
	}

	// Changes and writes new bot log to storage
	guildSettings = guildSettings.SetBotLog(guildBotLog)
	db.SetGuildSettings(m.GuildID, guildSettings)

	if guildBotLog == (entities.Cha{}) {
		_, err := s.ChannelMessageSend(m.ChannelID, "Success! BotLog has been disabled!")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Bot Log is: `%s - %s`", guildBotLog.GetName(), guildBotLog.GetID()))
	if err != nil {
		guildSettings.BotLog = guildBotLog
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Adds a voice channel ID with a role to the voice channel list
func addVoiceChaRole(s *discordgo.Session, m *discordgo.Message) {

	var (
		cha   entities.VoiceCha
		role  entities.Role
		merge bool
	)

	guildSettings := db.GetGuildSettings(m.GuildID)
	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 3)

	if len(commandStrings) < 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"addvoice [channel ID] [role]` \n\n")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parse channel
	chaID, chaName := common.ChannelParser(s, commandStrings[1], m.GuildID)
	cha = cha.SetID(chaID)
	cha = cha.SetName(chaName)
	if cha.GetID() == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such channel exists.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	voiceCheck, err := s.Channel(cha.GetID())
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	if voiceCheck.Type != discordgo.ChannelTypeGuildVoice {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: That is not a voice channel. Please use a voice channel.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	// Parse role
	roleID, roleName := common.RoleParser(s, commandStrings[2], m.GuildID)
	role = role.SetID(roleID)
	role = role.SetName(roleName)
	if role.GetID() == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role exists.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if the role is already set
	for i, voiceCha := range guildSettings.GetVoiceChas() {
		if voiceCha.GetID() == cha.GetID() {
			for _, roleIteration := range voiceCha.GetRoles() {
				if roleIteration.GetID() == role.GetID() {
					_, err := s.ChannelMessageSend(m.ChannelID, "Error: That role is already set to that channel.")
					if err != nil {
						common.LogError(s, guildSettings.BotLog, err)
						return
					}
					return
				}
			}
			// Adds the voice channel and role to the guild voice channels
			cha = cha.SetRoles(voiceCha.GetRoles())
			cha = cha.AppendToRoles(role)
			voiceChas := guildSettings.GetVoiceChas()
			voiceChas[i] = voiceChas[i].SetRoles(cha.GetRoles())
			guildSettings = guildSettings.SetVoiceChas(voiceChas)

			merge = true
			break
		}
	}

	// Adds the voice channel and role to the guild voice channels
	if !merge {
		cha = cha.AppendToRoles(role)
		guildSettings = guildSettings.AppendToVoiceChas(cha)
	}

	db.SetGuildSettings(m.GuildID, guildSettings)

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Channel `%v` will now give role `%v` when a user joins and take it away when they leave.", cha.GetName(), role.GetName()))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Removes a voice channel or role from the voice channel list
func removeVoiceChaRole(s *discordgo.Session, m *discordgo.Message) {

	var (
		message string

		roleExistsInCmd bool
		chaExists       bool
		chaDeleted      bool

		cha  entities.VoiceCha
		role entities.Role
	)

	guildSettings := db.GetGuildSettings(m.GuildID)
	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 3)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"removevoice [channel ID] [role]*`\n\n***** is optional")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parse channel
	chaID, chaName := common.ChannelParser(s, commandStrings[1], m.GuildID)
	cha = cha.SetID(chaID)
	cha = cha.SetName(chaName)
	if cha.GetID() == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such channel exists.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	// Parse role
	if len(commandStrings) == 3 {
		roleID, roleName := common.RoleParser(s, commandStrings[2], m.GuildID)
		role = role.SetID(roleID)
		role = role.SetName(roleName)
		if role.GetID() == "" {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role exists.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
	}
	if role.GetID() != "" {
		roleExistsInCmd = true
	}

	// Checks if that channel exists in the voice channel list
	for _, voiceCha := range guildSettings.GetVoiceChas() {
		if voiceCha.GetID() == cha.GetID() {
			chaExists = true
			break
		}
	}

	if !chaExists {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such voice channel has been set.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Delete only the role if there, else delete the entire channel
	if roleExistsInCmd {
		for _, voiceCha := range guildSettings.GetVoiceChas() {
			if voiceCha.GetID() == cha.GetID() {
				for j, roleIteration := range voiceCha.GetRoles() {
					if roleIteration.GetID() == role.GetID() {

						if len(voiceCha.GetRoles()) == 1 {
							chaDeleted = true
						}

						voiceCha = voiceCha.RemoveFromRoles(j)
						break
					}
				}
			}
		}
	} else {
		for i, voiceCha := range guildSettings.GetVoiceChas() {
			if voiceCha.GetID() == cha.GetID() {
				guildSettings = guildSettings.RemoveFromVoiceChas(i)
				chaDeleted = true
				break
			}
		}
	}

	db.SetGuildSettings(m.GuildID, guildSettings)

	if chaDeleted {
		message = fmt.Sprintf("Success! Entire channel`%v` and all associated roles has been removed from the voice channel list.", cha.GetName())
	} else {
		message = fmt.Sprintf("Success! Removed `%s` from voice channel `%s` in the voice channel list.", role.GetName(), cha.GetName())
	}

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Prints all set voice channels and their associated role
func viewVoiceChaRoles(s *discordgo.Session, m *discordgo.Message) {

	var (
		message      string
		splitMessage []string
	)

	guildSettings := db.GetGuildSettings(m.GuildID)
	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"voiceroles`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	if len(guildSettings.GetVoiceChas()) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no set voice channel roles.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	for _, cha := range guildSettings.GetVoiceChas() {
		message += fmt.Sprintf("**%v : %v**\n\n", cha.GetName(), cha.GetID())
		for _, role := range cha.GetRoles() {
			message += fmt.Sprintf("`%s - %s`\n", role.GetName(), role.GetID())
		}
		message += "——————\n"
	}

	// Splits the message if it's over 1900 characters
	if len(message) > 1900 {
		splitMessage = common.SplitLongMessage(message)
	}

	// Prints split or unsplit
	if splitMessage == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	for i := 0; i < len(splitMessage); i++ {
		_, err := s.ChannelMessageSend(m.ChannelID, splitMessage[i])
		if err != nil {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot send voice channel roles message.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
		}
	}
}

// Handles react module disable or enable
func reactModuleCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		message string
		module  bool
	)

	guildSettings := db.GetGuildSettings(m.GuildID)
	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	// Displays current module setting
	if len(commandStrings) == 1 {
		if !guildSettings.GetReactsModule() {
			message = fmt.Sprintf("Reacts module is disabled. Please use `%vreactmodule true` to enable it.", guildSettings.GetPrefix())
		} else {
			message = fmt.Sprintf("Reacts module is enabled. Please use `%vreactmodule false` to disable it.", guildSettings.GetPrefix())
		}
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	} else if len(commandStrings) > 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vreactmodule [true/false]`", guildSettings.GetPrefix()))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parses bool
	if commandStrings[1] == "true" ||
		commandStrings[1] == "1" ||
		commandStrings[1] == "enable" {
		module = true
		message = "Success! Reacts module was enabled."
	} else if commandStrings[1] == "false" ||
		commandStrings[1] == "0" ||
		commandStrings[1] == "disable" {
		module = false
		message = "Success! Reacts module was disabled."
	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: That is not a valid value. Please use `true` or `false`.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Changes and writes module bool to guild
	guildSettings = guildSettings.SetReactsModule(module)
	db.SetGuildSettings(m.GuildID, guildSettings)

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Handles ping message view or change
func pingMessageCommand(s *discordgo.Session, m *discordgo.Message) {

	guildSettings := db.GetGuildSettings(m.GuildID)

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	// Displays current prefix
	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current ping message is: `%s` \n\n To change ping message please use `%spingmessage [new ping]`", guildSettings.GetPingMessage(), guildSettings.GetPrefix()))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Changes and writes new ping message to storage
	guildSettings = guildSettings.SetPingMessage(commandStrings[1])
	db.SetGuildSettings(m.GuildID, guildSettings)

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New ping message is: `%s`", guildSettings.GetPingMessage()))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Handles mod only disable or enable
func modOnlyCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		message string
		mode    bool
	)

	guildSettings := db.GetGuildSettings(m.GuildID)

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	// Displays current mode setting
	if len(commandStrings) == 1 {
		if !guildSettings.ModOnly {
			message = fmt.Sprintf("Mod-only mode is disabled. Please use `%smodonly true` to enable it.", guildSettings.GetPrefix())
		} else {
			message = fmt.Sprintf("Mod-only mode is enabled. Please use `%smodonly false` to disable it.", guildSettings.GetPrefix())
		}

		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	} else if len(commandStrings) > 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%smodonly [true/false]`", guildSettings.GetPrefix()))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parses bool
	if commandStrings[1] == "true" ||
		commandStrings[1] == "1" ||
		commandStrings[1] == "enable" {
		mode = true
		message = "Success! Mod-only mode was enabled."
	} else if commandStrings[1] == "false" ||
		commandStrings[1] == "0" ||
		commandStrings[1] == "disable" {
		mode = false
		message = "Success! Mod-only mode was disabled."
	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: That is not a valid value. Please use `true` or `false`.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Changes and writes mode bool to guild
	guildSettings = guildSettings.SetModOnly(mode)
	db.SetGuildSettings(m.GuildID, guildSettings)

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

func init() {
	Add(&Command{
		Execute:    addCommandRole,
		Trigger:    "addcommandrole",
		Aliases:    []string{"setcommandrole"},
		Desc:       "Adds a privileged role",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	Add(&Command{
		Execute:    removeCommandRole,
		Trigger:    "removecommandrole",
		Aliases:    []string{"killcommandrole"},
		Desc:       "Removes a privileged role",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	Add(&Command{
		Execute:    viewCommandRoles,
		Trigger:    "commandroles",
		Aliases:    []string{"vcommandroles", "viewcommandrole", "commandrole", "viewcommandroles", "showcommandroles"},
		Desc:       "Prints all privileged roles",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	Add(&Command{
		Execute:    prefixCommand,
		Trigger:    "prefix",
		Desc:       "Views or changes the current prefix.",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	Add(&Command{
		Execute:    botLogCommand,
		Trigger:    "botlog",
		Desc:       "Views or changes the current Bot Log.",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	Add(&Command{
		Execute:    addVoiceChaRole,
		Trigger:    "addvoice",
		Aliases:    []string{"addvoicechannelrole", "addvoicecharole", "addvoicerole", "addvoicerole", "addvoicechannelrole", "addvoicerole"},
		Desc:       "Sets a voice channel as one that will give users the specified role when they join it",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	Add(&Command{
		Execute:    removeVoiceChaRole,
		Trigger:    "removevoice",
		Aliases:    []string{"removevoicechannelrole", "removevoicechannelrole", "killvoicecharole", "killvoicechannelrole", "killvoicechannelidrole", "removevoicechannelidrole", "removevoicecharole", "removevoicerole", "removevoicerole", "killvoice"},
		Desc:       "Stops a voice channel from giving its associated role on user join",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	Add(&Command{
		Execute:    viewVoiceChaRoles,
		Trigger:    "voiceroles",
		Aliases:    []string{"viewvoicechannels", "viewvoicechannel", "viewvoicechaids", "viewvoicechannelids", "viewvoivechannelid", "viewvoicecharole", "voicerole", "voicechannelroles", "viewvoicecharoles", "voice", "voices"},
		Desc:       "Prints all set voice channels and their associated roles",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	Add(&Command{
		Execute:    reactModuleCommand,
		Trigger:    "reactmodule",
		Aliases:    []string{"reactumod", "reactsmodule", "reactsmod"},
		Desc:       "React Module. [REACTS]",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	Add(&Command{
		Execute:    pingMessageCommand,
		Trigger:    "pingmessage",
		Aliases:    []string{"pingmsg"},
		Desc:       "Views or changes the current ping message.",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	Add(&Command{
		Execute:    modOnlyCommand,
		Trigger:    "modonly",
		Desc:       "Allow only Mods and Admins to use BOT commands in the entire server",
		Permission: functionality.Admin,
		Module:     "settings",
	})
}
