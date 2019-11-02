package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Adds a warning to a specific user in memberInfo.json without telling them
func addWarningCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		warning          string
		warningTimestamp functionality.Punishment
	)

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 3)

	if len(commandStrings) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"addwarning [@user, userID, or username#discrim] [warning]`\n\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
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

	warning = commandStrings[2]
	// Checks if the warning contains a mention and finds the actual name instead of ID
	warning = functionality.MentionParser(s, warning, m.GuildID)

	// Pulls info on user
	userMem, err := s.State.Member(m.GuildID, userID)
	if err != nil {
		userMem, _ = s.GuildMember(m.GuildID, userID)
	}

	// Checks if user is in memberInfo and handles them
	functionality.Mutex.Lock()
	if _, ok := functionality.GuildMap[m.GuildID].MemberInfoMap[userID]; !ok || len(functionality.GuildMap[m.GuildID].MemberInfoMap) == 0 {
		if userMem == nil {
			functionality.Mutex.Unlock()
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in server _and_ memberInfo. Cannot warn user until they join the server.")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		// Initializes user if he doesn't exist in memberInfo but is in server
		functionality.InitializeMember(userMem, m.GuildID)
	}

	// Appends warning to user in memberInfo
	functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Warnings = append(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Warnings, warning)

	// Adds timestamp for that warning
	t, err := m.Timestamp.Parse()
	if err != nil {
		functionality.Mutex.Unlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	warningTimestamp.Timestamp = t
	warningTimestamp.Punishment = warning
	warningTimestamp.Type = "Warning"
	functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps = append(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps, &warningTimestamp)

	// Writes to memberInfo.json
	_ = functionality.WriteMemberInfo(functionality.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	functionality.Mutex.Unlock()

	// Sends warning embed message to channel
	if userMem == nil {
		return
	}
	err = functionality.WarningEmbed(s, m, userMem.User, warning, m.ChannelID, true)
	if err != nil {
		return
	}
}

// Issues a warning to a specific user in memberInfo.json wand tells them
func issueWarningCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		warning          string
		warningTimestamp functionality.Punishment
	)

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 3)

	if len(commandStrings) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"issuewarning [@user, userID, or username#discrim] [warning]`\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls the guild name early on purpose
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	userID, err := functionality.GetUserID(m, commandStrings)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	warning = commandStrings[2]
	// Checks if the warning contains a mention and finds the actual name instead of ID
	warning = functionality.MentionParser(s, warning, m.GuildID)

	// Pulls info on user
	userMem, err := s.State.Member(m.GuildID, userID)
	if err != nil {
		userMem, _ = s.GuildMember(m.GuildID, userID)
	}

	// Checks if user is in memberInfo and handles them
	functionality.Mutex.Lock()
	if _, ok := functionality.GuildMap[m.GuildID].MemberInfoMap[userID]; !ok || len(functionality.GuildMap[m.GuildID].MemberInfoMap) == 0 {
		if userMem == nil {
			functionality.Mutex.Unlock()
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in server _and_ memberInfo. Cannot warn user until they join the server.")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		// Initializes user if he doesn't exist in memberInfo but is in server
		functionality.InitializeMember(userMem, m.GuildID)
	}

	// Appends warning to user in memberInfo
	functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Warnings = append(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Warnings, warning)

	// Adds timestamp for that warning
	t, err := m.Timestamp.Parse()
	if err != nil {
		functionality.Mutex.Unlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	warningTimestamp.Timestamp = t
	warningTimestamp.Punishment = warning
	warningTimestamp.Type = "Warning"
	functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps = append(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps, &warningTimestamp)

	// Writes to memberInfo.json
	_ = functionality.WriteMemberInfo(functionality.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	functionality.Mutex.Unlock()

	// Sends message in DMs that they have been warned if able
	dm, _ := s.UserChannelCreate(userID)
	_, _ = s.ChannelMessageSend(dm.ID, "You have been warned on "+guild.Name+":\n`"+warning+"`")

	// Sends warning embed message to channel
	if userMem == nil {
		return
	}
	err = functionality.WarningEmbed(s, m, userMem.User, warning, m.ChannelID, false)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    addWarningCommand,
		Trigger:    "addwarning",
		Desc:       "Adds a warning to a user without messaging them",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
	functionality.Add(&functionality.Command{
		Execute:    issueWarningCommand,
		Trigger:    "issuewarning",
		Desc:       "Issues a warning to a user and messages them",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
}
