package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Returns user avatar in channel as message
func avatarCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		err           error
		guildSettings = entities.GuildSettings{Prefix: "."}
	)

	if m.GuildID != "" {
		guildSettings = db.GetGuildSettings(m.GuildID)
	}

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) > 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage `%vavatar [user]`", guildSettings.GetPrefix()))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}
	if len(commandStrings) == 1 {
		// Fetches user
		mem, err := s.User(m.Author.ID)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		// Sends user who used the command's avatar
		_, err = s.ChannelMessageSend(m.ChannelID, mem.AvatarURL("256"))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings
	userID, err := common.GetUserID(m, commandStrings)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Fetches user
	mem, err := s.User(userID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Sends avatar
	_, err = s.ChannelMessageSend(m.ChannelID, mem.AvatarURL("256"))
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

func init() {
	Add(&Command{
		Execute: avatarCommand,
		Trigger: "avatar",
		Desc:    "Show user avatar. Add a @mention or userID to specify a user",
		Module:  "normal",
		DMAble:  true,
	})
}
