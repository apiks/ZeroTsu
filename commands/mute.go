package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"ZeroTsu/functionality"
)

// Mutes a user for a set period with a reason
func muteCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		userID             string
		length             string
		reason             string
		success            string
		remaining          string
		commandStringsCopy []string
		guildMutedRoleID   string

		validSlice         bool
		punishedUserExists bool
		gaveRole           bool

		temp functionality.PunishedUsers

		muteTimestamp functionality.Punishment

		user *discordgo.User
	)

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	if guildSettings.MutedRole != nil {
		if guildSettings.MutedRole.ID != "" {
			guildMutedRoleID = guildSettings.MutedRole.ID
		}
	}

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 4)
	commandStringsCopy = commandStrings

	if len(commandStrings) != 4 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"mute [@user, userID, or username#discrim] [time] [reason]` format. \n\n"+
			"Time is in #w#d#h#m format, such as 2w1d12h30m for 2 weeks, 1 day, 12 hours, 30 minutes. Use 0d for permanent.\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Check if the time is the 2nd parameter and handle that
	_, _, err := functionality.ResolveTimeFromString(commandStrings[1])
	if err == nil {
		length = commandStrings[1]
		commandStrings = append(commandStrings[:1], commandStrings[1+1:]...)
	}
	// Handle userID, reason and length
	userID, err = functionality.GetUserID(m, commandStrings)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	if length == "" {
		length = commandStrings[2]
	} else {
		commandStrings = commandStringsCopy
	}

	reason = commandStrings[3]
	// Checks if the reason contains a mention and finds the actual name instead of ID
	reason = functionality.MentionParser(s, reason, m.GuildID)

	// Checks if a number is contained in length var. Fixes some cases of invalid length
	lengthSlice := strings.Split(length, "")
	for i := 0; i < len(lengthSlice); i++ {
		if _, err := strconv.ParseInt(lengthSlice[i], 10, 64); err == nil || lengthSlice[i] == "∞" {
			validSlice = true
			break
		}
	}
	if !validSlice {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid length. \n Usage: `"+guildSettings.Prefix+"mute [@user or userID] [time] [reason]` format. \n\n"+
			"Time is in #w#d#h#m format, such as 2w1d12h30m for 2 weeks, 1 day, 12 hours, 30 minutes. Use 0d for permanent.\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls info on user
	userMem, err := s.State.Member(m.GuildID, userID)
	if err != nil {
		userMem, _ = s.GuildMember(m.GuildID, userID)
	}
	// Checks if user has a privileged role
	if userMem != nil {
		if functionality.HasElevatedPermissions(s, userMem.User.ID, m.GuildID) {
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: Target user has a privileged role. Cannot mute.")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
	}

	// Checks if user is in memberInfo and handles them
	functionality.Mutex.Lock()
	if _, ok := functionality.GuildMap[m.GuildID].MemberInfoMap[userID]; !ok {
		if userMem != nil {
			// Initializes user if he doesn't exist in memberInfo but is in server
			functionality.InitializeMember(userMem, m.GuildID)
		}
	}

	// Handles user if he's not in the server
	if userMem == nil {
		functionality.Mutex.Unlock()
		user, err = s.User(userID)
		if err != nil {
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: Cannot get this user. Cannot mute.")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		functionality.Mutex.Lock()
	}

	// Adds mute date to memberInfo and checks if perma
	functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Mutes = append(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Mutes, reason)
	UnmuteDate, perma, err := functionality.ResolveTimeFromString(length)
	if err != nil {
		functionality.Mutex.Unlock()
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid time given.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	if commandStrings[2] == "∞" || commandStrings[1] == "∞" {
		perma = true
	}
	if !perma {
		functionality.GuildMap[m.GuildID].MemberInfoMap[userID].UnmuteDate = UnmuteDate.Format("2006-01-02 15:04:05.999999999 -0700 MST")
	} else {
		functionality.GuildMap[m.GuildID].MemberInfoMap[userID].UnmuteDate = "_Never_"
	}

	// Adds timestamp for that mute
	t, err := m.Timestamp.Parse()
	if err != nil {
		functionality.Mutex.Unlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	muteTimestamp.Timestamp = t
	muteTimestamp.Punishment = reason
	muteTimestamp.Type = "Mute"
	functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps = append(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps, &muteTimestamp)

	// Writes to memberInfo.json
	_ = functionality.WriteMemberInfo(functionality.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)

	// Saves the details in temp
	temp.ID = userID
	if userMem != nil {
		temp.User = userMem.User.Username
	} else {
		temp.User = user.Username
	}

	if perma {
		temp.UnmuteDate = time.Date(9999, 9, 9, 9, 9, 9, 9, time.Local)
	} else {
		temp.UnmuteDate = UnmuteDate
	}

	// Adds or updates the now muted user in PunishedUsers
	for index, val := range functionality.GuildMap[m.GuildID].PunishedUsers {
		if val.ID == userID {
			temp.UnbanDate = val.UnbanDate
			functionality.GuildMap[m.GuildID].PunishedUsers[index] = &temp
			punishedUserExists = true
		}
	}
	if !punishedUserExists {
		functionality.GuildMap[m.GuildID].PunishedUsers = append(functionality.GuildMap[m.GuildID].PunishedUsers, &temp)
	}
	_ = functionality.PunishedUsersWrite(functionality.GuildMap[m.GuildID].PunishedUsers, m.GuildID)
	functionality.Mutex.Unlock()

	// Parses how long is left of the mute
	now := time.Now()
	remainingUnformatted := temp.UnmuteDate.Sub(now)
	if remainingUnformatted.Hours() < 1 {
		remaining = strconv.FormatFloat(remainingUnformatted.Minutes(), 'f', 0, 64) + " minutes"
	} else if remainingUnformatted.Hours() < 24 {
		remaining = strconv.FormatFloat(remainingUnformatted.Hours(), 'f', 0, 64) + " hours"
	} else {
		remaining = strconv.FormatFloat(remainingUnformatted.Hours()/24, 'f', 0, 64) + " days"
	}

	// Pulls the guild name
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Assigns success mute print string for user
	if perma && m.GuildID == "267799767843602452" {
		success = "You have been muted from " + guild.Name + ": **" + reason + "**\n\nUntil: _Forever_ \n\nIf you would like to appeal, use modmail at <https://reddit.com/r/anime>"
	} else if perma {
		success = "You have been muted from " + guild.Name + ": **" + reason + "**\n\nUntil: _Forever_"
	} else {
		z, _ := time.Now().Zone()
		success = "You have been muted from " + guild.Name + ": **" + reason + "**\n\nUntil: _" + UnmuteDate.Format("2006-01-02 15:04:05") + " " + z + "_\n" +
			"Remaining: " + remaining
	}

	// Checks if the muted role is set and gives it to the user. If it's not then tries to find a muted role on its own
	if guildMutedRoleID != "" {
		_ = s.GuildMemberRoleAdd(m.GuildID, userID, guildMutedRoleID)
		gaveRole = true
	} else {
		// Pulls info on server roles
		deb, err := s.GuildRoles(m.GuildID)
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}

		// Checks by string for a muted role
		for _, role := range deb {
			if strings.ToLower(role.Name) == "muted" || strings.ToLower(role.Name) == "t-mute" {
				_ = s.GuildMemberRoleAdd(m.GuildID, userID, role.ID)
				gaveRole = true
				break
			}
		}
	}

	if !gaveRole {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: This server does not have a set muted role. Please use `%ssetmuted [Role ID]` before trying this command again.", guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Sends success string to user in DMs if possible
	dm, _ := s.UserChannelCreate(userID)
	_, _ = s.ChannelMessageSend(dm.ID, success)

	// Sends embed channel message
	if userMem != nil {
		err = functionality.MuteEmbed(s, m, userMem.User, reason, UnmuteDate, perma, m.ChannelID)
		if err != nil {
			return
		}
	} else {
		err = functionality.MuteEmbed(s, m, user, reason, UnmuteDate, perma, m.ChannelID)
		if err != nil {
			return
		}
	}

	// Sends embed bot-log message
	if userMem != nil {
		if guildSettings.BotLog == nil {
			return
		}
		if guildSettings.BotLog.ID == "" {
			return
		}
		err = functionality.MuteEmbed(s, m, userMem.User, reason, UnmuteDate, perma, guildSettings.BotLog.ID)
		if err != nil {
			return
		}
	} else {
		if guildSettings.BotLog == nil {
			return
		}
		if guildSettings.BotLog.ID == "" {
			return
		}
		err = functionality.MuteEmbed(s, m, user, reason, UnmuteDate, perma, guildSettings.BotLog.ID)
		if err != nil {
			return
		}
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    muteCommand,
		Trigger:    "mute",
		Aliases:    []string{"m", "muted", "shut"},
		Desc:       "Mutes a user for a set period of time",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
}
