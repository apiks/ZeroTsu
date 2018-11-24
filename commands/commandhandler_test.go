package commands

import (
	"fmt"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/config"
)

var (
	executedCommand bool

	fakeCommand = func(s *discordgo.Session, m *discordgo.Message) {
		fmt.Printf("Successfully executed the command %s\n", m.Content)
		executedCommand = true
	}

	commandMetadata = &command{
		execute: fakeCommand,
		trigger: "command",
		aliases: []string{"cmd", "alias"},
	}

	elevatedCommandMetadata = &command{
		execute:  fakeCommand,
		trigger:  "elevated",
		elevated: true,
	}

	genericUser = &discordgo.User{ID: "1"}
	session     = &discordgo.Session{}

	fakeElevatedRole = "12345"
)

func TestGoodCommandMessage(t *testing.T) {
	commandText := fmt.Sprintf("%scommand", config.BotPrefix)
	teardown := setupTest(t)
	defer teardown(t)
	messageCreate := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			ID:      "1",
			Content: commandText,
			Author:  genericUser,
		},
	}
	HandleCommand(session, messageCreate)
	if !executedCommand {
		t.Errorf("Did not execute command")
	}
}

func TestBadCommandMessage(t *testing.T) {
	commandText := "no prefix command"
	teardown := setupTest(t)
	defer teardown(t)
	messageCreate := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			ID:      "1",
			Content: commandText,
			Author:  genericUser,
		},
	}
	HandleCommand(session, messageCreate)
	if executedCommand {
		t.Errorf("Executed malformed commmand")
	}
}

func TestCommandAliases(t *testing.T) {
	commandText := fmt.Sprintf("%salias", config.BotPrefix)
	commandText2 := fmt.Sprintf("%scmd", config.BotPrefix)
	teardown := setupTest(t)
	defer teardown(t)
	messageCreate := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			ID:      "1",
			Content: commandText,
			Author:  genericUser,
		},
	}
	HandleCommand(session, messageCreate)
	if !executedCommand {
		t.Errorf("Did not execute command")
	}
	teardown(t)
	setupTest(t)
	messageCreate.Message.Content = commandText2
	HandleCommand(session, messageCreate)
	if !executedCommand {
		t.Errorf("Did not execute command")
	}
}

func TestElevatedGood(t *testing.T) {
	commandText := fmt.Sprintf("%selevated", config.BotPrefix)
	teardown := setupTest(t)
	defer teardown(t)
	messageCreate := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			ID:      "1",
			Content: commandText,
			Author:  genericUser,
		},
	}
	HandleCommand(session, messageCreate)
	if !executedCommand {
		t.Error("Did not execute elevated command")
	}
}

func TestElevatedBad(t *testing.T) {
	commandText := fmt.Sprintf("%selevated", config.BotPrefix)
	teardown := setupTest(t)
	defer teardown(t)
	messageCreate := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			ID:      "1",
			Content: commandText,
			Author: &discordgo.User{
				ID: "2",
			},
		},
	}
	HandleCommand(session, messageCreate)
	if executedCommand {
		t.Error("Did not execute elevated command")
	}
}

func setupTest(t *testing.T) func(t *testing.T) {
	executedCommand = false
	return func(t *testing.T) {
		executedCommand = false
	}
}

func init() {
	config.BotPrefix = "!"
	session.State = discordgo.NewState()
	config.CommandRoles = append(config.CommandRoles, fakeElevatedRole)
	add(commandMetadata)
	add(elevatedCommandMetadata)
	session.State.GuildAdd(&discordgo.Guild{
		ID: config.ServerID,
	})
	session.State.MemberAdd(&discordgo.Member{
		GuildID: config.ServerID,
		Roles:   []string{fakeElevatedRole},
		User: &discordgo.User{
			ID: "1",
		},
	})
	session.State.MemberAdd(&discordgo.Member{
		GuildID: config.ServerID,
		Roles:   []string{},
		User: &discordgo.User{
			ID: "2",
		},
	})
	session.State.User = &discordgo.User{
		ID: "801",
	}
}
