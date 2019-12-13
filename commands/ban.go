package commands

import (
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Bans a user for a set period with a reason
func banCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		userID             string
		length             string
		reason             string
		success            string
		remaining          string
		commandStringsCopy []string

		validSlice         bool
		punishedUserExists bool

		temp functionality.PunishedUsers

		banTimestamp functionality.Punishment

		user *discordgo.User
	)

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 4)
	commandStringsCopy = commandStrings

	if len(commandStrings) != 4 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"ban [@user, userID, or username#discrim] [time] [reason]` format. \n\n"+
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
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid length. \n Usage: `"+guildSettings.Prefix+"ban [@user or userID] [time] [reason]` format. \n\n"+
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
		if functionality.HasElevatedPermissions(*s, userID, m.GuildID) {
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: Target user has a privileged role. Cannot ban.")
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
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: Cannot get this user. Cannot ban.")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		functionality.Mutex.Lock()
		functionality.InitializeUser(user, m.GuildID)
	}

	// Adds ban date to memberInfo and checks if perma
	functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Bans = append(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Bans, reason)
	UnbanDate, perma, err := functionality.ResolveTimeFromString(length)
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
		functionality.GuildMap[m.GuildID].MemberInfoMap[userID].UnbanDate = UnbanDate.Format("2006-01-02 15:04:05.999999999 -0700 MST")
	} else {
		functionality.GuildMap[m.GuildID].MemberInfoMap[userID].UnbanDate = "_Never_"
	}

	// Adds timestamp for that ban
	t, err := m.Timestamp.Parse()
	if err != nil {
		functionality.Mutex.Unlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	banTimestamp.Timestamp = t
	banTimestamp.Punishment = reason
	banTimestamp.Type = "Ban"
	functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps = append(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps, &banTimestamp)

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
		temp.UnbanDate = time.Date(9999, 9, 9, 9, 9, 9, 9, time.Local)
	} else {
		temp.UnbanDate = UnbanDate
	}

	// Adds or updates the now banned user in PunishedUsers
	for index, val := range functionality.GuildMap[m.GuildID].PunishedUsers {
		if val.ID == userID {
			temp.UnmuteDate = val.UnmuteDate
			functionality.GuildMap[m.GuildID].PunishedUsers[index] = &temp
			punishedUserExists = true
		}
	}
	if !punishedUserExists {
		functionality.GuildMap[m.GuildID].PunishedUsers = append(functionality.GuildMap[m.GuildID].PunishedUsers, &temp)
	}
	_ = functionality.PunishedUsersWrite(functionality.GuildMap[m.GuildID].PunishedUsers, m.GuildID)
	functionality.Mutex.Unlock()

	// Parses how long is left of the ban
	now := time.Now()
	remainingUnformatted := temp.UnbanDate.Sub(now)
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

	// Assigns success ban print string for user
	if perma && m.GuildID == "267799767843602452" {
		success = "You have been banned from " + guild.Name + ": **" + reason + "**\n\nUntil: _Forever_ \n\nIf you would like to appeal, use modmail at <https://reddit.com/r/anime>"
	} else if perma {
		success = "You have been banned from " + guild.Name + ": **" + reason + "**\n\nUntil: _Forever_"
	} else {
		z, _ := time.Now().Zone()
		success = "You have been banned from " + guild.Name + ": **" + reason + "**\n\nUntil: _" + UnbanDate.Format("2006-01-02 15:04:05") + " " + z + "_\n" +
			"Remaining: " + remaining
	}

	// Sends success string to user in DMs if able
	dm, _ := s.UserChannelCreate(userID)
	_, _ = s.ChannelMessageSend(dm.ID, success)

	// Bans the user
	err = s.GuildBanCreateWithReason(m.GuildID, userID, reason, 0)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Sends embed channel message
	if userMem != nil {
		err = functionality.BanEmbed(s, m, userMem.User, reason, UnbanDate, perma, m.ChannelID)
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	} else {
		err = functionality.BanEmbed(s, m, user, reason, UnbanDate, perma, m.ChannelID)
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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
		err = functionality.BanEmbed(s, m, userMem.User, reason, UnbanDate, perma, guildSettings.BotLog.ID)
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
		err = functionality.BanEmbed(s, m, user, reason, UnbanDate, perma, guildSettings.BotLog.ID)
		if err != nil {
			return
		}
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    banCommand,
		Trigger:    "ban",
		Aliases:    []string{"b", "hammer"},
		Desc:       "Bans a user for a set period of time",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
}
