package commands

import (
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

// Bans a user for a set period with a reason
func banCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		userID    string
		length    string
		reason    string
		success   string
		remaining string
		commandStringsCopy []string

		validSlice bool
		punishedUserExists bool

		temp misc.PunishedUsers

		banTimestamp misc.Punishment

		user	*discordgo.User
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	commandStrings := strings.SplitN(m.Content, " ", 4)
	commandStringsCopy = commandStrings

	if len(commandStrings) != 4 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"ban [@user, userID, or username#discrim] [time] [reason]` format. \n\n"+
			"Time is in #w#d#h#m format, such as 2w1d12h30m for 2 weeks, 1 day, 12 hours, 30 minutes. Use 0d for permanent.\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
		return
	}

	// Check if the time is the 2nd parameter and handle that
	_, _, err := misc.ResolveTimeFromString(commandStrings[1])
	if err == nil {
		length = commandStrings[1]
		commandStrings = append(commandStrings[:1], commandStrings[1+1:]...)
	}
	// Handle userID, reason and length
	userID, err = misc.GetUserID(m, commandStrings)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	if length == "" {
		length = commandStrings[2]
	} else {
		commandStrings = commandStringsCopy
	}

	reason = commandStrings[3]
	// Checks if the reason contains a mention and finds the actual name instead of ID
	reason = misc.MentionParser(s, reason, m.GuildID)

	// Checks if a number is contained in length var. Fixes some cases of invalid length
	lengthSlice := strings.Split(length, "")
	for i := 0; i < len(lengthSlice); i++ {
		if _, err := strconv.ParseInt(lengthSlice[i], 10, 64); err == nil || lengthSlice[i] == "∞" {
			validSlice = true
			break
		}
	}
	if !validSlice {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid length. \n Usage: `"+guildPrefix+"ban [@user or userID] [time] [reason]` format. \n\n"+
			"Time is in #w#d#h#m format, such as 2w1d12h30m for 2 weeks, 1 day, 12 hours, 30 minutes. Use 0d for permanent.\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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
		misc.MapMutex.Lock()
		if HasElevatedPermissions(s, userMem.User.ID, m.GuildID) {
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: Target user has a privileged role. Cannot ban.")
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
	}

	// Checks if user is in memberInfo and handles them
	misc.MapMutex.Lock()
	if _, ok := misc.GuildMap[m.GuildID].MemberInfoMap[userID]; !ok {
		if userMem != nil {
			// Initializes user if he doesn't exist in memberInfo but is in server
			misc.InitializeUser(userMem, m.GuildID)
		}
	}

	// Handles user if he's not in the server
	if userMem == nil {
		user, err = s.User(userID)
		if err != nil {
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: Cannot get this user. Cannot ban.")
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error())
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
	}

	// Adds ban date to memberInfo and checks if perma
	misc.GuildMap[m.GuildID].MemberInfoMap[userID].Bans = append(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Bans, reason)
	UnbanDate, perma, err := misc.ResolveTimeFromString(length)
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid time given.")
		if err != nil {
			_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	if commandStrings[2] == "∞" || commandStrings[1] == "∞" {
		perma = true
	}
	if !perma {
		misc.GuildMap[m.GuildID].MemberInfoMap[userID].UnbanDate = UnbanDate.Format("2006-01-02 15:04:05.999999999 -0700 MST")
	} else {
		misc.GuildMap[m.GuildID].MemberInfoMap[userID].UnbanDate = "_Never_"
	}

	// Adds timestamp for that ban
	t, err := m.Timestamp.Parse()
	if err != nil {
		misc.MapMutex.Unlock()
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
	banTimestamp.Timestamp = t
	banTimestamp.Punishment = reason
	banTimestamp.Type = "Ban"
	misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps = append(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps, &banTimestamp)

	// Writes to memberInfo.json
	misc.WriteMemberInfo(misc.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)

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
	for index, val := range misc.GuildMap[m.GuildID].PunishedUsers {
		if val.ID == userID {
			temp.UnmuteDate = val.UnmuteDate
			misc.GuildMap[m.GuildID].PunishedUsers[index] = &temp
			punishedUserExists = true
		}
	}
	if !punishedUserExists {
		misc.GuildMap[m.GuildID].PunishedUsers = append(misc.GuildMap[m.GuildID].PunishedUsers, &temp)
	}
	_ = misc.PunishedUsersWrite(misc.GuildMap[m.GuildID].PunishedUsers, m.GuildID)
	misc.MapMutex.Unlock()

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
		misc.CommandErrorHandler(s, m, err, guildBotLog)
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
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Sends embed channel message
	if userMem != nil {
		err = BanEmbed(s, m, userMem.User, reason, UnbanDate, perma, m.ChannelID)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
	} else {
		err = BanEmbed(s, m, user, reason, UnbanDate, perma, m.ChannelID)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
	}

	// Sends embed bot-log message
	if userMem != nil {
		err = BanEmbed(s, m, userMem.User, reason, UnbanDate, perma, guildBotLog)
		if err != nil {
			return
		}
	} else {
		err = BanEmbed(s, m, user, reason, UnbanDate, perma, guildBotLog)
		if err != nil {
			return
		}
	}
}

func BanEmbed(s *discordgo.Session, m *discordgo.Message, mem *discordgo.User, reason string, length time.Time, perma bool, channelID string) error {

	var (
		embedMess      discordgo.MessageEmbed
		embedThumbnail discordgo.MessageEmbedThumbnail
		embedFooter    discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embedField       []*discordgo.MessageEmbedField
		embedFieldUserID discordgo.MessageEmbedField
		embedFieldReason discordgo.MessageEmbedField
	)

	// Sets timestamp for unban date and footer
	banDate := length.Format(time.RFC3339)
	embedMess.Timestamp = banDate
	embedFooter.Text = "Unban Date"
	embedMess.Footer = &embedFooter

	// Sets ban embed color
	embedMess.Color = 0xff0000

	// Saves user avatar as thumbnail
	embedThumbnail.URL = mem.AvatarURL("128")

	// Sets field titles
	embedFieldUserID.Name = "User ID:"
	embedFieldReason.Name = "Reason:"

	// Sets field content
	embedFieldUserID.Value = mem.ID
	embedFieldReason.Value = reason

	// Adds the two fields to embedField slice (because embedMess.Fields requires slice input)
	embedField = append(embedField, &embedFieldUserID)
	embedField = append(embedField, &embedFieldReason)

	// Sets embed title and its description (which it uses the same way as a field)
	if !perma {
		embedMess.Title = mem.Username + "#" + mem.Discriminator + " was banned by " + m.Author.Username
	} else {
		embedMess.Title = mem.Username + "#" + mem.Discriminator + " was permabanned by " + m.Author.Username
	}

	// Adds user thumbnail and the two other fields as well
	embedMess.Thumbnail = &embedThumbnail
	embedMess.Fields = embedField

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(channelID, &embedMess)
	return err
}

func init() {
	add(&command{
		execute:  banCommand,
		trigger:  "ban",
		aliases:  []string{"b", "hammer"},
		desc:     "Bans a user for a set period of time",
		elevated: true,
		category: "moderation",
	})
}
