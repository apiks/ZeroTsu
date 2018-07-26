package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

var (
	whitelist = [...]string{"png", "gif", "gifv",
		"jpeg", "jpg", "bmp", "tif", "tiff"}
)

// Checks messages with uploads if they're uploading a whitelisted file type. If not it removes them
func MessageAttachmentsHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if len(m.Attachments) == 0 {
		return
	}
	// Checks if it's within the /r/anime server
	ch, err := s.State.Channel(m.ChannelID)
	if err != nil {
		ch, err = s.Channel(m.ChannelID)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}
	if ch.GuildID != config.ServerID {
		return
	}
	// Pulls info on message author
	mem, err := s.State.Member(config.ServerID, m.Author.ID)
	if err != nil {
		mem, err = s.GuildMember(config.ServerID, m.Author.ID)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
	// Checks if user is mod before checking the message
	if misc.HasPermissions(mem) {
		return
	}

	// Iterates through all the attachments (since more than one can be posted in one go)
	// and checks if it's an allowed file type. If it isn't sends error message for each file
	for _, attachment := range m.Attachments {
		if isAllowed(attachment.Filename) {
			continue
		}

		// Deletes the message that was sent if has a non-whitelisted attachment
		err = s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {

			fmt.Println("Error:", err)
		}

		// Stores time of removal
		now := time.Now().Format("2006-01-02 15:04:05")

		// Prints success in bot-log channel
		_, err = s.ChannelMessageSend(config.BotLogID, m.Author.Mention() + " had their message removed for uploading non-whitelisted `" +
			attachment.Filename + "` in " + "<#" + m.ChannelID + "> on [_" + now + "_]")
		if err != nil {
			fmt.Println("Error: ", err)
		}

		// Sends a message to the user in their DMs
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			return
		}
		_, err = s.ChannelMessageSend(dm.ID, "Your message upload `" + attachment.Filename + "` was removed for using a non-whitelisted file type. Only gifs and images are allowed.")
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