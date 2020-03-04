package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"github.com/r-anime/ZeroTsu/functionality"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	CommandMap = make(map[string]*Command)
	aliasMap   = make(map[string]string)
)

type Command struct {
	Execute     func(*discordgo.Session, *discordgo.Message)
	Trigger     string
	Aliases     []string
	Desc        string
	DeleteAfter bool
	Permission  functionality.Permission
	Module      string
	DMAble      bool
}

func Add(c *Command) {
	CommandMap[c.Trigger] = c
	for _, alias := range c.Aliases {
		aliasMap[alias] = c.Trigger
	}
	log.Printf("Added command %s | %d aliases | %v module", c.Trigger, len(c.Aliases), c.Module)
}

// HandleCommand handles the incoming message
func HandleCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in HandleCommand with message: " + m.Content)
			if m.GuildID != "" {
				log.Println("Guild ID: " + m.GuildID)
				guildSettings := db.GetGuildSettings(m.GuildID)
				log.Println(guildSettings.GetPrefix())
			}
		}
	}()

	if m == nil {
		return
	}
	if m.Author == nil {
		return
	}
	if m.Author.Bot {
		return
	}
	if m.Author.ID == "" {
		return
	}
	if m.Message == nil {
		return
	}
	if m.Message.Content == "" {
		return
	}

	// Handle guild command if it's coming from a server
	if m.GuildID != "" {
		handleGuild(s, m)
		return
	}

	// Parse the command
	var guildPrefix = "."
	if len(m.Message.Content) <= 1 || m.Message.Content[0:len(guildPrefix)] != guildPrefix {
		return
	}
	cmdTrigger := strings.Split(m.Content, " ")[0][len(guildPrefix):]
	cmdTrigger = strings.ToLower(cmdTrigger)
	cmd, ok := CommandMap[cmdTrigger]
	if !ok {
		cmd, ok = CommandMap[aliasMap[cmdTrigger]]
		if !ok {
			return
		}
	}

	// Allow only normal DMable commands
	if !cmd.DMAble {
		return
	}

	// Execute the command
	cmd.Execute(s, m.Message)
}

// Handles a command from a guild
func handleGuild(s *discordgo.Session, m *discordgo.MessageCreate) {
	entities.HandleNewGuild(m.GuildID)

	guildSettings := db.GetGuildSettings(m.GuildID)

	if len(m.Message.Content) <= len(guildSettings.GetPrefix()) || m.Message.Content[0:len(guildSettings.GetPrefix())] != guildSettings.GetPrefix() {
		return
	}
	cmdTrigger := strings.Split(m.Content, " ")[0][len(guildSettings.GetPrefix()):]
	cmdTrigger = strings.ToLower(cmdTrigger)
	cmd, ok := CommandMap[cmdTrigger]
	if !ok {
		cmd, ok = CommandMap[aliasMap[cmdTrigger]]
		if !ok {
			return
		}
	}
	if cmd.Trigger == "votecategory" ||
		cmd.Trigger == "startvote" {
		if !guildSettings.GetVoteModule() {
			return
		}
	}
	if cmd.Module == "waifus" {
		if !guildSettings.GetWaifuModule() {
			return
		}
	}
	if cmd.Module == "reacts" {
		if !guildSettings.GetReactsModule() {
			return
		}
	}
	if cmd.Permission != functionality.User || guildSettings.GetModOnly() {
		if !functionality.HasElevatedPermissions(s, m.Author.ID, m.GuildID) {
			return
		}
	}

	// Sanitize mentions
	m.Content = strings.ReplaceAll(m.Content, "@here", "@\u200Bhere")
	if m.MentionEveryone {
		m.Content = strings.ReplaceAll(m.Content, "@everyone", "@\u200Beveryone")
	}
	if len(m.MentionRoles) != 0 {
		for _, roleID := range m.MentionRoles {
			m.Content = strings.ReplaceAll(m.Content, fmt.Sprintf("<@&%s>", roleID), fmt.Sprintf("<@\u200B&%s>", roleID))
		}
	}

	cmd.Execute(s, m.Message)
	if cmd.DeleteAfter {
		err := s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			return
		}
	}
}

// Handles a command from DMs
func handleDM(s *discordgo.Session, m *discordgo.MessageCreate) {

}
