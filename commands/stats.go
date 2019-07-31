package commands

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

// Adds to message count on every message for that channel
func OnMessageChannel(s *discordgo.Session, m *discordgo.MessageCreate) {

	var channelStatsVar misc.Channel
	t := time.Now()

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

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	// Pull channel info
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	// Sets channel params if it didn't exist before in database
	misc.MapMutex.Lock()
	if misc.GuildMap[m.GuildID] == nil {
		misc.MapMutex.Unlock()
		return
	}

	if _, ok := misc.GuildMap[m.GuildID].ChannelStats[m.ChannelID]; !ok {
		// Fetches all guild users
		guild, err := s.Guild(m.GuildID)
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		// Fetches all server roles
		roles, err := s.GuildRoles(m.GuildID)
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}

		channelStatsVar.ChannelID = channel.ID
		channelStatsVar.Name = channel.Name
		channelStatsVar.RoleCount = make(map[string]int)
		channelStatsVar.RoleCount[channel.Name] = misc.GetRoleUserAmount(guild, roles, channel.Name)

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
		flag			   bool
	)

	t := time.Now()

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
	channels := make([]*misc.Channel, len(misc.GuildMap[m.GuildID].ChannelStats))
	for i := 0; i < len(misc.GuildMap[m.GuildID].ChannelStats); i++ {
		for _, channel := range misc.GuildMap[m.GuildID].ChannelStats {
			channels[i] = channel
			i++
		}
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
	err = misc.OptInsHandler(s, m.GuildID)
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

	// Pull guild info
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
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
			message += lineSpaceFormatChannel(channel.ChannelID, false, m.GuildID)
			message += "\n"
		}
		msgs, message = splitStatMessages(msgs, message)
	}

	message += fmt.Sprintf("\nNormal Total: %d\n\n------", normalChannelTotal)
	message += "\n\nOpt-in Name:                     ([Daily Messages] | [Total Messages] | [Role Members]) \n\n"

	for _, channel := range channels {
		if channel.Optin {

			// Checks if channel exists and sets optin status
			channel, ok := isChannelUsable(*channel, guild)
			if !ok {
				continue
			}
			// Formats  and splits message
			message += lineSpaceFormatChannel(channel.ChannelID, true, m.GuildID)
			msgs, message = splitStatMessages(msgs, message)
		}
	}

	message += fmt.Sprintf("\nOpt-in Total: %d\n\n------\n", optinChannelTotal)
	message += fmt.Sprintf("\nGrand Total Messages: %d\n\n", optinChannelTotal+normalChannelTotal)
	message += fmt.Sprintf("\nDaily User Change: %d\n\n", misc.GuildMap[m.GuildID].UserChangeStats[t.Format(misc.DateFormat)])
	if len(misc.GuildMap[m.GuildID].VerifiedStats) != 0 {
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
func lineSpaceFormatChannel(id string, optin bool, guildID string) string {

	var totalMessages int
	t := time.Now()

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
	misc.GuildMap[u.GuildID].UserChangeStats[t.Format(misc.DateFormat)]++
	misc.MapMutex.Unlock()
}

// Removes 1 from User Change on member removal and also resets variables
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
	misc.GuildMap[u.GuildID].UserChangeStats[t.Format(misc.DateFormat)]--
	misc.WriteMemberInfo(misc.GuildMap[u.GuildID].MemberInfoMap, u.GuildID)
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
func splitStatMessages (msgs []string, message string) ([]string, string) {
	const maxMsgLength = 1700
	if len(message) > maxMsgLength {
		msgs = append(msgs, message)
		message = ""
	}
	return msgs, message
}

// Posts daily stats and update schedule command
func dailyStats(s *discordgo.Session, e *discordgo.Ready) {

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
	hour := t.Hour()
	minute := t.Minute()

	for _, guild := range e.Guilds {

		misc.MapMutex.Lock()
		guildPrefix := misc.GuildMap[guild.ID].GuildConfig.Prefix
		guildBotLog := misc.GuildMap[guild.ID].GuildConfig.BotLog.ID
		guildDailyStats := misc.GuildMap[guild.ID].GuildConfig.DailyStats
		misc.MapMutex.Unlock()

		if hour == 23 && minute == 59 && !guildDailyStats {
			_, err := s.ChannelMessageSend(guildBotLog, fmt.Sprintf("Update for **%v %v, %v**", t.Month(), t.Day(), t.Year()))
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}

			author.ID = s.State.User.ID
			message.GuildID = guild.ID
			message.Author = &author
			message.Content = guildPrefix + "stats"
			message.ChannelID = guildBotLog
			guildDailyStats = true
			showStats(s, &message)
			misc.MapMutex.Lock()
			misc.GuildMap[guild.ID].GuildConfig.DailyStats = true
			misc.GuildSettingsWrite(misc.GuildMap[guild.ID].GuildConfig, guild.ID)
			misc.MapMutex.Unlock()
		}
		if hour == 0 && minute == 0 && guildDailyStats {
			misc.MapMutex.Lock()
			misc.GuildMap[guild.ID].GuildConfig.DailyStats = false
			misc.GuildSettingsWrite(misc.GuildMap[guild.ID].GuildConfig, guild.ID)
			misc.MapMutex.Unlock()
		}
	}

	// Update daily anime schedule command
	if hour == 0 && minute == 0 {
		UpdateAnimeSchedule()
	}
}

// Daily stat update timer
func DailyStatsTimer(s *discordgo.Session, e *discordgo.Ready) {
	for range time.NewTicker(40 * time.Second).C {
		dailyStats(s, e)
	}
}

// Adds channel stats command to the commandHandler
func init() {
	add(&command{
		execute:   showStats,
		trigger:  "stats",
		aliases:  []string{"channelstats", "channels"},
		desc:     "Prints all channel stats.",
		elevated: true,
		category: "stats",
	})
}