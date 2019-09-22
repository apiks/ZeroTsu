package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

// Adds a role to the command role list
func addCommandRole(s *discordgo.Session, m *discordgo.Message) {

	var role misc.Role

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(messageLowercase, " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"addocommandrole [Role ID]`")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Parse role for roleID
	role.ID, role.Name = misc.RoleParser(s, commandStrings[1], m.GuildID)
	if role.ID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role exists.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Checks if the role already exists as a command role
	misc.MapMutex.Lock()
	for _, commandRole := range misc.GuildMap[m.GuildID].GuildConfig.CommandRoles {
		if commandRole.ID == role.ID {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: That role is already a command role.")
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
				if err != nil {
					misc.MapMutex.Unlock()
					return
				}
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
	}

	// Adds the role to the guild command roles
	misc.GuildMap[m.GuildID].GuildConfig.CommandRoles = append(misc.GuildMap[m.GuildID].GuildConfig.CommandRoles, role)
	_ = misc.GuildSettingsWrite(misc.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	misc.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Role `%v` is now a privileged role.", role.Name))
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Removes a role from the command role list
func removeCommandRole(s *discordgo.Session, m *discordgo.Message) {

	var roleExists bool

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(messageLowercase, " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"removecommandrole [Role ID]`")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Parse role for roleID
	roleID, roleName := misc.RoleParser(s, commandStrings[1], m.GuildID)
	if roleID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role exists.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Checks if that role is in the command role list
	misc.MapMutex.Lock()
	for _, commandRole := range misc.GuildMap[m.GuildID].GuildConfig.CommandRoles {
		if commandRole.ID == roleID {
			roleExists = true
			break
		}
	}

	if !roleExists {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role in the command role list.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}

	for i, role := range misc.GuildMap[m.GuildID].GuildConfig.CommandRoles {
		if role.ID == roleID {
			misc.GuildMap[m.GuildID].GuildConfig.CommandRoles = append(misc.GuildMap[m.GuildID].GuildConfig.CommandRoles[:i], misc.GuildMap[m.GuildID].GuildConfig.CommandRoles[i+1:]...)
			break
		}
	}
	misc.GuildSettingsWrite(misc.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	misc.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Role `%v` has been removed from the command role list.", roleName))
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Prints all command roles
func viewCommandRoles(s *discordgo.Session, m *discordgo.Message) {

	var (
		message      string
		splitMessage []string
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	commandStrings := strings.SplitN(m.Content, " ", 2)

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"commandroles`")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	misc.MapMutex.Lock()
	if len(misc.GuildMap[m.GuildID].GuildConfig.CommandRoles) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no privileged command roles.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}

	for _, role := range misc.GuildMap[m.GuildID].GuildConfig.CommandRoles {
		message += fmt.Sprintf("**Name:** `%v` | **ID:** `%v`\n", role.Name, role.ID)
	}
	misc.MapMutex.Unlock()

	// Splits the message if it's over 1900 characters
	if len(message) > 1900 {
		splitMessage = misc.SplitLongMessage(message)
	}

	// Prints split or unsplit
	if splitMessage == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}
	for i := 0; i < len(splitMessage); i++ {
		_, err := s.ChannelMessageSend(m.ChannelID, splitMessage[i])
		if err != nil {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot send commandroles message.")
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
		}
	}
}

// Handles prefix view or change
func prefixCommand(s *discordgo.Session, m *discordgo.Message) {

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	mLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(mLowercase, " ", 2)

	// Displays current prefix
	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current prefix is: `%v` \n\n To change prefix please use `%vprefix [new prefix]`", guildPrefix, guildPrefix))
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Changes and writes new prefix to storage
	misc.MapMutex.Lock()
	misc.GuildMap[m.GuildID].GuildConfig.Prefix = commandStrings[1]
	_ = misc.GuildSettingsWrite(misc.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	misc.DynamicNicknameChange(s, m.GuildID, guildPrefix)
	misc.MapMutex.Unlock()

	guildPrefix = commandStrings[1]

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New prefix is: `%v`", guildPrefix))
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Handles botlog view or change
func botLogCommand(s *discordgo.Session, m *discordgo.Message) {

	var message string

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog
	misc.MapMutex.Unlock()

	mLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(mLowercase, " ", 2)

	// Displays current botlog channel
	if len(commandStrings) == 1 {
		if guildBotLog.ID == "" {
			message = fmt.Sprintf("Error: Bot Log is currently not set. Please use `%vbotlog [channel]`", guildPrefix)
		} else {
			message = fmt.Sprintf("Current Bot Log is: `%v - %v` \n\n To change Bot Log please use `%vbotlog [channel]`", guildBotLog.Name, guildBotLog.ID, guildPrefix)
		}

		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Parses channel
	chaID, chaName := misc.ChannelParser(s, commandStrings[1], m.GuildID)
	if chaID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such channel exists.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Changes and writes new bot log to storage
	misc.MapMutex.Lock()
	misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID = chaID
	misc.GuildMap[m.GuildID].GuildConfig.BotLog.Name = chaName
	misc.GuildSettingsWrite(misc.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	misc.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Bot Log is: `%v - %v`", chaName, chaID))
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Handles optInUnder view or change
func optInUnderCommand(s *discordgo.Session, m *discordgo.Message) {

	var message string

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog
	guildOptInUnder := misc.GuildMap[m.GuildID].GuildConfig.OptInUnder
	misc.MapMutex.Unlock()

	mLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(mLowercase, " ", 2)

	// Displays current optinunder role
	if len(commandStrings) == 1 {
		if guildOptInUnder.ID == "" {
			message = fmt.Sprintf("Error: 'Opt In Under' role is currently not set. Please use `%voptinunder [role]`", guildPrefix)
		} else {
			message = fmt.Sprintf("Current 'Opt In Under' role is: `%v - %v` \n\n To change 'Opt In Under' role please use `%voptinunder [role]`", guildOptInUnder.Name, guildOptInUnder.ID, guildPrefix)
		}

		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Parses role
	roleID, roleName := misc.RoleParser(s, commandStrings[1], m.GuildID)
	if roleID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role exists.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Changes and writes new optinunder role to storage
	misc.MapMutex.Lock()
	misc.GuildMap[m.GuildID].GuildConfig.OptInUnder.ID = roleID
	misc.GuildMap[m.GuildID].GuildConfig.OptInUnder.Name = roleName
	misc.GuildSettingsWrite(misc.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	misc.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! 'Opt In Under' role is: `%v - %v`", roleName, roleID))
	if err != nil {
		_, _ = s.ChannelMessageSend(guildBotLog.ID, err.Error())
	}
}

// Handles optInAbove view or change
func optInAboveCommand(s *discordgo.Session, m *discordgo.Message) {

	var message string

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog
	guildOptInAbove := misc.GuildMap[m.GuildID].GuildConfig.OptInAbove
	misc.MapMutex.Unlock()

	mLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(mLowercase, " ", 2)

	// Displays current optinabove role
	if len(commandStrings) == 1 {
		if guildOptInAbove.ID == "" {
			message = fmt.Sprintf("Error: 'Opt In Above' role is currently not set. Please use `%voptinunder [role]`", guildPrefix)
		} else {
			message = fmt.Sprintf("Current 'Opt In Above' role is: `%v - %v` \n\n To change 'Opt In Above' role please use `%voptinabove [role]`", guildOptInAbove.Name, guildOptInAbove.ID, guildPrefix)
		}

		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Parses role
	roleID, roleName := misc.RoleParser(s, commandStrings[1], m.GuildID)

	// Changes and writes new optinabove role to storage
	misc.MapMutex.Lock()
	misc.GuildMap[m.GuildID].GuildConfig.OptInAbove.ID = roleID
	misc.GuildMap[m.GuildID].GuildConfig.OptInAbove.Name = roleName
	misc.GuildSettingsWrite(misc.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	misc.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! 'Opt In Above' role is: `%v - %v`", roleName, roleID))
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Adds a voice channel ID with a role to the voice channel list
func addVoiceChaRole(s *discordgo.Session, m *discordgo.Message) {

	var (
		cha   misc.VoiceCha
		role  misc.Role
		merge bool
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(messageLowercase, " ", 3)

	if len(commandStrings) < 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"addvoice [channel ID] [role]` \n\n")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Parse channel
	cha.ID, cha.Name = misc.ChannelParser(s, commandStrings[1], m.GuildID)
	if cha.ID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such channel exists.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}
	voiceCheck, err := s.Channel(cha.ID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
	if voiceCheck.Type != discordgo.ChannelTypeGuildVoice {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: That is not a voice channel. Please use a voice channel.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}
	// Parse role
	role.ID, role.Name = misc.RoleParser(s, commandStrings[2], m.GuildID)
	if role.ID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role exists.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	misc.MapMutex.Lock()

	// Checks if the role is already set
	for i, voiceCha := range misc.GuildMap[m.GuildID].GuildConfig.VoiceChas {
		if voiceCha.ID == cha.ID {
			for _, roleIteration := range voiceCha.Roles {
				if roleIteration.ID == role.ID {
					_, err := s.ChannelMessageSend(m.ChannelID, "Error: That role is already set to that channel.")
					if err != nil {
						_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
						if err != nil {
							misc.MapMutex.Unlock()
							return
						}
						misc.MapMutex.Unlock()
						return
					}
					misc.MapMutex.Unlock()
					return
				}
			}
			// Adds the voice channel and role to the guild voice channels
			cha.Roles = voiceCha.Roles
			cha.Roles = append(cha.Roles, role)
			misc.GuildMap[m.GuildID].GuildConfig.VoiceChas[i].Roles = cha.Roles
			merge = true
			break
		}
	}

	// Adds the voice channel and role to the guild voice channels
	if !merge {
		cha.Roles = append(cha.Roles, role)
		misc.GuildMap[m.GuildID].GuildConfig.VoiceChas = append(misc.GuildMap[m.GuildID].GuildConfig.VoiceChas, cha)
	}
	misc.GuildSettingsWrite(misc.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	misc.MapMutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Channel `%v` will now give role `%v` when a user joins and take it away when they leave.", cha.Name, role.Name))
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
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

		cha  misc.VoiceCha
		role misc.Role
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(messageLowercase, " ", 3)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"removevoice [channel ID] [role]*`\n\n***** is optional")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Parse channel
	cha.ID, cha.Name = misc.ChannelParser(s, commandStrings[1], m.GuildID)
	if cha.ID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such channel exists.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}
	// Parse role
	if len(commandStrings) == 3 {
		role.ID, role.Name = misc.RoleParser(s, commandStrings[2], m.GuildID)
		if role.ID == "" {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such role exists.")
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
			return
		}
	}
	if role.ID != "" {
		roleExistsInCmd = true
	}

	// Checks if that channel exists in the voice channel list
	misc.MapMutex.Lock()
	for _, voiceCha := range misc.GuildMap[m.GuildID].GuildConfig.VoiceChas {
		if voiceCha.ID == cha.ID {
			chaExists = true
			break
		}
	}

	if !chaExists {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such voice channel has been set.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}

	// Delete only the role if there, else delete the entire channel
	if roleExistsInCmd {
		for i, voiceCha := range misc.GuildMap[m.GuildID].GuildConfig.VoiceChas {
			if voiceCha.ID == cha.ID {
				for j, roleIteration := range voiceCha.Roles {
					if roleIteration.ID == role.ID {

						if len(voiceCha.Roles) == 1 {
							chaDeleted = true
						}

						misc.GuildMap[m.GuildID].GuildConfig.VoiceChas[i].Roles = append(misc.GuildMap[m.GuildID].GuildConfig.VoiceChas[i].Roles[:j], misc.GuildMap[m.GuildID].GuildConfig.VoiceChas[i].Roles[j+1:]...)
						break
					}
				}
			}
		}
	} else {
		for i, voiceCha := range misc.GuildMap[m.GuildID].GuildConfig.VoiceChas {
			if voiceCha.ID == cha.ID {
				misc.GuildMap[m.GuildID].GuildConfig.VoiceChas = append(misc.GuildMap[m.GuildID].GuildConfig.VoiceChas[:i], misc.GuildMap[m.GuildID].GuildConfig.VoiceChas[i+1:]...)
				chaDeleted = true
				break
			}
		}
	}

	misc.GuildSettingsWrite(misc.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	misc.MapMutex.Unlock()

	if chaDeleted {
		message = fmt.Sprintf("Success! Entire channel`%v` and all associated roles has been removed from the voice channel list.", cha.Name)
	} else {
		message = fmt.Sprintf("Success! Removed `%v` from voice channel `%v` in the voice channel list.", role.Name, cha.Name)
	}

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Prints all set voice channels and their associated role
func viewVoiceChaRoles(s *discordgo.Session, m *discordgo.Message) {

	var (
		message      string
		splitMessage []string
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	commandStrings := strings.SplitN(m.Content, " ", 2)

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"voiceroles`")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	misc.MapMutex.Lock()
	if len(misc.GuildMap[m.GuildID].GuildConfig.VoiceChas) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no set voice channel roles.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}

	for _, cha := range misc.GuildMap[m.GuildID].GuildConfig.VoiceChas {
		message += fmt.Sprintf("**%v : %v**\n\n", cha.Name, cha.ID)
		for _, role := range cha.Roles {
			message += fmt.Sprintf("`%v - %v`\n", role.Name, role.ID)
		}
		message += "——————\n"
	}
	misc.MapMutex.Unlock()

	// Splits the message if it's over 1900 characters
	if len(message) > 1900 {
		splitMessage = misc.SplitLongMessage(message)
	}

	// Prints split or unsplit
	if splitMessage == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}
	for i := 0; i < len(splitMessage); i++ {
		_, err := s.ChannelMessageSend(m.ChannelID, splitMessage[i])
		if err != nil {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot send voice channel roles message.")
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
		}
	}
}

// Handles vote category view or change
func voteCategoryCommand(s *discordgo.Session, m *discordgo.Message) {

	var message string

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog
	guildVoteCategory := misc.GuildMap[m.GuildID].GuildConfig.VoteChannelCategory
	misc.MapMutex.Unlock()

	mLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(mLowercase, " ", 2)

	// Displays current vote category
	if len(commandStrings) == 1 {
		if guildVoteCategory.ID == "" {
			message = fmt.Sprintf("Error: Vote Category is currently not set. Please use `%vvotecategory [category]`", guildPrefix)
		} else {
			message = fmt.Sprintf("Current Vote Category is: `%v - %v` \n\n To change Vote Category please use `%vvotecategory [category]`", guildVoteCategory.Name, guildVoteCategory.ID, guildPrefix)
		}

		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Parses category
	catID, catName := misc.CategoryParser(s, commandStrings[1], m.GuildID)
	if catID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such category exists.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Changes and writes new vote category to storage
	misc.MapMutex.Lock()
	misc.GuildMap[m.GuildID].GuildConfig.VoteChannelCategory.ID = catID
	misc.GuildMap[m.GuildID].GuildConfig.VoteChannelCategory.Name = catName
	misc.GuildSettingsWrite(misc.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	misc.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Vote Category is: `%v - %v`", catName, catID))
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Handles vote module disable or enable
func voteModuleCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		message string
		module  bool
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog
	guildVoteModule := misc.GuildMap[m.GuildID].GuildConfig.VoteModule
	misc.MapMutex.Unlock()

	mLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(mLowercase, " ", 2)

	// Displays current module setting
	if len(commandStrings) == 1 {
		if !guildVoteModule {
			message = fmt.Sprintf("Vote module is disabled. Please use `%vvotemodule true` to enable it.", guildPrefix)
		} else {
			message = fmt.Sprintf("Vote module is enabled. Please use `%vvotemodule false` to disable it.", guildPrefix)
		}
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	} else if len(commandStrings) > 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vvotemodule [true/false]`", guildPrefix))
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
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
			_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Changes and writes module bool to guild
	misc.MapMutex.Lock()
	misc.GuildMap[m.GuildID].GuildConfig.VoteModule = module
	misc.GuildSettingsWrite(misc.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	misc.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Handles waifu module disable or enable
func waifuModuleCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		message string
		module  bool
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog
	guildWaifuModule := misc.GuildMap[m.GuildID].GuildConfig.WaifuModule
	misc.MapMutex.Unlock()

	mLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(mLowercase, " ", 2)

	// Displays current module setting
	if len(commandStrings) == 1 {
		if !guildWaifuModule {
			message = fmt.Sprintf("Waifus module is disabled. Please use `%vwaifumodule true` to enable it.", guildPrefix)
		} else {
			message = fmt.Sprintf("Waifus module is enabled. Please use `%vwaifumodule false` to disable it.", guildPrefix)
		}
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	} else if len(commandStrings) > 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vwaifumodule [true/false]`", guildPrefix))
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
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
			_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Changes and writes module bool to guild
	misc.MapMutex.Lock()
	misc.GuildMap[m.GuildID].GuildConfig.WaifuModule = module
	misc.GuildSettingsWrite(misc.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	misc.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Handles react module disable or enable
func reactModuleCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		message string
		module  bool
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog
	guildReactsModule := misc.GuildMap[m.GuildID].GuildConfig.ReactsModule
	misc.MapMutex.Unlock()

	mLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(mLowercase, " ", 2)

	// Displays current module setting
	if len(commandStrings) == 1 {
		if !guildReactsModule {
			message = fmt.Sprintf("Reacts module is disabled. Please use `%vreactmodule true` to enable it.", guildPrefix)
		} else {
			message = fmt.Sprintf("Reacts module is enabled. Please use `%vreactmodule false` to disable it.", guildPrefix)
		}
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	} else if len(commandStrings) > 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vreactmodule [true/false]`", guildPrefix))
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
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
			_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Changes and writes module bool to guild
	misc.MapMutex.Lock()
	misc.GuildMap[m.GuildID].GuildConfig.ReactsModule = module
	misc.GuildSettingsWrite(misc.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	misc.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Handles attachment removal disable or enable
func whitelistFileFilter(s *discordgo.Session, m *discordgo.Message) {

	var (
		message   string
		module    bool
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog
	guildFileFilter := misc.GuildMap[m.GuildID].GuildConfig.WhitelistFileFilter
	misc.MapMutex.Unlock()

	mLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(mLowercase, " ", 2)

	// Displays current module setting
	if len(commandStrings) == 1 {
		if !guildFileFilter {
			message = fmt.Sprintf("Whitelist File Filter version is disabled. Using a Blacklist File Filter. Please use `%vwhitelist true` to enable it.", guildPrefix)
		} else {
			message = fmt.Sprintf("WhitelistFile Filter version is enabled. Please use `%vwhitelist false` to disable it and enable the Blacklist File Filter instead.", guildPrefix)
		}
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error())
			if err != nil {
				return
			}
			return
		}
		return
	} else if len(commandStrings) > 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vwhitelist [true/false]`", guildPrefix))
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error())
			if err != nil {
				return
			}
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
			_, err = s.ChannelMessageSend(guildBotLog.ID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Changes and writes module bool to guild
	misc.MapMutex.Lock()
	misc.GuildMap[m.GuildID].GuildConfig.WhitelistFileFilter = module
	misc.GuildSettingsWrite(misc.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	misc.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		_, _ = s.ChannelMessageSend(guildBotLog.ID, err.Error())
	}
}

// Handles ping message view or change
func pingMessageCommand(s *discordgo.Session, m *discordgo.Message) {

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	guildPingMessage := misc.GuildMap[m.GuildID].GuildConfig.PingMessage
	misc.MapMutex.Unlock()

	commandStrings := strings.SplitN(m.Content, " ", 2)

	// Displays current prefix
	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current ping message is: `%v` \n\n To change ping message please use `%vpingmessage [new ping]`", guildPingMessage, guildPrefix))
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Changes and writes new ping message to storage
	misc.MapMutex.Lock()
	misc.GuildMap[m.GuildID].GuildConfig.PingMessage = commandStrings[1]
	misc.GuildSettingsWrite(misc.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	misc.MapMutex.Unlock()

	guildPingMessage = commandStrings[1]

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! New ping message is: `%v`", guildPingMessage))
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

func init() {
	add(&command{
		execute:  addCommandRole,
		trigger:  "addcommandrole",
		aliases:  []string{"setcommandrole"},
		desc:     "Adds a privileged role",
		elevated: true,
		admin:    true,
		category: "settings",
	})
	add(&command{
		execute:  removeCommandRole,
		trigger:  "removecommandrole",
		aliases:  []string{"killcommandrole"},
		desc:     "Removes a privileged role",
		elevated: true,
		admin:    true,
		category: "settings",
	})
	add(&command{
		execute:  viewCommandRoles,
		trigger:  "commandroles",
		aliases:  []string{"vcommandroles", "viewcommandrole", "commandrole", "viewcommandroles", "showcommandroles"},
		desc:     "Prints all privileged roles",
		elevated: true,
		admin:    true,
		category: "settings",
	})
	add(&command{
		execute:  prefixCommand,
		trigger:  "prefix",
		desc:     "Views or changes the current prefix.",
		elevated: true,
		admin:    true,
		category: "settings",
	})
	add(&command{
		execute:  botLogCommand,
		trigger:  "botlog",
		desc:     "Views or changes the current Bot Log.",
		elevated: true,
		admin:    true,
		category: "settings",
	})
	add(&command{
		execute:  optInUnderCommand,
		trigger:  "optinunder",
		desc:     "Views or changes the current `Opt In Under` role.",
		elevated: true,
		admin:    true,
		category: "settings",
	})
	add(&command{
		execute:  optInAboveCommand,
		trigger:  "optinabove",
		desc:     "Views or changes the current `Opt In Above` role.",
		elevated: true,
		admin:    true,
		category: "settings",
	})
	add(&command{
		execute:  addVoiceChaRole,
		trigger:  "addvoice",
		aliases:  []string{"addvoicechannelrole", "addvoicecharole", "addvoicerole", "addvoicerole", "addvoicechannelrole", "addvoicerole"},
		desc:     "Sets a voice channel as one that will give users the specified role when they join it",
		elevated: true,
		admin:    true,
		category: "settings",
	})
	add(&command{
		execute:  removeVoiceChaRole,
		trigger:  "removevoice",
		aliases:  []string{"removevoicechannelrole", "removevoicechannelrole", "killvoicecharole", "killvoicechannelrole", "killvoicechannelidrole", "removevoicechannelidrole", "removevoicecharole", "removevoicerole", "removevoicerole", "killvoice"},
		desc:     "Stops a voice channel from giving its associated role on user join",
		elevated: true,
		admin:    true,
		category: "settings",
	})
	add(&command{
		execute:  viewVoiceChaRoles,
		trigger:  "voiceroles",
		aliases:  []string{"vvoicerole", "viewvoicechannels", "viewvoicechannel", "viewvoicechaids", "viewvoicechannelids", "viewvoivechannelid", "viewvoicecharole", "voicerole", "voicechannelroles", "viewvoicecharoles", "voice", "voices"},
		desc:     "Prints all set voice channels and their associated roles",
		elevated: true,
		admin:    true,
		category: "settings",
	})
	add(&command{
		execute:  voteCategoryCommand,
		trigger:  "votecategory",
		desc:     "Views or changes the current Vote Category where non-admin temp vote channels will be auto placed and sorted [VOTE]",
		elevated: true,
		admin:    true,
		category: "settings",
	})
	add(&command{
		execute:  voteModuleCommand,
		trigger:  "votemodule",
		aliases:  []string{"votemod"},
		desc:     "Vote Module. [VOTE]",
		elevated: true,
		admin:    true,
		category: "settings",
	})
	add(&command{
		execute:  waifuModuleCommand,
		trigger:  "waifumodule",
		aliases:  []string{"waifumod"},
		desc:     "Waifu Module. [WAIFU]",
		elevated: true,
		admin:    true,
		category: "settings",
	})
	add(&command{
		execute:  reactModuleCommand,
		trigger:  "reactmodule",
		aliases:  []string{"reactumod", "reactsmodule", "reactsmod"},
		desc:     "React Module. [REACTS]",
		elevated: true,
		admin:    true,
		category: "settings",
	})
	add(&command{
		execute:  whitelistFileFilter,
		trigger:  "whitelist",
		aliases:  []string{"filefilter", "attachmentremove", "attachremoval", "fileremove", "fileremoval", "attachmentremoval", "filesfilter", "whitelistfilter", "whitelistfile", "filewhitelist"},
		desc:     "Switch between a whitelist attachment file filter (removes all attachments except whitelisted ones) and a blacklist attachment file filter (remove only specified file extensions)",
		elevated: true,
		admin:    true,
		category: "settings",
	})
	add(&command{
		execute:  pingMessageCommand,
		trigger:  "pingmessage",
		desc:     "Views or changes the current ping message.",
		aliases:  []string{"pingmsg"},
		elevated: true,
		admin:    true,
		category: "settings",
	})
}
