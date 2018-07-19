package commands

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
)

// Returns a message on "!about" for general information
func aboutCommand(s *discordgo.Session, m *discordgo.Message) {
	_, err := s.ChannelMessageSend(m.ChannelID, "Hello, darling. I'm ZeroTsu and was made by Professor Apiks for /r/anime. I'm written in Go. "+
		"He says I'm from Darling in the Franxx but that's just a bunch of nonsense to me. Use `!help` to list what commands are available to you. I hope you brought sweets.")
	if err != nil {

		fmt.Println("Error:", err)
	}
}

func init() {
	add(&command{
		execute: aboutCommand,
		trigger: "about",
		desc:    "Get info about me!",
	})
}
