package commands

import (
	"fmt"
	"strings"

	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"github.com/r-anime/ZeroTsu/events"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// addModeratorRole adds a role to the command role list
func addModeratorRole(role *discordgo.Role, guildID string) string {
	guildSettings := db.GetGuildSettings(guildID)

	// Checks if the role already exists as a command role
	for _, commandRole := range guildSettings.GetCommandRoles() {
		if commandRole.GetID() == role.ID {
			return "Error: That role is already a moderator role."
		}
	}

	// Adds the role to the guild command roles
	guildSettings = guildSettings.AppendToCommandRoles(entities.Role{
		Name:     role.Name,
		ID:       role.ID,
		Position: role.Position,
	})
	db.SetGuildSettings(guildID, guildSettings)

	return fmt.Sprintf("Success! Role `%s` is now a moderator role.", role.Name)
}

// addModeratorRoleHandler adds a role to the command role list
func addModeratorRoleHandler(s *discordgo.Session, m *discordgo.Message) {
	var role entities.Role

	guildSettings := db.GetGuildSettings(m.GuildID)
	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"add-moderator-role [Role ID]`")
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
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: That role is already a moderator role.")
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

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Role `%s` is now a moderator role.", role.GetName()))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// removeModeratorRole removes a role from the command role list
func removeModeratorRole(role *discordgo.Role, guildID string) string {
	var (
		roleExists    bool
		guildSettings = db.GetGuildSettings(guildID)
	)

	// Checks if that role is in the command role list
	for _, commandRole := range guildSettings.GetCommandRoles() {
		if commandRole.GetID() == role.ID {
			roleExists = true
			break
		}
	}

	if !roleExists {
		return "Error: No such role in the moderator role list."
	}

	for i, moderatorRole := range guildSettings.GetCommandRoles() {
		if moderatorRole.GetID() == role.ID {
			guildSettings = guildSettings.RemoveFromCommandRoles(i)
			break
		}
	}
	db.SetGuildSettings(guildID, guildSettings)

	return fmt.Sprintf("Success! Role `%s` has been removed from the moderator role list.", role.Name)
}

// removeModeratorRoleHandler removes a role from the command role list
func removeModeratorRoleHandler(s *discordgo.Session, m *discordgo.Message) {
	var roleExists bool

	guildSettings := db.GetGuildSettings(m.GuildID)
	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"remove-moderator-role [Role ID]`")
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
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role in the moderator role list.")
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

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Role `%s` has been removed from the moderator role list.", roleName))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// viewModeratorRoles prints all command roles
func viewModeratorRoles(guildID string) []string {
	var (
		message       string
		guildSettings = db.GetGuildSettings(guildID)
	)

	if len(guildSettings.GetCommandRoles()) == 0 {
		return []string{"Error: There are no moderator roles."}
	}

	for _, role := range guildSettings.GetCommandRoles() {
		message += fmt.Sprintf("**Name:** `%s` | **ID:** `%s`\n", role.GetName(), role.GetID())
	}

	return common.SplitLongMessage(message)
}

// viewModeratorRolesHandler prints all command roles
func viewModeratorRolesHandler(s *discordgo.Session, m *discordgo.Message) {
	var (
		message      string
		splitMessage []string
	)

	guildSettings := db.GetGuildSettings(m.GuildID)
	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"moderator-roles`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	if len(guildSettings.GetCommandRoles()) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no moderator roles.")
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
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot send moderator roles message.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
		}
	}
}

// prefixCommandHandler handles prefix view or change
func prefixCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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

// botLogCommand handles botlog view or change
func botLogCommand(targetChannel *discordgo.Channel, enabled bool, guildID string) string {
	var (
		guildBotLog   entities.Cha
		guildSettings = db.GetGuildSettings(guildID)
	)

	if guildSettings.BotLog != (entities.Cha{}) {
		guildBotLog = guildSettings.BotLog
	}

	if targetChannel == nil && guildBotLog == (entities.Cha{}) {
		return "Error: Bot log is currently not set."
	} else if targetChannel == nil && guildBotLog != (entities.Cha{}) && enabled {
		return fmt.Sprintf("Current bot log channel is: `%s - %s`", guildBotLog.GetName(), guildBotLog.GetID())
	}

	if guildBotLog == (entities.Cha{}) {
		guildBotLog = entities.NewCha("", "")
	}

	// Parse and save the target channel
	if !enabled {
		guildBotLog = entities.Cha{}
	} else {
		guildBotLog = entities.NewCha(targetChannel.Name, targetChannel.ID)
	}

	// Write
	guildSettings = guildSettings.SetBotLog(guildBotLog)
	db.SetGuildSettings(guildID, guildSettings)

	if guildBotLog == (entities.Cha{}) {
		return "Success: Bot log has been disabled!"
	} else {
		return fmt.Sprintf("Success: New bot log channel is: `%s - %s`", guildBotLog.GetName(), guildBotLog.GetID())
	}
}

// botLogCommandHandler handles botlog view or change
func botLogCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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

// addVoiceChaRole adds a voice channel ID with a role to the voice channel list
func addVoiceChaRole(targetChannel *discordgo.Channel, targetRole *discordgo.Role) string {
	var (
		cha           entities.VoiceCha
		role          entities.Role
		merge         bool
		guildSettings = db.GetGuildSettings(targetChannel.GuildID)
	)

	// Set cha
	cha = cha.SetID(targetChannel.ID)
	cha = cha.SetName(targetChannel.Name)

	if targetChannel.Type != discordgo.ChannelTypeGuildVoice {
		return "Error: That is not a voice channel. Please use a voice channel."
	}

	// Set role
	role = role.SetID(targetRole.ID)
	role = role.SetName(targetRole.Name)

	// Checks if the role is already set
	for i, voiceCha := range guildSettings.GetVoiceChas() {
		if voiceCha.GetID() == cha.GetID() {
			for _, roleIteration := range voiceCha.GetRoles() {
				if roleIteration.GetID() == role.GetID() {
					return "Error: That role is already set to that voice channel."
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

	// Adds the voice channel and role to the guild voice channels and updates db
	if !merge {
		cha = cha.AppendToRoles(role)
		guildSettings = guildSettings.AppendToVoiceChas(cha)
	}

	db.SetGuildSettings(targetChannel.GuildID, guildSettings)

	return fmt.Sprintf("Success! Channel `%s` will now give role `%s` when a user joins and take it away when they leave.", cha.GetName(), role.GetName())
}

// addVoiceChaRoleHandler adds a voice channel ID with a role to the voice channel list
func addVoiceChaRoleHandler(s *discordgo.Session, m *discordgo.Message) {
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

// removeVoiceChaRole removes a voice channel or role from the voice channel list
func removeVoiceChaRole(targetChannel *discordgo.Channel, targetRole *discordgo.Role) string {
	var (
		roleExistsInCmd bool
		chaExists       bool
		allDeleted      bool
		roleDeleted     bool

		cha  entities.VoiceCha
		role entities.Role

		guildSettings = db.GetGuildSettings(targetChannel.GuildID)
	)

	// Set cha
	cha = cha.SetID(targetChannel.ID)
	cha = cha.SetName(targetChannel.Name)

	// Set role
	if targetRole != nil {
		role = role.SetID(targetRole.ID)
		role = role.SetName(targetRole.Name)
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
		return "Error: No such voice channel has been set."
	}

	// Delete only the role if there, else delete all roles associated with that voice channel
	if roleExistsInCmd {
		for _, voiceCha := range guildSettings.GetVoiceChas() {
			if voiceCha.GetID() == cha.GetID() {
				for j, roleIteration := range voiceCha.GetRoles() {
					if roleIteration.GetID() == role.GetID() {
						guildSettings = guildSettings.RemoveFromVoiceChas(j)
						roleDeleted = true
						break
					}
				}
			}
		}
	} else {
		for i, voiceCha := range guildSettings.GetVoiceChas() {
			if voiceCha.GetID() == cha.GetID() {
				guildSettings = guildSettings.RemoveFromVoiceChas(i)
				allDeleted = true
				roleDeleted = true
				break
			}
		}
	}

	if !roleDeleted {
		return "Error: No such role channel needs removal."
	}

	db.SetGuildSettings(targetChannel.GuildID, guildSettings)

	if allDeleted {
		return fmt.Sprintf("Success! All roles associated with `%s` have been removed.", cha.GetName())
	} else {
		return fmt.Sprintf("Success! Removed the role `%s` from the voice channel `%s`.", role.GetName(), cha.GetName())
	}
}

// removeVoiceChaRoleHandler removes a voice channel or role from the voice channel list
func removeVoiceChaRoleHandler(s *discordgo.Session, m *discordgo.Message) {
	var (
		message string

		roleExistsInCmd bool
		chaExists       bool
		chaDeleted      bool
		roleDeleted     bool

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

	// Delete only the role if there, else delete all roles associated with that voice channel
	if roleExistsInCmd {
		for _, voiceCha := range guildSettings.GetVoiceChas() {
			if voiceCha.GetID() == cha.GetID() {
				for j, roleIteration := range voiceCha.GetRoles() {
					if roleIteration.GetID() == role.GetID() {
						guildSettings = guildSettings.RemoveFromVoiceChas(j)
						roleDeleted = true
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
				roleDeleted = true
				break
			}
		}
	}

	if !roleDeleted {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role channel needs removal.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	db.SetGuildSettings(m.GuildID, guildSettings)

	if chaDeleted {
		message = fmt.Sprintf(fmt.Sprintf("Success! All roles associated with `%s` have been removed.", cha.GetName()), cha.GetName())
	} else {
		message = fmt.Sprintf("Success! Removed the role `%s` from the voice channel `%s`.", role.GetName(), cha.GetName())
	}

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// viewVoiceChaRoles prints all set voice channels and their associated role
func viewVoiceChaRoles(guildID string) []string {
	var (
		message       string
		guildSettings = db.GetGuildSettings(guildID)
	)

	if len(guildSettings.GetVoiceChas()) == 0 {
		return []string{"Error: There are no set voice channel roles."}
	}

	for _, cha := range guildSettings.GetVoiceChas() {
		message += fmt.Sprintf("**%v : %v**\n\n", cha.GetName(), cha.GetID())
		for _, role := range cha.GetRoles() {
			message += fmt.Sprintf("`%s - %s`\n", role.GetName(), role.GetID())
		}
		message += "——————\n"
	}

	return common.SplitLongMessage(message)
}

// viewVoiceChaRolesHandler prints all set voice channels and their associated role
func viewVoiceChaRolesHandler(s *discordgo.Session, m *discordgo.Message) {
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

// reactModuleCommand handles react module disable or enable
func reactModuleCommand(print bool, enabled bool, guildID string) string {
	guildSettings := db.GetGuildSettings(guildID)

	if print {
		if guildSettings.GetReactsModule() {
			return "Reacts module is enabled."
		} else {
			return "Reacts module is disabled."
		}
	}

	// Changes and writes module bool to guild
	guildSettings = guildSettings.SetReactsModule(enabled)
	db.SetGuildSettings(guildID, guildSettings)

	if !enabled {
		return "Success! Reacts module was disabled."
	}

	return "Success! Reacts module was enabled."
}

// reactModuleCommandHandler handles react module disable or enable
func reactModuleCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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

// pingMessageCommand handles ping message view or change
func pingMessageCommand(pingMsg, guildID string) string {
	guildSettings := db.GetGuildSettings(guildID)

	// Displays current ping message
	if pingMsg == "" {
		return fmt.Sprintf("Current ping message is: `%s`", guildSettings.GetPingMessage())
	}

	// Changes and writes new ping message to storage
	guildSettings = guildSettings.SetPingMessage(pingMsg)
	db.SetGuildSettings(guildID, guildSettings)

	return fmt.Sprintf("Success! New ping message is: `%s`", guildSettings.GetPingMessage())
}

// pingMessageCommandHandler handles ping message view or change
func pingMessageCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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

// modOnlyCommand handles mod only disable or enable
func modOnlyCommand(print bool, enabled bool, guildID string) string {
	guildSettings := db.GetGuildSettings(guildID)

	// Displays current mode setting
	if print {
		if guildSettings.ModOnly {
			return "Mod-only mode is enabled."
		} else {
			return "Mod-only mode is disabled."
		}
	}

	// Changes and writes enabled bool to guild
	guildSettings = guildSettings.SetModOnly(enabled)
	db.SetGuildSettings(guildID, guildSettings)

	if !enabled {
		return "Success! Mod-only mode was disabled."
	}

	return "Success! Mod-only mode was enabled."
}

// modOnlyCommandHandler handles mod only disable or enable
func modOnlyCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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
		Execute:    addModeratorRoleHandler,
		Name:       "add-moderator-role",
		Aliases:    []string{"setmoderatorrole", "addmoderatorrole", "addcommandrole", "setcommandrole", "add-command-role", "addcommandrole"},
		Desc:       "Adds a moderator role",
		Permission: functionality.Admin,
		Module:     "settings",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        "role",
				Description: "The role you want to elevate to moderator.",
				Required:    true,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "add-moderator-role", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			var role *discordgo.Role
			if i.Data.Options == nil {
				return
			}

			for _, option := range i.Data.Options {
				if option.Name == "role" {
					role = option.RoleValue(s, i.GuildID)
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionApplicationCommandResponseData{
					Content: addModeratorRole(role, i.GuildID),
				},
			})
		},
	})
	Add(&Command{
		Execute:    removeModeratorRoleHandler,
		Name:       "remove-moderator-role",
		Aliases:    []string{"killmoderatorrole", "removemoderatorrole", "removecommandrole", "killcommandrole", "remove-command-role"},
		Desc:       "Removes a moderator role",
		Permission: functionality.Admin,
		Module:     "settings",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        "role",
				Description: "The role you want to remove from being a moderator.",
				Required:    true,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "remove-moderator-role", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			var role *discordgo.Role
			if i.Data.Options == nil {
				return
			}

			for _, option := range i.Data.Options {
				if option.Name == "role" {
					role = option.RoleValue(s, i.GuildID)
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionApplicationCommandResponseData{
					Content: removeModeratorRole(role, i.GuildID),
				},
			})
		},
	})
	Add(&Command{
		Execute:    viewModeratorRolesHandler,
		Name:       "moderator-roles",
		Aliases:    []string{"vmoderatorroles", "viewmoderatorrole", "moderatorrole", "viewmoderatorroles", "showmoderatorroles", "moderatorroles", "commandroles", "commandrole", "viewcommandroles", "viewcommandrole"},
		Desc:       "Prints all moderator roles",
		Permission: functionality.Admin,
		Module:     "settings",
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "moderator-roles", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			messages := viewModeratorRoles(i.GuildID)
			if messages == nil {
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionApplicationCommandResponseData{
					Content: messages[0],
				},
			})

			if len(messages) > 1 {
				for j, message := range messages {
					if j == 0 {
						continue
					}

					s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
						Content: message,
					})
				}
			}
		},
	})
	Add(&Command{
		Execute:    prefixCommandHandler,
		Name:       "prefix",
		Desc:       "Views or changes the current prefix.",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	Add(&Command{
		Execute:    botLogCommandHandler,
		Name:       "bot-log",
		Aliases:    []string{"botlog"},
		Desc:       "Views or changes the current bot log.",
		Permission: functionality.Admin,
		Module:     "settings",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "The channel in which you want to set the new bot log to.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "enabled",
				Description: "Whether the bot log should be enabled or disabled.",
				Required:    false,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "bot-log", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			var targetChannel *discordgo.Channel
			enabled := true
			if i.Data.Options != nil {
				for _, option := range i.Data.Options {
					if option.Name == "channel" {
						targetChannel = option.ChannelValue(s)
					} else if option.Name == "enabled" {
						enabled = option.BoolValue()
					}
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionApplicationCommandResponseData{
					Content: botLogCommand(targetChannel, enabled, i.GuildID),
				},
			})
		},
	})
	Add(&Command{
		Execute:    addVoiceChaRoleHandler,
		Name:       "add-voice",
		Aliases:    []string{"addvoicechannelrole", "addvoicecharole", "addvoicerole", "addvoicerole", "addvoicechannelrole", "addvoicerole", "addvoice"},
		Desc:       "Sets a voice channel as one that will give users the specified role when they join it",
		Permission: functionality.Admin,
		Module:     "settings",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "The voice channel in which you want it to give and remove the role from.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        "role",
				Description: "The role you want it to give and remove.",
				Required:    true,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "add-voice", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			var (
				targetChannel *discordgo.Channel
				targetRole    *discordgo.Role
			)
			if i.Data.Options == nil {
				return
			}
			for _, option := range i.Data.Options {
				if option.Name == "channel" {
					targetChannel = option.ChannelValue(s)
				} else if option.Name == "role" {
					targetRole = option.RoleValue(s, i.GuildID)
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionApplicationCommandResponseData{
					Content: addVoiceChaRole(targetChannel, targetRole),
				},
			})
		},
	})
	Add(&Command{
		Execute:    removeVoiceChaRoleHandler,
		Name:       "remove-voice",
		Aliases:    []string{"removevoicechannelrole", "removevoicechannelrole", "killvoicecharole", "killvoicechannelrole", "killvoicechannelidrole", "removevoicechannelidrole", "removevoicecharole", "removevoicerole", "removevoicerole", "killvoice", "removevoice"},
		Desc:       "Stops a voice channel from toggling a specified role on user join/leave",
		Permission: functionality.Admin,
		Module:     "settings",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "The voice channel in which you want to remove associated roles.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        "role",
				Description: "The role you want to remove.",
				Required:    false,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "remove-voice", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			var (
				targetChannel *discordgo.Channel
				targetRole    *discordgo.Role
			)
			if i.Data.Options == nil {
				return
			}

			for _, option := range i.Data.Options {
				if option.Name == "channel" {
					targetChannel = option.ChannelValue(s)
				} else if option.Name == "role" {
					targetRole = option.RoleValue(s, i.GuildID)
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionApplicationCommandResponseData{
					Content: removeVoiceChaRole(targetChannel, targetRole),
				},
			})
		},
	})
	Add(&Command{
		Execute:    viewVoiceChaRolesHandler,
		Name:       "voice-roles",
		Aliases:    []string{"viewvoicechannels", "viewvoicechannel", "viewvoicechaids", "viewvoicechannelids", "viewvoivechannelid", "viewvoicecharole", "voicerole", "voicechannelroles", "viewvoicecharoles", "voice", "voices", "voiceroles"},
		Desc:       "Prints all set voice channels and their associated roles",
		Permission: functionality.Admin,
		Module:     "settings",
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "voice-roles", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			messages := viewVoiceChaRoles(i.GuildID)
			if messages == nil {
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionApplicationCommandResponseData{
					Content: messages[0],
				},
			})

			if len(messages) > 1 {
				for j, message := range messages {
					if j == 0 {
						continue
					}

					s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
						Content: message,
					})
				}
			}
		},
	})
	Add(&Command{
		Execute:    reactModuleCommandHandler,
		Name:       "react-module",
		Aliases:    []string{"reactumod", "reactsmodule", "reactsmod", "reactmodule", "reactsmodule", "reacts-module"},
		Desc:       "Display or change whether the reacts module is enabled",
		Permission: functionality.Admin,
		Module:     "settings",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "enabled",
				Description: "Whether the reacts module should be enabled or disabled.",
				Required:    false,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "react-module", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			enabled := true
			printModule := true
			if i.Data.Options != nil {
				for _, option := range i.Data.Options {
					if option.Name == "enabled" {
						enabled = option.BoolValue()
						printModule = false
					}
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionApplicationCommandResponseData{
					Content: reactModuleCommand(printModule, enabled, i.GuildID),
				},
			})
		},
	})
	Add(&Command{
		Execute:    pingMessageCommandHandler,
		Name:       "ping-message",
		Aliases:    []string{"pingmsg", "pingmessage"},
		Desc:       "Display or change the current ping message",
		Permission: functionality.Admin,
		Module:     "settings",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message",
				Description: "The new text you want to change the ping message to.",
				Required:    false,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "ping-message", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			var pingMsg string
			if i.Data.Options != nil {
				for _, option := range i.Data.Options {
					if option.Name == "message" {
						pingMsg = option.StringValue()
					}
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionApplicationCommandResponseData{
					Content: pingMessageCommand(pingMsg, i.GuildID),
				},
			})
		},
	})
	Add(&Command{
		Execute:    modOnlyCommandHandler,
		Name:       "mod-only",
		Aliases:    []string{"modonly", "adminonly", "admin-only"},
		Desc:       "Allow only Mods and Admins to use BOT commands in the entire server",
		Permission: functionality.Admin,
		Module:     "settings",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "enabled",
				Description: "Whether to make it so only Mods and Admins can use this BOT's commands in the server.",
				Required:    false,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "mod-only", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			enabled := true
			printModule := true
			if i.Data.Options != nil {
				for _, option := range i.Data.Options {
					if option.Name == "enabled" {
						enabled = option.BoolValue()
						printModule = false
					}
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionApplicationCommandResponseData{
					Content: modOnlyCommand(printModule, enabled, i.GuildID),
				},
			})
		},
	})
}
