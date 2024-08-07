package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"

	"github.com/bwmarrin/discordgo"
)

// remindMeCommand sets a remindMe note for after the target time has passed to be sent to the user
func remindMeCommand(channelID, userID, time, message string) string {
	var (
		remindMeObject entities.RemindMe
		flag           bool
		dummySlice     entities.RemindMeSlice
	)

	// Figures out the date to show the message
	date, perma, err := common.ResolveTimeFromString(time)
	if err != nil {
		return "Error: Invalid time given."
	}
	if perma {
		return "Error: Cannot use that time. Please use another."
	}

	// Saves the remindMe data to an object of type remindMe
	maxId := 1
	if _, ok := entities.SharedInfo.GetRemindMesMap()[userID]; ok {
		for _, remind := range entities.SharedInfo.GetRemindMesMap()[userID].GetRemindMeSlice() {
			if remind.GetRemindID() <= maxId {
				continue
			}
			maxId = remind.GetRemindID()
		}
	}
	if _, ok := entities.SharedInfo.GetRemindMesMap()[userID]; ok {
		maxId++
		flag = true
	}
	remindMeObject.SetCommandChannel(channelID)
	remindMeObject.SetRemindID(maxId)
	remindMeObject.SetDate(date)
	remindMeObject.SetMessage(message)

	// Adds the above object to the remindMe map where all of the remindMes are kept and writes them to disk
	entities.SharedInfo.Lock()
	if !flag {
		entities.SharedInfo.RemindMes[userID] = &dummySlice
	}
	entities.SharedInfo.RemindMes[userID].AppendToRemindMeSlice(&remindMeObject)
	entities.SharedInfo.Unlock()

	err = entities.RemindMeWrite(entities.SharedInfo.GetRemindMesMap())
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("Success! You will be reminded of the message <t:%d:R>. Make sure your DMs are open.", date.UTC().Unix())
}

// remindMeCommandHandler sets a remindMe note for after the target time has passed to be sent to the user
func remindMeCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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
	date, perma, err := common.ResolveTimeFromString(commandStrings[1])
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
	maxId := 1
	if _, ok := entities.SharedInfo.GetRemindMesMap()[userID]; ok {
		for _, remind := range entities.SharedInfo.GetRemindMesMap()[userID].GetRemindMeSlice() {
			if remind.GetRemindID() <= maxId {
				continue
			}
			maxId = remind.GetRemindID()
		}
	}
	if _, ok := entities.SharedInfo.GetRemindMesMap()[userID]; ok {
		maxId++
		flag = true
	}
	remindMeObject.SetCommandChannel(m.ChannelID)
	remindMeObject.SetRemindID(maxId)
	remindMeObject.SetDate(date)
	remindMeObject.SetMessage(commandStrings[2])

	// Adds the above object to the remindMe map where all of the remindMes are kept and writes them to disk
	entities.SharedInfo.Lock()
	if !flag {
		entities.SharedInfo.RemindMes[userID] = &dummySlice
	}
	entities.SharedInfo.RemindMes[userID].AppendToRemindMeSlice(&remindMeObject)
	entities.SharedInfo.Unlock()

	err = entities.RemindMeWrite(entities.SharedInfo.GetRemindMesMap())
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! You will be reminded of the message <t:%d:R>. Make sure your DMs are open.", date.UTC().Unix()))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

func removeRemindMe(id int, userID string) string {
	// Checks if the user has any reminds
	_, ok := entities.SharedInfo.GetRemindMesMap()[userID]
	if !ok {
		return "Error: No saved reminds found for you to delete."
	}

	if id <= 0 {
		return "Error: Invalid ID."
	}

	// Deletes the remind from the map and writes to disk
	flag := false
	for i, remind := range entities.SharedInfo.GetRemindMesMap()[userID].GetRemindMeSlice() {
		if remind == nil {
			continue
		}

		if remind.GetRemindID() == id {
			entities.SharedInfo.Lock()
			entities.SharedInfo.RemindMes[userID].RemoveFromRemindMeSlice(i)
			entities.SharedInfo.Unlock()
			flag = true

			err := entities.RemindMeWrite(entities.SharedInfo.GetRemindMesMap())
			if err != nil {
				return err.Error()
			}
			break
		}
	}

	if !flag {
		return "Error: No such remind with that ID found."
	}

	return fmt.Sprintf("Sucesss: Deleted remind with ID `%d`", id)
}

func removeRemindMeHandler(s *discordgo.Session, m *discordgo.Message) {
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
	_, ok := entities.SharedInfo.GetRemindMesMap()[userID]
	if !ok {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No saved reminds found for you to delete.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

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
	for i, remind := range entities.SharedInfo.GetRemindMesMap()[userID].GetRemindMeSlice() {
		if remind == nil {
			continue
		}

		if remind.GetRemindID() == remindID {
			entities.SharedInfo.Lock()
			entities.SharedInfo.RemindMes[userID].RemoveFromRemindMeSlice(i)
			entities.SharedInfo.Unlock()
			flag = true

			err := entities.RemindMeWrite(entities.SharedInfo.GetRemindMesMap())
			if err != nil {
				common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
				return
			}
			break
		}
	}

	// Prints success or error based on whether it deleted anything above
	if flag {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Sucesss: Deleted remind with ID `%d`.", remindID))
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

func viewRemindMes(userID string) []string {
	var message string

	if entities.SharedInfo.GetRemindMesMap()[userID] == nil || len(entities.SharedInfo.GetRemindMesMap()[userID].GetRemindMeSlice()) == 0 {
		return []string{"No saved reminds for you found."}
	}

	for _, remind := range entities.SharedInfo.GetRemindMesMap()[userID].GetRemindMeSlice() {
		if remind == nil {
			continue
		}

		message += fmt.Sprintf("`%s` - <t:%d:R> - ID: %d\n", remind.GetMessage(), remind.GetDate().UTC().Unix(), remind.GetRemindID())
	}

	return common.SplitLongMessage(message)
}

func viewRemindMesHandler(s *discordgo.Session, m *discordgo.Message) {
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

	if entities.SharedInfo.GetRemindMesMap()[userID] == nil || len(entities.SharedInfo.GetRemindMesMap()[userID].GetRemindMeSlice()) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No saved reminds for you found.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"reminds`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	for _, remind := range entities.SharedInfo.GetRemindMesMap()[userID].GetRemindMeSlice() {
		if remind == nil {
			continue
		}

		formattedMessage := fmt.Sprintf("`%s` - <t:%d:R> - ID: %d", remind.GetMessage(), remind.GetDate().UTC().Unix(), remind.GetRemindID())
		remindMes = append(remindMes, formattedMessage)
	}

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
		Execute: remindMeCommandHandler,
		Name:    "add-remind",
		Aliases: []string{"remind", "setremind", "addremind", "remindme", "remind-me"},
		Desc:    "Reminds you with a message in DMs after a period of time",
		Module:  "normal",
		DMAble:  true,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "time",
				Description: "The amount of time to wait before sending the message. Format: 2w1d12h30m.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message",
				Description: "The message you want to be sent.",
				Required:    true,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.ApplicationCommandData().Options == nil {
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			userID := ""
			if i.Member == nil {
				userID = i.User.ID
			} else {
				userID = i.Member.User.ID
			}

			time := ""
			message := ""
			for _, option := range i.ApplicationCommandData().Options {
				if option.Name == "time" {
					time = option.StringValue()
				} else if option.Name == "message" {
					message = option.StringValue()
				}
			}

			respStr := remindMeCommand(i.ChannelID, userID, time, message)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &respStr,
			})
		},
	})
	Add(&Command{
		Execute: removeRemindMeHandler,
		Name:    "remove-remind",
		Aliases: []string{"removeremindme", "deleteremind", "deleteremindme", "killremind", "stopremind", "removeremind"},
		Desc:    "Removes a previously set remind",
		Module:  "normal",
		DMAble:  true,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "id",
				Description: "The remind ID.",
				Required:    true,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var id int
			if i.ApplicationCommandData().Options == nil {
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			userID := ""
			if i.Member == nil {
				userID = i.User.ID
			} else {
				userID = i.Member.User.ID
			}

			for _, option := range i.ApplicationCommandData().Options {
				if option.Name == "id" {
					id = int(option.IntValue())
				}
			}

			respStr := removeRemindMe(id, userID)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &respStr,
			})
		},
	})
	Add(&Command{
		Execute: viewRemindMesHandler,
		Name:    "reminds",
		Aliases: []string{"viewremindmes", "viewremindme", "viewremind", "viewreminds", "remindmes"},
		Desc:    "Prints all currently set reminds by you.",
		Module:  "normal",
		DMAble:  true,
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			userID := ""
			if i.Member == nil {
				userID = i.User.ID
			} else {
				userID = i.Member.User.ID
			}

			messages := viewRemindMes(userID)
			if messages == nil {
				return
			}

			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &messages[0],
			})

			if len(messages) > 1 {
				for j, message := range messages {
					if j == 0 {
						continue
					}

					s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
						Content: message,
					})
				}
			}
		},
	})
}
