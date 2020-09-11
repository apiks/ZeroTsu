package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/embeds"
	"github.com/r-anime/ZeroTsu/entities"
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

		validSlice bool

		punishedUserObject = entities.NewPunishedUsers("", "", time.Time{}, time.Time{})

		banTimestamp = entities.NewPunishment("", "", time.Time{})
	)

	guildSettings := db.GetGuildSettings(m.GuildID)

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 4)
	commandStringsCopy = commandStrings

	if len(commandStrings) != 4 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"ban [@user, userID, or username#discrim] [time] [reason]` format. \n\n"+
			"Time is in #w#d#h#m format, such as 2w1d12h30m for 2 weeks, 1 day, 12 hours, 30 minutes. Use 0d for permanent.\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Check if the time is the 2nd parameter and handle that
	_, _, err := common.ResolveTimeFromString(commandStrings[1])
	if err == nil {
		length = commandStrings[1]
		commandStrings = append(commandStrings[:1], commandStrings[1+1:]...)
	}
	// Handle userID, reason and length
	userID, err = common.GetUserID(m, commandStrings)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	if length == "" {
		length = commandStrings[2]
	} else {
		commandStrings = commandStringsCopy
	}

	reason = commandStrings[3]
	// Checks if the reason contains a mention and finds the actual name instead of ID
	reason = common.MentionParser(s, reason, m.GuildID)

	// Checks if a number is contained in length var. Fixes some cases of invalid length
	lengthSlice := strings.Split(length, "")
	for i := 0; i < len(lengthSlice); i++ {
		if _, err := strconv.ParseInt(lengthSlice[i], 10, 64); err == nil || lengthSlice[i] == "∞" {
			validSlice = true
			break
		}
	}
	if !validSlice {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid length. \n Usage: `"+guildSettings.GetPrefix()+"ban [@user or userID] [time] [reason]` format. \n\n"+
			"Time is in #w#d#h#m format, such as 2w1d12h30m for 2 weeks, 1 day, 12 hours, 30 minutes. Use 0d for permanent.\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
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
		if functionality.HasElevatedPermissions(s, userID, m.GuildID) {
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: Target user has a privileged role. Cannot ban.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
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
				_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in the server, internal database and cannot fetch manually either. Cannot ban until user joins the server.")
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

	// Adds ban date to memberInfo and checks if perma
	mem = mem.AppendToBans(reason)
	UnbanDate, perma, err := common.ResolveTimeFromString(length)
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid time given.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	if commandStrings[2] == "∞" || commandStrings[1] == "∞" {
		perma = true
	}
	if !perma {
		mem = mem.SetUnbanDate(UnbanDate.Format("2006-01-02 15:04:05.999999999 -0700 MST"))
	} else {
		mem = mem.SetUnbanDate("_Never_")
	}

	// Adds timestamp for that ban
	t, err := m.Timestamp.Parse()
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	banTimestamp = banTimestamp.SetTimestamp(t)
	banTimestamp = banTimestamp.SetPunishment(reason)
	banTimestamp = banTimestamp.SetPunishmentType("Ban")
	mem = mem.AppendToTimestamps(banTimestamp)

	// Write
	db.SetGuildMember(m.GuildID, mem)

	// Saves the details in punishedUserObject
	punishedUserObject = punishedUserObject.SetID(userID)
	if userMem != nil {
		punishedUserObject = punishedUserObject.SetUsername(userMem.User.Username)
	} else {
		punishedUserObject = punishedUserObject.SetUsername(mem.GetUsername())
	}

	if perma {
		punishedUserObject = punishedUserObject.SetUnbanDate(time.Date(9999, 9, 9, 9, 9, 9, 9, time.Local))
	} else {
		unbanDate, err := time.Parse(common.LongDateFormat, UnbanDate.Format(common.LongDateFormat))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		punishedUserObject = punishedUserObject.SetUnbanDate(unbanDate)
	}

	// Adds or updates the now banned user in PunishedUsers
	err = db.SetGuildPunishedUser(m.GuildID, punishedUserObject)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Parses how long is left of the ban
	now := time.Now()
	remainingUnformatted := punishedUserObject.GetUnbanDate().Sub(now)
	if remainingUnformatted.Hours() < 1 {
		remaining = strconv.FormatFloat(remainingUnformatted.Minutes(), 'f', 0, 64) + " minutes"
	} else if remainingUnformatted.Hours() < 24 {
		remaining = strconv.FormatFloat(remainingUnformatted.Hours(), 'f', 0, 64) + " hours"
	} else {
		remaining = strconv.FormatFloat(remainingUnformatted.Hours()/24, 'f', 0, 64) + " days"
	}

	// Pulls the guild name
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		guild, err = s.Guild(m.GuildID)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}

	// Assigns success ban print string for user
	if perma {
		success = fmt.Sprintf("You have been banned from **%s**:\n**\"**%s**\"**\n\nUntil: `Forever`", guild.Name, reason)
	} else {
		z, _ := time.Now().Zone()
		success = fmt.Sprintf("You have been banned from **%s**:\n**\"**%s**\"**\n\nUntil: `%s` %s\nRemaining: `%s`", guild.Name, reason, UnbanDate.Format("2006-01-02 15:04:05"), z, remaining)
	}

	// Sends success string to user in DMs if able
	dm, _ := s.UserChannelCreate(userID)
	_, _ = s.ChannelMessageSend(dm.ID, success)

	// Bans the user
	err = s.GuildBanCreateWithReason(m.GuildID, userID, reason, 0)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	if userMem == nil {
		userMem = &discordgo.Member{GuildID: m.GuildID, User: &discordgo.User{ID: mem.GetID(), Username: mem.GetUsername(), Discriminator: mem.GetDiscrim()}}
	}

	// Sends embed channel message
	err = embeds.PunishmentAddition(s, m, userMem, "ban", "banned", reason, m.ChannelID, &UnbanDate, perma)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Sends embed bot-log message
	if guildSettings.BotLog == (entities.Cha{}) {
		return
	}
	if guildSettings.BotLog.GetID() == "" {
		return
	}
	err = embeds.PunishmentAddition(s, m, userMem, "ban", "banned", reason, guildSettings.BotLog.GetID(), &UnbanDate, perma)
	if err != nil {
		return
	}
}

func init() {
	Add(&Command{
		Execute:    banCommand,
		Trigger:    "ban",
		Aliases:    []string{"b", "hammer"},
		Desc:       "Bans a user for a set period of time",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
}
