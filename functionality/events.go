package functionality

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/config"
)

var darlingTrigger int

// Periodic events such as Unbanning and RSS timer every 30 sec
func StatusReady(s *discordgo.Session, e *discordgo.Ready) {

	MapMutex.Lock()
	for _, guild := range e.Guilds {

		// Initialize guild if missing
		HandleNewGuild(s, guild.ID)

		// Clean up SpoilerRoles.json in each guild
		err := cleanSpoilerRoles(s, guild.ID)
		if err != nil {
			guildSettings := GuildMap[guild.ID].GetGuildSettings()
			LogError(s, guildSettings.BotLog, err)
		}

		DynamicNicknameChange(s, guild.ID)
	}

	// Updates playing status
	if len(config.PlayingMsg) > 1 {
		rand.Seed(time.Now().UnixNano())
		randInt := rand.Intn(len(config.PlayingMsg))
		_ = s.UpdateStatus(0, config.PlayingMsg[randInt])
	} else if len(config.PlayingMsg) == 1 {
		_ = s.UpdateStatus(0, config.PlayingMsg[0])
	} else {
		_ = s.UpdateStatus(0, "")
	}
	MapMutex.Unlock()

	// Sends server count to bot list sites if it's the public ZeroTsu
	sendServers(s)
}

// Adds the voice role whenever a user joins the config voice chat
func VoiceRoleHandler(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in VoiceRoleHandler")
		}
	}()

	if v.GuildID == "" {
		return
	}

	MapMutex.Lock()
	HandleNewGuild(s, v.GuildID)
	guildSettings := GuildMap[v.GuildID].GetGuildSettings()
	MapMutex.Unlock()

	if len(guildSettings.VoiceChas) == 0 {
		return
	}

	var (
		voiceChannels  = guildSettings.VoiceChas
		noRemovalRoles []*Role
		dontRemove     bool
	)

	// Goes through each guild voice channel and removes/adds roles
	for _, cha := range voiceChannels {
		for _, chaRole := range cha.Roles {

			// Resets value
			dontRemove = false

			// Adds role
			if v.ChannelID == cha.ID {
				err := s.GuildMemberRoleAdd(v.GuildID, v.UserID, chaRole.ID)
				if err != nil {
					return
				}
				noRemovalRoles = append(noRemovalRoles, chaRole)
			}

			// Checks if this role should be removable
			for _, role := range noRemovalRoles {
				if chaRole.ID == role.ID {
					dontRemove = true
				}
			}
			if dontRemove {
				continue
			}

			// Removes role
			err := s.GuildMemberRoleRemove(v.GuildID, v.UserID, chaRole.ID)
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

	var guildSettings = &GuildSettings{
		Prefix: ".",
	}

	if m.GuildID != "" {
		MapMutex.Lock()
		HandleNewGuild(s, m.GuildID)
		*guildSettings = GuildMap[m.GuildID].GetGuildSettings()
		MapMutex.Unlock()
	}

	if strings.ToLower(m.Content) == fmt.Sprintf("<@%v> good bot", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v> good bot", s.State.User.ID) {
		_, err := s.ChannelMessageSend(m.ChannelID, "Thank you ❤")
		if err != nil {
			LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && m.Author.ID == "128312718779219968" {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Professor!\n\nPrefix: `%v`", guildSettings.Prefix))
		if err != nil {
			LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && m.Author.ID == "66207186417627136" {

		randomNum := rand.Intn(5)
		if randomNum == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Bug hunter!\n\nPrefix: `%v`", guildSettings.Prefix))
			if err != nil {
				LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Player!\n\nPrefix: `%v`", guildSettings.Prefix))
			if err != nil {
				LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Big brain!\n\nPrefix: `%v`", guildSettings.Prefix))
			if err != nil {
				LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 3 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Poster expert!\n\nPrefix: `%v`", guildSettings.Prefix))
			if err != nil {
				LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 4 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Idiot!\n\nPrefix: `%v`", guildSettings.Prefix))
			if err != nil {
				LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		return
	}

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && m.Author.ID == "365245718866427904" {

		randomNum := rand.Intn(5)
		if randomNum == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Begone ethot.\n\nPrefix: `%v`", guildSettings.Prefix))
			if err != nil {
				LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Humph!\n\nPrefix: `%v`", guildSettings.Prefix))
			if err != nil {
				LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Wannabe ethot!\n\nPrefix: `%v`", guildSettings.Prefix))
			if err != nil {
				LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 3 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Not even worth my time.\n\nPrefix: `%v`", guildSettings.Prefix))
			if err != nil {
				LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 4 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Okay, maybe you're not that bad.\n\nPrefix: `%v`", guildSettings.Prefix))
			if err != nil {
				LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		return
	}

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && m.Author.ID == "315201054377771009" {

		randomNum := rand.Intn(5)
		if randomNum == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("https://cdn.discordapp.com/attachments/618463738504151086/619090216329674800/uiz31mhq12k11.gif\n\nPrefix: `%v`", guildSettings.Prefix))
			if err != nil {
				LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Onii-chan no ecchi!\n\nPrefix: `%v`", guildSettings.Prefix))
			if err != nil {
				LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Kusuguttai Neiru-kun.\n\nPrefix: `%v`", guildSettings.Prefix))
			if err != nil {
				LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 3 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Liking lolis isn't a crime, but I'll still visit you in prison.\n\nPrefix: `%v`", guildSettings.Prefix))
			if err != nil {
				LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		if randomNum == 4 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Iris told me you wanted her to meow at you while she was still young.\n\nPrefix: `%v`", guildSettings.Prefix))
			if err != nil {
				LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
		return
	}

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && darlingTrigger > 10 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Daaarling~\n\nPrefix: `%v`", guildSettings.Prefix))
		if err != nil {
			LogError(s, guildSettings.BotLog, err)
			return
		}
		darlingTrigger = 0
		return
	}

	if m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID) {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Baka!\n\nPrefix: `%v`", guildSettings.Prefix))
		if err != nil {
			LogError(s, guildSettings.BotLog, err)
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

	MapMutex.Lock()
	HandleNewGuild(s, e.GuildID)

	for _, user := range GuildMap[e.GuildID].PunishedUsers {
		if user.ID == e.User.ID {
			MapMutex.Unlock()
			return
		}
	}

	guildSettings := GuildMap[e.GuildID].GetGuildSettings()
	MapMutex.Unlock()
	if guildSettings.BotLog == nil {
		return
	}
	if guildSettings.BotLog.ID == "" {
		return
	}
	_, _ = s.ChannelMessageSend(guildSettings.BotLog.ID, fmt.Sprintf("%v#%v was manually permabanned. ID: %v", e.User.Username, e.User.Discriminator, e.User.ID))
}

// Sends a message to a channel to log whenever a user joins. Intended use was to catch spambots for r/anime
// Now also serves for the mute command
func GuildJoin(s *discordgo.Session, u *discordgo.GuildMemberAdd) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in GuildJoin")
		}
	}()

	if u.GuildID == "" {
		return
	}

	// Gives the user the muted role if he is muted and has rejoined the server
	MapMutex.Lock()
	for _, punishedUser := range GuildMap[u.GuildID].PunishedUsers {
		if punishedUser.ID == u.User.ID {
			t := time.Now()
			zeroTimeValue := time.Time{}
			if punishedUser.UnmuteDate == zeroTimeValue {
				continue
			}
			muteDifference := t.Sub(punishedUser.UnmuteDate)
			if muteDifference > 0 {
				continue
			}

			if GuildMap[u.GuildID].GuildConfig.MutedRole != nil {
				if GuildMap[u.GuildID].GuildConfig.MutedRole.ID != "" {
					_ = s.GuildMemberRoleAdd(u.GuildID, punishedUser.ID, GuildMap[u.GuildID].GuildConfig.MutedRole.ID)
				}
			} else {
				// Pulls info on server roles
				deb, _ := s.GuildRoles(u.GuildID)

				// Checks by string for a muted role
				for _, role := range deb {
					if strings.ToLower(role.Name) == "muted" || strings.ToLower(role.Name) == "t-mute" {
						_ = s.GuildMemberRoleAdd(u.GuildID, punishedUser.ID, role.ID)
						break
					}
				}
			}
		}
	}

	if u.GuildID != "267799767843602452" {
		MapMutex.Unlock()
		return
	}
	if _, ok := GuildMap[u.GuildID]; !ok {
		InitDB(s, u.GuildID)
		LoadGuilds()
	}
	MapMutex.Unlock()

	creationDate, err := CreationTime(u.User.ID)
	if err != nil {

		MapMutex.Lock()
		guildSettings := GuildMap[u.GuildID].GetGuildSettings()
		MapMutex.Unlock()

		LogError(s, guildSettings.BotLog, err)
		return
	}

	// Sends user join message for r/anime discord server
	if u.GuildID == "267799767843602452" {
		_, _ = s.ChannelMessageSend("566233292026937345", fmt.Sprintf("User joined the server: %v\nAccount age: %v", u.User.Mention(), creationDate.String()))
	}
}

// Sends a message to suspected spambots to verify and bans them immediately after. Only does it for accounts younger than 3 days
func SpambotJoin(s *discordgo.Session, u *discordgo.GuildMemberAdd) {
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in SpambotJoin")
		}
	}()

	if u.GuildID == "" {
		return
	}

	var (
		creationDate time.Time
		now          time.Time

		temp    PunishedUsers
		tempMem UserInfo

		dmMessage string
	)

	MapMutex.Lock()
	HandleNewGuild(s, u.GuildID)
	guildSettings := GuildMap[u.GuildID].GetGuildSettings()
	MapMutex.Unlock()

	// Fetches date of account creation and checks if it's younger than 14 days
	creationDate, err := CreationTime(u.User.ID)
	if err != nil {
		LogError(s, guildSettings.BotLog, err)
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

	MapMutex.Lock()

	// Initializes user if he's not in memberInfo
	if _, ok := GuildMap[u.GuildID].MemberInfoMap[u.User.ID]; !ok {
		InitializeUser(u.Member, u.GuildID)
	}

	// Checks if the user is verified
	if _, ok := GuildMap[u.GuildID].MemberInfoMap[u.User.ID]; ok {
		if GuildMap[u.GuildID].MemberInfoMap[u.User.ID].RedditUsername != "" {
			MapMutex.Unlock()
			return
		}
	}

	// Checks if they're using a default avatar
	if u.User.Avatar != "" {
		MapMutex.Unlock()
		return
	}

	// Adds the spambot ban to PunishedUsers so it doesn't Trigger the OnGuildBan func
	temp.ID = u.User.ID
	temp.User = u.User.Username
	temp.UnbanDate = time.Date(9999, 9, 9, 9, 9, 9, 9, time.Local)
	for index, val := range GuildMap[u.GuildID].PunishedUsers {
		if val.ID == u.User.ID {
			GuildMap[u.GuildID].PunishedUsers = append(GuildMap[u.GuildID].PunishedUsers[:index], GuildMap[u.GuildID].PunishedUsers[index+1:]...)
		}
	}
	GuildMap[u.GuildID].PunishedUsers = append(GuildMap[u.GuildID].PunishedUsers, &temp)
	_ = PunishedUsersWrite(GuildMap[u.GuildID].PunishedUsers, u.GuildID)

	// Adds a bool to memberInfo that it's a suspected spambot account in case they try to reverify
	tempMem = *GuildMap[u.GuildID].MemberInfoMap[u.User.ID]
	tempMem.SuspectedSpambot = true
	GuildMap[u.GuildID].MemberInfoMap[u.User.ID] = &tempMem
	_ = WriteMemberInfo(GuildMap[u.GuildID].MemberInfoMap, u.GuildID)
	MapMutex.Unlock()

	// Sends a message to the user warning them in case it's a false positive
	dmMessage = "You have been suspected of being a spambot and banned."
	if u.GuildID == "267799767843602452" {
		dmMessage += "\nTo get unbanned please do our mandatory verification process at https://%v/verification and then rejoin the server."
	}

	dm, _ := s.UserChannelCreate(u.User.ID)
	_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf(dmMessage, config.Website))

	// Bans the suspected account
	err = s.GuildBanCreateWithReason(u.GuildID, u.User.ID, "Autoban Spambot Account", 0)
	if err != nil {
		LogError(s, guildSettings.BotLog, err)
		return
	}

	// Botlog message
	if guildSettings.BotLog == nil {
		return
	}
	if guildSettings.BotLog.ID == "" {
		return
	}
	_, _ = s.ChannelMessageSend(guildSettings.BotLog.ID, fmt.Sprintf("Suspected spambot was banned. User: <@!%v>", u.User.ID))
}

// Cleans spoilerroles.json
func cleanSpoilerRoles(s *discordgo.Session, guildID string) error {

	var shouldDelete bool

	// Pulls all of the server roles
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		guildSettings := GuildMap[guildID].GetGuildSettings()
		LogError(s, guildSettings.BotLog, err)
		return err
	}

	// Removes roles not found in spoilerRoles.json
	for _, spoilerRole := range GuildMap[guildID].SpoilerMap {
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
			SpoilerRolesDelete(spoilerRole.ID, guildID)
		}
	}

	SpoilerRolesWrite(GuildMap[guildID].SpoilerMap, guildID)
	LoadGuildFile(guildID, "spoilerRoles.json")

	return nil
}

// Handles BOT joining a server
func GuildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	MapMutex.Lock()
	HandleNewGuild(s, g.ID)
	MapMutex.Unlock()

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

	guildPrefix := GuildMap[guildID].GuildConfig.Prefix

	// Set custom nickname based on guild prefix if there is no existing nickname
	me, err := s.State.Member(guildID, s.State.User.ID)
	if err != nil {
		me, err = s.GuildMember(guildID, s.State.User.ID)
		if err != nil {
			return
		}
	}

	if me.Nick != "" {
		targetPrefix := guildPrefix
		if len(oldPrefix) > 0 {
			if oldPrefix[0] != "" {
				targetPrefix = oldPrefix[0]
			}
		}
		if me.Nick != fmt.Sprintf("%s %s", targetPrefix, s.State.User.Username) && me.Nick != "" {
			return
		}
	}

	err = s.GuildMemberNickname(guildID, "@me", fmt.Sprintf("%s %s", guildPrefix, s.State.User.Username))
	if err != nil {
		if _, ok := err.(*discordgo.RESTError); ok {
			if err.(*discordgo.RESTError).Response.Status == "400 Bad Request" {
				_ = s.GuildMemberNickname(guildID, "@me", fmt.Sprintf("%s", s.State.User.Username))
			}
		}
	}
}
