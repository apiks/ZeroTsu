package events

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"github.com/r-anime/ZeroTsu/functionality"
	"log"
	"strings"
)

var Key = []byte("VfBhgLzmD4QH3W94pjgdbH8Tyv2HPRzq")

// Checks if user exists in memberInfo on joining server and adds him if he doesn't
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

		// Encrypts id
		ciphertext := common.Encrypt(Key, e.User.ID)

		// Sends verification message to user in DMs if possible
		if config.Website != "" && e.GuildID == "267799767843602452" {
			dm, _ := s.UserChannelCreate(e.User.ID)
			_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("You have joined the /r/anime discord. We require a reddit account verification with an at least 1 week old account. \n"+
				"Please verify your reddit account at http://%s/verification?reqvalue=%s", config.Website, ciphertext))
		}
	} else if mem.GetRedditUsername() == "" && config.Website != "" && e.GuildID == "267799767843602452" {
		// If user is already in memberInfo and lacks a reddit username and site is enabled, tell user to verify

		// Encrypts id
		ciphertext := common.Encrypt(Key, e.User.ID)

		// Sends verification message to user in DMs if possible
		dm, _ := s.UserChannelCreate(e.User.ID)
		_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("You have joined the /r/anime discord. We require a reddit account verification with an at least 1 week old account. \n"+
			"Please verify your reddit account at http://%s/verification?reqvalue=%s", config.Website, ciphertext))
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
	}

	// Checks if the discrim in database is the same as the discrim used by the memberInfoUser. If not it changes it
	if mem.GetDiscrim() != e.User.Discriminator && e.User.Discriminator != "" {
		mem = mem.SetDiscrim(e.User.Discriminator)
	}
}

// OnPresenceUpdate listens for user updates to compare usernames and discrim
func OnPresenceUpdate(_ *discordgo.Session, e *discordgo.PresenceUpdate) {
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in OnPresenceUpdate")
		}
	}()

	if e.GuildID == "" {
		return
	}

	entities.HandleNewGuild(e.GuildID)

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
	}

	// Checks if the discrim in database is the same as the discrim used by the memberInfoUser. If not it changes it
	if mem.GetDiscrim() != e.User.Discriminator && e.User.Discriminator != "" {
		mem = mem.SetDiscrim(e.User.Discriminator)
	}
}
