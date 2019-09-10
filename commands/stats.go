package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
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

	var channelStatsVar misc.Channel
	t := time.Now()

	misc.MapMutex.Lock()
	if _, ok := misc.GuildMap[m.GuildID]; !ok {
		misc.InitDB(m.GuildID)
		misc.LoadGuilds()
	}

	if misc.GuildMap[m.GuildID] == nil {
		misc.MapMutex.Unlock()
		return
	}
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID

	// Sets channel params if it didn't exist before in database
	if _, ok := misc.GuildMap[m.GuildID].ChannelStats[m.ChannelID]; !ok {
		// Fetches all guild info
		guild, err := s.State.Guild(m.GuildID)
		if err != nil {
			guild, err = s.Guild(m.GuildID)
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
				if err != nil {
					misc.MapMutex.Unlock()
					return
				}
				misc.MapMutex.Unlock()
				return
			}
		}
		// Fetches channel info
		channel, err := s.State.Channel(m.ChannelID)
		if err != nil {
			channel, err = s.Channel(m.ChannelID)
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
				if err != nil {
					misc.MapMutex.Unlock()
					return
				}
				misc.MapMutex.Unlock()
				return
			}
		}

		channelStatsVar.ChannelID = channel.ID
		channelStatsVar.Name = channel.Name
		channelStatsVar.RoleCount = make(map[string]int)
		channelStatsVar.RoleCount[channel.Name] = misc.GetRoleUserAmount(guild, guild.Roles, channel.Name)

		// Removes role stat for channels without associated roles. Else turns bool to true
		if channelStatsVar.RoleCount[channel.Name] == 0 {
			channelStatsVar.RoleCount = nil
		} else {
			channelStatsVar.Optin = true
		}

		channelStatsVar.Messages = make(map[string]int)
		channelStatsVar.Exists = true
		misc.GuildMap[m.GuildID].ChannelStats[m.ChannelID] = &channelStatsVar
	}
	if misc.GuildMap[m.GuildID].ChannelStats[m.ChannelID].ChannelID == "" {
		misc.GuildMap[m.GuildID].ChannelStats[m.ChannelID].ChannelID = m.ChannelID
	}

	misc.GuildMap[m.GuildID].ChannelStats[m.ChannelID].Messages[t.Format(misc.DateFormat)]++
	misc.MapMutex.Unlock()
}

// Prints all channel stats
func showStats(s *discordgo.Session, m *discordgo.Message) {

	var (
		msgs               []string
		normalChannelTotal int
		optinChannelTotal  int
		flag               bool
		channels	   	   []*misc.Channel
		t				   time.Time
	)

	// Print either Today or yesterday based on whether it's the bot that called the func
	if m.Author.ID == s.State.User.ID {
		t = Today
	} else {
		t = time.Now()
	}

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID

	// Fixes channels without ID param
	for id := range misc.GuildMap[m.GuildID].ChannelStats {
		if misc.GuildMap[m.GuildID].ChannelStats[id].ChannelID == "" {
			misc.GuildMap[m.GuildID].ChannelStats[id].ChannelID = id
			flag = true
		}
	}

	// Writes channel stats to disk if IDs were fixed
	if flag {
		_, err := misc.ChannelStatsWrite(misc.GuildMap[m.GuildID].ChannelStats, m.GuildID)
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
	}

	// Sorts channel by their message use
	for _, channel := range misc.GuildMap[m.GuildID].ChannelStats {
		channels = append(channels, channel)
	}
	sort.Sort(byFrequencyChannel(channels))

	// Calculates normal channels and optin channels message totals
	for chas := range misc.GuildMap[m.GuildID].ChannelStats {
		if !misc.GuildMap[m.GuildID].ChannelStats[chas].Optin {
			for date := range misc.GuildMap[m.GuildID].ChannelStats[chas].Messages {
				normalChannelTotal += misc.GuildMap[m.GuildID].ChannelStats[chas].Messages[date]
			}
		} else {
			for date := range misc.GuildMap[m.GuildID].ChannelStats[chas].Messages {
				optinChannelTotal += misc.GuildMap[m.GuildID].ChannelStats[chas].Messages[date]
			}
		}
	}
	misc.MapMutex.Unlock()

	// Fetches info on server roles from the server and puts it in deb
	deb, err := s.GuildRoles(m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Confirms whether optins exist
	err = misc.OptInsHandler(s, m.ChannelID, m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Updates opt-in-under and opt-in-above position for use later in isChannelUsable func
	misc.MapMutex.Lock()
	for i := 0; i < len(deb); i++ {
		if deb[i].ID == misc.GuildMap[m.GuildID].GuildConfig.OptInUnder.ID {
			misc.GuildMap[m.GuildID].GuildConfig.OptInUnder.Position = deb[i].Position
		} else if deb[i].ID == misc.GuildMap[m.GuildID].GuildConfig.OptInAbove.ID {
			misc.GuildMap[m.GuildID].GuildConfig.OptInAbove.Position = deb[i].Position
		}
	}
	misc.MapMutex.Unlock()

	// Fetches all guild info
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		guild, err = s.Guild(m.GuildID)
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
	}

	// Adds the channels and their stats to message and formats it
	message := "```CSS\nName:                            ([Daily Messages] | [Total Messages]) \n\n"
	misc.MapMutex.Lock()
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
	message += fmt.Sprintf("\nDaily User Change: %d\n\n", misc.GuildMap[m.GuildID].UserChangeStats[t.Format(misc.DateFormat)])
	if len(misc.GuildMap[m.GuildID].VerifiedStats) != 0 && config.Website != "" {
		message += fmt.Sprintf("\nDaily Verified Change: %d\n\n", misc.GuildMap[m.GuildID].VerifiedStats[t.Format(misc.DateFormat)])
	}
	misc.MapMutex.Unlock()

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
		_, err := s.ChannelMessageSend(m.ChannelID, msgs[j])
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
	}
}

// Sort functions for emoji use by message use
type byFrequencyChannel []*misc.Channel

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

	for date := range misc.GuildMap[guildID].ChannelStats[id].Messages {
		totalMessages += misc.GuildMap[guildID].ChannelStats[id].Messages[date]
	}
	line := fmt.Sprintf("%v", misc.GuildMap[guildID].ChannelStats[id].Name)
	spacesRequired := 33 - len(misc.GuildMap[guildID].ChannelStats[id].Name)
	for i := 0; i < spacesRequired; i++ {
		line += " "
	}
	line += fmt.Sprintf("([%d])", misc.GuildMap[guildID].ChannelStats[id].Messages[t.Format(misc.DateFormat)])
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
		line += fmt.Sprintf("| [%d])\n", misc.GuildMap[guildID].ChannelStats[id].RoleCount[t.Format(misc.DateFormat)])
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

	misc.MapMutex.Lock()
	if _, ok := misc.GuildMap[u.GuildID]; !ok {
		misc.InitDB(u.GuildID)
		misc.LoadGuilds()
	}

	misc.GuildMap[u.GuildID].UserChangeStats[t.Format(misc.DateFormat)]++
	misc.MapMutex.Unlock()
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

	misc.MapMutex.Lock()
	if _, ok := misc.GuildMap[u.GuildID]; !ok {
		misc.InitDB(u.GuildID)
		misc.LoadGuilds()
	}

	misc.GuildMap[u.GuildID].UserChangeStats[t.Format(misc.DateFormat)]--
	misc.MapMutex.Unlock()
}

// Checks if specific channel stat should be printed
func isChannelUsable(channel misc.Channel, guild *discordgo.Guild) (misc.Channel, bool) {

	// Checks if channel exists and if it's optin
	for guildIndex := range guild.Channels {
		for roleIndex := range guild.Roles {
			if guild.Roles[roleIndex].Position < misc.GuildMap[guild.ID].GuildConfig.OptInUnder.Position &&
				guild.Roles[roleIndex].Position > misc.GuildMap[guild.ID].GuildConfig.OptInAbove.Position &&
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
	misc.GuildMap[guild.ID].ChannelStats[channel.ChannelID] = &channel

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

	if Today.Day() == t.Day() {
		return
	}

	// Update daily anime schedule
	UpdateAnimeSchedule()
	ResetSubscriptions()

	folders, err := ioutil.ReadDir("database/guilds")
	if err != nil {
		log.Panicln(err)
	}

	misc.MapMutex.Lock()
	for _, f := range folders {
		if !f.IsDir() {
			continue
		}
		guildID := f.Name()

		// Sends daily schedule if need be
		DailySchedule(s, guildID)

		guildPrefix := misc.GuildMap[guildID].GuildConfig.Prefix
		guildBotLog := misc.GuildMap[guildID].GuildConfig.BotLog.ID
		var guildDailyStatsID string
		if dailystats, ok := misc.GuildMap[guildID].Autoposts["dailystats"]; !ok {
			continue
		} else {
			guildDailyStatsID = dailystats.ID
		}
		if guildDailyStatsID == "" {
			continue
		}

		_, err := s.ChannelMessageSend(guildDailyStatsID, fmt.Sprintf("Stats for **%v %v, %v**", Today.Month(), Today.Day(), Today.Year()))
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error())
			if err != nil {
				continue
			}
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

		misc.MapMutex.Unlock()
		showStats(s, &message)
		misc.MapMutex.Lock()
	}
	misc.MapMutex.Unlock()

	Today = t
}

// Daily stat update timer
func DailyStatsTimer(s *discordgo.Session, e *discordgo.Ready) {
	for range time.NewTicker(1 * time.Minute).C {
		dailyStats(s)
	}
}

// Adds channel stats command to the commandHandler
func init() {
	add(&command{
		execute:  showStats,
		trigger:  "stats",
		aliases:  []string{"channelstats", "channels"},
		desc:     "Prints all channel stats.",
		elevated: true,
		category: "stats",
	})
}
