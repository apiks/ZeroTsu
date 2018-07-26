package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// Returns a message on "ping" to see if bot is alive
func pingCommand(s *discordgo.Session, m *discordgo.Message) {
	_, err := s.ChannelMessageSend(m.ChannelID, "Hmm? Do you want some honey, darling? Open wide~~")
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func init() {
	add(&command{
		execute:  pingCommand,
		trigger:  "ping",
		aliases:  []string{"pingme"},
		desc:     "Am I alive?",
		elevated: true,
	})
}