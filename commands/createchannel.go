package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

type channel struct {
	Name        string
	Category    string
	Type        string
	Description string
}

// Creates a named channel and a named role with parameters and checks for mod perms
func createChannelCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		muted            string
		airing           string
		tmute            string
		roleName         string
		descriptionSlice []string
		fixed            bool
		guildMutedRoleID string

		categoryNum = 0

		channel channel
		newRole *discordgo.Role

		descriptionEdit discordgo.ChannelEdit
		channelEdit     discordgo.ChannelEdit

		channelCreationData discordgo.GuildChannelCreateData
	)

	guildSettings := db.GetGuildSettings(m.GuildID)
	if guildSettings.GetMutedRole() != (entities.Role{}) && guildSettings.GetMutedRole().GetID() != "" {
		guildMutedRoleID = guildSettings.GetMutedRole().GetID()
	}

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vcreate [name] OPTIONAL[type] [categoryID] [description; must have at least one other non-name parameter]`\n\nFour type of parameters exist: `airing`, `temp`, `general` and `optin`. `Optin` is the default one. `Temp` gets auto-deleted after three hours of inactivity. Only `general` does not create a role", guildSettings.GetPrefix()))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	command := strings.Replace(strings.ToLower(m.Content), guildSettings.GetPrefix()+"create ", "", 1)
	commandStrings = strings.Split(command, " ")

	// Confirms whether optins exist
	err := common.OptInsHandler(s, m.ChannelID, m.GuildID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Checks if [category] and [type] exist and assigns them if they do and removes them from slice and command string
	for i := 0; i < len(commandStrings); i++ {
		_, err := strconv.Atoi(commandStrings[i])
		if len(commandStrings[i]) >= 17 && err == nil {
			channel.Category = commandStrings[i]
			commandStrings = append(commandStrings[:i], commandStrings[i+1:]...)
			command = strings.Join(commandStrings, " ")
			categoryNum = i
		}
	}
	for i := 0; i < len(commandStrings); i++ {
		if commandStrings[i] == "airing" ||
			commandStrings[i] == "general" ||
			commandStrings[i] == "opt-in" ||
			commandStrings[i] == "optin" ||
			commandStrings[i] == "temp" ||
			commandStrings[i] == "temporary" {
			if categoryNum != 0 {
				if categoryNum-1 != i {
					continue
				}
			}

			channel.Type = commandStrings[i]
			commandStrings = append(commandStrings[:i], commandStrings[i+1:]...)
			command = strings.Join(commandStrings, " ")
			fixed = true
		}
	}

	// If no other parameters exist, fixes a bug where it deletes [type] even if it's a channel name and not at the end of name
	if !fixed {
		if commandStrings[len(commandStrings)-1] == "airing" ||
			commandStrings[len(commandStrings)-1] == "general" ||
			commandStrings[len(commandStrings)-1] == "opt-in" ||
			commandStrings[len(commandStrings)-1] == "optin" ||
			commandStrings[len(commandStrings)-1] == "temp" ||
			commandStrings[len(commandStrings)-1] == "temporary" {

			channel.Type = commandStrings[len(commandStrings)-1]

			commandStrings = append(commandStrings[:len(commandStrings)-1], commandStrings[len(commandStrings):]...)
			command = strings.Join(commandStrings, " ")
		}
	}

	// If either [description] or [type] exist then checks if a description is also present
	if channel.Type != "" || channel.Category != "" {
		if channel.Category != "" {
			descriptionSlice = strings.SplitAfter(m.Content, channel.Category)
		} else {
			descriptionSlice = strings.SplitAfter(m.Content, channel.Type)
		}

		// Makes the description the second element of the slice above
		channel.Description = descriptionSlice[1]
		// Makes a copy of description that it puts to lowercase
		descriptionLowercase := strings.ToLower(channel.Description)
		// Removes description from command variable
		command = strings.Replace(command, descriptionLowercase, "", -1)
	}

	// Removes all hyphen prefixes and suffixes because discord cannot handle them
	for strings.HasPrefix(command, "-") || strings.HasSuffix(command, "-") {
		command = strings.TrimPrefix(command, "-")
		command = strings.TrimSuffix(command, "-")
	}

	// Creates the new channel of type text
	channelCreationData.Name = command
	if channel.Category != "" {
		channelCreationData.ParentID = channel.Category
	}
	if channel.Description != "" {
		channelCreationData.Topic = channel.Description
	}

	newCha, err := s.GuildChannelCreateComplex(m.GuildID, channelCreationData)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Handles role creation if not general type
	if channel.Type != "general" {

		// Creates the new role
		newRole, err = s.GuildRoleCreate(m.GuildID)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}

		// Sets role name to hyphenated form
		roleName = newCha.Name

		// Edits the new role with proper hyphenated name
		_, err = s.GuildRoleEdit(m.GuildID, newRole.ID, roleName, 0, false, 0, false)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}

		// Adds the role to the SpoilerMap and writes to storage
		db.SetGuildSpoilerRole(m.GuildID, &discordgo.Role{ID: newRole.ID, Name: command})
	}

	// Pulls info on server roles
	deb, err := s.GuildRoles(m.GuildID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Finds ID of muted role
	if guildMutedRoleID != "" {
		muted = guildMutedRoleID
	} else {
		for _, role := range deb {
			if strings.ToLower(role.Name) == "muted" || strings.ToLower(role.Name) == "t-mute" {
				muted = role.ID
				break
			}
		}
	}

	// Finds ID Airing role
	for i := 0; i < len(deb); i++ {
		if channel.Type == "airing" && strings.ToLower(deb[i].Name) == "airing" {
			airing = deb[i].ID
		}
	}

	// If it can't find airing role then create it
	if channel.Type == "airing" && airing == "" {
		_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Warning: Airing channel type was specified but no airing role was found. Creating role `airing`. Please use `%vsortroles` afterwards or manually put it between the dummy opt-in roles.", guildSettings.GetPrefix()))

		airingRole, err := s.GuildRoleCreate(m.GuildID)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		_, err = s.GuildRoleEdit(m.GuildID, airingRole.ID, "airing", 0, false, 0, false)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		airing = airingRole.ID

		// Adds the role to the SpoilerMap and writes to storage
		db.SetGuildSpoilerRole(m.GuildID, &discordgo.Role{ID: airingRole.ID, Name: "airing"})
	}

	// Assigns channel permission overwrites
	for _, goodRole := range guildSettings.GetCommandRoles() {
		// Mod perms
		_ = s.ChannelPermissionSet(newCha.ID, goodRole.GetID(), "role", common.FullSpoilerPerms, 0)
	}
	// Assign perms for the BOT
	err = s.ChannelPermissionSet(newCha.ID, s.State.User.ID, "member", common.FullSpoilerPerms, 0)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	if channel.Type != "general" {
		// Everyone perms
		err = s.ChannelPermissionSet(newCha.ID, m.GuildID, "role", discordgo.PermissionSendMessages, common.ReadSpoilerPerms)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		// Spoiler role perms
		err = s.ChannelPermissionSet(newCha.ID, newRole.ID, "role", common.ReadSpoilerPerms, 0)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}

	// Muted perms
	if muted != "" {
		err = s.ChannelPermissionSet(newCha.ID, muted, "role", 0, discordgo.PermissionSendMessages)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}
	if tmute != "" {
		err = s.ChannelPermissionSet(newCha.ID, tmute, "role", 0, discordgo.PermissionSendMessages)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}

	// Airing perms
	if channel.Type == "airing" && airing != "" {
		err = s.ChannelPermissionSet(newCha.ID, airing, "role", common.ReadSpoilerPerms, 0)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}

	// Category Permissions that overwrite if needed
	if channel.Category != "" {
		category, err := s.Channel(channel.Category)
		if err == nil {
			for _, catPerm := range category.PermissionOverwrites {

				// Special behavior for everyone perm
				if catPerm.ID == m.GuildID {
					err = s.ChannelPermissionSet(newCha.ID, catPerm.ID, "role", catPerm.Allow, catPerm.Deny|discordgo.PermissionReadMessages)
					if err != nil {
						common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
						return
					}
					continue
				}

				err = s.ChannelPermissionSet(newCha.ID, catPerm.ID, "role", catPerm.Allow, catPerm.Deny)
				if err != nil {
					err = s.ChannelPermissionSet(newCha.ID, catPerm.ID, "member", catPerm.Allow, catPerm.Deny)
					if err != nil {
						common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
						return
					}
				}
			}
		}
	}

	// Sets channel description if it exists
	if channel.Description != "" {
		descriptionEdit.Topic = channel.Description
		_, err = s.ChannelEditComplex(newCha.ID, &descriptionEdit)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}

	if channel.Type == "temp" || channel.Type == "temporary" {
		t := time.Now()
		temp := entities.NewTempChaInfo(t, roleName, true)

		guildVoteInfoMap := db.GetGuildVoteInfo(m.GuildID)
		for _, vote := range guildVoteInfoMap {
			if vote == nil {
				continue
			}

			if roleName == vote.GetChannel() {
				if !functionality.HasElevatedPermissions(s, vote.GetUser().ID, m.GuildID) {
					temp.SetElevated(false)
					break
				}
			}
		}

		err = db.SetGuildTempChannel(m.GuildID, newRole.ID, temp)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}

	// Parses category from name or ID
	if channel.Category != "" {
		// Pulls info on server channel
		chaAll, err := s.GuildChannels(m.GuildID)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		for i := 0; i < len(chaAll); i++ {
			// Puts channel name to lowercase
			nameLowercase := strings.ToLower(chaAll[i].Name)
			// Compares if Category is either a valid category name or ID
			if nameLowercase == channel.Category || chaAll[i].ID == channel.Category {
				if chaAll[i].Type == discordgo.ChannelTypeGuildCategory {
					channel.Category = chaAll[i].ID
				}
			}
		}

		// Sets categoryID to the parentID
		channelEdit.ParentID = channel.Category

		// Pushes new parentID to channel
		_, err = s.ChannelEditComplex(newCha.ID, &channelEdit)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}

	// Mod-only message for non-startvote channels
	if m.Author.ID != s.State.User.ID {
		if roleName != "" {
			_, err = s.ChannelMessageSend(m.ChannelID, "Channel and role `"+roleName+"` created. If opt-in please sort in the roles list between the dummy roles or with `"+guildSettings.GetPrefix()+"sortroles` (warning, lags in big servers)."+
				" If you do not do this you cannot join the role with reacts or `"+guildSettings.GetPrefix()+"join`. Sort category separately.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
		} else {
			_, err = s.ChannelMessageSend(m.ChannelID, "Success! Channel created.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
		}
	}
}

func init() {
	Add(&Command{
		Execute:    createChannelCommand,
		Trigger:    "create",
		Desc:       "Creates a text channel with settings",
		Permission: functionality.Mod,
		Module:     "channel",
	})
}
