package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Kicks a user from the server with a reason
func kickCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		userID        string
		reason        string
		kickTimestamp functionality.Punishment
	)

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 3)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"kick [@user, userID, or username#discrim] [reason]*` format.\n\n* is optional"+
			"\n\nNote: If using username#discrim you cannot have spaces in the username. It must be a single word.")
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
	if len(commandStrings) == 3 {
		reason = commandStrings[2]
		// Checks if the reason contains a mention and finds the actual name instead of ID
		reason = functionality.MentionParser(s, reason, m.GuildID)
	} else {
		reason = "[No reason given]"
	}

	// Pulls info on user
	userMem, err := s.State.Member(m.GuildID, userID)
	if err != nil {
		userMem, err = s.GuildMember(m.GuildID, userID)
		if err != nil {
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: Invalid user or user not found in server. Cannot kick user.")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		return
	}

	// Checks if user has a privileged role
	if functionality.HasElevatedPermissions(s, userID, m.GuildID) {
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: Target user has a privileged role. Cannot kick.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Initialize user if they are not in memberInfo but is in server
	functionality.Mutex.Lock()
	memberInfoUser, ok := functionality.GuildMap[m.GuildID].MemberInfoMap[userID]
	if !ok {
		functionality.InitializeMember(userMem, m.GuildID)
	}

	// Adds kick reason to user memberInfo info
	memberInfoUser.Kicks = append(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Kicks, reason)

	// Adds timestamp for that kick
	t, err := m.Timestamp.Parse()
	if err != nil {
		functionality.Mutex.Unlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	kickTimestamp.Timestamp = t
	kickTimestamp.Punishment = reason
	kickTimestamp.Type = "Kick"
	memberInfoUser.Timestamps = append(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps, &kickTimestamp)

	// Writes memberInfo.json
	_ = functionality.WriteMemberInfo(functionality.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	functionality.Mutex.Unlock()

	// Fetches the guild for the Name
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Sends message to user DMs if possible
	dm, _ := s.UserChannelCreate(userID)
	_, _ = s.ChannelMessageSend(dm.ID, "You have been kicked from "+guild.Name+":\n**"+reason+"**")

	// Kicks the user from the server with a reason
	err = s.GuildMemberDeleteWithReason(m.GuildID, userID, reason)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Sends embed channel message
	err = functionality.KickEmbed(s, m, userMem.User, reason, m.ChannelID)
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}

	// Sends embed bot-log message
	if guildSettings.BotLog == nil {
		return
	}
	if guildSettings.BotLog.ID == "" {
		return
	}
	_ = functionality.KickEmbed(s, m, userMem.User, reason, guildSettings.BotLog.ID)
}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    kickCommand,
		Trigger:    "kick",
		Aliases:    []string{"k", "yeet", "yut"},
		Desc:       "Kicks a user from the server and logs reason",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
}
