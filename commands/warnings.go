package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

// Adds a warning to a specific user in memberInfo.json without telling them
func addWarningCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		warning          string
		warningTimestamp misc.Punishment
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(messageLowercase, " ", 3)

	if len(commandStrings) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"addwarning [@user, userID, or username#discrim] [warning]`\n\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	userID, err := misc.GetUserID(m, commandStrings)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	warning = commandStrings[2]
	// Checks if the warning contains a mention and finds the actual name instead of ID
	warning = misc.MentionParser(s, warning, m.GuildID)

	// Pulls info on user
	userMem, err := s.State.Member(m.GuildID, userID)
	if err != nil {
		userMem, _ = s.GuildMember(m.GuildID, userID)
	}

	// Checks if user is in memberInfo and handles them
	misc.MapMutex.Lock()
	if _, ok := misc.GuildMap[m.GuildID].MemberInfoMap[userID]; !ok || len(misc.GuildMap[m.GuildID].MemberInfoMap) == 0 {
		if userMem == nil {
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in server _and_ memberInfo. Cannot warn user until they join the server.")
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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
		// Initializes user if he doesn't exist in memberInfo but is in server
		misc.InitializeUser(userMem, m.GuildID)
	}

	// Appends warning to user in memberInfo
	misc.GuildMap[m.GuildID].MemberInfoMap[userID].Warnings = append(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Warnings, warning)

	// Adds timestamp for that warning
	t, err := m.Timestamp.Parse()
	if err != nil {
		misc.MapMutex.Unlock()
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
	warningTimestamp.Timestamp = t
	warningTimestamp.Punishment = warning
	warningTimestamp.Type = "Warning"
	misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps = append(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps, warningTimestamp)

	// Writes to memberInfo.json
	misc.WriteMemberInfo(misc.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	misc.MapMutex.Unlock()

	// Sends warning embed message to channel
	if userMem == nil {
		return
	}
	err = WarningEmbed(s, m, userMem.User, warning, m.ChannelID, true)
	if err != nil {
		return
	}
}

// Issues a warning to a specific user in memberInfo.json wand tells them
func issueWarningCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		warning          string
		warningTimestamp misc.Punishment
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(messageLowercase, " ", 3)

	if len(commandStrings) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"issuewarning [@user, userID, or username#discrim] [warning]`\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Pulls the guild name early on purpose
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	userID, err := misc.GetUserID(m, commandStrings)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	warning = commandStrings[2]
	// Checks if the warning contains a mention and finds the actual name instead of ID
	warning = misc.MentionParser(s, warning, m.GuildID)

	// Pulls info on user
	userMem, err := s.State.Member(m.GuildID, userID)
	if err != nil {
		userMem, _ = s.GuildMember(m.GuildID, userID)
	}

	// Checks if user is in memberInfo and handles them
	misc.MapMutex.Lock()
	if _, ok := misc.GuildMap[m.GuildID].MemberInfoMap[userID]; !ok || len(misc.GuildMap[m.GuildID].MemberInfoMap) == 0 {
		if userMem == nil {
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in server _and_ memberInfo. Cannot warn user until they join the server.")
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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
		// Initializes user if he doesn't exist in memberInfo but is in server
		misc.InitializeUser(userMem, m.GuildID)
	}

	// Appends warning to user in memberInfo
	misc.GuildMap[m.GuildID].MemberInfoMap[userID].Warnings = append(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Warnings, warning)

	// Adds timestamp for that warning
	t, err := m.Timestamp.Parse()
	if err != nil {
		misc.MapMutex.Unlock()
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
	warningTimestamp.Timestamp = t
	warningTimestamp.Punishment = warning
	warningTimestamp.Type = "Warning"
	misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps = append(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps, warningTimestamp)

	// Writes to memberInfo.json
	misc.WriteMemberInfo(misc.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	misc.MapMutex.Unlock()

	// Sends message in DMs that they have been warned if able
	dm, err := s.UserChannelCreate(userID)
	if err != nil {
		return
	}
	_, _ = s.ChannelMessageSend(dm.ID, "You have been warned on "+guild.Name+":\n`"+warning+"`")

	// Sends warning embed message to channel
	if userMem == nil {
		return
	}
	err = WarningEmbed(s, m, userMem.User, warning, m.ChannelID, false)
	if err != nil {
		return
	}
}

func WarningEmbed(s *discordgo.Session, m *discordgo.Message, mem *discordgo.User, reason string, channelID string, discrete bool) error {

	var (
		embedMess      discordgo.MessageEmbed
		embedThumbnail discordgo.MessageEmbedThumbnail

		// Embed slice and its fields
		embedField       []*discordgo.MessageEmbedField
		embedFieldUserID discordgo.MessageEmbedField
		embedFieldReason discordgo.MessageEmbedField
	)
	t := time.Now()

	// Sets timestamp for warning
	embedMess.Timestamp = t.Format(time.RFC3339)

	// Sets warning embed color
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

	if discrete {
		embedMess.Title = fmt.Sprintf("Added warning to %v#%v", mem.Username, mem.Discriminator)
	} else {
		embedMess.Title = mem.Username + "#" + mem.Discriminator + " was warned by " + m.Author.Username
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
		execute:  addWarningCommand,
		trigger:  "addwarning",
		desc:     "Adds a warning to a user without messaging them",
		elevated: true,
		category: "moderation",
	})
	add(&command{
		execute:  issueWarningCommand,
		trigger:  "issuewarning",
		desc:     "Issues a warning to a user and messages them",
		elevated: true,
		category: "moderation",
	})
}
