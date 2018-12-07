package commands

import (
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Bans a user for a set period with a reason
func banCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		userID  		string
		length  		string
		reason  		string
		success 		string
		remaining		string

		validSlice 		bool

		temp 			misc.BannedUsers

		banTimestamp 	misc.Punishment
	)
	z, _ := time.Now().Zone()

	commandStrings := strings.SplitN(m.Content, " ", 4)

	if len(commandStrings) != 4 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "ban [@user, userID, or username#discrim] [time] [reason]` format. \n\n"+
			"Time is in #w#d#h#m format, such as 2w1d12h30m for 2 weeks, 1 day, 12 hours, 30 minutes. Use 0d for permanent.\n" +
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	userID, err := misc.GetUserID(s, m, commandStrings)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}
	length = commandStrings[2]
	reason = commandStrings[3]
	// Checks if the reason contains a mention and finds the actual name instead of ID
	reason = misc.MentionParser(s, reason)

	// Checks if a number is contained in length var. Fixes some cases of invalid length
	lengthSlice := strings.Split(length, "")
	for i := 0; i < len(lengthSlice); i++ {
		if _, err := strconv.ParseInt(lengthSlice[i], 10, 64); err == nil || lengthSlice[i] == "∞" {
			validSlice = true
			break
		}
	}
	if !validSlice {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid length. \n Usage: `" + config.BotPrefix + "ban [@user or userID] [time] [reason]` format. \n\n"+
			"Time is in #w#d#h#m format, such as 2w1d12h30m for 2 weeks, 1 day, 12 hours, 30 minutes. Use 0d for permanent.\n" +
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Fetches user
	mem, err := s.User(userID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Checks if user is in memberInfo and handles them
	misc.MapMutex.Lock()
	if len(misc.MemberInfoMap) == 0 || misc.MemberInfoMap[userID] == nil {
		// Pulls info on user if they're in the server
		userMem, err := s.State.Member(config.ServerID, mem.ID)
		if err != nil {
			userMem, err = s.GuildMember(config.ServerID, mem.ID)
			if err != nil {
				_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in server _and_ memberInfo. Cannot ban user until they join the server.")
				if err != nil {
					_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
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
		// Initializes user if he doesn't exist in memberInfo but is in server
		misc.InitializeUser(userMem)
	}

	// Adds unban date to memberInfo and checks if perma
	misc.MemberInfoMap[userID].Bans = append(misc.MemberInfoMap[userID].Bans, reason)
	UnbanDate, perma := misc.ResolveTimeFromString(length)
	if commandStrings[2] == "∞" {
		perma = true
	}
	if !perma {
		misc.MemberInfoMap[userID].UnbanDate = UnbanDate.Format("2006-01-02 15:04:05.999999999 -0700 MST")
	} else {
		misc.MemberInfoMap[userID].UnbanDate = "_Never_"
	}

	// Adds timestamp for that ban
	t, err := m.Timestamp.Parse()
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}
	banTimestamp.Timestamp = t
	banTimestamp.Punishment = reason
	banTimestamp.Type = "Ban"
	misc.MemberInfoMap[userID].Timestamps = append(misc.MemberInfoMap[userID].Timestamps, banTimestamp)
	misc.MapMutex.Unlock()

	// Writes to memberInfo.json
	misc.MemberInfoWrite(misc.MemberInfoMap)

	// Saves the details in temp
	temp.ID = userID
	temp.User = mem.Username

	if perma {
		temp.UnbanDate = time.Date(9999, 9, 9, 9, 9, 9, 9, time.Local)
		UnbanDate = time.Date(9999, 9, 9, 9, 9, 9, 9, time.Local)
	} else {
		temp.UnbanDate = UnbanDate
	}

	// Adds the now banned user to BannedUsersSlice
	misc.MapMutex.Lock()
	misc.BannedUsersSlice = append(misc.BannedUsersSlice, temp)
	misc.MapMutex.Unlock()

	// Pulls the guild Name
	guild, err := s.Guild(config.ServerID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

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

	// Assigns success ban print string for user
	if !perma {
		success = "You have been banned from " + guild.Name + ": **" + reason + "**\n\nUntil: _" + UnbanDate.Format("2006-01-02 15:04:05") + " " + z + "_\n" +
			"Remaining: " + remaining
	} else {
		success = "You have been banned from " + guild.Name + ": **" + reason + "**\n\nUntil: _Forever_ \n\nIf you would like to appeal, use modmail at <https://reddit.com/r/anime>"
	}

	// Sends success string to user in DMs if able
	dm, _ := s.UserChannelCreate(userID)
	_, _ = s.ChannelMessageSend(dm.ID, success)

	// Bans the user
	err = s.GuildBanCreateWithReason(config.ServerID, userID, reason, 0)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Sends embed bot-log message
	err = BanEmbed(s, m, mem, reason, UnbanDate, perma, config.BotLogID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Sends embed channel message
	err = BanEmbed(s, m, mem, reason, UnbanDate, perma, m.ChannelID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}
}

func BanEmbed(s *discordgo.Session, m *discordgo.Message, mem *discordgo.User, reason string, length time.Time, perma bool, channelID string) error {

	var (
		embedMess      discordgo.MessageEmbed
		embedThumbnail discordgo.MessageEmbedThumbnail
		embedFooter	   discordgo.MessageEmbedFooter

		// Embed slice and its fields
		embedField         []*discordgo.MessageEmbedField
		embedFieldUserID   discordgo.MessageEmbedField
		embedFieldReason   discordgo.MessageEmbedField
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
		desc:     "Bans a user for a set period of time.",
		elevated: true,
		category: "punishment",
	})
}