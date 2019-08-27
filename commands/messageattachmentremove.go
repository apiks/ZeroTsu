package commands

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

var (
	whitelist = [...]string{"png", "gif", "gifv",
		"jpeg", "jpg", "bmp", "tif", "tiff", "webm", "webps", "webp",
		"mp4", "ogg", "wmv", "3gp", "avi", "flv", "wav"}
	whitelistString = ".png, .gif, .gifv, .jpeg, .jpg, .bmp, .tif, .tiff, .webm, .webps, .webp, .mp4, .ogg, .wmv, .3gp, .avi, .flv, .wav"
)

// Checks messages with uploads if they're uploading a whitelisted file type. If not it removes them
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

	if m.Author.ID == s.State.User.ID {
		misc.InitDB(m.GuildID)
		misc.LoadGuilds()
	}

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	guildFileFilter := misc.GuildMap[m.GuildID].GuildConfig.FileFilter
	misc.MapMutex.Unlock()

	if !guildFileFilter {
		return
	}

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
	// and checks if it's an allowed file type. If it isn't sends error message for each file
	for _, attachment := range m.Attachments {
		if isAllowed(attachment.Filename) {
			continue
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
		_, _ = s.ChannelMessageSend(guildBotLog, m.Author.Mention()+" had their message removed for uploading non-whitelisted `"+
			attachment.Filename+"` in "+"<#"+m.ChannelID+"> on [_"+now+"_]")

		// Sends a message to the user in their DMs if possible
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			return
		}
		_, _ = s.ChannelMessageSend(dm.ID, "Your message upload `"+attachment.Filename+"` was removed for using a non-whitelisted file type.\n\nAllowed file types: `"+whitelistString+"`")
	}

}

// Checks if it's an allowed file type and returns true if it is, else false. By Kagumi
func isAllowed(filename string) bool {
	filename = strings.ToLower(filename)
	for _, ext := range whitelist {
		if strings.HasSuffix(filename, fmt.Sprintf(".%s", ext)) {
			return true
		}
	}
	return false
}
