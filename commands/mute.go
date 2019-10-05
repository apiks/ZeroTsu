package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/misc"
	"strconv"
	"strings"
	"time"
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

		validSlice bool
		punishedUserExists bool
		gaveRole bool

		temp misc.PunishedUsers

		muteTimestamp misc.Punishment

		user *discordgo.User
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	guildMutedRoleID := misc.GuildMap[m.GuildID].GuildConfig.MutedRole.ID
	misc.MapMutex.Unlock()

	commandStrings := strings.SplitN(m.Content, " ", 4)
	commandStringsCopy = commandStrings

	if len(commandStrings) != 4 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"mute [@user, userID, or username#discrim] [time] [reason]` format. \n\n"+
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
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid length. \n Usage: `"+guildPrefix+"mute [@user or userID] [time] [reason]` format. \n\n"+
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
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: Target user has a privileged role. Cannot mute.")
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
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: Cannot get this user. Cannot mute.")
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error())
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
	}

	// Adds mute date to memberInfo and checks if perma
	misc.GuildMap[m.GuildID].MemberInfoMap[userID].Mutes = append(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Mutes, reason)
	UnmuteDate, perma, err := misc.ResolveTimeFromString(length)
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
		misc.GuildMap[m.GuildID].MemberInfoMap[userID].UnmuteDate = UnmuteDate.Format("2006-01-02 15:04:05.999999999 -0700 MST")
	} else {
		misc.GuildMap[m.GuildID].MemberInfoMap[userID].UnmuteDate = "_Never_"
	}

	// Adds timestamp for that mute
	t, err := m.Timestamp.Parse()
	if err != nil {
		misc.MapMutex.Unlock()
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
	muteTimestamp.Timestamp = t
	muteTimestamp.Punishment = reason
	muteTimestamp.Type = "Mute"
	misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps = append(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps, muteTimestamp)

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
		temp.UnmuteDate = time.Date(9999, 9, 9, 9, 9, 9, 9, time.Local)
	} else {
		temp.UnmuteDate = UnmuteDate
	}

	// Adds or updates the now muted user in PunishedUsers
	for index, val := range misc.GuildMap[m.GuildID].PunishedUsers {
		if val.ID == userID {
			temp.UnbanDate = val.UnbanDate
			misc.GuildMap[m.GuildID].PunishedUsers[index] = temp
			punishedUserExists = true
		}
	}
	if !punishedUserExists {
		misc.GuildMap[m.GuildID].PunishedUsers = append(misc.GuildMap[m.GuildID].PunishedUsers, temp)
	}
	_ = misc.PunishedUsersWrite(misc.GuildMap[m.GuildID].PunishedUsers, m.GuildID)
	misc.MapMutex.Unlock()

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
		misc.CommandErrorHandler(s, m, err, guildBotLog)
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
			misc.CommandErrorHandler(s, m, err, guildBotLog)
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
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: This server does not have a set muted role. Please use `%vsetmuted [Role ID]` before trying this command again.", guildPrefix))
		if err != nil {
			_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			return
		}
		return
	}

	// Sends success string to user in DMs if possible
	dm, _ := s.UserChannelCreate(userID)
	_, _ = s.ChannelMessageSend(dm.ID, success)

	// Sends embed channel message
	if userMem != nil {
		err = MuteEmbed(s, m, userMem.User, reason, UnmuteDate, perma, m.ChannelID)
		if err != nil {
			return
		}
	} else {
		err = MuteEmbed(s, m, user, reason, UnmuteDate, perma, m.ChannelID)
		if err != nil {
			return
		}
	}
}

func MuteEmbed(s *discordgo.Session, m *discordgo.Message, mem *discordgo.User, reason string, length time.Time, perma bool, channelID string) error {

	var (
		embedMess      discordgo.MessageEmbed
		embedThumbnail discordgo.MessageEmbedThumbnail
		embedFooter    discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embedField       []*discordgo.MessageEmbedField
		embedFieldUserID discordgo.MessageEmbedField
		embedFieldReason discordgo.MessageEmbedField
	)

	// Sets timestamp for unmute date and footer
	unmuteDate := length.Format(time.RFC3339)
	embedMess.Timestamp = unmuteDate
	embedFooter.Text = "Unmute Date"
	embedMess.Footer = &embedFooter

	// Sets mute embed color
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
		embedMess.Title = mem.Username + "#" + mem.Discriminator + " was muted by " + m.Author.Username
	} else {
		embedMess.Title = mem.Username + "#" + mem.Discriminator + " was permamutted by " + m.Author.Username
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
		execute:  muteCommand,
		trigger:  "mute",
		aliases:  []string{"m", "muted", "shut"},
		desc:     "Mutes a user for a set period of time",
		elevated: true,
		category: "moderation",
	})
}
