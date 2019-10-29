package functionality

import (
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
	Permission  permission
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
	if m.Message.Content[0:len(guildPrefix)] != guildPrefix {
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
	HandleNewGuild(s, m.GuildID)

	Mutex.RLock()
	guildSettings := GuildMap[m.GuildID].GetGuildSettings()
	Mutex.RUnlock()

	if m.Message.Content[0:len(guildSettings.Prefix)] != guildSettings.Prefix {
		return
	}
	cmdTrigger := strings.Split(m.Content, " ")[0][len(guildSettings.Prefix):]
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
		if !guildSettings.VoteModule {
			return
		}
	}
	if cmd.Module == "waifus" {
		if !guildSettings.WaifuModule {
			return
		}
	}
	if cmd.Module == "reacts" {
		if !guildSettings.ReactsModule {
			return
		}
	}
	if cmd.Permission != User {
		if !HasElevatedPermissions(s, m.Author.ID, m.GuildID) {
			return
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

// Inits guild if it's not in memory
func HandleNewGuild(s *discordgo.Session, guildID string) {
	Mutex.RLock()
	if _, ok := GuildMap[guildID]; !ok {
		Mutex.RUnlock()
		InitDB(s, guildID)
		Mutex.Lock()
		LoadGuild(guildID)
		Mutex.Unlock()
		return
	}
	Mutex.RUnlock()
}
