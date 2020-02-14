package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"ZeroTsu/functionality"
)

// Sets a remindMe note for after the target time has passed to be sent to the user
func remindMeCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		remindMeObject functionality.RemindMe
		userID         string
		flag           bool
		dummySlice     functionality.RemindMeSlice

		guildSettings = &functionality.GuildSettings{
			Prefix: ".",
		}
	)

	if m.GuildID != "" {
		functionality.Mutex.RLock()
		guildSettings = functionality.GuildMap[m.GuildID].GetGuildSettings()
		functionality.Mutex.RUnlock()
	}

	// Checks if message contains filtered words, which would not be allowed as a remind
	badWordExists, _ := isFiltered(s, m)
	if badWordExists {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Usage of server filtered words in the remindMe command is not allowed. Please use remindMe in another server I am in or DMs.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 3)

	if len(commandStrings) < 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"remindme [time] [message]`\n\n"+
			"Time is in #w#d#h#m format, such as 2w1d12h30m for 2 weeks, 1 day, 12 hours, 30 minutes.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Figures out the date to show the message
	Date, perma, err := functionality.ResolveTimeFromString(commandStrings[1])
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid time given.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	if perma {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Cannot use that time. Please use another.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Saves the userID in a separate variable
	userID = m.Author.ID

	// Saves the remindMe data to an object of type remindMe
	remindMeObject.CommandChannel = m.ChannelID
	functionality.Mutex.Lock()
	if _, ok := functionality.SharedInfo.RemindMes[userID]; ok {
		remindMeObject.RemindID = len(functionality.SharedInfo.RemindMes[userID].RemindMeSlice) + 1
		flag = true
	} else {
		remindMeObject.RemindID++
	}
	remindMeObject.Date = Date
	remindMeObject.Message = commandStrings[2]

	// Adds the above object to the remindMe map where all of the remindMes are kept and writes them to disk
	if !flag {
		functionality.SharedInfo.RemindMes[userID] = &dummySlice
	}
	functionality.SharedInfo.RemindMes[userID].RemindMeSlice = append(functionality.SharedInfo.RemindMes[userID].RemindMeSlice, &remindMeObject)
	err = functionality.RemindMeWrite(functionality.SharedInfo.RemindMes)
	if err != nil {
		functionality.Mutex.Unlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	functionality.Mutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, "Success! You will be reminded of the message on _"+Date.Format("2006-01-02 15:04 MST")+"_.")
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

func viewRemindMe(s *discordgo.Session, m *discordgo.Message) {
	var (
		userID    string
		remindMes []string
		message   string

		guildSettings = &functionality.GuildSettings{
			Prefix: ".",
		}
	)

	userID = m.Author.ID

	// Checks if the user has any reminds
	functionality.Mutex.RLock()
	if m.GuildID != "" {
		guildSettings = functionality.GuildMap[m.GuildID].GetGuildSettings()
	}

	if functionality.SharedInfo.RemindMes[userID] == nil || len(functionality.SharedInfo.RemindMes[userID].RemindMeSlice) == 0 {
		functionality.Mutex.RUnlock()
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No saved reminds for you found.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	functionality.Mutex.RUnlock()

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"reminds`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	functionality.Mutex.RLock()
	for _, remind := range functionality.SharedInfo.RemindMes[userID].RemindMeSlice {
		formattedMessage := fmt.Sprintf("`%s` - _%s_ - ID: %d", remind.Message, remind.Date.Format("2006-01-02 15:04"), remind.RemindID)
		remindMes = append(remindMes, formattedMessage)
	}
	functionality.Mutex.RUnlock()

	// Splits the message objects into multiple messages if it's too big
	remindMes, message = splitRemindsMessages(remindMes, message)

	// Limits the size it can display so it isn't abused
	if len(remindMes) > 4 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: The message size of all of the reminds is too big to display."+
			" Please wait them out or never use this command again.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	for _, remind := range remindMes {
		_, err := s.ChannelMessageSend(m.ChannelID, remind)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
	}
}

func removeRemindMe(s *discordgo.Session, m *discordgo.Message) {
	var (
		userID   string
		remindID int
		flag     bool

		guildSettings = &functionality.GuildSettings{
			Prefix: ".",
		}
	)

	userID = m.Author.ID

	functionality.Mutex.RLock()
	if m.GuildID != "" {
		guildSettings = functionality.GuildMap[m.GuildID].GetGuildSettings()
	}

	// Checks if the user has any reminds
	_, ok := functionality.SharedInfo.RemindMes[userID]
	if !ok {
		functionality.Mutex.RUnlock()
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No saved reminds found for you to delete.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	functionality.Mutex.RUnlock()

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"removeremind [ID]`\n\nID is from the `"+guildSettings.Prefix+"reminds` command.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	remindID, err := strconv.Atoi(commandStrings[1])
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Please input only a number as the second parameter.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Deletes the remind from the map and writes to disk
	functionality.Mutex.Lock()
	for i, remind := range functionality.SharedInfo.RemindMes[userID].RemindMeSlice {
		if remind.RemindID == remindID {

			if i < len(functionality.SharedInfo.RemindMes[userID].RemindMeSlice)-1 {
				copy(functionality.SharedInfo.RemindMes[userID].RemindMeSlice[i:], functionality.SharedInfo.RemindMes[userID].RemindMeSlice[i+1:])
			}
			functionality.SharedInfo.RemindMes[userID].RemindMeSlice[len(functionality.SharedInfo.RemindMes[userID].RemindMeSlice)-1] = nil
			functionality.SharedInfo.RemindMes[userID].RemindMeSlice = functionality.SharedInfo.RemindMes[userID].RemindMeSlice[:len(functionality.SharedInfo.RemindMes[userID].RemindMeSlice)-1]

			flag = true

			err := functionality.RemindMeWrite(functionality.SharedInfo.RemindMes)
			if err != nil {
				functionality.Mutex.Unlock()
				functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
				return
			}
			break
		}
	}
	functionality.Mutex.Unlock()

	// Prints success or error based on whether it deleted anything above
	if flag {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Sucesss: Deleted remind with ID %d.", remindID))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: No such remind with that ID found. ID is from the `"+guildSettings.Prefix+"reminds` command."))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
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
	functionality.Add(&functionality.Command{
		Execute: remindMeCommand,
		Trigger: "remindme",
		Aliases: []string{"remind", "setremind", "addremind"},
		Desc:    "Reminds you of the set message after a period of time",
		Module:  "normal",
		DMAble:  true,
	})
	functionality.Add(&functionality.Command{
		Execute: viewRemindMe,
		Trigger: "reminds",
		Aliases: []string{"viewremindmes", "viewremindme", "viewremind", "viewreminds", "remindmes"},
		Desc:    "Shows you what reminds you have currently set",
		Module:  "normal",
		DMAble:  true,
	})
	functionality.Add(&functionality.Command{
		Execute: removeRemindMe,
		Trigger: "removeremind",
		Aliases: []string{"removeremindme", "deleteremind", "deleteremindme", "killremind", "stopremind"},
		Desc:    "Removes a previously set remind",
		Module:  "normal",
		DMAble:  true,
	})
}
