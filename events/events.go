package events

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"github.com/r-anime/ZeroTsu/functionality"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
)

var darlingTrigger int
var GuildIds = &GuildIdsStruct{
	Ids: make(map[string]bool),
}

type GuildIdsStruct struct {
	sync.RWMutex
	Ids map[string]bool
}

func StatusReady(s *discordgo.Session, _ *discordgo.Ready) {
	var guildIds []string

	GuildIds.RLock()
	for gID := range GuildIds.Ids {
		guildIds = append(guildIds, gID)
	}
	GuildIds.RUnlock()

	for _, guildID := range guildIds {
		// Initialize guild if missing
		entities.HandleNewGuild(guildID)

		// Reload null guild anime subs
		fixGuildSubsCommand(guildID)
	}

	// Updates playing status
	var randomPlayingMsg string
	rand.Seed(time.Now().UnixNano())
	entities.Mutex.RLock()
	if len(config.PlayingMsg) > 1 {
		randomPlayingMsg = config.PlayingMsg[rand.Intn(len(config.PlayingMsg))]
	}
	entities.Mutex.RUnlock()
	if randomPlayingMsg != "" {
		_ = s.UpdateGameStatus(0, randomPlayingMsg)
	}

	// Sends server count to bot list sites if it's the public ZeroTsu
	guildCountStr := strconv.Itoa(config.Mgr.GuildCount())
	functionality.SendServers(guildCountStr, s)
}

// VoiceRoleHandler toggles a role whenever a user join/leave the specified voice channel
func VoiceRoleHandler(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	var (
		noRemovalRoles []entities.Role
		dontRemove     bool
	)

	if v.GuildID == "" {
		return
	}

	entities.HandleNewGuild(v.GuildID)

	guildSettings := db.GetGuildSettings(v.GuildID)
	if guildSettings.GetVoiceChas() == nil || len(guildSettings.GetVoiceChas()) == 0 {
		return
	}

	// Goes through each guild voice channel and removes/adds roles
	for _, cha := range guildSettings.GetVoiceChas() {
		// Check if the BOT has the necessary permissions
		perms, err := s.State.UserChannelPermissions(s.State.User.ID, cha.GetID())
		if err != nil {
			continue
		}
		if perms&discordgo.PermissionViewChannel != discordgo.PermissionViewChannel {
			continue
		}
		if perms&discordgo.PermissionManageRoles != discordgo.PermissionManageRoles {
			continue
		}

		for _, chaRole := range cha.GetRoles() {

			// Resets value
			dontRemove = false

			// Adds role
			if v.ChannelID == cha.GetID() {
				err := s.GuildMemberRoleAdd(v.GuildID, v.UserID, chaRole.GetID())
				if err != nil {
					continue
				}
				noRemovalRoles = append(noRemovalRoles, chaRole)
			}

			// Checks if this role should be removable
			for _, role := range noRemovalRoles {
				if chaRole.GetID() == role.GetID() {
					dontRemove = true
				}
			}
			if dontRemove {
				continue
			}

			// Removes role
			_ = s.GuildMemberRoleRemove(v.GuildID, v.UserID, chaRole.GetID())
		}
	}
}

// OnBotPing Prints fluff message on bot ping
func OnBotPing(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID == "" {
		return
	}
	if m.Author.Bot {
		return
	}

	var guildSettings = entities.GuildSettings{
		Prefix: ".",
	}

	if m.GuildID != "" {
		entities.HandleNewGuild(m.GuildID)
		guildSettings = db.GetGuildSettings(m.GuildID)
	}

	if strings.ToLower(m.Content) == fmt.Sprintf("<@%v> good bot", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v> good bot", s.State.User.ID) {
		_, err := s.ChannelMessageSend(m.ChannelID, "Thank you ❤")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && m.Author.ID == "128312718779219968" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Professor!")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && m.Author.ID == "66207186417627136" {
		randomNum := rand.Intn(5)
		if randomNum == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Bug hunter!")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Player!")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Big brain!")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 3 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Poster expert!")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 4 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Idiot!")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		return
	}

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && m.Author.ID == "365245718866427904" {
		randomNum := rand.Intn(5)
		if randomNum == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Begone ethot.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Humph!")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Wannabe ethot!")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 3 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Not even worth my time.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 4 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Okay, maybe you're not that bad.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		return
	}

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && m.Author.ID == "315201054377771009" {
		randomNum := rand.Intn(5)
		if randomNum == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, "https://cdn.discordapp.com/attachments/618463738504151086/619090216329674800/uiz31mhq12k11.gif")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Onii-chan no ecchi!")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Kusuguttai Neiru-kun.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 3 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Liking lolis isn't a crime, but I'll still visit you in prison.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 4 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Iris told me you wanted her to meow at you while she was still young.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		return
	}

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && darlingTrigger > 10 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Daaarling~")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		darlingTrigger = 0
		return
	}

	if m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID) {
		_, err := s.ChannelMessageSend(m.ChannelID, "Baka!")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		darlingTrigger++
	}
}

// GuildCreate Handles BOT joining a server
func GuildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	isNew, _ := entities.Guilds.Load(g.Guild.ID)
	if isNew && s.State.User.ID == "614495694769618944" {
		_, _ = s.ChannelMessageSend("619899424428130315", fmt.Sprintf("A DB entry has been created for guild: %s", g.Name))
	}

	entities.HandleNewGuild(g.ID)
	GuildIds.Lock()
	GuildIds.Ids[g.Guild.ID] = true
	GuildIds.Unlock()
	// log.Println(fmt.Sprintf("Joined guild %s", g.Guild.Name))
}

// GuildDelete logs BOT leaving a server
func GuildDelete(_ *discordgo.Session, g *discordgo.GuildDelete) {
	GuildIds.Lock()
	entities.Guilds.Lock()
	delete(GuildIds.Ids, g.Guild.ID)
	delete(entities.Guilds.DB, g.Guild.ID)
	entities.Guilds.Unlock()
	GuildIds.Unlock()
	log.Println(fmt.Sprintf("Left guild with ID: %s", g.Guild.ID))
}

// Fixes broken anime guild subs that are null
func fixGuildSubsCommand(guildID string) {
	for ID, subs := range entities.SharedInfo.GetAnimeSubsMap() {
		if subs != nil || ID != guildID {
			continue
		}

		entities.SetupGuildSub(guildID)
		_ = entities.AnimeSubsWrite(entities.SharedInfo.GetAnimeSubsMap())
		break
	}
}
