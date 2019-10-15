package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/functionality"
)

// Sends memberInfo user information to channel
func whoisCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		pastUsernames string
		pastNicknames string
		warnings      string
		mutes         string
		kicks         string
		bans          string
		unbanDate     string
		splitMessage  []string
		isInsideGuild = true
		creationDate  time.Time
	)

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(messageLowercase, " ", 2)

	if len(commandStrings) < 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"whois [@user, userID, or username#discrim]`\n\n"+
			"Note: this command supports username#discrim where username contains spaces.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	userID, err := functionality.GetUserID(m, commandStrings)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Fetches user from server if possible and sets whether they're inside the server
	mem, err := s.State.Member(m.GuildID, userID)
	if err != nil {
		mem, err = s.GuildMember(m.GuildID, userID)
		if err != nil {
			isInsideGuild = false
		}
	}

	// Checks if user is in MemberInfo and assigns to user variable. Else initializes user.
	functionality.MapMutex.Lock()
	user, ok := functionality.GuildMap[m.GuildID].MemberInfoMap[userID]
	if !ok {
		if mem == nil {
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in server and internal database. Cannot whois until user joins the server.")
			if err != nil {
				functionality.MapMutex.Unlock()
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			functionality.MapMutex.Unlock()
			return
		}

		// Initializes user if he doesn't exist and is in server
		functionality.InitializeUser(mem, m.GuildID)
		user = functionality.GuildMap[m.GuildID].MemberInfoMap[userID]
		functionality.WriteMemberInfo(functionality.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	}
	functionality.MapMutex.Unlock()

	// Puts past usernames into a string
	if len(user.PastUsernames) != 0 {
		for i := 0; i < len(user.PastUsernames); i++ {
			if len(pastUsernames) == 0 {
				pastUsernames = user.PastUsernames[i]
			} else {
				pastUsernames = pastUsernames + ", " + user.PastUsernames[i]
			}
		}
	} else {
		pastUsernames = "None"
	}

	// Puts past nicknames into a string
	if len(user.PastNicknames) != 0 {
		for i := 0; i < len(user.PastNicknames); i++ {
			if len(pastNicknames) == 0 {
				pastNicknames = user.PastNicknames[i]
			} else {
				pastNicknames = pastNicknames + ", " + user.PastNicknames[i]
			}
		}
	} else {
		pastNicknames = "None"
	}

	// Puts warnings into a slice
	if len(user.Warnings) != 0 {
		for i := 0; i < len(user.Warnings); i++ {
			if len(warnings) == 0 {
				// Converts index to string and appends warning
				iStr := strconv.Itoa(i + 1)
				warnings = user.Warnings[i] + " [" + iStr + "]"
			} else {
				// Converts index to string and appends new warning to old ones
				iStr := strconv.Itoa(i + 1)
				warnings = warnings + ", " + user.Warnings[i] + " [" + iStr + "]"

			}
		}
	} else {
		warnings = "None"
	}

	// Puts mutes into a slice
	if len(user.Mutes) != 0 {
		for i := 0; i < len(user.Mutes); i++ {
			if len(mutes) == 0 {
				// Converts index to string and appends warning
				iStr := strconv.Itoa(i + 1)
				mutes = user.Mutes[i] + " [" + iStr + "]"
			} else {
				// Converts index to string and appends new warning to old ones
				iStr := strconv.Itoa(i + 1)
				mutes = mutes + ", " + user.Mutes[i] + " [" + iStr + "]"

			}
		}
	} else {
		mutes = "None"
	}

	// Puts kicks into a slice
	if len(user.Kicks) != 0 {
		for i := 0; i < len(user.Kicks); i++ {
			if len(kicks) == 0 {
				// Converts index to string and appends kick
				iStr := strconv.Itoa(i + 1)
				kicks = user.Kicks[i] + " [" + iStr + "]"
			} else {
				// Converts index to string and appends new kick to old ones
				iStr := strconv.Itoa(i + 1)
				kicks = kicks + ", " + user.Kicks[i] + " [" + iStr + "]"
			}
		}
	} else {
		kicks = "None"
	}

	// Puts bans into a slice
	if len(user.Bans) != 0 {
		for i := 0; i < len(user.Bans); i++ {
			if len(bans) == 0 {
				// Converts index to string and appends ban
				iStr := strconv.Itoa(i + 1)
				bans = user.Bans[i] + " [" + iStr + "]"
			} else {
				// Converts index to string and appends new ban to old ones
				iStr := strconv.Itoa(i + 1)
				bans = bans + ", " + user.Bans[i] + " [" + iStr + "]"
			}
		}
	} else {
		bans = "None"
	}

	// Puts unban Date into a separate string variable
	unbanDate = user.UnbanDate
	if unbanDate == "" {
		unbanDate = "No Ban"
	}

	// Fetches account creation time
	creationDate, err = functionality.CreationTime(userID)
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}

	// Sets whois message
	message := "**User:** " + user.Username + "#" + user.Discrim + " | **ID:** " + user.ID +
		"\n\n**Past Usernames:** " + pastUsernames +
		"\n\n**Past Nicknames:** " + pastNicknames + "\n\n**Warnings:** " + warnings +
		"\n\n**Mutes:** " + mutes +
		"\n\n**Kicks:** " + kicks + "\n\n**Bans:** " + bans +
		"\n\n**Join Date:** " + user.JoinDate
	if config.Website != "" {
		message += "\n\n**Verification Date:** " + user.VerifiedDate
	}
	message += "\n\n**Account Creation Date:** " + creationDate.String()

	// Sets reddit Username if it exists
	if config.Website != "" {
		if user.RedditUsername != "" {
			message = message + "\n\n**Reddit Account:** " + "<https://reddit.com/u/" + user.RedditUsername + ">"
		} else {
			message += "\n\n**Reddit Account:** " + "None"
		}
	}

	// Sets unban date if it exists
	if user.UnbanDate != "" {
		message += "\n\n**Unban Date:** " + user.UnbanDate
	}

	if !isInsideGuild {
		message += "\n\n**_User is not in the server._**"
	}

	// Alt check
	if config.Website != "" {
		functionality.MapMutex.Lock()
		alts := CheckAltAccountWhois(userID, m.GuildID)

		// If there's more than one account with the same reddit username add to whois message
		if len(alts) > 1 {
			// Forms the alts string
			success := "\n\n**Alts:**\n"
			for _, altID := range alts {
				success += fmt.Sprintf("%v#%v | %v\n", functionality.GuildMap[m.GuildID].MemberInfoMap[altID].Username, functionality.GuildMap[m.GuildID].MemberInfoMap[altID].Discrim, altID)
			}

			// Adds the alts to the whois message
			message += success
			alts = nil
		}
		functionality.MapMutex.Unlock()
	}

	// Checks if the message contains a mention and finds the actual name instead of ID
	message = functionality.MentionParser(s, message, m.GuildID)

	// Splits the message if it's over 1900 characters
	if len(message) > 1900 {
		splitMessage = functionality.SplitLongMessage(message)
	}

	// Prints split or unsplit whois
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
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot send whois message.")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
		}
	}
}

// Function that iterates through memberInfo.json and checks for any alt accounts for that ID. Whois version
func CheckAltAccountWhois(id string, guildID string) []string {

	var alts []string

	// Stops func if target reddit username is nil
	if functionality.GuildMap[guildID].MemberInfoMap[id].RedditUsername == "" {
		return nil
	}

	// Iterates through all users in memberInfo.json
	for _, user := range functionality.GuildMap[guildID].MemberInfoMap {
		// Skips iteration if iteration reddit username is nil
		if user.RedditUsername == "" {
			continue
		}
		// Checks if the current user has the same reddit username as the entry parameter and adds to alts string slice if so
		if user.RedditUsername == functionality.GuildMap[guildID].MemberInfoMap[id].RedditUsername {
			alts = append(alts, user.ID)
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

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"timestamps [@user, userID, or username#discrim]`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	userID, err := functionality.GetUserID(m, commandStrings)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Checks if user is in MemberInfo and assigns to user variable. Else initializes user.
	functionality.MapMutex.Lock()
	user, ok := functionality.GuildMap[m.GuildID].MemberInfoMap[userID]
	if !ok {

		// Fetches user from server if possible
		mem, err := s.State.Member(m.GuildID, userID)
		if err != nil {
			mem, _ = s.GuildMember(m.GuildID, userID)
		}

		if mem == nil {
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in server and internal database. Cannot timestamp until they rejoin server.")
			if err != nil {
				functionality.MapMutex.Unlock()
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			functionality.MapMutex.Unlock()
			return
		}

		// Initializes user if he doesn't exist and is in server
		functionality.InitializeUser(mem, m.GuildID)
		user = functionality.GuildMap[m.GuildID].MemberInfoMap[userID]
		functionality.WriteMemberInfo(functionality.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	}
	functionality.MapMutex.Unlock()

	// Check if timestamps exist
	if len(user.Timestamps) == 0 {
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: No saved timestamps for that user.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Formats message
	for _, timestamp := range user.Timestamps {
		timezone, displacement := timestamp.Timestamp.Zone()
		message += fmt.Sprintf("**%v:** `%v` - _%v %v %v, %v:%v:%v %v+%v_\n", timestamp.Type, timestamp.Punishment, timestamp.Timestamp.Day(),
			timestamp.Timestamp.Month(), timestamp.Timestamp.Year(), timestamp.Timestamp.Hour(), timestamp.Timestamp.Minute(), timestamp.Timestamp.Second(), timezone, displacement)
	}

	// Splits messsage if too long
	msgs := functionality.SplitLongMessage(message)

	// Prints timestamps
	for index := range msgs {
		_, err = s.ChannelMessageSend(m.ChannelID, msgs[index])
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    whoisCommand,
		Trigger:    "whois",
		Desc:       "Print mod information about a user",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
	functionality.Add(&functionality.Command{
		Execute:    showTimestampsCommand,
		Trigger:    "timestamp",
		Aliases:    []string{"timestamps"},
		Desc:       "Prints all punishments for a user and their timestamps",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
}
