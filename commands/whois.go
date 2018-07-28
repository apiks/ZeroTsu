package commands

import (
	"strings"
	"math"
	"strconv"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Sends memberInfo user information to channel
func whoisCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		pastUsernames string
		pastNicknames string
		warnings      string
		kicks         string
		bans          string
		unbanDate     string
		splitMessage []string
	)

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	if len(commandStrings) != 2 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+config.BotPrefix+"whois [@user or userID]`")
		if err != nil {

			_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
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

	// Fetches user from server if possible
	mem, err := s.State.Member(config.ServerID, userID)
	if err != nil {
		mem, err = s.GuildMember(config.ServerID, userID)
		if err != nil {

			_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in memberInfo. Cannot whois until they join server.")
			if err != nil {

				_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
				if err != nil {

					return
				}
				return
			}
			return
		}
	}

	misc.MapMutex.Lock()
	// Checks if user is in MemberInfo and assigns to user variable. Else initializes user.
	user, ok := misc.MemberInfoMap[userID]
	if !ok {
		misc.MapMutex.Unlock()

		// Initializes user if he doesn't exist and is in server
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in memberInfo. Initializing and whoising empty user.")
		if err != nil {

			_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
			if err != nil {

				return
			}
			return
		}

		misc.InitializeUser(mem)
		misc.MemberInfoWrite(misc.MemberInfoMap)
	}
	misc.MapMutex.Unlock()

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
				warnings = user.Warnings[i] + "[" + iStr + "]"

			} else {

				// Converts index to string and appends new warning to old ones
				iStr := strconv.Itoa(i + 1)
				warnings = warnings + ", " + user.Warnings[i] + "[" + iStr + "]"

			}
		}
	} else {

		warnings = "None"
	}

	// Puts kicks into a slice
	if len(user.Kicks) != 0 {
		for i := 0; i < len(user.Kicks); i++ {

			if len(kicks) == 0 {

				// Converts index to string and appends kick
				iStr := strconv.Itoa(i + 1)
				kicks = user.Kicks[i] + "[" + iStr + "]"

			} else {

				// Converts index to string and appends new kick to old ones
				iStr := strconv.Itoa(i + 1)
				kicks = kicks + ", " + user.Kicks[i] + "[" + iStr + "]"

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
				bans = user.Bans[i] + "[" + iStr + "]"

			} else {

				// Converts index to string and appends new ban to old ones
				iStr := strconv.Itoa(i + 1)
				bans = bans + ", " + user.Bans[i] + "[" + iStr + "]"
			}
		}
	} else {

		bans = "None"
	}

	// Puts unban Date into a separate string variable
	unbanDate = user.UnbanDate
	if unbanDate == "" {

		unbanDate = "User has never been banned."
	}

	// Sets whois message
	message := "**User:** " + user.Username + "#" + user.Discrim + "\n\n**Past Usernames:** " + pastUsernames +
		"\n\n**Past Nicknames:** " + pastNicknames + "\n\n**Warnings:** " + warnings +
		"\n\n**Kicks:** " + kicks + "\n\n**Bans:** " + bans +
		"\n\n**Join Date:** " + user.JoinDate + "\n\n**Verification Date:** " +
		user.VerifiedDate

	// Sets reddit Username if it exists
	if user.RedditUsername != "" {

		message = message + "\n\n**Reddit Account:** " + "<https://reddit.com/u/" + user.RedditUsername + ">"
	} else {

		message = message + "\n\n**Reddit Account:** " + "None"
	}

	// Sets unban date if it exists
	if user.UnbanDate != "" {

		message = message + "\n\n**Unban Date:** " + user.UnbanDate
	}

	// Alt check
	alts := CheckAltAccountWhois(userID)

	// If there's more than one account with that reddit username print a message
	if len(alts) > 1 {

		// Forms the success string
		success := "\n\n**Alts:** \n\n"
		for i := 0; i < len(alts); i++ {

			success = success + "<@" + alts[i] + "> \n"
		}

		// Adds the alts to the whois message
		message = message + success

		// Resets alts variable
		alts = nil
	}

	// Splits the message if it's over 1950 characters
	if len(message) > 1950 {

		splitMessage = SplitLongMessage(message)
	}

	// Prints split or unsplit whois
	if splitMessage == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
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
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot send whois message.")
			if err != nil {

				_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
				if err != nil {

					return
				}
				return
			}
		}
	}
}

// SplitLongMessage takes a message and splits it if it's longer than 1950. By Kagumi
func SplitLongMessage(message string) (split []string) {
	const maxLength = 1950
	if len(message) > maxLength {
		partitions := len(message) / maxLength
		if math.Mod(float64(len(message)), maxLength) > 0 {
			partitions++
		}
		split = make([]string, partitions)
		for i := 0; i < partitions; i++ {
			if i == partitions-1 {
				split[i] = message[i*maxLength:]
				break
			}
			split[i] = message[i*maxLength : (i+1)*maxLength]
		}
	} else {
		split = make([]string, 1)
		split[0] = message
	}
	return
}

// Function that iterates through memberInfo.json and checks for any alt accounts for that ID. Whois version
func CheckAltAccountWhois(id string) []string {

	var alts []string

	// Iterates through all users in memberInfo.json
	for userOne := range misc.MemberInfoMap {

		// Checks if the current user has the same reddit username as id string user
		if misc.MemberInfoMap[userOne].RedditUsername == misc.MemberInfoMap[id].RedditUsername &&
			misc.MemberInfoMap[userOne].RedditUsername != "" && misc.MemberInfoMap[id].RedditUsername != "" {

			alts = append(alts, misc.MemberInfoMap[userOne].ID)
		}
	}

	if len(alts) > 1 {

		return alts
	} else {

		return nil
	}
}

//func init() {
//	add(&command{
//		execute:  whoisCommand,
//		trigger:  "whois",
//		desc:     "Pulls memberInfo information about a user.",
//		elevated: true,
//	})
//}