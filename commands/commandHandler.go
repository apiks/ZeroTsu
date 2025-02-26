package commands

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"github.com/r-anime/ZeroTsu/functionality"
)

var (
	SlashCommands         []*discordgo.ApplicationCommand
	SlashCommandsHandlers = make(map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate))
	CommandMap            = make(map[string]*Command)
	aliasMap              = make(map[string]string)
)

type Command struct {
	Execute     func(*discordgo.Session, *discordgo.Message)
	Name        string
	Aliases     []string
	Desc        string
	DeleteAfter bool
	Permission  functionality.Permission
	Module      string
	DMAble      bool
	Options     []*discordgo.ApplicationCommandOption
	Handler     func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

func Add(c *Command) {
	CommandMap[c.Name] = c
	for _, alias := range c.Aliases {
		aliasMap[alias] = c.Name
	}
	if c.Handler != nil {
		SlashCommands = append(SlashCommands, &discordgo.ApplicationCommand{
			Name:        c.Name,
			Description: c.Desc,
			Options:     c.Options,
		})
		SlashCommandsHandlers[c.Name] = c.Handler
	}

	log.Printf("Added command %s | %d aliases | %v module", c.Name, len(c.Aliases), c.Module)
}

// HandleCommand handles the incoming message
func HandleCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in HandleCommand with message: " + m.Content)
			if m.GuildID != "" {
				log.Println("Guild ID: " + m.GuildID)
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
	err := entities.InitGuildIfNotExists(m.GuildID)
	if err != nil {
		return
	}
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

func IsValidSlashCommand(s *discordgo.Session, cmdTrigger, authorID, guildID string) bool {
	cmd, ok := CommandMap[cmdTrigger]
	if !ok {
		cmd, ok = CommandMap[aliasMap[cmdTrigger]]
		if !ok {
			return false
		}
	}
	guildSettings := db.GetGuildSettings(guildID)
	if cmd.Module == "reacts" {
		if !guildSettings.GetReactsModule() {
			return false
		}
	}
	if cmd.Permission != functionality.User || guildSettings.GetModOnly() {
		if !functionality.HasElevatedPermissions(s, authorID, guildID) {
			return false
		}
	}

	return true
}

func VerifySlashCommand(s *discordgo.Session, cmdTrigger string, i *discordgo.InteractionCreate) error {
	if i.GuildID == "" {
		return errors.New("Error: This command is available only for moderators or admins in servers, not DMs.")
	}

	userID := ""
	if i.Member == nil {
		userID = i.User.ID
	} else {
		userID = i.Member.User.ID
	}

	if !IsValidSlashCommand(s, cmdTrigger, userID, i.GuildID) {
		return errors.New("Error: You do not have permissions to do this command.")
	}

	return nil
}

func RegisterSlashCommands(_ *discordgo.Session, _ *discordgo.Ready) {
	for _, v := range SlashCommands {
		err := config.Mgr.ApplicationCommandCreate("", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
	}
}
