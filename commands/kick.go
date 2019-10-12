package commands

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

// Kicks a user from the server with a reason
func kickCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		userID        string
		reason        string
		kickTimestamp misc.Punishment
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	commandStrings := strings.SplitN(m.Content, " ", 3)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"kick [@user, userID, or username#discrim] [reason]*` format.\n\n* is optional"+
			"\n\nNote: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			return
		}
		return
	}

	userID, err := misc.GetUserID(m, commandStrings)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
	if len(commandStrings) == 3 {
		reason = commandStrings[2]
		// Checks if the reason contains a mention and finds the actual name instead of ID
		reason = misc.MentionParser(s, reason, m.GuildID)
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
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
				return
			}
			return
		}
		return
	}

	// Checks if user has a privileged role
	misc.MapMutex.Lock()
	if HasElevatedPermissions(s, userID, m.GuildID) {
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: Target user has a privileged role. Cannot kick.")
		if err != nil {
			_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}

	// Initialize user if they are not in memberInfo but is in server
	if _, ok := misc.GuildMap[m.GuildID].MemberInfoMap[userID]; !ok {
		misc.InitializeUser(userMem, m.GuildID)
	}

	// Adds kick reason to user memberInfo info
	misc.GuildMap[m.GuildID].MemberInfoMap[userID].Kicks = append(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Kicks, reason)

	// Adds timestamp for that kick
	t, err := m.Timestamp.Parse()
	if err != nil {
		misc.MapMutex.Unlock()
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
	kickTimestamp.Timestamp = t
	kickTimestamp.Punishment = reason
	kickTimestamp.Type = "Kick"
	misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps = append(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps, &kickTimestamp)

	// Writes memberInfo.json
	misc.WriteMemberInfo(misc.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	misc.MapMutex.Unlock()

	// Fetches the guild for the Name
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Sends message to user DMs if possible
	dm, _ := s.UserChannelCreate(userID)
	_, _ = s.ChannelMessageSend(dm.ID, "You have been kicked from "+guild.Name+":\n**"+reason+"**")

	// Kicks the user from the server with a reason
	err = s.GuildMemberDeleteWithReason(m.GuildID, userID, reason)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Sends embed channel message
	err = KickEmbed(s, m, userMem.User, reason, m.ChannelID)
	if err != nil {
		_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		return
	}

	// Sends embed bot-log message
	_ = KickEmbed(s, m, userMem.User, reason, guildBotLog)
}

func KickEmbed(s *discordgo.Session, m *discordgo.Message, mem *discordgo.User, reason string, channelID string) error {

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

	// Sets field inline
	embedFieldUserID.Inline = true
	embedFieldReason.Inline = true

	// Adds the two fields to embedField slice (because embedMess.Fields requires slice input)
	embedField = append(embedField, &embedFieldUserID)
	embedField = append(embedField, &embedFieldReason)

	// Sets embed title and its description (which it uses the same way as a field)
	embedMess.Title = mem.Username + "#" + mem.Discriminator + " was kicked by " + m.Author.Username

	// Adds user thumbnail and the two other fields as well
	embedMess.Thumbnail = &embedThumbnail
	embedMess.Fields = embedField

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(channelID, &embedMess)
	return err
}

func init() {
	add(&command{
		execute:  kickCommand,
		trigger:  "kick",
		aliases:  []string{"k", "yeet", "yut"},
		desc:     "Kicks a user from the server and logs reason",
		elevated: true,
		category: "moderation",
	})
}
