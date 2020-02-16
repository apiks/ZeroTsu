package commands

import (
	"fmt"
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

	if m.GuildID == "" {
		return
	}

	if m.Author.ID == s.State.User.ID || len(m.Attachments) == 0 {
		return
	}

	functionality.HandleNewGuild(s, m.GuildID)

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	// Checks if user is mod before checking the message
	if functionality.HasElevatedPermissions(s, m.Author.ID, m.GuildID) {
		return
	}

	// Iterates through all the attachments (since more than one can be posted in one go)
	// and checks if it's a banned file type. If it is then remove them
	for _, attachment := range m.Attachments {

		if guildSettings.WhitelistFileFilter {
			if isBannedExtension(attachment.Filename, m.GuildID) {
				continue
			}
		} else {
			if !isBannedExtension(attachment.Filename, m.GuildID) {
				continue
			}
		}

		// Deletes the message that was sent if has a non-whitelisted attachment
		err := s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}

		now := time.Now().Format("2006-01-02 15:04:05")

		// Prints success in bot-log channel
		if guildSettings.BotLog != nil {
			if guildSettings.BotLog.ID != "" {
				if guildSettings.WhitelistFileFilter {
					_, _ = s.ChannelMessageSend(guildSettings.BotLog.ID, m.Author.Mention()+" had their message removed for uploading a non-whitelisted file type `"+
						attachment.Filename+"` in "+"<#"+m.ChannelID+"> on [_"+now+"_]")
				} else {
					_, _ = s.ChannelMessageSend(guildSettings.BotLog.ID, m.Author.Mention()+" had their message removed for uploading a blacklisted file type `"+
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
		functionality.Mutex.RLock()
		guildExtensionList := functionality.GuildMap[m.GuildID].ExtensionList
		functionality.Mutex.RUnlock()
		for ext := range guildExtensionList {
			extensions += fmt.Sprintf("%v, ", ext)
		}
		extensions = strings.TrimSuffix(extensions, ", ")

		if len(extensions) == 0 {
			_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("Your message upload `%s` was removed for using a non-whitelisted file type.\n\nNo file attachments are allowed on that server.", attachment.Filename))
			return
		}

		if guildSettings.WhitelistFileFilter {
			_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("Your message upload `%s` was removed for using a non-whitelisted file type.\n\nAllowed file types are: `%v`", attachment.Filename, extensions))
		} else {
			_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("Your message upload `%s` was removed for using a blacklisted file type.", attachment.Filename))
		}
	}

}

// Checks if it's an banned file type and returns true if it is, else false
func isBannedExtension(filename, guildID string) bool {
	filename = strings.ToLower(filename)

	functionality.Mutex.RLock()
	guildExtensionList := functionality.GuildMap[guildID].ExtensionList
	functionality.Mutex.RUnlock()

	for ext := range guildExtensionList {
		if strings.HasSuffix(filename, fmt.Sprintf(".%s", ext)) {
			return true
		}
	}
	return false
}

// Blacklists a file extension
func filterExtensionCommand(s *discordgo.Session, m *discordgo.Message) {

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%sextension [file extension]`\n\n[file extension] is the file extension (e.g. .exe or .jpeg).", guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Remove double spaces
	for i := 0; i < len(commandStrings); i++ {
		commandStrings[i] = strings.Replace(commandStrings[i], "  ", " ", -1)
	}

	// Writes the extension to extensionList.json and checks if the extension was already in storage
	err := functionality.ExtensionsWrite(commandStrings[1], m.GuildID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("`%s` has been added to the file extension list.", commandStrings[1]))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Removes a blacklisted file extension from the blacklist
func unfilterExtensionCommand(s *discordgo.Session, m *discordgo.Message) {

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	guildExtensionListLen := len(functionality.GuildMap[m.GuildID].ExtensionList)
	functionality.Mutex.RUnlock()

	if guildExtensionListLen == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no blacklisted file extensions.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%sremoveextension [file extension]`\n\n[file extension] is the file extension you want to remove from the blacklist (e.g. .exe or .jpeg)", guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Remove double spaces
	for i := 0; i < len(commandStrings); i++ {
		commandStrings[i] = strings.Replace(commandStrings[i], "  ", " ", -1)
	}

	// Removes extension from storage and memory
	err := functionality.ExtensionsRemove(commandStrings[1], m.GuildID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("`%s` has been removed from the file extension list.", commandStrings[1]))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Print file extensions from memory
func viewExtensionsCommand(s *discordgo.Session, m *discordgo.Message) {

	var extensions string

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	guildExtensionList := functionality.GuildMap[m.GuildID].ExtensionList
	functionality.Mutex.RUnlock()

	if len(guildExtensionList) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no file extensions saved.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Iterates through all the file extensions in memory and adds them to the extensions string
	for ext := range guildExtensionList {
		extensions += fmt.Sprintf("**.%s**\n", ext)
	}

	// Splits and sends message
	splitMessage := functionality.SplitLongMessage(extensions)
	for i := 0; i < len(splitMessage); i++ {
		_, err := s.ChannelMessageSend(m.ChannelID, splitMessage[i])
		if err != nil {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Cannot send file extensions message.")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
		}
	}
}

// Adds file extension commands to the commandHandler
func init() {
	functionality.Add(&functionality.Command{
		Execute:    filterExtensionCommand,
		Trigger:    "addextension",
		Aliases:    []string{"filterextension", "extension"},
		Desc:       "Adds a file extension to the extension blacklist/whitelist",
		Permission: functionality.Mod,
		Module:     "filters",
	})
	functionality.Add(&functionality.Command{
		Execute:    unfilterExtensionCommand,
		Trigger:    "removeextension",
		Aliases:    []string{"killextension", "unextension"},
		Desc:       "Removes a file extension from the extension blacklist/whitelist",
		Permission: functionality.Mod,
		Module:     "filters",
	})
	functionality.Add(&functionality.Command{
		Execute:    viewExtensionsCommand,
		Trigger:    "extensions",
		Aliases:    []string{"filextensions", "filteredextensions", "printextensions"},
		Desc:       "Prints the file extension blacklist/whitelist",
		Permission: functionality.Mod,
		Module:     "filters",
	})
}
