package events

import (
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"github.com/r-anime/ZeroTsu/functionality"
	"log"
	"strings"
)

var Key = []byte("VfBhgLzmD4QH3W94pjgdbH8Tyv2HPRzq")

// Checks if user exists in memberInfo on joining server
// Also updates usernames and/or nicknames
// Also updates discriminator
// Also verifies them if they're already verified in memberinfo
func OnMemberJoinGuild(s *discordgo.Session, e *discordgo.GuildMemberAdd) {
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in OnMemberJoinGuild")
		}
	}()

	if e.GuildID == "" {
		return
	}

	entities.HandleNewGuild(e.GuildID)

	// Initializes and handles user if he's new
	mem := db.GetGuildMember(e.GuildID, e.User.ID)
	if mem.GetID() == "" {
		functionality.InitializeUser(e.User, e.GuildID)
		mem = db.GetGuildMember(e.GuildID, e.User.ID)
		if mem.GetID() == "" {
			return
		}
	}

	// Checks if the user's current username is the same as the one in the database. Otherwise updates
	if e.User.Username != mem.GetUsername() && e.User.Username != "" {
		flag := true

		for _, names := range mem.GetPastUsernames() {
			if strings.ToLower(names) == strings.ToLower(e.User.Username) {
				flag = false
				break
			}
		}

		if flag {
			mem = mem.AppendToPastUsernames(e.User.Username)
		}
		mem = mem.SetUsername(e.User.Username)
	}

	// Checks if the user's current nickname is the same as the one in the database. Otherwise updates
	if mem.GetNickname() != e.Nick && e.Nick != "" {
		flag := true

		for _, names := range mem.GetPastNicknames() {
			if strings.ToLower(names) == strings.ToLower(e.Nick) {
				flag = false
				break
			}
		}

		if flag {
			mem = mem.AppendToPastNicknames(e.Nick)
		}
		mem = mem.SetNickname(e.Nick)
	}

	// Checks if the discrim in database is the same as the discrim used by the user. If not it changes it
	if e.User.Discriminator != mem.GetDiscrim() && e.User.Discriminator != "" {
		mem = mem.SetDiscrim(e.User.Discriminator)
	}

	db.SetGuildMember(e.GuildID, mem)
}

// OnMemberUpdate listens for member updates to compare usernames, nicknames and discrim
func OnMemberUpdate(_ *discordgo.Session, e *discordgo.GuildMemberUpdate) {
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in OnMemberUpdate")
		}
	}()

	if e.GuildID == "" {
		return
	}
	entities.HandleNewGuild(e.GuildID)

	var writeFlag bool

	// Fetches user from memberInfo if possible
	mem := db.GetGuildMember(e.GuildID, e.User.ID)
	if mem.GetID() == "" {
		return
	}

	// Checks usernames and updates if needed
	if mem.GetUsername() != e.User.Username && e.User.Username != "" {
		flag := true

		for _, names := range mem.GetPastUsernames() {
			if strings.ToLower(names) == strings.ToLower(e.User.Username) {
				flag = false
				break
			}
		}

		if flag {
			mem = mem.AppendToPastUsernames(e.User.Username)
		}
		mem = mem.SetUsername(e.User.Username)
		writeFlag = true
	}

	// Checks nicknames and updates if needed
	if mem.GetNickname() != e.Nick && e.Nick != "" {
		flag := true

		for _, names := range mem.GetPastNicknames() {
			if strings.ToLower(names) == strings.ToLower(e.Nick) {
				flag = false
				break
			}
		}

		if flag {
			mem = mem.AppendToPastNicknames(e.Nick)
		}
		mem = mem.SetNickname(e.Nick)
		writeFlag = true
	}

	// Checks if the discrim in database is the same as the discrim used by the memberInfoUser. If not it changes it
	if mem.GetDiscrim() != e.User.Discriminator && e.User.Discriminator != "" {
		mem = mem.SetDiscrim(e.User.Discriminator)
		writeFlag = true
	}

	if !writeFlag {
		return
	}

	db.SetGuildMember(e.GuildID, mem)
}
