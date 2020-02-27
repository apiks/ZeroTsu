package events

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"github.com/r-anime/ZeroTsu/functionality"
	"log"
	"math/rand"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
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

	var wg sync.WaitGroup
	wg.Add(len(guildIds))

	for _, guildID := range guildIds {
		go func(guildID string) {
			defer wg.Done()

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

			// Changes nickname dynamically based on prefix
			DynamicNicknameChange(s, guildID)
		}(guildID)
	}

	wg.Wait()

	// Handles Reddit Feeds
	err := feedHandler(s, guildIds)
	if err != nil {
		log.Println(err)
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
		noRemovalRoles []*entities.Role
		dontRemove     bool
	)

	// Goes through each guild voice channel and removes/adds roles
	for _, cha := range guildSettings.GetVoiceChas() {
		if cha == nil {
			continue
		}

		for _, chaRole := range cha.GetRoles() {
			if chaRole == nil {
				continue
			}

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
				if role == nil {
					continue
				}

				if chaRole.GetID() == role.GetID() {
					dontRemove = true
				}
			}
			if dontRemove {
				continue
			}

			// Removes role
			err := s.GuildMemberRoleRemove(v.GuildID, v.UserID, chaRole.GetID())
			if err != nil {
				return
			}
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

	var guildSettings = &entities.GuildSettings{
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
	auditLog, err := s.GuildAuditLog(e.GuildID, e.User.ID, "", discordgo.AuditLogActionMemberBanAdd, 10)
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
	if err == nil || user != nil {
		return
	}

	guildSettings := db.GetGuildSettings(e.GuildID)
	if guildSettings.BotLog == nil || guildSettings.BotLog.GetID() == "" {
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
	if punishedUser != nil && punishedUser.GetUnmuteDate() != (time.Time{}) {
		t := time.Now()
		muteDifference := t.Sub(punishedUser.GetUnmuteDate())

		if muteDifference <= 0 {
			guildSettings := db.GetGuildSettings(u.GuildID)
			if guildSettings == nil || guildSettings.GetMutedRole() == nil || guildSettings.GetMutedRole().GetID() == "" {
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

	if u.GuildID != "267799767843602452" {
		return
	}

	creationDate, err := common.CreationTime(u.User.ID)
	if err != nil {
		guildSettings := db.GetGuildSettings(u.GuildID)
		if guildSettings != nil {
			common.LogError(s, guildSettings.BotLog, err)
		}
		return
	}

	// Sends user join message for r/anime discord server
	if u.GuildID == "267799767843602452" {
		_, _ = s.ChannelMessageSend("566233292026937345", fmt.Sprintf("Username joined the server: %v\nAccount age: %s", u.User.Mention(), creationDate.String()))
	}
}

// Sends a message to suspected spambots to verify and bans them immediately after. Only does it for accounts younger than 3 days
func SpambotJoin(s *discordgo.Session, u *discordgo.GuildMemberAdd) {
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in SpambotJoin")
			log.Println("stacktrace from panic: \n" + string(debug.Stack()))
		}
	}()

	if u.GuildID == "" {
		return
	}

	var (
		creationDate time.Time
		now          time.Time

		dmMessage string
	)

	entities.HandleNewGuild(u.GuildID)

	guildSettings := db.GetGuildSettings(u.GuildID)
	guildPunishedUsers := db.GetGuildPunishedUsers(u.GuildID)

	// Fetches date of account creation and checks if it's younger than 14 days
	creationDate, err := common.CreationTime(u.User.ID)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
	now = time.Now()
	difference := now.Sub(creationDate)
	if difference.Hours() > 384 {
		return
	}

	// Matches known spambot patterns with regex
	regexCases := regexp.MustCompile(`(?im)(^[a-zA-Z]+\d{2,4}[a-zA-Z]+$)|(^[a-zA-Z]+\d{5}$)|(^[a-zA-Z]+\d{2,5}$)`)
	spambotMatches := regexCases.FindAllString(u.User.Username, 1)
	if len(spambotMatches) == 0 {
		return
	}

	// Checks if they're using a default avatar
	if u.User.Avatar != "" {
		return
	}

	// Initializes user if he's not in memberInfo
	memberInfoUser := db.GetGuildMember(u.GuildID, u.User.ID)
	if memberInfoUser == nil || memberInfoUser.GetID() == "" {
		functionality.InitializeUser(u.User, u.GuildID)
	}
	memberInfoUser = db.GetGuildMember(u.GuildID, u.User.ID)

	// Checks if the user is verified
	if memberInfoUser.GetRedditUsername() != "" {
		return
	}

	// Adds the spambot ban to PunishedUsers so it doesn't Trigger the OnGuildBan func
	temp := entities.NewPunishedUsers(u.User.ID, u.User.Username, time.Date(9999, 9, 9, 9, 9, 9, 9, time.Local), time.Time{})
	for _, punishedUser := range guildPunishedUsers {
		if punishedUser == nil {
			continue
		}

		if punishedUser.GetID() == u.User.ID {
			_ = db.SetGuildPunishedUser(u.GuildID, temp, true)
		}
	}
	err = db.SetGuildPunishedUser(u.GuildID, temp, true)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}

	// Adds a bool to memberInfo that it's a suspected spambot account in case they try to reverify
	memberInfoUser.SetSuspectedSpambot(true)
	db.SetGuildMember(u.GuildID, memberInfoUser)

	// Sends a message to the user warning them in case it's a false positive
	dmMessage = "You have been suspected of being a spambot and banned."
	if u.GuildID == "267799767843602452" {
		dmMessage += fmt.Sprintf("\nTo get unbanned please do our mandatory verification process at https://%s/verification and then rejoin the server.", config.Website)
	}

	dm, _ := s.UserChannelCreate(u.User.ID)
	_, _ = s.ChannelMessageSend(dm.ID, dmMessage)

	// Bans the suspected account
	err = s.GuildBanCreateWithReason(u.GuildID, u.User.ID, "Autoban Suspected Spambot", 0)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}

	// Botlog message
	if guildSettings.BotLog == nil || guildSettings.BotLog.GetID() == "" {
		return
	}
	_, _ = s.ChannelMessageSend(guildSettings.BotLog.GetID(), fmt.Sprintf("Suspected spambot was banned. Username: <@!%s>", u.User.ID))
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
			_ = db.SetGuildSpoilerRole(guildID, spoilerRole, true)
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
func DynamicNicknameChange(s *discordgo.Session, guildID string, oldPrefix ...string) {
	guildSettings := db.GetGuildSettings(guildID)

	// Set custom nickname based on guild prefix if there is no existing nickname
	me, err := s.State.Member(guildID, s.State.User.ID)
	if err != nil {
		me, err = s.GuildMember(guildID, s.State.User.ID)
		if err != nil {
			return
		}
	}

	if me.Nick != "" {
		targetPrefix := guildSettings.GetPrefix()
		if len(oldPrefix) > 0 {
			if oldPrefix[0] != "" {
				targetPrefix = oldPrefix[0]
			}
		}
		if me.Nick != fmt.Sprintf("%s %s", targetPrefix, s.State.User.Username) && me.Nick != "" {
			return
		}
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
