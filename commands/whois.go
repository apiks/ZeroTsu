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

// Sends memberInfo user information to channel
func whoisCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		pastUsernames  string
		pastNicknames  string
		warnings       []string
		mutes          []string
		kicks          []string
		bans           []string
		splitMessage   []string
		isInsideGuild  = true
		creationDate   time.Time
		messageBuilder strings.Builder
	)

	guildSettings := db.GetGuildSettings(m.GuildID)
	cmdStrs := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(cmdStrs) < 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"whois [@mem, userID, or username#discrim]`\n\n"+
			"Note: this command supports username#discrim where username contains spaces.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	userID, err := common.GetUserID(m, cmdStrs)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Fetches user from server if possible and sets whether they're inside the server
	userMem, err := s.State.Member(m.GuildID, userID)
	if err != nil {
		userMem, err = s.GuildMember(m.GuildID, userID)
		if err != nil {
			isInsideGuild = false
		}
	}

	// Checks if user is in memberInfo and fetches them
	mem := db.GetGuildMember(m.GuildID, userID)
	if mem.GetID() == "" {
		var user *discordgo.User
		if userMem != nil {
			user = userMem.User
		} else {
			user, err = s.User(userID)
			if err != nil {
				_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in the server, internal database and cannot fetch manually either. Cannot whois until user joins the server.")
				if err != nil {
					common.LogError(s, guildSettings.BotLog, err)
					return
				}
				return
			}
		}

		// Initializes user if he doesn't exist in memberInfo but is in server
		functionality.InitializeUser(user, m.GuildID)

		mem = db.GetGuildMember(m.GuildID, userID)
		if mem.GetID() == "" {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, fmt.Errorf("error: member object is empty"))
			return
		}
	}

	// Puts past usernames into a string
	if len(mem.GetPastUsernames()) != 0 {
		pastUsernames = strings.Join(mem.GetPastUsernames(), ", ")
	} else {
		pastUsernames = "None"
	}

	// Puts past nicknames into a string
	if len(mem.GetPastNicknames()) != 0 {
		pastNicknames = strings.Join(mem.GetPastNicknames(), ", ")
	} else {
		pastNicknames = "None"
	}

	// Puts warnings into a slice
	if len(mem.GetWarnings()) != 0 {
		for i, warning := range mem.GetWarnings() {
			var warningBuilder strings.Builder
			iStr := strconv.Itoa(i + 1)

			warningBuilder.WriteString(warning)
			warningBuilder.WriteString(" [")
			warningBuilder.WriteString(iStr)
			warningBuilder.WriteString("]")
			warnings = append(warnings, warningBuilder.String())
		}
	} else {
		warnings = append(warnings, "None")
	}

	// Puts mutes into a slice
	if len(mem.GetMutes()) != 0 {
		for i, mute := range mem.GetMutes() {
			var muteBuilder strings.Builder
			iStr := strconv.Itoa(i + 1)

			muteBuilder.WriteString(mute)
			muteBuilder.WriteString(" [")
			muteBuilder.WriteString(iStr)
			muteBuilder.WriteString("]")
			mutes = append(mutes, muteBuilder.String())
		}
	} else {
		mutes = append(mutes, "None")
	}

	// Puts kicks into a slice
	if len(mem.GetKicks()) != 0 {
		for i, kick := range mem.GetKicks() {
			var kickBuilder strings.Builder
			iStr := strconv.Itoa(i + 1)

			kickBuilder.WriteString(kick)
			kickBuilder.WriteString(" [")
			kickBuilder.WriteString(iStr)
			kickBuilder.WriteString("]")
			kicks = append(kicks, kickBuilder.String())
		}
	} else {
		kicks = append(kicks, "None")
	}

	// Puts bans into a slice
	if len(mem.GetBans()) != 0 {
		for i, ban := range mem.GetBans() {
			var banBuilder strings.Builder
			iStr := strconv.Itoa(i + 1)

			banBuilder.WriteString(ban)
			banBuilder.WriteString(" [")
			banBuilder.WriteString(iStr)
			banBuilder.WriteString("]")
			bans = append(bans, banBuilder.String())
		}
	} else {
		bans = append(bans, "None")
	}

	// Fetches account creation time
	creationDate, err = common.CreationTime(userID)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}

	// Sets whois message
	messageBuilder.WriteString("**User:** ")
	messageBuilder.WriteString(mem.GetUsername())
	messageBuilder.WriteString("#")
	messageBuilder.WriteString(mem.GetDiscrim())
	messageBuilder.WriteString(" | **ID:** ")
	messageBuilder.WriteString(mem.GetID())
	messageBuilder.WriteString("\n\n**Past Usernames:** ")
	messageBuilder.WriteString(pastUsernames)
	messageBuilder.WriteString("\n\n**Past Nicknames:** ")
	messageBuilder.WriteString(pastNicknames)
	messageBuilder.WriteString("\n\n**Warnings:** ")
	messageBuilder.WriteString(strings.Join(warnings, ", "))
	messageBuilder.WriteString("\n\n**Mutes:** ")
	messageBuilder.WriteString(strings.Join(mutes, ", "))
	messageBuilder.WriteString("\n\n**Kicks:** ")
	messageBuilder.WriteString(strings.Join(kicks, ", "))
	messageBuilder.WriteString("\n\n**Bans:** ")
	messageBuilder.WriteString(strings.Join(bans, ", "))
	messageBuilder.WriteString("\n\n**Join Date:** ")
	messageBuilder.WriteString(mem.GetJoinDate())
	messageBuilder.WriteString("\n\n**Account Creation Date:** ")
	messageBuilder.WriteString(creationDate.String())

	// Sets unban date if it exists
	if mem.GetUnbanDate() != "" {
		messageBuilder.WriteString("\n\n**Unban Date:** ")
		messageBuilder.WriteString(mem.GetUnbanDate())
	}

	// Sets unmute date if it exists
	if mem.GetUnmuteDate() != "" {
		messageBuilder.WriteString("\n\n**Unmute Date:** ")
		messageBuilder.WriteString(mem.GetUnmuteDate())
	}

	if !isInsideGuild {
		messageBuilder.WriteString("\n\n**_User is not in the server._**")
	}

	// Checks if the message contains a mention and finds the actual name instead of ID
	message := messageBuilder.String()
	message = common.MentionParser(s, message, m.GuildID)

	// Splits the message if it's over 1900 characters
	if len(message) > 1900 {
		splitMessage = common.SplitLongMessage(message)
	}

	// Prints split or unsplit whois
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
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot send whois message.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
	}
}

// Function that iterates through memberInfo.json and checks for any alt accounts for that ID. Whois version
func CheckAltAccountWhois(guildID string, user entities.UserInfo) []string {
	var alts []string

	// Stops func if target reddit username is nil
	if user.GetRedditUsername() == "" {
		return nil
	}

	// Iterates through all users in memberInfo.json
	memberInfo := db.GetGuildMemberInfo(guildID)
	if memberInfo == nil {
		return nil
	}

	for _, memberInfoUser := range memberInfo {
		// Skips iteration if iteration reddit username is nil
		if memberInfoUser.GetRedditUsername() == "" {
			continue
		}
		// Checks if the current user has the same reddit username as the entry parameter and adds to alts string slice if so
		if user.GetRedditUsername() == memberInfoUser.GetRedditUsername() {
			alts = append(alts, memberInfoUser.GetID())
		}
	}
	if len(alts) > 1 {
		return alts
	} else {
		return nil
	}
}

// Displays all punishments for that user with timestamps and type of punishment
func showTimestampsCommand(s *discordgo.Session, m *discordgo.Message) {
	var message string

	guildSettings := db.GetGuildSettings(m.GuildID)
	cmdStrs := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	if len(cmdStrs) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"timestamps [@user, userID, or username#discrim]`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	userID, err := common.GetUserID(m, cmdStrs)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Checks if user is in memberInfo and fetches them
	mem := db.GetGuildMember(m.GuildID, userID)
	if mem.GetID() == "" {
		var user *discordgo.User

		// Fetches user from server if possible and sets whether they're inside the server
		userMem, err := s.State.Member(m.GuildID, userID)
		if err != nil {
			userMem, _ = s.GuildMember(m.GuildID, userID)
		}

		if userMem != nil {
			user = userMem.User
		} else {
			user, err = s.User(userID)
			if err != nil {
				_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in the server, internal database and cannot fetch manually either. Cannot timestamp until user joins the server.")
				if err != nil {
					common.LogError(s, guildSettings.BotLog, err)
					return
				}
				return
			}
		}

		// Initializes user if he doesn't exist in memberInfo but is in server
		functionality.InitializeUser(user, m.GuildID)

		mem = db.GetGuildMember(m.GuildID, userID)
		if mem.GetID() == "" {
			err = fmt.Errorf("error: member object is empty")
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}

	// Check if timestamps exist
	if len(mem.GetTimestamps()) == 0 {
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: No saved timestamps for that user.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Formats message
	for _, timestamp := range mem.GetTimestamps() {
		timezone, displacement := timestamp.GetTimestamp().Zone()
		message += fmt.Sprintf("**%v:** `%v` - _%v %v %v, %v:%v:%v %v+%v_\n", timestamp.GetPunishmentType(), timestamp.GetPunishment(), timestamp.GetTimestamp().Day(),
			timestamp.GetTimestamp().Month(), timestamp.GetTimestamp().Year(), timestamp.GetTimestamp().Hour(), timestamp.GetTimestamp().Minute(), timestamp.GetTimestamp().Second(), timezone, displacement)
	}

	// Splits messsage if too long
	msgs := common.SplitLongMessage(message)

	// Prints timestamps
	for index := range msgs {
		_, err = s.ChannelMessageSend(m.ChannelID, msgs[index])
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
	}
}

func init() {
	Add(&Command{
		Execute:    whoisCommand,
		Trigger:    "whois",
		Desc:       "Print mod information about a user",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
	Add(&Command{
		Execute:    showTimestampsCommand,
		Trigger:    "timestamp",
		Aliases:    []string{"timestamps"},
		Desc:       "Prints all punishments for a user and their timestamps",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
}
