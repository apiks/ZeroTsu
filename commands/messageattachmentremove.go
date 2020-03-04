package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Checks messages with uploads if they're uploading a blacklisted file type. If so it removvvevs them
func MessageAttachmentsHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in MessageAttachmentsHandler")
		}
	}()

	if m.GuildID == "" || len(m.Attachments) == 0 || m.Author.ID == s.State.User.ID {
		return
	}

	entities.HandleNewGuild(m.GuildID)

	// Checks if user is mod before checking the message
	if functionality.HasElevatedPermissions(s, m.Author.ID, m.GuildID) {
		return
	}

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildExtensions := db.GetGuildExtensions(m.GuildID)

	// Iterates through all the attachments (since more than one can be posted in one go)
	// and checks if it's a banned file type. If it is then remove them
	for _, attachment := range m.Attachments {

		if guildSettings.GetWhitelistFileFilter() {
			if isBannedExtension(attachment.Filename, guildExtensions) {
				continue
			}
		} else if !isBannedExtension(attachment.Filename, guildExtensions) {
			continue
		}

		// Deletes the message that was sent if has a non-whitelisted attachment
		err := s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}

		now := time.Now().Format("2006-01-02 15:04:05")

		// Prints success in bot-log channel
		if guildSettings.BotLog != (entities.Cha{}) {
			if guildSettings.BotLog.GetID() != "" {
				if guildSettings.GetWhitelistFileFilter() {
					_, _ = s.ChannelMessageSend(guildSettings.BotLog.GetID(), m.Author.Mention()+" had their message removed for uploading a non-whitelisted file type `"+
						attachment.Filename+"` in "+"<#"+m.ChannelID+"> on [_"+now+"_]")
				} else {
					_, _ = s.ChannelMessageSend(guildSettings.BotLog.GetID(), m.Author.Mention()+" had their message removed for uploading a blacklisted file type `"+
						attachment.Filename+"` in "+"<#"+m.ChannelID+"> on [_"+now+"_]")
				}
			}
		}

		// Sends a message to the user in their DMs if possible
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			return
		}

		// Fetches all file extensions
		var extensions string
		for ext := range guildExtensions {
			extensions += fmt.Sprintf("%v, ", ext)
		}
		extensions = strings.TrimSuffix(extensions, ", ")

		if len(extensions) == 0 {
			_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("Your message upload `%s` was removed for using a non-whitelisted file type.\n\nNo file attachments are allowed on that server.", attachment.Filename))
			return
		}

		if guildSettings.GetWhitelistFileFilter() {
			_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("Your message upload `%s` was removed for using a non-whitelisted file type.\n\nAllowed file types are: `%v`", attachment.Filename, extensions))
		} else {
			_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("Your message upload `%s` was removed for using a blacklisted file type.", attachment.Filename))
		}
	}

}

// Checks if it's an banned file type and returns true if it is, else false
func isBannedExtension(filename string, extensions map[string]string) bool {
	filename = strings.ToLower(filename)

	for ext := range extensions {
		if strings.HasSuffix(filename, fmt.Sprintf(".%s", ext)) {
			return true
		}
	}
	return false
}

// Blacklists a file extension
func filterExtensionCommand(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)
	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%sextension [file extension]`\n\n[file extension] is the file extension (e.g. .exe or .jpeg).", guildSettings.GetPrefix()))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Remove double spaces
	for i := 0; i < len(commandStrings); i++ {
		commandStrings[i] = strings.Replace(commandStrings[i], "  ", " ", -1)
	}

	// Writes the extension to extensionList.json and checks if the extension was already in storage
	err := db.SetGuildExtension(m.GuildID, commandStrings[1])
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("`%s` has been added to the file extension list.", commandStrings[1]))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Removes a blacklisted file extension from the blacklist
func unfilterExtensionCommand(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)
	guildExtensions := db.GetGuildExtensions(m.GuildID)

	if len(guildExtensions) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no blacklisted file extensions.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%sremoveextension [file extension]`\n\n[file extension] is the file extension you want to remove from the blacklist (e.g. .exe or .jpeg)", guildSettings.GetPrefix()))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Remove double spaces
	for i := 0; i < len(commandStrings); i++ {
		commandStrings[i] = strings.Replace(commandStrings[i], "  ", " ", -1)
	}

	// Removes extension from storage and memory
	err := db.SetGuildExtension(m.GuildID, commandStrings[1], true)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("`%s` has been removed from the file extension list.", commandStrings[1]))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Print file extensions from memory
func viewExtensionsCommand(s *discordgo.Session, m *discordgo.Message) {
	var extensions string

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildExtensions := db.GetGuildExtensions(m.GuildID)

	if len(guildExtensions) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no file extensions saved.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Iterates through all the file extensions in memory and adds them to the extensions string
	for ext := range guildExtensions {
		extensions += fmt.Sprintf("**.%s**\n", ext)
	}

	// Splits and sends message
	splitMessage := common.SplitLongMessage(extensions)
	for i := 0; i < len(splitMessage); i++ {
		_, err := s.ChannelMessageSend(m.ChannelID, splitMessage[i])
		if err != nil {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Cannot send file extensions message.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
		}
	}
}

// Adds file extension commands to the commandHandler
func init() {
	Add(&Command{
		Execute:    filterExtensionCommand,
		Trigger:    "addextension",
		Aliases:    []string{"filterextension", "extension"},
		Desc:       "Adds a file extension to the extension blacklist/whitelist",
		Permission: functionality.Mod,
		Module:     "filters",
	})
	Add(&Command{
		Execute:    unfilterExtensionCommand,
		Trigger:    "removeextension",
		Aliases:    []string{"killextension", "unextension"},
		Desc:       "Removes a file extension from the extension blacklist/whitelist",
		Permission: functionality.Mod,
		Module:     "filters",
	})
	Add(&Command{
		Execute:    viewExtensionsCommand,
		Trigger:    "extensions",
		Aliases:    []string{"filextensions", "filteredextensions", "printextensions"},
		Desc:       "Prints the file extension blacklist/whitelist",
		Permission: functionality.Mod,
		Module:     "filters",
	})
}
