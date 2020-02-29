package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
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

	entities.HandleNewGuild(m.GuildID)

	var channelStatsVar entities.Channel
	t := time.Now()

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildChannelStats := db.GetGuildChannelStats(m.GuildID)

	// Sets channel params if it didn't exist before in database
	if _, ok := guildChannelStats[m.ChannelID]; !ok {

		// Fetches all guild info
		guild, err := s.State.Guild(m.GuildID)
		if err != nil {
			guild, err = s.Guild(m.GuildID)
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
		}
		// Fetches channel info
		channel, err := s.State.Channel(m.ChannelID)
		if err != nil {
			channel, err = s.Channel(m.ChannelID)
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
		}

		channelStatsVar = channelStatsVar.SetChannelID(channel.ID)
		channelStatsVar = channelStatsVar.SetName(channel.Name)
		channelStatsVar = channelStatsVar.SetRoleCount(channel.Name, common.GetRoleUserAmount(guild, guild.Roles, channel.Name))

		// Removes role stat for channels without associated roles. Else turns bool to true
		if channelStatsVar.GetRoleCount(channel.Name) == 0 {
			channelStatsVar = channelStatsVar.SetRoleCountMap(nil)
		} else {
			channelStatsVar = channelStatsVar.SetOptin(true)
		}

		channelStatsVar = channelStatsVar.SetMessagesMap(make(map[string]int))
		channelStatsVar = channelStatsVar.SetExists(true)

		guildChannelStats[m.ChannelID] = channelStatsVar
	}
	if guildChannelStats[m.ChannelID].GetChannelID() == "" {
		channelStats := guildChannelStats[m.ChannelID].SetChannelID(m.ChannelID)
		guildChannelStats[m.ChannelID] = channelStats
	}

	channelMessages := guildChannelStats[m.ChannelID].AddMessages(t.Format(common.ShortDateFormat), 1)
	guildChannelStats[m.ChannelID] = channelMessages
}

// Prints all channel stats
func showStats(s *discordgo.Session, m *discordgo.Message) {

	var (
		msgs               []string
		normalChannelTotal int
		optinChannelTotal  int
		flag               bool
		channels           []entities.Channel
		t                  time.Time
	)

	// Print either Today or yesterday based on whether it's the bot that called the func
	if m.Author.ID == s.State.User.ID {
		t = Today
	} else {
		t = time.Now()
	}

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildChannelStats := db.GetGuildChannelStats(m.GuildID)

	// Fixes channels without ID param
	for id, channel := range guildChannelStats {
		if channel.GetChannelID() == "" {
			channel = channel.SetChannelID(id)
			flag = true
		}
	}

	// Writes channel stats to disk if IDs were fixed
	if flag {
		db.SetGuildChannelStats(m.GuildID, guildChannelStats)
	}

	// Sorts channel by their message use
	for _, channel := range guildChannelStats {
		channels = append(channels, channel)
	}
	sort.Sort(byFrequencyChannel(channels))

	// Calculates normal channels and optin channels message totals
	for _, channel := range guildChannelStats {
		if !channel.GetOptin() {
			for date := range channel.GetMessagesMap() {
				normalChannelTotal += channel.GetMessages(date)
			}
		} else {
			for date := range channel.GetMessagesMap() {
				optinChannelTotal += channel.GetMessages(date)
			}
		}
	}

	// Fetches info on server roles from the server and puts it in deb
	deb, err := s.GuildRoles(m.GuildID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Confirms whether optins exist
	optInExist := common.OptInsExist(s, m.GuildID)
	//err = common.OptInsHandler(s, m.ChannelID, m.GuildID)
	//if err != nil {
	//	common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
	//	return
	//}

	// Fetches all guild info
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		guild, err = s.Guild(m.GuildID)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}

	// Updates opt-in-under and opt-in-above position for use later in isChannelUsable func
	if optInExist {
		for i := 0; i < len(deb); i++ {
			if deb[i].ID == guildSettings.GetOptInUnder().GetID() {
				optInUnder := guildSettings.GetOptInUnder()
				optInUnder = optInUnder.SetPosition(deb[i].Position)
				guildSettings = guildSettings.SetOptInUnder(optInUnder)
			} else if deb[i].ID == guildSettings.GetOptInAbove().GetID() {
				optInAbove := guildSettings.GetOptInAbove()
				optInAbove = optInAbove.SetPosition(deb[i].Position)
				guildSettings = guildSettings.SetOptInUnder(optInAbove)
			}
		}
	}

	// Adds the channels and their stats to message and formats it
	message := "```CSS\nName:                            ([Daily Messages] | [Total Messages]) \n\n"
	for _, channel := range channels {

		// Checks if channel exists and sets optin status
		channel, ok := isChannelUsable(channel, guild, guildSettings, optInExist)
		if !ok {
			continue
		}
		// Formats  and splits message
		if !channel.GetOptin() {
			message += lineSpaceFormatChannel(channel.GetChannelID(), false, t, guildChannelStats)
			message += "\n"
		}
		msgs, message = splitStatMessages(msgs, message)
	}

	message += fmt.Sprintf("\nNormal Total: %d\n\n------", normalChannelTotal)
	if optInExist {
		message += "\n\nOpt-in Name:                     ([Daily Messages] | [Total Messages] | [Role Members]) \n\n"
	}

	for _, channel := range channels {

		// Checks if channel exists and sets optin status
		channel, ok := isChannelUsable(channel, guild, guildSettings, optInExist)
		if !ok {
			continue
		}

		if channel.GetOptin() {
			// Formats  and splits message
			message += lineSpaceFormatChannel(channel.GetChannelID(), true, t, guildChannelStats)
			msgs, message = splitStatMessages(msgs, message)
		}
	}

	userChangeStat := db.GetGuildUserChangeStat(m.GuildID, t.Format(common.ShortDateFormat))
	verifiedStats := db.GetGuildVerifiedStats(m.GuildID)

	if optInExist {
		message += fmt.Sprintf("\nOpt-in Total: %d\n\n------\n", optinChannelTotal)
	}
	message += fmt.Sprintf("\nGrand Total Messages: %d\n\n", optinChannelTotal+normalChannelTotal)
	message += fmt.Sprintf("\nDaily Username Change: %d\n\n", userChangeStat)
	if len(verifiedStats) != 0 && config.Website != "" {
		verifiedStat := db.GetGuildVerifiedStat(m.GuildID, t.Format(common.ShortDateFormat))
		message += fmt.Sprintf("\nDaily Verified Change: %d\n\n", verifiedStat)
	}

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
		_, err = s.ChannelMessageSend(m.ChannelID, msgs[j])
		if err != nil {
			log.Println(err)
		}
	}
}

// Sort functions for channels by message use
type byFrequencyChannel []entities.Channel

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
	for date := range e[j].GetMessagesMap() {
		jTotalMessages += e[j].GetMessages(date)
	}
	for date := range e[i].GetMessagesMap() {
		iTotalMessages += e[i].GetMessages(date)
	}
	return jTotalMessages < iTotalMessages
}

// Formats the line space length for the above to keep level spacing
func lineSpaceFormatChannel(id string, optin bool, t time.Time, guildChannelStats map[string]entities.Channel) string {

	var totalMessages int

	for date := range guildChannelStats[id].GetMessagesMap() {
		totalMessages += guildChannelStats[id].GetMessages(date)
	}
	line := fmt.Sprintf("%s", guildChannelStats[id].GetName())
	spacesRequired := 33 - len(guildChannelStats[id].GetName())
	for i := 0; i < spacesRequired; i++ {
		line += " "
	}
	line += fmt.Sprintf("([%d])", guildChannelStats[id].GetMessages(t.Format(common.ShortDateFormat)))
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
		line += fmt.Sprintf("| [%d])\n", guildChannelStats[id].GetRoleCount(t.Format(common.ShortDateFormat)))
	}

	return line
}

// Adds 1 to Username Change on member join
func OnMemberJoin(_ *discordgo.Session, u *discordgo.GuildMemberAdd) {
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

	entities.HandleNewGuild(u.GuildID)

	t := time.Now()

	db.AddGuildUserChangeStat(u.GuildID, t.Format(common.ShortDateFormat), 1)
}

// Removes 1 from Username Change on member removal
func OnMemberRemoval(_ *discordgo.Session, u *discordgo.GuildMemberRemove) {
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

	entities.HandleNewGuild(u.GuildID)

	t := time.Now()

	db.AddGuildUserChangeStat(u.GuildID, t.Format(common.ShortDateFormat), -1)
}

// Checks if specific channel stat should be printed
func isChannelUsable(channel entities.Channel, guild *discordgo.Guild, guildSettings entities.GuildSettings, optInExist bool) (*entities.Channel, bool) {

	// Checks if channel exists and if it's optin
	for guildIndex := range guild.Channels {
		if optInExist {
			for roleIndex := range guild.Roles {
				if guild.Roles[roleIndex].Position < guildSettings.GetOptInUnder().GetPosition() &&
					guild.Roles[roleIndex].Position > guildSettings.GetOptInAbove().GetPosition() &&
					guild.Channels[guildIndex].Name == guild.Roles[roleIndex].Name {
					channel = channel.SetOptin(true)
					break
				} else {
					channel = channel.SetOptin(false)
				}
			}
		}
		if guild.Channels[guildIndex].Name == channel.GetName() &&
			guild.Channels[guildIndex].ID == channel.GetChannelID() {
			channel = channel.SetExists(true)
			break
		} else {
			channel = channel.SetExists(false)
		}
	}

	db.SetGuildChannelStat(guild.ID, channel)

	if channel.GetExists() {
		return &channel, true
	}
	return &channel, false
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

	entities.Mutex.RLock()
	if Today.Day() == t.Day() {
		entities.Mutex.RUnlock()
		return
	}
	entities.Mutex.RUnlock()

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

	for _, f := range folders {
		if !f.IsDir() {
			continue
		}
		guildID := f.Name()

		// Sends daily schedule if need be
		DailySchedule(s, guildID)

		guildSettings := db.GetGuildSettings(guildID)

		dailyStats := db.GetGuildAutopost(guildID, "dailystats")
		if dailyStats == (entities.Cha{}) || dailyStats.GetID() == "" {
			continue
		}
		guildDailyStatsID := dailyStats.GetID()

		_, err := s.ChannelMessageSend(guildDailyStatsID, fmt.Sprintf("Stats for **%v %v, %v**", Today.Month(), Today.Day(), Today.Year()))
		if err != nil {
			continue
		}

		author.ID = s.State.User.ID
		message.GuildID = guildID
		message.Author = &author
		message.Content = guildSettings.GetPrefix() + "stats"
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

		showStats(s, &message)
	}

	entities.Mutex.Lock()
	Today = t
	entities.Mutex.Unlock()
}

// Daily stat update timer
func DailyStatsTimer(s *discordgo.Session, _ *discordgo.Ready) {
	for range time.NewTicker(1 * time.Minute).C {
		dailyStats(s)
	}
}

// Adds channel stats command to the commandHandler
func init() {
	Add(&Command{
		Execute:    showStats,
		Trigger:    "stats",
		Aliases:    []string{"channelstats", "channels", "stat", "chanstat", "chanstats", "statss"},
		Desc:       "Prints all channel stats",
		Permission: functionality.Mod,
		Module:     "stats",
	})
}
