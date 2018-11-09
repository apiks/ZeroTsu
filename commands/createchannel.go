package commands

import (
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

type channel struct {
	Name		string
	Category	string
	Type		string
	Description	string
}

// Creates a named channel and a named role with parameters and checks for mod perms
func createChannelCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		muted            string
		airing           string
		roleName         string
		descriptionSlice []string
		fixed            bool

		categoryNum int

		channel channel

		descriptionEdit discordgo.ChannelEdit
		channelEdit     discordgo.ChannelEdit
	)

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+config.BotPrefix+"create [name] OPTIONAL[type] [category] [description; must have at least one other non-name parameter]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	command := strings.Replace(messageLowercase, config.BotPrefix+"create ", "", 1)
	commandStrings = strings.Split(command, " ")

	// Checks if [category] and [type] exist and assigns them if they do and removes them from slice and command string
	for i := 0; i < len(commandStrings); i++ {
		_, err := strconv.Atoi(commandStrings[i])
		if len(commandStrings[i]) >= 17 && err == nil {
			channel.Category = commandStrings[i]
			categoryNum = i
		}
	}
	for i := 0; i < len(commandStrings); i++ {
		if categoryNum != 0 {
			if (commandStrings[i] == "airing" ||
				commandStrings[i] == "general" ||
				commandStrings[i] == "opt-in" ||
				commandStrings[i] == "optin" ||
				commandStrings[i] == "temp" ||
				commandStrings[i] == "temporary") &&
				i == categoryNum-1 {

				channel.Type = commandStrings[i]
				commandStrings = append(commandStrings[:i], commandStrings[i+1:]...)
				commandStrings = append(commandStrings[:i], commandStrings[i+1:]...)
				command = strings.Join(commandStrings, " ")
				fixed = true
			}
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
	newCha, err := s.GuildChannelCreate(config.ServerID, command, 0)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	time.Sleep(500 * time.Millisecond)

	// Creates the new role
	newRole, err := s.GuildRoleCreate(config.ServerID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Sets role name to hyphenated form
	roleName = newCha.Name

	// Edits the new role with proper hyphenated name
	_, err = s.GuildRoleEdit(config.ServerID, newRole.ID, roleName, 0, false, 0, false)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Adds the role to the SpoilerMap and writes to storage
	tempRole := discordgo.Role{
		ID:   newRole.ID,
		Name: command,
	}

	if m.Author.ID == s.State.User.ID {
		misc.SpoilerMap[newRole.ID] = &tempRole
		misc.SpoilerRolesWrite(misc.SpoilerMap)
	} else {
		misc.MapMutex.Lock()
		misc.SpoilerMap[newRole.ID] = &tempRole
		misc.SpoilerRolesWrite(misc.SpoilerMap)
		misc.MapMutex.Unlock()
	}

	// Pulls info on server roles
	deb, err := s.GuildRoles(config.ServerID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}
	// Finds ID of Muted role and Airing role
	for i := 0; i < len(deb); i++ {
		if deb[i].Name == "Muted" {
			muted = deb[i].ID
		} else if channel.Type == "airing" && deb[i].Name == "airing" {
			airing = deb[i].ID
		}
	}

	// Assigns channel permission overwrites
	for _, goodRole := range config.CommandRoles {
		// Mod perms
		err = s.ChannelPermissionSet(newCha.ID, goodRole, "role", misc.SpoilerPerms, 0)
		if err != nil {
			misc.CommandErrorHandler(s, m, err)
			return
		}
	}

	time.Sleep(100 * time.Millisecond)

	if channel.Type != "general" {
		// Everyone perms
		err = s.ChannelPermissionSet(newCha.ID, config.ServerID, "role", 0, misc.SpoilerPerms)
		if err != nil {
			misc.CommandErrorHandler(s, m, err)
			return
		}
		// Spoiler role perms
		err = s.ChannelPermissionSet(newCha.ID, newRole.ID, "role", misc.SpoilerPerms, 0)
		if err != nil {
			misc.CommandErrorHandler(s, m, err)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Muted perms
	err = s.ChannelPermissionSet(newCha.ID, muted, "role", 0, discordgo.PermissionSendMessages)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	time.Sleep(100 * time.Millisecond)

	// Airing perms
	if channel.Type == "airing" {
		err = s.ChannelPermissionSet(newCha.ID, airing, "role", misc.SpoilerPerms, 0)
		if err != nil {
			misc.CommandErrorHandler(s, m, err)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Sets channel description if it exists
	if channel.Description != "" {
		descriptionEdit.Topic = channel.Description
		_, err = s.ChannelEditComplex(newCha.ID, &descriptionEdit)
		if err != nil {
			misc.CommandErrorHandler(s, m, err)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	if channel.Type == "temp" || channel.Type == "temporary" {
		t := time.Now()
		var temp TempChaInfo
		temp.RoleName = roleName
		temp.CreationDate = t
		temp.Elevated = true

		// Locks mutex based on whether the bot called the command or not
		if m.Author.ID == s.State.User.ID {
			for _, v := range VoteInfoMap {
				if roleName == v.Channel {
					if !hasElevatedPermissions(s, v.User) {
						temp.Elevated = false
						break
					}
				}
			}
			TempChaMap[newRole.ID] = &temp
			TempChaWrite(TempChaMap)
		} else {
			misc.MapMutex.Lock()
			for _, v := range VoteInfoMap {
				if roleName == v.Channel {
					if !hasElevatedPermissions(s, v.User) {
						temp.Elevated = false
						break
					}
				}
			}
			TempChaMap[newRole.ID] = &temp
			TempChaWrite(TempChaMap)
			misc.MapMutex.Unlock()
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Parses category from name or ID
	if channel.Category != "" {
		// Pulls info on server channel
		chaAll, err := s.GuildChannels(config.ServerID)
		if err != nil {
			misc.CommandErrorHandler(s, m, err)
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
			misc.CommandErrorHandler(s, m, err)
			return
		}
	}

	// If the message was called from StartVote, prints it for all to see, else mod-only message
	if m.Author.ID != s.State.User.ID {
		_, err = s.ChannelMessageSend(m.ChannelID, "Channel and role `"+roleName+"` created. If opt-in please sort in the roles list. Sort category separately.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		}
	}
}

func init() {
	add(&command{
		execute:  createChannelCommand,
		trigger:  "create",
		desc:     "Creates a channel.",
		elevated: true,
		category: "channel",
	})
}