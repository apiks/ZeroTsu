package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Adds a role to the command role list
func addCommandRole(s *discordgo.Session, m *discordgo.Message) {

	var role functionality.Role

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"addocommandrole [Role ID]`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parse role for roleID
	role.ID, role.Name = functionality.RoleParser(s, commandStrings[1], m.GuildID)
	if role.ID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role exists.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if the role already exists as a command role
	for _, commandRole := range guildSettings.CommandRoles {
		if commandRole.ID == role.ID {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: That role is already a command role.")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
	}

	// Adds the role to the guild command roles
	guildSettings.CommandRoles = append(guildSettings.CommandRoles, &role)
	functionality.Mutex.Lock()
	functionality.GuildMap[m.GuildID].GuildConfig = guildSettings
	_ = functionality.GuildSettingsWrite(functionality.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	functionality.Mutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Role `%v` is now a privileged role.", role.Name))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Removes a role from the command role list
func removeCommandRole(s *discordgo.Session, m *discordgo.Message) {

	var roleExists bool

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"removecommandrole [Role ID]`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parse role for roleID
	roleID, roleName := functionality.RoleParser(s, commandStrings[1], m.GuildID)
	if roleID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role exists.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if that role is in the command role list
	for _, commandRole := range guildSettings.CommandRoles {
		if commandRole.ID == roleID {
			roleExists = true
			break
		}
	}

	if !roleExists {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role in the command role list.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	for i, role := range guildSettings.CommandRoles {
		if role.ID == roleID {
			guildSettings.CommandRoles = append(guildSettings.CommandRoles[:i], guildSettings.CommandRoles[i+1:]...)
			break
		}
	}

	functionality.Mutex.Lock()
	functionality.GuildMap[m.GuildID].GuildConfig = guildSettings
	_ = functionality.GuildSettingsWrite(functionality.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	functionality.Mutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Role `%v` has been removed from the command role list.", roleName))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Prints all command roles
func viewCommandRoles(s *discordgo.Session, m *discordgo.Message) {

	var (
		message      string
		splitMessage []string
	)

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"commandroles`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	if len(guildSettings.CommandRoles) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no privileged command roles.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	for _, role := range guildSettings.CommandRoles {
		message += fmt.Sprintf("**Name:** `%v` | **ID:** `%v`\n", role.Name, role.ID)
	}

	// Splits the message if it's over 1900 characters
	if len(message) > 1900 {
		splitMessage = functionality.SplitLongMessage(message)
	}

	// Prints split or unsplit
	if splitMessage == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	for i := 0; i < len(splitMessage); i++ {
		_, err := s.ChannelMessageSend(m.ChannelID, splitMessage[i])
		if err != nil {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot send commandroles message.")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
		}
	}
}

// Handles prefix view or change
func prefixCommand(s *discordgo.Session, m *discordgo.Message) {

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	// Displays current prefix
	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current prefix is: `%v` \n\n To change prefix please use `%vprefix [new prefix]`", guildSettings.Prefix, guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Changes and writes new prefix to storage
	functionality.Mutex.Lock()
	functionality.GuildMap[m.GuildID].GuildConfig.Prefix = commandStrings[1]
	_ = functionality.GuildSettingsWrite(functionality.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	functionality.Mutex.Unlock()
	functionality.DynamicNicknameChange(s, m.GuildID, guildSettings.Prefix)

	guildSettings.Prefix = commandStrings[1]

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New prefix is: `%v`", guildSettings.Prefix))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Handles botlog view or change
func botLogCommand(s *discordgo.Session, m *discordgo.Message) {

	var message string

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	// Displays current botlog channel
	if len(commandStrings) == 1 {
		if guildSettings.BotLog == nil {
			message = fmt.Sprintf("Error: Bot Log is currently not set. Please use `%sbotlog [channel]`", guildSettings.Prefix)
		} else if guildSettings.BotLog.ID == "" {
			message = fmt.Sprintf("Error: Bot Log is currently not set. Please use `%sbotlog [channel]`", guildSettings.Prefix)
		} else {
			message = fmt.Sprintf("Current Bot Log is: `%s - %s` \n\n To change Bot Log please use `%sbotlog [channel]`", guildSettings.BotLog.Name, guildSettings.BotLog.ID, guildSettings.Prefix)
		}

		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parses channel
	chaID, chaName := functionality.ChannelParser(s, commandStrings[1], m.GuildID)
	if chaID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such channel exists.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Changes and writes new bot log to storage
	functionality.Mutex.Lock()
	functionality.GuildMap[m.GuildID].GuildConfig.BotLog = &functionality.Cha{Name: chaName, ID: chaID}
	_ = functionality.GuildSettingsWrite(functionality.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	functionality.Mutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Bot Log is: `%v - %v`", chaName, chaID))
	if err != nil {
		guildSettings.BotLog = &functionality.Cha{Name: chaName, ID: chaID}
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Handles optInUnder view or change
func optInUnderCommand(s *discordgo.Session, m *discordgo.Message) {

	var message string

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	// Displays current optinunder role
	if len(commandStrings) == 1 {
		if guildSettings.OptInUnder == nil {
			message = fmt.Sprintf("Error: 'Opt In Under' role is currently not set. Please use `%soptinunder [role]`", guildSettings.Prefix)
		} else if guildSettings.OptInUnder.ID == "" {
			message = fmt.Sprintf("Error: 'Opt In Under' role is currently not set. Please use `%soptinunder [role]`", guildSettings.Prefix)
		} else {
			message = fmt.Sprintf("Current 'Opt In Under' role is: `%s - %s` \n\n To change 'Opt In Under' role please use `%soptinunder [role]`", guildSettings.OptInUnder.Name, guildSettings.OptInUnder.ID, guildSettings.Prefix)
		}

		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parses role
	roleID, roleName := functionality.RoleParser(s, commandStrings[1], m.GuildID)
	if roleID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role exists.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Changes and writes new optinunder role to storage
	functionality.Mutex.Lock()
	functionality.GuildMap[m.GuildID].GuildConfig.OptInAbove = &functionality.Role{Name: roleName, ID: roleID}
	_ = functionality.GuildSettingsWrite(functionality.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	functionality.Mutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! 'Opt In Under' role is: `%v - %v`", roleName, roleID))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
	}
}

// Handles optInAbove view or change
func optInAboveCommand(s *discordgo.Session, m *discordgo.Message) {

	var message string

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	// Displays current optinabove role
	if len(commandStrings) == 1 {
		if guildSettings.OptInAbove == nil {
			message = fmt.Sprintf("Error: 'Opt In Above' role is currently not set. Please use `%soptinunder [role]`", guildSettings.Prefix)
		} else if guildSettings.OptInAbove.ID == "" {
			message = fmt.Sprintf("Error: 'Opt In Above' role is currently not set. Please use `%soptinunder [role]`", guildSettings.Prefix)
		} else {
			message = fmt.Sprintf("Current 'Opt In Above' role is: `%s - %s` \n\n To change 'Opt In Above' role please use `%soptinabove [role]`", guildSettings.OptInAbove.Name, guildSettings.OptInAbove.ID, guildSettings.Prefix)
		}

		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parses role
	roleID, roleName := functionality.RoleParser(s, commandStrings[1], m.GuildID)

	// Changes and writes new optinabove role to storage
	functionality.Mutex.Lock()
	functionality.GuildMap[m.GuildID].GuildConfig.OptInAbove = &functionality.Role{Name: roleName, ID: roleID}
	_ = functionality.GuildSettingsWrite(functionality.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	functionality.Mutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! 'Opt In Above' role is: `%v - %v`", roleName, roleID))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Adds a voice channel ID with a role to the voice channel list
func addVoiceChaRole(s *discordgo.Session, m *discordgo.Message) {

	var (
		cha   functionality.VoiceCha
		role  functionality.Role
		merge bool
	)

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 3)

	if len(commandStrings) < 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"addvoice [channel ID] [role]` \n\n")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parse channel
	cha.ID, cha.Name = functionality.ChannelParser(s, commandStrings[1], m.GuildID)
	if cha.ID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such channel exists.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	voiceCheck, err := s.Channel(cha.ID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	if voiceCheck.Type != discordgo.ChannelTypeGuildVoice {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: That is not a voice channel. Please use a voice channel.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	// Parse role
	role.ID, role.Name = functionality.RoleParser(s, commandStrings[2], m.GuildID)
	if role.ID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role exists.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if the role is already set
	for i, voiceCha := range guildSettings.VoiceChas {
		if voiceCha.ID == cha.ID {
			for _, roleIteration := range voiceCha.Roles {
				if roleIteration.ID == role.ID {
					_, err := s.ChannelMessageSend(m.ChannelID, "Error: That role is already set to that channel.")
					if err != nil {
						functionality.LogError(s, guildSettings.BotLog, err)
						return
					}
					return
				}
			}
			// Adds the voice channel and role to the guild voice channels
			cha.Roles = voiceCha.Roles
			cha.Roles = append(cha.Roles, &role)
			guildSettings.VoiceChas[i].Roles = cha.Roles
			merge = true
			break
		}
	}

	// Adds the voice channel and role to the guild voice channels
	if !merge {
		cha.Roles = append(cha.Roles, &role)
		guildSettings.VoiceChas = append(guildSettings.VoiceChas, &cha)
	}

	functionality.Mutex.Lock()
	functionality.GuildMap[m.GuildID].GuildConfig = guildSettings
	_ = functionality.GuildSettingsWrite(functionality.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	functionality.Mutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Channel `%v` will now give role `%v` when a user joins and take it away when they leave.", cha.Name, role.Name))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
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

		cha  functionality.VoiceCha
		role functionality.Role
	)

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 3)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"removevoice [channel ID] [role]*`\n\n***** is optional")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parse channel
	cha.ID, cha.Name = functionality.ChannelParser(s, commandStrings[1], m.GuildID)
	if cha.ID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such channel exists.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	// Parse role
	if len(commandStrings) == 3 {
		role.ID, role.Name = functionality.RoleParser(s, commandStrings[2], m.GuildID)
		if role.ID == "" {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role exists.")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
	}
	if role.ID != "" {
		roleExistsInCmd = true
	}

	// Checks if that channel exists in the voice channel list
	for _, voiceCha := range guildSettings.VoiceChas {
		if voiceCha.ID == cha.ID {
			chaExists = true
			break
		}
	}

	if !chaExists {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such voice channel has been set.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Delete only the role if there, else delete the entire channel
	if roleExistsInCmd {
		for i, voiceCha := range guildSettings.VoiceChas {
			if voiceCha.ID == cha.ID {
				for j, roleIteration := range voiceCha.Roles {
					if roleIteration.ID == role.ID {

						if len(voiceCha.Roles) == 1 {
							chaDeleted = true
						}

						guildSettings.VoiceChas[i].Roles = append(guildSettings.VoiceChas[i].Roles[:j], guildSettings.VoiceChas[i].Roles[j+1:]...)
						break
					}
				}
			}
		}
	} else {
		for i, voiceCha := range guildSettings.VoiceChas {
			if voiceCha.ID == cha.ID {
				guildSettings.VoiceChas = append(guildSettings.VoiceChas[:i], guildSettings.VoiceChas[i+1:]...)
				chaDeleted = true
				break
			}
		}
	}

	functionality.Mutex.RLock()
	functionality.GuildMap[m.GuildID].GuildConfig = guildSettings
	_ = functionality.GuildSettingsWrite(functionality.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	functionality.Mutex.RUnlock()

	if chaDeleted {
		message = fmt.Sprintf("Success! Entire channel`%v` and all associated roles has been removed from the voice channel list.", cha.Name)
	} else {
		message = fmt.Sprintf("Success! Removed `%v` from voice channel `%v` in the voice channel list.", role.Name, cha.Name)
	}

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Prints all set voice channels and their associated role
func viewVoiceChaRoles(s *discordgo.Session, m *discordgo.Message) {

	var (
		message      string
		splitMessage []string
	)

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"voiceroles`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	if len(guildSettings.VoiceChas) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no set voice channel roles.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	for _, cha := range guildSettings.VoiceChas {
		message += fmt.Sprintf("**%v : %v**\n\n", cha.Name, cha.ID)
		for _, role := range cha.Roles {
			message += fmt.Sprintf("`%v - %v`\n", role.Name, role.ID)
		}
		message += "——————\n"
	}

	// Splits the message if it's over 1900 characters
	if len(message) > 1900 {
		splitMessage = functionality.SplitLongMessage(message)
	}

	// Prints split or unsplit
	if splitMessage == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	for i := 0; i < len(splitMessage); i++ {
		_, err := s.ChannelMessageSend(m.ChannelID, splitMessage[i])
		if err != nil {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot send voice channel roles message.")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
		}
	}
}

// Handles vote category view or change
func voteCategoryCommand(s *discordgo.Session, m *discordgo.Message) {

	var message string

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	// Displays current vote category
	if len(commandStrings) == 1 {
		if guildSettings.VoteChannelCategory == nil {
			message = fmt.Sprintf("Error: Vote Module is currently not set. Please use `%vvotecategory [category]`", guildSettings.Prefix)
		} else if guildSettings.VoteChannelCategory.ID == "" {
			message = fmt.Sprintf("Error: Vote Module is currently not set. Please use `%vvotecategory [category]`", guildSettings.Prefix)
		} else {
			message = fmt.Sprintf("Current Vote Module is: `%v - %v` \n\n To change Vote Module please use `%vvotecategory [category]`", guildSettings.VoteChannelCategory.Name, guildSettings.VoteChannelCategory.ID, guildSettings.Prefix)
		}

		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parses category
	catID, catName := functionality.CategoryParser(s, commandStrings[1], m.GuildID)
	if catID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such category exists.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Changes and writes new vote category to storage
	functionality.Mutex.Lock()
	functionality.GuildMap[m.GuildID].GuildConfig.VoteChannelCategory = &functionality.Cha{Name: catName, ID: catID}
	_ = functionality.GuildSettingsWrite(functionality.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	functionality.Mutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Vote Module is: `%v - %v`", catName, catID))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Handles vote module disable or enable
func voteModuleCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		message string
		module  bool
	)

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	// Displays current module setting
	if len(commandStrings) == 1 {
		if guildSettings.VoteModule {
			message = fmt.Sprintf("Vote module is disabled. Please use `%vvotemodule true` to enable it.", guildSettings.Prefix)
		} else {
			message = fmt.Sprintf("Vote module is enabled. Please use `%vvotemodule false` to disable it.", guildSettings.Prefix)
		}

		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	} else if len(commandStrings) > 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vvotemodule [true/false]`", guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parses bool
	if commandStrings[1] == "true" ||
		commandStrings[1] == "1" ||
		commandStrings[1] == "enable" {
		module = true
		message = "Success! Vote module was enabled."
	} else if commandStrings[1] == "false" ||
		commandStrings[1] == "0" ||
		commandStrings[1] == "disable" {
		module = false
		message = "Success! Vote module was disabled."
	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: That is not a valid value. Please use `true` or `false`.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Changes and writes module bool to guild
	functionality.Mutex.Lock()
	functionality.GuildMap[m.GuildID].GuildConfig.VoteModule = module
	_ = functionality.GuildSettingsWrite(functionality.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	functionality.Mutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Handles waifu module disable or enable
func waifuModuleCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		message string
		module  bool
	)

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	// Displays current module setting
	if len(commandStrings) == 1 {
		if !guildSettings.WaifuModule {
			message = fmt.Sprintf("Waifus module is disabled. Please use `%vwaifumodule true` to enable it.", guildSettings.Prefix)
		} else {
			message = fmt.Sprintf("Waifus module is enabled. Please use `%vwaifumodule false` to disable it.", guildSettings.Prefix)
		}
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	} else if len(commandStrings) > 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vwaifumodule [true/false]`", guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parses bool
	if commandStrings[1] == "true" ||
		commandStrings[1] == "1" ||
		commandStrings[1] == "enable" {
		module = true
		message = "Success! Waifu module was enabled."
	} else if commandStrings[1] == "false" ||
		commandStrings[1] == "0" ||
		commandStrings[1] == "disable" {
		module = false
		message = "Success! Waifu module was disabled."
	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: That is not a valid value. Please use `true` or `false`.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Changes and writes module bool to guild
	functionality.Mutex.Lock()
	functionality.GuildMap[m.GuildID].GuildConfig.WaifuModule = module
	_ = functionality.GuildSettingsWrite(functionality.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	functionality.Mutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Handles react module disable or enable
func reactModuleCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		message string
		module  bool
	)

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	// Displays current module setting
	if len(commandStrings) == 1 {
		if !guildSettings.ReactsModule {
			message = fmt.Sprintf("Reacts module is disabled. Please use `%vreactmodule true` to enable it.", guildSettings.Prefix)
		} else {
			message = fmt.Sprintf("Reacts module is enabled. Please use `%vreactmodule false` to disable it.", guildSettings.Prefix)
		}
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	} else if len(commandStrings) > 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vreactmodule [true/false]`", guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
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
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Changes and writes module bool to guild
	functionality.Mutex.Lock()
	functionality.GuildMap[m.GuildID].GuildConfig.ReactsModule = module
	_ = functionality.GuildSettingsWrite(functionality.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	functionality.Mutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Handles attachment removal disable or enable
func whitelistFileFilter(s *discordgo.Session, m *discordgo.Message) {

	var (
		message string
		module  bool
	)

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	// Displays current module setting
	if len(commandStrings) == 1 {
		if !guildSettings.WhitelistFileFilter {
			message = fmt.Sprintf("Whitelist File Filter version is disabled. Using a Blacklist File Filter. Please use `%vwhitelist true` to enable it.", guildSettings.Prefix)
		} else {
			message = fmt.Sprintf("WhitelistFile Filter version is enabled. Please use `%vwhitelist false` to disable it and enable the Blacklist File Filter instead.", guildSettings.Prefix)
		}
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	} else if len(commandStrings) > 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vwhitelist [true/false]`", guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parses bool
	if commandStrings[1] == "true" ||
		commandStrings[1] == "1" ||
		commandStrings[1] == "enable" {
		module = true
		message = "Success! Whitelist File Filter was enabled."
	} else if commandStrings[1] == "false" ||
		commandStrings[1] == "0" ||
		commandStrings[1] == "disable" {
		module = false
		message = "Success! File Filter has reverted to a blacklist."
	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: That is not a valid value. Please use `true` or `false`.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Changes and writes module bool to guild
	functionality.Mutex.Lock()
	functionality.GuildMap[m.GuildID].GuildConfig.WhitelistFileFilter = module
	_ = functionality.GuildSettingsWrite(functionality.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	functionality.Mutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Handles ping message view or change
func pingMessageCommand(s *discordgo.Session, m *discordgo.Message) {

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	// Displays current prefix
	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current ping message is: `%v` \n\n To change ping message please use `%vpingmessage [new ping]`", guildSettings.PingMessage, guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Changes and writes new ping message to storage
	functionality.Mutex.Lock()
	functionality.GuildMap[m.GuildID].GuildConfig.PingMessage = commandStrings[1]
	_ = functionality.GuildSettingsWrite(functionality.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	functionality.Mutex.Unlock()

	guildSettings.PingMessage = commandStrings[1]

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New ping message is: `%s`", guildSettings.PingMessage))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Adds a role as the muted role
func setMutedRole(s *discordgo.Session, m *discordgo.Message) {

	var role functionality.Role

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		if guildSettings.MutedRole == nil || guildSettings.MutedRole.ID == "" {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("The muted role is not set. Please use `%ssetmuted [Role ID]` to set it.", guildSettings.Prefix))
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}

		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("The current muted role is `%s - %s`.\nPlease use `%ssetmuted [Role ID]` to change it.", guildSettings.MutedRole.Name, guildSettings.MutedRole.ID, guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parse role for roleID
	role.ID, role.Name = functionality.RoleParser(s, commandStrings[1], m.GuildID)
	if role.ID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role exists.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if the role already exists as a muted role
	if guildSettings.MutedRole != nil {
		if guildSettings.MutedRole.ID == role.ID {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: That role is already the muted role.")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
	}

	// Sets the role as the muted role and writes to disk
	functionality.Mutex.Lock()
	functionality.GuildMap[m.GuildID].GuildConfig.MutedRole = &role
	_ = functionality.GuildSettingsWrite(functionality.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	functionality.Mutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Role `%v` is now the muted role.", role.Name))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    addCommandRole,
		Trigger:    "addcommandrole",
		Aliases:    []string{"setcommandrole"},
		Desc:       "Adds a privileged role",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	functionality.Add(&functionality.Command{
		Execute:    removeCommandRole,
		Trigger:    "removecommandrole",
		Aliases:    []string{"killcommandrole"},
		Desc:       "Removes a privileged role",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	functionality.Add(&functionality.Command{
		Execute:    viewCommandRoles,
		Trigger:    "commandroles",
		Aliases:    []string{"vcommandroles", "viewcommandrole", "commandrole", "viewcommandroles", "showcommandroles"},
		Desc:       "Prints all privileged roles",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	functionality.Add(&functionality.Command{
		Execute:    prefixCommand,
		Trigger:    "prefix",
		Desc:       "Views or changes the current prefix.",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	functionality.Add(&functionality.Command{
		Execute:    botLogCommand,
		Trigger:    "botlog",
		Desc:       "Views or changes the current Bot Log.",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	functionality.Add(&functionality.Command{
		Execute:    optInUnderCommand,
		Trigger:    "optinunder",
		Desc:       "Views or changes the current `Opt In Under` role.",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	functionality.Add(&functionality.Command{
		Execute:    optInAboveCommand,
		Trigger:    "optinabove",
		Desc:       "Views or changes the current `Opt In Above` role.",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	functionality.Add(&functionality.Command{
		Execute:    addVoiceChaRole,
		Trigger:    "addvoice",
		Aliases:    []string{"addvoicechannelrole", "addvoicecharole", "addvoicerole", "addvoicerole", "addvoicechannelrole", "addvoicerole"},
		Desc:       "Sets a voice channel as one that will give users the specified role when they join it",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	functionality.Add(&functionality.Command{
		Execute:    removeVoiceChaRole,
		Trigger:    "removevoice",
		Aliases:    []string{"removevoicechannelrole", "removevoicechannelrole", "killvoicecharole", "killvoicechannelrole", "killvoicechannelidrole", "removevoicechannelidrole", "removevoicecharole", "removevoicerole", "removevoicerole", "killvoice"},
		Desc:       "Stops a voice channel from giving its associated role on user join",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	functionality.Add(&functionality.Command{
		Execute:    viewVoiceChaRoles,
		Trigger:    "voiceroles",
		Aliases:    []string{"viewvoicechannels", "viewvoicechannel", "viewvoicechaids", "viewvoicechannelids", "viewvoivechannelid", "viewvoicecharole", "voicerole", "voicechannelroles", "viewvoicecharoles", "voice", "voices"},
		Desc:       "Prints all set voice channels and their associated roles",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	functionality.Add(&functionality.Command{
		Execute:    voteCategoryCommand,
		Trigger:    "votecategory",
		Desc:       "Views or changes the current Vote Module where non-admin temp vote channels will be auto placed and sorted [VOTE]",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	functionality.Add(&functionality.Command{
		Execute:    voteModuleCommand,
		Trigger:    "votemodule",
		Aliases:    []string{"votemod"},
		Desc:       "Vote Module. [VOTE]",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	functionality.Add(&functionality.Command{
		Execute:    waifuModuleCommand,
		Trigger:    "waifumodule",
		Aliases:    []string{"waifumod"},
		Desc:       "Waifu Module. [WAIFU]",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	functionality.Add(&functionality.Command{
		Execute:    reactModuleCommand,
		Trigger:    "reactmodule",
		Aliases:    []string{"reactumod", "reactsmodule", "reactsmod"},
		Desc:       "React Module. [REACTS]",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	functionality.Add(&functionality.Command{
		Execute:    whitelistFileFilter,
		Trigger:    "whitelist",
		Aliases:    []string{"filefilter", "attachmentremove", "attachremoval", "fileremove", "fileremoval", "attachmentremoval", "filesfilter", "whitelistfilter", "whitelistfile", "filewhitelist"},
		Desc:       "Switch between a whitelist attachment file filter (removes all attachments except whitelisted ones) and a blacklist attachment file filter (remove only specified file extensions)",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	functionality.Add(&functionality.Command{
		Execute:    pingMessageCommand,
		Trigger:    "pingmessage",
		Aliases:    []string{"pingmsg"},
		Desc:       "Views or changes the current ping message.",
		Permission: functionality.Admin,
		Module:     "settings",
	})
	functionality.Add(&functionality.Command{
		Execute:    setMutedRole,
		Trigger:    "setmuted",
		Aliases:    []string{"setmutedrole", "addmuted", "addmutedrole"},
		Desc:       "Sets a role as the muted role",
		Permission: functionality.Admin,
		Module:     "settings",
	})
}
