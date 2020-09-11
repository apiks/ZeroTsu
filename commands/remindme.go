package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Sets a remindMe note for after the target time has passed to be sent to the user
func remindMeCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		remindMeObject entities.RemindMe
		userID         string
		flag           bool
		dummySlice     entities.RemindMeSlice

		guildSettings = entities.GuildSettings{Prefix: "."}
	)

	if m.GuildID != "" {
		guildSettings = db.GetGuildSettings(m.GuildID)
	}

	// Checks if message contains filtered words, which would not be allowed as a remind
	badWordExists, _, err := isFiltered(s, m)
	if err != nil {
		guildSettings := db.GetGuildSettings(m.GuildID)
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
	if badWordExists {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Usage of server filtered words in the remindMe command is not allowed. Please use remindMe in another server I am in or DMs.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 3)

	if len(commandStrings) < 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"remindme [time] [message]`\n\n"+
			"Time is in #w#d#h#m format, such as 2w1d12h30m for 2 weeks, 1 day, 12 hours, 30 minutes.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Figures out the date to show the message
	Date, perma, err := common.ResolveTimeFromString(commandStrings[1])
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid time given.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	if perma {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Cannot use that time. Please use another.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Saves the userID in a separate variable
	userID = m.Author.ID

	// Saves the remindMe data to an object of type remindMe
	entities.Mutex.Lock()
	remindMeObject.SetCommandChannel(m.ChannelID)
	if _, ok := entities.SharedInfo.GetRemindMesMap()[userID]; ok {
		remindMeObject.SetRemindID(len(entities.SharedInfo.GetRemindMesMap()[userID].GetRemindMeSlice()) + 1)
		flag = true
	} else {
		remindMeObject.AddToRemindID(1)
	}
	remindMeObject.SetDate(Date)
	remindMeObject.SetMessage(commandStrings[2])

	// Adds the above object to the remindMe map where all of the remindMes are kept and writes them to disk
	if !flag {
		entities.SharedInfo.GetRemindMesMap()[userID] = &dummySlice
	}
	entities.SharedInfo.GetRemindMesMap()[userID].AppendToRemindMeSlice(&remindMeObject)
	err = entities.RemindMeWrite(entities.SharedInfo.GetRemindMesMap())
	if err != nil {
		entities.Mutex.Unlock()
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	entities.Mutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, "Success! You will be reminded of the message on _"+Date.Format("2006-01-02 15:04 MST")+"_.")
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

func viewRemindMe(s *discordgo.Session, m *discordgo.Message) {
	var (
		userID    string
		remindMes []string
		message   string

		guildSettings = entities.GuildSettings{Prefix: "."}
	)

	userID = m.Author.ID

	// Checks if the user has any reminds
	if m.GuildID != "" {
		guildSettings = db.GetGuildSettings(m.GuildID)
	}

	entities.Mutex.RLock()
	if entities.SharedInfo.GetRemindMesMap()[userID] == nil || len(entities.SharedInfo.GetRemindMesMap()[userID].GetRemindMeSlice()) == 0 {
		entities.Mutex.RUnlock()
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No saved reminds for you found.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	entities.Mutex.RUnlock()

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"reminds`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	entities.Mutex.RLock()
	for _, remind := range entities.SharedInfo.GetRemindMesMap()[userID].GetRemindMeSlice() {
		if remind == nil {
			continue
		}

		formattedMessage := fmt.Sprintf("`%s` - _%s_ - ID: %d", remind.GetMessage(), remind.GetDate().Format("2006-01-02 15:04"), remind.GetRemindID())
		remindMes = append(remindMes, formattedMessage)
	}
	entities.Mutex.RUnlock()

	// Splits the message objects into multiple messages if it's too big
	remindMes, message = splitRemindsMessages(remindMes, message)

	// Limits the size it can display so it isn't abused
	if len(remindMes) > 4 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: The message size of all of the reminds is too big to display."+
			" Please wait them out or never use this command again.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	for _, remind := range remindMes {
		_, err := s.ChannelMessageSend(m.ChannelID, remind)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
	}
}

func removeRemindMe(s *discordgo.Session, m *discordgo.Message) {
	var (
		userID   string
		remindID int
		flag     bool

		guildSettings = entities.GuildSettings{Prefix: "."}
	)

	userID = m.Author.ID

	if m.GuildID != "" {
		guildSettings = db.GetGuildSettings(m.GuildID)
	}

	// Checks if the user has any reminds
	entities.Mutex.RLock()
	_, ok := entities.SharedInfo.GetRemindMesMap()[userID]
	if !ok {
		entities.Mutex.RUnlock()
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No saved reminds found for you to delete.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	entities.Mutex.RUnlock()

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"removeremind [ID]`\n\nID is from the `"+guildSettings.GetPrefix()+"reminds` command.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	remindID, err := strconv.Atoi(commandStrings[1])
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Please input only a number as the second parameter.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Deletes the remind from the map and writes to disk
	entities.Mutex.Lock()
	for i, remind := range entities.SharedInfo.GetRemindMesMap()[userID].GetRemindMeSlice() {
		if remind == nil {
			continue
		}

		if remind.GetRemindID() == remindID {
			entities.SharedInfo.GetRemindMesMap()[userID].RemoveFromRemindMeSlice(i)
			flag = true

			err := entities.RemindMeWrite(entities.SharedInfo.GetRemindMesMap())
			if err != nil {
				entities.Mutex.Unlock()
				common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
				return
			}
			break
		}
	}
	entities.Mutex.Unlock()

	// Prints success or error based on whether it deleted anything above
	if flag {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Sucesss: Deleted remind with ID %d.", remindID))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: No such remind with that ID found. ID is from the `"+guildSettings.GetPrefix()+"reminds` command."))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
	}
}

// Splits the view reminds messages into blocks
func splitRemindsMessages(msgs []string, message string) ([]string, string) {
	const maxMsgLength = 1900
	if len(message) > maxMsgLength {
		msgs = append(msgs, message)
		message = ""
	}
	return msgs, message
}

func init() {
	Add(&Command{
		Execute: remindMeCommand,
		Trigger: "remindme",
		Aliases: []string{"remind", "setremind", "addremind"},
		Desc:    "Reminds you of the set message after a period of time",
		Module:  "normal",
		DMAble:  true,
	})
	Add(&Command{
		Execute: viewRemindMe,
		Trigger: "reminds",
		Aliases: []string{"viewremindmes", "viewremindme", "viewremind", "viewreminds", "remindmes"},
		Desc:    "Shows you what reminds you have currently set",
		Module:  "normal",
		DMAble:  true,
	})
	Add(&Command{
		Execute: removeRemindMe,
		Trigger: "removeremind",
		Aliases: []string{"removeremindme", "deleteremind", "deleteremindme", "killremind", "stopremind"},
		Desc:    "Removes a previously set remind",
		Module:  "normal",
		DMAble:  true,
	})
}
