package commands

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
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

	if _, ok := misc.GuildMap[m.GuildID]; !ok {
		misc.InitDB(m.GuildID)
		misc.LoadGuilds()
	}

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	guildExtensionFilterType := misc.GuildMap[m.GuildID].GuildConfig.WhitelistFileFilter
	misc.MapMutex.Unlock()

	if m.Author.ID == s.State.User.ID {
		return
	}
	if len(m.Attachments) == 0 {
		return
	}
	// Pulls info on message author
	mem, err := s.State.Member(m.GuildID, m.Author.ID)
	if err != nil {
		mem, err = s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			return
		}
	}
	// Checks if user is mod before checking the message
	misc.MapMutex.Lock()
	if HasElevatedPermissions(s, mem.User.ID, m.GuildID) {
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()

	// Iterates through all the attachments (since more than one can be posted in one go)
	// and checks if it's a banned file type. If it is then remove them
	for _, attachment := range m.Attachments {

		if guildExtensionFilterType {
			if isBannedExtension(attachment.Filename, m.GuildID) {
				continue
			}
		} else {
			if !isBannedExtension(attachment.Filename, m.GuildID) {
				continue
			}
		}

		// Deletes the message that was sent if has a non-whitelisted attachment
		err = s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}

		now := time.Now().Format("2006-01-02 15:04:05")

		// Prints success in bot-log channel
		if guildExtensionFilterType {
			_, _ = s.ChannelMessageSend(guildBotLog, m.Author.Mention()+" had their message removed for uploading a non-whitelisted file type `"+
				attachment.Filename+"` in "+"<#"+m.ChannelID+"> on [_"+now+"_]")
		} else {
			_, _ = s.ChannelMessageSend(guildBotLog, m.Author.Mention()+" had their message removed for uploading a blacklisted file type `"+
				attachment.Filename+"` in "+"<#"+m.ChannelID+"> on [_"+now+"_]")
		}

		// Sends a message to the user in their DMs if possible
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			return
		}

		// Fetches all file extensions
		var extensions string
		misc.MapMutex.Lock()
		for ext := range misc.GuildMap[m.GuildID].ExtensionList {
			extensions += fmt.Sprintf("%v, ", ext)
		}
		misc.MapMutex.Unlock()
		extensions = strings.TrimSuffix(extensions, ", ")

		if len(extensions) == 0 {
			_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("Your message upload `%v` was removed for using a non-whitelisted file type.\n\nNo file attachments are allowed on that server.", attachment.Filename))
			return
		}

		if guildExtensionFilterType {
			_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("Your message upload `%v` was removed for using a non-whitelisted file type.\n\nAllowed file types are: `%v`", attachment.Filename, extensions))
		} else {
			_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("Your message upload `%v` was removed for using a blacklisted file type.", attachment.Filename))
		}
	}

}

// Checks if it's an banned file type and returns true if it is, else false
func isBannedExtension(filename, guildID string) bool {
	filename = strings.ToLower(filename)
	misc.MapMutex.Lock()
	for ext := range misc.GuildMap[guildID].ExtensionList {
		if strings.HasSuffix(filename, fmt.Sprintf(".%s", ext)) {
			misc.MapMutex.Unlock()
			return true
		}
	}
	misc.MapMutex.Unlock()
	return false
}

// Blacklists a file extension
func filterExtensionCommand(s *discordgo.Session, m *discordgo.Message) {

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	commandStrings := strings.SplitN(strings.ToLower(m.Content), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vextension [file extension]`\n\n[file extension] is the file extension (e.g. .exe or .jpeg).", guildPrefix))
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error())
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Remove double spaces
	for i := 0; i < len(commandStrings); i++ {
		commandStrings[i] = strings.Replace(commandStrings[i], "  ", " ", -1)
	}

	// Writes the extension to extensionList.json and checks if the extension was already in storage
	err := misc.ExtensionsWrite(commandStrings[1], m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("`%v` has been added to the file extension list.", commandStrings[1]))
	if err != nil {
		_, _ = s.ChannelMessageSend(guildBotLog, err.Error())
	}
}

// Removes a blacklisted file extension from the blacklist
func unfilterExtensionCommand(s *discordgo.Session, m *discordgo.Message) {

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID

	if len(misc.GuildMap[m.GuildID].ExtensionList) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no blacklisted file extensions.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error())
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

	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	misc.MapMutex.Unlock()

	commandStrings := strings.SplitN(strings.ToLower(m.Content), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vremoveextension [file extension]`\n\n[file extension] is the file extension you want to remove from the blacklist (e.g. .exe or .jpeg)", guildPrefix))
		if err != nil {
			_, _ = s.ChannelMessageSend(guildBotLog, err.Error())
			return
		}
		return
	}

	// Remove double spaces
	for i := 0; i < len(commandStrings); i++ {
		commandStrings[i] = strings.Replace(commandStrings[i], "  ", " ", -1)
	}

	// Removes extension from storage and memory
	err := misc.ExtensionsRemove(commandStrings[1], m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("`%v` has been removed from the file extension list.", commandStrings[1]))
	if err != nil {
		_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
	}
}

// Print file extensions from memory
func viewExtensionsCommand(s *discordgo.Session, m *discordgo.Message) {

	var extensions string

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID

	if len(misc.GuildMap[m.GuildID].ExtensionList) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no file extensions saved.")
		if err != nil {
			_, err := s.ChannelMessageSend(guildBotLog, err.Error())
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

	// Iterates through all the file extensions in memory and adds them to the extensions string
	for ext := range misc.GuildMap[m.GuildID].ExtensionList {
		extensions += fmt.Sprintf(".%v\n", ext)
	}
	misc.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, extensions)
	if err != nil {
		_, _ = s.ChannelMessageSend(guildBotLog, err.Error())
	}
}

// Adds file extension commands to the commandHandler
func init() {
	add(&command{
		execute:  filterExtensionCommand,
		trigger:  "addextension",
		aliases:  []string{"filterextension", "extension"},
		desc:     "Adds a file extension to the extension blacklist/whitelist.",
		elevated: true,
		category: "filters",
	})
	add(&command{
		execute:  unfilterExtensionCommand,
		trigger:  "removeextension",
		aliases:  []string{"killextension", "unextension"},
		desc:     "Removes a file extension from the extension blacklist/whitelist",
		elevated: true,
		category: "filters",
	})
	add(&command{
		execute:  viewExtensionsCommand,
		trigger:  "extensions",
		aliases:  []string{"filextensions", "filteredextensions", "printextensions"},
		desc:     "Prints the file extension blacklist/whitelist.",
		elevated: true,
		category: "filters",
	})
}
