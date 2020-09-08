package events

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"github.com/r-anime/ZeroTsu/functionality"
	"log"
	"math/rand"
	"runtime/debug"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
)

var darlingTrigger int

// Status Ready Events
func StatusReady(s *discordgo.Session, e *discordgo.Ready) {
	var guildIds []string

	for _, guild := range e.Guilds {
		guildIds = append(guildIds, guild.ID)
	}

	for _, guildID := range guildIds {
		// Initialize guild if missing
		entities.HandleNewGuild(guildID)

		// Clean up SpoilerRoles.json in each guild
		err := cleanSpoilerRoles(s, guildID)
		if err != nil {
			log.Println(err)
		}

		// Handles Unbans and Unmutes
		punishmentHandler(s, guildID)

		// Handles RemindMes
		remindMeHandler(s, guildID)

		// Reload null guild anime subs
		fixGuildSubsCommand(guildID)
	}

	// Handles Reddit Feeds
	feedHandler(s, guildIds)

	// Updates playing status
	var randomPlayingMsg string
	rand.Seed(time.Now().UnixNano())
	entities.Mutex.RLock()
	if len(config.PlayingMsg) > 1 {
		randomPlayingMsg = config.PlayingMsg[rand.Intn(len(config.PlayingMsg))]
	}
	entities.Mutex.RUnlock()
	if randomPlayingMsg != "" {
		_ = s.UpdateStatus(0, randomPlayingMsg)
	}

	// Sends server count to bot list sites if it's the public ZeroTsu
	functionality.SendServers(s)
}

// Adds the voice role whenever a user joins the config voice chat
func VoiceRoleHandler(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in VoiceRoleHandler")
			log.Println("stacktrace from panic: \n" + string(debug.Stack()))
		}
	}()

	if v.GuildID == "" {
		return
	}

	entities.HandleNewGuild(v.GuildID)

	guildSettings := db.GetGuildSettings(v.GuildID)
	if guildSettings.GetVoiceChas() == nil || len(guildSettings.GetVoiceChas()) == 0 {
		return
	}

	var (
		noRemovalRoles []entities.Role
		dontRemove     bool
	)

	// Goes through each guild voice channel and removes/adds roles
	for _, cha := range guildSettings.GetVoiceChas() {
		for _, chaRole := range cha.GetRoles() {

			// Resets value
			dontRemove = false

			// Adds role
			if v.ChannelID == cha.GetID() {
				err := s.GuildMemberRoleAdd(v.GuildID, v.UserID, chaRole.GetID())
				if err != nil {
					return
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

// Print fluff message on bot ping
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
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Professor!\n\nPrefix: `%v`", guildSettings.GetPrefix()))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && m.Author.ID == "66207186417627136" {

		randomNum := rand.Intn(5)
		if randomNum == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Bug hunter!\n\nPrefix: `%v`", guildSettings.GetPrefix()))
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Player!\n\nPrefix: `%v`", guildSettings.GetPrefix()))
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Big brain!\n\nPrefix: `%v`", guildSettings.GetPrefix()))
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 3 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Poster expert!\n\nPrefix: `%v`", guildSettings.GetPrefix()))
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 4 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Idiot!\n\nPrefix: `%v`", guildSettings.GetPrefix()))
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
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Begone ethot.\n\nPrefix: `%v`", guildSettings.GetPrefix()))
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Humph!\n\nPrefix: `%v`", guildSettings.GetPrefix()))
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Wannabe ethot!\n\nPrefix: `%v`", guildSettings.GetPrefix()))
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 3 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Not even worth my time.\n\nPrefix: `%v`", guildSettings.GetPrefix()))
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 4 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Okay, maybe you're not that bad.\n\nPrefix: `%v`", guildSettings.GetPrefix()))
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
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("https://cdn.discordapp.com/attachments/618463738504151086/619090216329674800/uiz31mhq12k11.gif\n\nPrefix: `%v`", guildSettings.GetPrefix()))
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Onii-chan no ecchi!\n\nPrefix: `%v`", guildSettings.GetPrefix()))
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Kusuguttai Neiru-kun.\n\nPrefix: `%v`", guildSettings.GetPrefix()))
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 3 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Liking lolis isn't a crime, but I'll still visit you in prison.\n\nPrefix: `%v`", guildSettings.GetPrefix()))
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 4 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Iris told me you wanted her to meow at you while she was still young.\n\nPrefix: `%v`", guildSettings.GetPrefix()))
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		return
	}

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && darlingTrigger > 10 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Daaarling~\n\nPrefix: `%v`", guildSettings.GetPrefix()))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		darlingTrigger = 0
		return
	}

	if m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID) {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Baka!\n\nPrefix: `%v`", guildSettings.GetPrefix()))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		darlingTrigger++
	}
}

// If there's a manual ban handle it
func OnGuildBan(s *discordgo.Session, e *discordgo.GuildBanAdd) {
	if e.GuildID == "" {
		return
	}

	entities.HandleNewGuild(e.GuildID)

	// Check if a bot did the banning
	auditLog, err := s.GuildAuditLog(e.GuildID, e.User.ID, "", int(discordgo.AuditLogActionMemberBanAdd), 10)
	if err == nil {
		for _, entry := range auditLog.AuditLogEntries {
			if entry.TargetID == e.User.ID {
				userBanning, err := s.User(entry.UserID)
				if err != nil {
					continue
				}
				if userBanning.Bot {
					return
				}
				break
			}
		}
	}

	user := db.GetGuildPunishedUser(e.GuildID, e.User.ID)
	if user != (entities.PunishedUsers{}) {
		return
	}

	guildSettings := db.GetGuildSettings(e.GuildID)
	if guildSettings.BotLog == (entities.Cha{}) || guildSettings.BotLog.GetID() == "" {
		return
	}

	_, _ = s.ChannelMessageSend(guildSettings.BotLog.GetID(), fmt.Sprintf("%s#%s was manually permabanned. ID: %s", e.User.Username, e.User.Discriminator, e.User.ID))
}

// Sends a message to a channel to log whenever a user joins. Intended use was to catch spambots for r/anime
// Now also serves for the mute command
func GuildJoin(s *discordgo.Session, u *discordgo.GuildMemberAdd) {
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in GuildJoin")
			log.Println("stacktrace from panic: \n" + string(debug.Stack()))
		}
	}()

	if u.GuildID == "" {
		return
	}

	entities.HandleNewGuild(u.GuildID)

	// Gives the user the muted role if he is muted and has rejoined the server
	punishedUser := db.GetGuildPunishedUser(u.GuildID, u.User.ID)
	if punishedUser != (entities.PunishedUsers{}) && punishedUser.GetUnmuteDate() != (time.Time{}) {
		t := time.Now()
		muteDifference := t.Sub(punishedUser.GetUnmuteDate())

		if muteDifference <= 0 {
			guildSettings := db.GetGuildSettings(u.GuildID)
			if guildSettings.GetMutedRole() == (entities.Role{}) || guildSettings.GetMutedRole().GetID() == "" {
				deb, _ := s.GuildRoles(u.GuildID)
				for _, role := range deb {
					if strings.ToLower(role.Name) == "muted" || strings.ToLower(role.Name) == "t-mute" {
						_ = s.GuildMemberRoleAdd(u.GuildID, punishedUser.GetID(), role.ID)
						break
					}
				}
			} else {
				_ = s.GuildMemberRoleAdd(u.GuildID, punishedUser.GetID(), guildSettings.GetMutedRole().GetID())
			}
		}
	}
}

// Cleans spoilerRoles.json
func cleanSpoilerRoles(s *discordgo.Session, guildID string) error {
	var shouldDelete bool

	// Pulls all of the server roles
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		guildSettings := db.GetGuildSettings(guildID)
		common.LogError(s, guildSettings.BotLog, err)
		return err
	}

	// Removes roles not found in spoilerRoles.json
	guildSpoilerMap := db.GetGuildSpoilerMap(guildID)
	for _, spoilerRole := range guildSpoilerMap {
		if spoilerRole == nil {
			continue
		}

		shouldDelete = true
		for _, role := range roles {
			if role.ID == spoilerRole.ID {
				shouldDelete = false

				// Updates names
				if strings.ToLower(role.Name) != strings.ToLower(spoilerRole.Name) {
					spoilerRole.Name = role.Name
				}
				break
			}
		}
		if shouldDelete {
			db.SetGuildSpoilerRole(guildID, spoilerRole, true)
		}
	}

	return nil
}

// Handles BOT joining a server
func GuildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	// Send message to support server mod log that a server has been created on the public ZeroTsu
	entities.Guilds.RLock()
	if _, ok := entities.Guilds.DB[g.Guild.ID]; !ok {
		if s.State.User.ID == "614495694769618944" {
			_, _ = s.ChannelMessageSend("619899424428130315", fmt.Sprintf("A DB entry has been created for guild: %s", g.Name))
		}
	}
	entities.Guilds.RUnlock()

	entities.HandleNewGuild(g.ID)
	log.Println(fmt.Sprintf("Joined guild %s", g.Guild.Name))
}

// Logs BOT leaving a server
func GuildDelete(_ *discordgo.Session, g *discordgo.GuildDelete) {
	if g.Name == "" {
		return
	}
	log.Println(fmt.Sprintf("Left guild %s", g.Guild.Name))
}

// Changes the BOT's nickname dynamically to a `prefix username` format if there is no existing custom nickname
func DynamicNicknameChange(s *discordgo.Session, guildID string) {
	guildSettings := db.GetGuildSettings(guildID)

	// Set custom nickname based on guild prefix if there is no existing nickname
	bot, err := s.State.Member(guildID, s.State.User.ID)
	if err != nil {
		bot, err = s.GuildMember(guildID, s.State.User.ID)
		if err != nil {
			return
		}
	}

	if bot.Nick != "" {
		return
	}
	if bot.Nick != fmt.Sprintf("%s %s", guildSettings.GetPrefix(), s.State.User.Username) {
		return
	}

	err = s.GuildMemberNickname(guildID, "@me", fmt.Sprintf("%s %s", guildSettings.GetPrefix(), s.State.User.Username))
	if err != nil {
		if _, ok := err.(*discordgo.RESTError); ok {
			if err.(*discordgo.RESTError).Response.Status == "400 Bad Request" {
				_ = s.GuildMemberNickname(guildID, "@me", fmt.Sprintf("%s", s.State.User.Username))
			}
		}
	}
}

// Fixes broken anime guild subs that are null
func fixGuildSubsCommand(guildID string) {
	entities.Mutex.Lock()
	for ID, subs := range entities.SharedInfo.GetAnimeSubsMap() {
		if subs != nil || ID != guildID {
			continue
		}

		entities.SetupGuildSub(guildID)
		break
	}
	entities.Mutex.Unlock()
}
