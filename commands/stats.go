package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/functionality"
)

var Today = time.Now()

// Adds to message count on every message for that channel
func OnMessageChannel(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in OnMessageChannel")
		}
	}()

	if m.GuildID == "" {
		return
	}

	functionality.HandleNewGuild(s, m.GuildID)

	var channelStatsVar functionality.Channel
	t := time.Now()

	functionality.Mutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()

	// Sets channel params if it didn't exist before in database
	if _, ok := functionality.GuildMap[m.GuildID].ChannelStats[m.ChannelID]; !ok {

		// Fetches all guild info
		guild, err := s.State.Guild(m.GuildID)
		if err != nil {
			guild, err = s.Guild(m.GuildID)
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
		}
		// Fetches channel info
		channel, err := s.State.Channel(m.ChannelID)
		if err != nil {
			channel, err = s.Channel(m.ChannelID)
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
		}

		channelStatsVar.ChannelID = channel.ID
		channelStatsVar.Name = channel.Name
		channelStatsVar.RoleCount = make(map[string]int)
		channelStatsVar.RoleCount[channel.Name] = functionality.GetRoleUserAmount(guild, guild.Roles, channel.Name)

		// Removes role stat for channels without associated roles. Else turns bool to true
		if channelStatsVar.RoleCount[channel.Name] == 0 {
			channelStatsVar.RoleCount = nil
		} else {
			channelStatsVar.Optin = true
		}

		channelStatsVar.Messages = make(map[string]int)
		channelStatsVar.Exists = true
		functionality.GuildMap[m.GuildID].ChannelStats[m.ChannelID] = &channelStatsVar
	}
	if functionality.GuildMap[m.GuildID].ChannelStats[m.ChannelID].ChannelID == "" {
		functionality.GuildMap[m.GuildID].ChannelStats[m.ChannelID].ChannelID = m.ChannelID
	}

	functionality.GuildMap[m.GuildID].ChannelStats[m.ChannelID].Messages[t.Format(functionality.DateFormat)]++
	functionality.Mutex.Unlock()
}

// Prints all channel stats
func showStats(s *discordgo.Session, m *discordgo.Message) {

	var (
		msgs               []string
		normalChannelTotal int
		optinChannelTotal  int
		flag               bool
		channels           []*functionality.Channel
		t                  time.Time
	)

	// Print either Today or yesterday based on whether it's the bot that called the func
	functionality.Mutex.Lock()
	if m.Author.ID == s.State.User.ID {
		t = Today
	} else {
		t = time.Now()
	}

	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()

	// Fixes channels without ID param
	for id := range functionality.GuildMap[m.GuildID].ChannelStats {
		if functionality.GuildMap[m.GuildID].ChannelStats[id].ChannelID == "" {
			functionality.GuildMap[m.GuildID].ChannelStats[id].ChannelID = id
			flag = true
		}
	}

	// Writes channel stats to disk if IDs were fixed
	if flag {
		_, err := functionality.ChannelStatsWrite(functionality.GuildMap[m.GuildID].ChannelStats, m.GuildID)
		if err != nil {
			functionality.Mutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
	}

	// Sorts channel by their message use
	for _, channel := range functionality.GuildMap[m.GuildID].ChannelStats {
		channels = append(channels, channel)
	}
	sort.Sort(byFrequencyChannel(channels))

	// Calculates normal channels and optin channels message totals
	for chas := range functionality.GuildMap[m.GuildID].ChannelStats {
		if !functionality.GuildMap[m.GuildID].ChannelStats[chas].Optin {
			for date := range functionality.GuildMap[m.GuildID].ChannelStats[chas].Messages {
				normalChannelTotal += functionality.GuildMap[m.GuildID].ChannelStats[chas].Messages[date]
			}
		} else {
			for date := range functionality.GuildMap[m.GuildID].ChannelStats[chas].Messages {
				optinChannelTotal += functionality.GuildMap[m.GuildID].ChannelStats[chas].Messages[date]
			}
		}
	}
	functionality.Mutex.Unlock()

	// Fetches info on server roles from the server and puts it in deb
	deb, err := s.GuildRoles(m.GuildID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Confirms whether optins exist
	err = functionality.OptInsHandler(s, m.ChannelID, m.GuildID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Fetches all guild info
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		guild, err = s.Guild(m.GuildID)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
	}

	// Updates opt-in-under and opt-in-above position for use later in isChannelUsable func
	functionality.Mutex.Lock()
	guildSettings = functionality.GuildMap[m.GuildID].GetGuildSettings()
	for i := 0; i < len(deb); i++ {
		if deb[i].ID == functionality.GuildMap[m.GuildID].GuildConfig.OptInUnder.ID {
			functionality.GuildMap[m.GuildID].GuildConfig.OptInUnder.Position = deb[i].Position
		} else if deb[i].ID == functionality.GuildMap[m.GuildID].GuildConfig.OptInAbove.ID {
			functionality.GuildMap[m.GuildID].GuildConfig.OptInAbove.Position = deb[i].Position
		}
	}

	// Adds the channels and their stats to message and formats it
	message := "```CSS\nName:                            ([Daily Messages] | [Total Messages]) \n\n"
	for _, channel := range channels {

		// Checks if channel exists and sets optin status
		channel, ok := isChannelUsable(*channel, guild)
		if !ok {
			continue
		}
		// Formats  and splits message
		if !channel.Optin {
			message += lineSpaceFormatChannel(channel.ChannelID, false, m.GuildID, t)
			message += "\n"
		}
		msgs, message = splitStatMessages(msgs, message)
	}

	message += fmt.Sprintf("\nNormal Total: %d\n\n------", normalChannelTotal)
	message += "\n\nOpt-in Name:                     ([Daily Messages] | [Total Messages] | [Role Members]) \n\n"

	for _, channel := range channels {

		// Checks if channel exists and sets optin status
		channel, ok := isChannelUsable(*channel, guild)
		if !ok {
			continue
		}

		if channel.Optin {

			// Formats  and splits message
			message += lineSpaceFormatChannel(channel.ChannelID, true, m.GuildID, t)
			msgs, message = splitStatMessages(msgs, message)
		}
	}

	message += fmt.Sprintf("\nOpt-in Total: %d\n\n------\n", optinChannelTotal)
	message += fmt.Sprintf("\nGrand Total Messages: %d\n\n", optinChannelTotal+normalChannelTotal)
	message += fmt.Sprintf("\nDaily User Change: %d\n\n", functionality.GuildMap[m.GuildID].UserChangeStats[t.Format(functionality.DateFormat)])
	if len(functionality.GuildMap[m.GuildID].VerifiedStats) != 0 && config.Website != "" {
		message += fmt.Sprintf("\nDaily Verified Change: %d\n\n", functionality.GuildMap[m.GuildID].VerifiedStats[t.Format(functionality.DateFormat)])
	}
	functionality.Mutex.Unlock()

	// Final message split for last block + formatting
	msgs, message = splitStatMessages(msgs, message)
	if message != "" {
		msgs = append(msgs, message)
	}
	msgs[0] += "```"
	for i := 1; i < len(msgs); i++ {
		msgs[i] = "```CSS\n" + msgs[i] + "\n```"
	}

	for j := 0; j < len(msgs); j++ {
		_, _ = s.ChannelMessageSend(m.ChannelID, msgs[j])
	}
}

// Sort functions for channels by message use
type byFrequencyChannel []*functionality.Channel

func (e byFrequencyChannel) Len() int {
	return len(e)
}
func (e byFrequencyChannel) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
func (e byFrequencyChannel) Less(i, j int) bool {
	var (
		jTotalMessages int
		iTotalMessages int
	)
	for date := range e[j].Messages {
		jTotalMessages += e[j].Messages[date]
	}
	for date := range e[i].Messages {
		iTotalMessages += e[i].Messages[date]
	}
	return jTotalMessages < iTotalMessages
}

// Formats the line space length for the above to keep level spacing
func lineSpaceFormatChannel(id string, optin bool, guildID string, t time.Time) string {

	var totalMessages int

	for date := range functionality.GuildMap[guildID].ChannelStats[id].Messages {
		totalMessages += functionality.GuildMap[guildID].ChannelStats[id].Messages[date]
	}
	line := fmt.Sprintf("%s", functionality.GuildMap[guildID].ChannelStats[id].Name)
	spacesRequired := 33 - len(functionality.GuildMap[guildID].ChannelStats[id].Name)
	for i := 0; i < spacesRequired; i++ {
		line += " "
	}
	line += fmt.Sprintf("([%d])", functionality.GuildMap[guildID].ChannelStats[id].Messages[t.Format(functionality.DateFormat)])
	spacesRequired = 51 - len(line)
	for i := 0; i < spacesRequired; i++ {
		line += " "
	}
	line += fmt.Sprintf("| ([%d]) ", totalMessages)
	spacesRequired = 70 - len(line)
	for i := 0; i < spacesRequired; i++ {
		line += " "
	}
	if optin {
		line += fmt.Sprintf("| [%d])\n", functionality.GuildMap[guildID].ChannelStats[id].RoleCount[t.Format(functionality.DateFormat)])
	}

	return line
}

// Adds 1 to User Change on member join
func OnMemberJoin(s *discordgo.Session, u *discordgo.GuildMemberAdd) {
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in OnMemberJoin")
		}
	}()

	if u.GuildID == "" {
		return
	}

	t := time.Now()

	functionality.Mutex.Lock()
	if _, ok := functionality.GuildMap[u.GuildID]; !ok {
		functionality.InitDB(s, u.GuildID)
		functionality.LoadGuilds()
	}

	functionality.GuildMap[u.GuildID].UserChangeStats[t.Format(functionality.DateFormat)]++
	functionality.Mutex.Unlock()
}

// Removes 1 from User Change on member removal
func OnMemberRemoval(s *discordgo.Session, u *discordgo.GuildMemberRemove) {
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in OnMemberRemoval")
		}
	}()

	if u.GuildID == "" {
		return
	}

	t := time.Now()

	functionality.Mutex.Lock()
	if _, ok := functionality.GuildMap[u.GuildID]; !ok {
		functionality.InitDB(s, u.GuildID)
		functionality.LoadGuilds()
	}

	functionality.GuildMap[u.GuildID].UserChangeStats[t.Format(functionality.DateFormat)]--
	functionality.Mutex.Unlock()
}

// Checks if specific channel stat should be printed
func isChannelUsable(channel functionality.Channel, guild *discordgo.Guild) (functionality.Channel, bool) {

	// Checks if channel exists and if it's optin
	for guildIndex := range guild.Channels {
		for roleIndex := range guild.Roles {
			if guild.Roles[roleIndex].Position < functionality.GuildMap[guild.ID].GuildConfig.OptInUnder.Position &&
				guild.Roles[roleIndex].Position > functionality.GuildMap[guild.ID].GuildConfig.OptInAbove.Position &&
				guild.Channels[guildIndex].Name == guild.Roles[roleIndex].Name {
				channel.Optin = true
				break
			} else {
				channel.Optin = false
			}
		}
		if guild.Channels[guildIndex].Name == channel.Name &&
			guild.Channels[guildIndex].ID == channel.ChannelID {
			channel.Exists = true
			break
		} else {
			channel.Exists = false
		}
	}
	functionality.GuildMap[guild.ID].ChannelStats[channel.ChannelID] = &channel

	if channel.Exists {
		return channel, true
	}
	return channel, false
}

// Splits the stat messages into blocks
func splitStatMessages(msgs []string, message string) ([]string, string) {
	const maxMsgLength = 1700
	if len(message) > maxMsgLength {
		msgs = append(msgs, message)
		message = ""
	}
	return msgs, message
}

// Posts daily stats and update schedule command
func dailyStats(s *discordgo.Session) {

	var (
		message discordgo.Message
		author  discordgo.User
	)

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in dailyStats")
		}
	}()

	t := time.Now()

	functionality.Mutex.RLock()
	if Today.Day() == t.Day() {
		functionality.Mutex.RUnlock()
		return
	}
	functionality.Mutex.RUnlock()

	// Update daily anime schedule
	UpdateAnimeSchedule()
	ResetSubscriptions()

	folders, err := ioutil.ReadDir("database/guilds")
	if err != nil {
		log.Panicln(err)
		return
	}

	// Sleeps until anime schedule is definitely updated
	time.Sleep(10 * time.Second)

	functionality.Mutex.Lock()
	for _, f := range folders {
		if !f.IsDir() {
			continue
		}
		guildID := f.Name()

		// Sends daily schedule if need be
		DailySchedule(s, guildID)

		guildPrefix := functionality.GuildMap[guildID].GuildConfig.Prefix
		var guildDailyStatsID string
		if dailystats, ok := functionality.GuildMap[guildID].Autoposts["dailystats"]; !ok {
			continue
		} else {
			if dailystats == nil {
				continue
			}
			if dailystats.ID == "" {
				continue
			}
			guildDailyStatsID = dailystats.ID
		}

		_, err := s.ChannelMessageSend(guildDailyStatsID, fmt.Sprintf("Stats for **%v %v, %v**", Today.Month(), Today.Day(), Today.Year()))
		if err != nil {
			continue
		}

		author.ID = s.State.User.ID
		message.GuildID = guildID
		message.Author = &author
		message.Content = guildPrefix + "stats"
		message.ChannelID = guildDailyStatsID

		// Check for when stats don't display possibly due to malformed message
		if author.ID == "" || message.GuildID == "" ||
			message.Author == nil || message.Content == "" ||
			message.ChannelID == "" {
			log.Println("ERROR: MALFORMED DAILY STATS MESSAGE")
			log.Println("author.ID: " + author.ID)
			log.Println("message.GuildID: " + message.GuildID)
			log.Println("message.Author: " + message.Author.String())
			log.Println("message.Content: " + message.Content)
			log.Println("message.ChannelID: " + message.ChannelID)
		}

		functionality.Mutex.Unlock()
		showStats(s, &message)
		functionality.Mutex.Lock()
	}

	Today = t
	functionality.Mutex.Unlock()
}

// Daily stat update timer
func DailyStatsTimer(s *discordgo.Session, e *discordgo.Ready) {
	for range time.NewTicker(1 * time.Minute).C {
		dailyStats(s)
	}
}

// Adds channel stats command to the commandHandler
func init() {
	functionality.Add(&functionality.Command{
		Execute:    showStats,
		Trigger:    "stats",
		Aliases:    []string{"channelstats", "channels", "stat", "chanstat", "chanstats", "statss"},
		Desc:       "Prints all channel stats",
		Permission: functionality.Mod,
		Module:     "stats",
	})
}
