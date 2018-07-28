package commands

import (
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

var (
	commandMap = make(map[string]*command)
	aliasMap   = make(map[string]string)
	l          = log.New(os.Stderr, "cmds: ", log.LstdFlags|log.Lshortfile)
)

type command struct {
	execute      func(*discordgo.Session, *discordgo.Message)
	trigger      string
	aliases      []string
	desc         string
	commandCount int
	deleteAfter  bool
	elevated     bool
}

func add(c *command) {
	commandMap[c.trigger] = c
	for _, alias := range c.aliases {
		aliasMap[alias] = c.trigger
	}
	l.Printf("Added command %s | %d aliases", c.trigger, len(c.aliases))
}

// HandleCommand handles the incoming message
func HandleCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string))
			if err != nil {

				l.Println(err.Error())
				l.Println(rec)
			}
		}
	}()

	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Message.Content[0:len(config.BotPrefix)] != config.BotPrefix {
		return
	}
	cmdTrigger := strings.Split(m.Content, " ")[0][len(config.BotPrefix):]
	cmdTrigger = strings.ToLower(cmdTrigger)
	cmd, ok := commandMap[cmdTrigger]
	if !ok {
		cmd, ok = commandMap[aliasMap[cmdTrigger]]
		if !ok {
			return
		}
	}
	if cmd.elevated && !hasElevatedPermissions(s, m.Author) {
		return
	}
	cmd.execute(s, m.Message)
	cmd.commandCount++
	if cmd.deleteAfter {
		s.ChannelMessageDelete(m.ChannelID, m.ID)
	}
}

func hasElevatedPermissions(s *discordgo.Session, u *discordgo.User) bool {
	mem, err := s.State.Member(config.ServerID, u.ID)
	if err != nil {
		mem, err = s.GuildMember(config.ServerID, u.ID)
		if err != nil {
			l.Panic(err)
		}
	}
	return misc.HasPermissions(mem)
}