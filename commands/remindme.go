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
func remindMeCommand(channelID, userID, timeStr, message string, isGuild bool, premium bool) string {
	// Parse the reminder time
	date, perma, err := common.ResolveTimeFromString(timeStr)
	if err != nil {
		return "Error: Invalid time given."
	}
	if perma {
		return "Error: Cannot use that time. Please use another."
	}

	// Load existing reminders for this user/guild
	remindMeSlice := db.GetReminders(userID)
	if remindMeSlice == nil {
		remindMeSlice = &entities.RemindMeSlice{
			RemindMeSlice: []*entities.RemindMe{},
			Guild:         isGuild,
			Premium:       premium,
		}
	}

	// Determine ID for the new reminder
	maxId := 1
	for _, remind := range remindMeSlice.GetRemindMeSlice() {
		if remind.GetRemindID() > maxId {
			maxId = remind.GetRemindID()
		}
	}
	maxId++

	// Create new reminder object
	remindMeObject := &entities.RemindMe{
		Message:        message,
		Date:           date,
		CommandChannel: channelID,
		RemindID:       maxId,
	}

	db.SetReminder(userID, remindMeObject, isGuild, premium)

	return fmt.Sprintf("Success! You will be reminded of the message <t:%d:R>. Make sure your DMs are open.", date.UTC().Unix())
}

// remindMeCommandHandler sets a remindMe note for after the target time has passed to be sent to the user
func remindMeCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	var guildSettings = entities.GuildSettings{Prefix: "."}

	// Fetch guild settings if the message is from a guild
	if m.GuildID != "" {
		guildSettings = db.GetGuildSettings(m.GuildID)
	}

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 3)
	if len(commandStrings) < 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"remindme [time] [message]`\n\n"+
			"Time is in #w#d#h#m format, such as 2w1d12h30m for 2 weeks, 1 day, 12 hours, 30 minutes.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
		}
		return
	}

	// Parse the reminder time
	date, perma, err := common.ResolveTimeFromString(commandStrings[1])
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error: Invalid time given.")
		return
	}
	if perma {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error: Cannot use that time. Please use another.")
		return
	}

	userID := m.Author.ID
	isGuild := m.GuildID != ""
	premium := false

	// Load existing reminders for this user/guild
	remindMeSlice := db.GetReminders(userID)
	if remindMeSlice == nil {
		remindMeSlice = &entities.RemindMeSlice{
			RemindMeSlice: []*entities.RemindMe{},
			Guild:         isGuild,
			Premium:       premium,
		}
	}

	// Determine ID for the new reminder
	maxId := 1
	for _, remind := range remindMeSlice.GetRemindMeSlice() {
		if remind.GetRemindID() > maxId {
			maxId = remind.GetRemindID()
		}
	}
	maxId++

	// Create new reminder object
	remindMeObject := &entities.RemindMe{
		Message:        commandStrings[2],
		Date:           date,
		CommandChannel: m.ChannelID,
		RemindID:       maxId,
	}

	db.SetReminder(userID, remindMeObject, isGuild, premium)

	// Notify the user
	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! You will be reminded of the message <t:%d:R>. Make sure your DMs are open.", date.UTC().Unix()))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
	}
}

func removeRemindMe(id int, userID string) string {
	// Validate input
	if id <= 0 {
		return "Error: Invalid ID."
	}

	// Fetch user's reminders from MongoDB
	remindMeSlice := db.GetReminders(userID)
	if remindMeSlice == nil || len(remindMeSlice.GetRemindMeSlice()) == 0 {
		return "Error: No saved reminds found for you to delete."
	}

	db.RemoveReminder(userID, id)

	return fmt.Sprintf("Success: Deleted remind with ID `%d`", id)
}

func removeRemindMeHandler(s *discordgo.Session, m *discordgo.Message) {
	var (
		userID        = m.Author.ID
		remindID      int
		guildSettings = entities.GuildSettings{Prefix: "."}
	)

	if m.GuildID != "" {
		guildSettings = db.GetGuildSettings(m.GuildID)
	}

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	// Validate command format
	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%sremoveremind [ID]`\n\nID is from the `%sreminds` command.", guildSettings.GetPrefix(), guildSettings.GetPrefix()))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
		}
		return
	}

	// Validate reminder ID
	remindID, err := strconv.Atoi(commandStrings[1])
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Please input only a number as the second parameter.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
		}
		return
	}

	// Fetch user's reminders from MongoDB
	remindMeSlice := db.GetReminders(userID)
	if remindMeSlice == nil || len(remindMeSlice.GetRemindMeSlice()) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No saved reminds found for you to delete.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
		}
		return
	}

	// Remove the reminder
	db.RemoveReminder(userID, remindID)

	// Confirm removal
	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success: Deleted remind with ID `%d`.", remindID))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
	}
}

func viewRemindMes(userID string) []string {
	var message string

	// Fetch user's reminders from MongoDB
	remindMeSlice := db.GetReminders(userID)
	if remindMeSlice == nil || len(remindMeSlice.GetRemindMeSlice()) == 0 {
		return []string{"No saved reminds for you found."}
	}

	// Format reminders
	for _, remind := range remindMeSlice.GetRemindMeSlice() {
		if remind == nil {
			continue
		}
		message += fmt.Sprintf("`%s` - <t:%d:R> - ID: %d\n", remind.GetMessage(), remind.GetDate().UTC().Unix(), remind.GetRemindID())
	}

	return common.SplitLongMessage(message)
}

func viewRemindMesHandler(s *discordgo.Session, m *discordgo.Message) {
	var (
		userID         = m.Author.ID
		remindMessages []string
		message        string
		guildSettings  = entities.GuildSettings{Prefix: "."}
	)

	// Fetch guild settings if in a server
	if m.GuildID != "" {
		guildSettings = db.GetGuildSettings(m.GuildID)
	}

	// Fetch user's reminders from MongoDB
	remindMeSlice := db.GetReminders(userID)
	if remindMeSlice == nil || len(remindMeSlice.GetRemindMeSlice()) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No saved reminds for you found.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
		}
		return
	}

	// Validate command format
	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")
	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"reminds`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
		}
		return
	}

	// Construct reminder messages
	for _, remind := range remindMeSlice.GetRemindMeSlice() {
		if remind == nil {
			continue
		}

		formattedMessage := fmt.Sprintf("`%s` - <t:%d:R> - ID: %d", remind.GetMessage(), remind.GetDate().UTC().Unix(), remind.GetRemindID())
		remindMessages = append(remindMessages, formattedMessage)
	}

	// Split long messages to fit Discord's limits
	remindMessages, message = splitRemindsMessages(remindMessages, message)

	// Prevent excessive output
	if len(remindMessages) > 4 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Too many reminds to display. Please wait them out or manage them differently.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
		}
		return
	}

	// Send reminder messages
	for _, remind := range remindMessages {
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

			isGuild := false
			userID := ""
			if i.Member == nil {
				userID = i.User.ID
			} else {
				userID = i.Member.User.ID
				isGuild = true
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

			respStr := remindMeCommand(i.ChannelID, userID, time, message, isGuild, false)
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
