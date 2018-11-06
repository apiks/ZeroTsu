package commands

import (
	"fmt"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

var (
	dailyFlag bool
)

// Adds to message count on every message for that channel
func OnMessageChannel(s *discordgo.Session, m *discordgo.MessageCreate) {

	var channelStatsVar misc.Channel
	t := time.Now()

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string))
			if err != nil {
				return
			}
		}
	}()

	// Checks if it's within the config server and whether it's the bot
	ch, err := s.State.Channel(m.ChannelID)
	if err != nil {
		ch, err = s.Channel(m.ChannelID)
		if err != nil {
			return
		}
	}
	if ch.GuildID != config.ServerID {
		return
	}

	// Pull channel info
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	// Sets channel params if it didn't exist before in database
	misc.MapMutex.Lock()
	if misc.ChannelStats[m.ChannelID] == nil {
		// Fetches all guild users
		guild, err := s.Guild(config.ServerID)
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		// Fetches all server roles
		roles, err := s.GuildRoles(config.ServerID)
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}

		channelStatsVar.ChannelID = channel.ID
		channelStatsVar.Name = channel.Name
		misc.MapMutex.Lock()
		channelStatsVar.RoleCount = make(map[string]int)
		misc.MapMutex.Unlock()
		channelStatsVar.RoleCount[channel.Name] = misc.GetRoleUserAmount(guild, roles, channel.Name)

		// Removes role stat for channels without associated roles. Else turns bool to true
		if channelStatsVar.RoleCount[channel.Name] == 0 {
			channelStatsVar.RoleCount = nil
		} else {
			channelStatsVar.Optin = true
		}

		channelStatsVar.Messages = make(map[string]int)
		channelStatsVar.Exists = true
		misc.ChannelStats[m.ChannelID] = &channelStatsVar
	}
	if misc.ChannelStats[m.ChannelID].ChannelID == "" {
		misc.ChannelStats[m.ChannelID].ChannelID = m.ChannelID
	}

	misc.ChannelStats[m.ChannelID].Messages[t.Format(misc.DateFormat)]++
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

	// Fixes channels without ID param
	misc.MapMutex.Lock()
	for id := range misc.ChannelStats {
		if misc.ChannelStats[id].ChannelID == "" {
			misc.ChannelStats[id].ChannelID = id
			flag = true
		}
	}

	// Writes channel stats to disk if IDs were fixed
	if flag {
		_, err := misc.ChannelStatsWrite(misc.ChannelStats)
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
	}


	// Sorts channel by their message use
	channels := make([]*misc.Channel, len(misc.ChannelStats))
	for i := 0; i < len(misc.ChannelStats); i++ {
		for _, channel := range misc.ChannelStats {
			channels[i] = channel
			i++
		}
	}
	misc.MapMutex.Unlock()
	sort.Sort(byFrequencyChannel(channels))

	// Calculates normal channels and optin channels message totals
	misc.MapMutex.Lock()
	for chas := range misc.ChannelStats {
		if !misc.ChannelStats[chas].Optin {
			for date := range misc.ChannelStats[chas].Messages {
				normalChannelTotal += misc.ChannelStats[chas].Messages[date]
			}
		} else {
			for date := range misc.ChannelStats[chas].Messages {
				optinChannelTotal += misc.ChannelStats[chas].Messages[date]
			}
		}
	}
	misc.MapMutex.Unlock()

	// Fetches info on server roles from the server and puts it in deb
	deb, err := s.GuildRoles(config.ServerID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Updates opt-in-under and opt-in-above position for use later in isChannlUsable func
	for i := 0; i < len(deb); i++ {
		if deb[i].Name == config.OptInUnder {
			misc.OptinUnderPosition = deb[i].Position
		} else if deb[i].Name == config.OptInAbove {
			misc.OptinAbovePosition = deb[i].Position
		}
	}

	// Pull guild info
	guild, err := s.State.Guild(config.ServerID)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
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
			misc.MapMutex.Lock()
			message += lineSpaceFormatChannel(channel.ChannelID, false, *s)
			misc.MapMutex.Unlock()
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
			misc.MapMutex.Lock()
			message += lineSpaceFormatChannel(channel.ChannelID, true, *s)
			misc.MapMutex.Unlock()
			msgs, message = splitStatMessages(msgs, message)
		}
	}

	message += fmt.Sprintf("\nOpt-in Total: %d\n\n------\n", optinChannelTotal)
	message += fmt.Sprintf("\nGrand Total Messages: %d\n\n", optinChannelTotal+normalChannelTotal)
	misc.MapMutex.Lock()
	message += fmt.Sprintf("\nDaily User Change: %d\n\n", misc.UserStats[t.Format(misc.DateFormat)])
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
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
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
func lineSpaceFormatChannel(id string, optin bool, s discordgo.Session) string {

	var totalMessages int
	t := time.Now()

	for date := range misc.ChannelStats[id].Messages {
		totalMessages += misc.ChannelStats[id].Messages[date]
	}
	line := fmt.Sprintf("%v", misc.ChannelStats[id].Name)
	spacesRequired := 33 - len(misc.ChannelStats[id].Name)
	for i := 0; i < spacesRequired; i++ {
		line += " "
	}
	line += fmt.Sprintf("([%d])", misc.ChannelStats[id].Messages[t.Format(misc.DateFormat)])
	spacesRequired = 51 - len(line)
	for i := 0; i < spacesRequired; i++ {
		line += " "
	}
	line += fmt.Sprintf("| ([%d])", totalMessages)
	spacesRequired = 59 - len(line)
	for i := 0; i < spacesRequired; i++ {
		line += " "
	}
	if optin {
		line += fmt.Sprintf("| [%d])\n", misc.ChannelStats[id].RoleCount[t.Format(misc.DateFormat)])
	}

	return line
}

// Adds 1 to User Change on member join
func OnMemberJoin(s *discordgo.Session, u *discordgo.GuildMemberAdd) {
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string))
			if err != nil {
				return
			}
		}
	}()

	t := time.Now()
	misc.MapMutex.Lock()
	misc.UserStats[t.Format(misc.DateFormat)]++
	misc.MapMutex.Unlock()
}

// Removes 1 from User Change on member removal and also resets variables
func OnMemberRemoval(s *discordgo.Session, u *discordgo.GuildMemberRemove) {
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string))
			if err != nil {
				return
			}
		}
	}()

	t := time.Now()
	misc.MapMutex.Lock()
	misc.UserStats[t.Format(misc.DateFormat)]--
	if misc.MemberInfoMap[u.User.ID] != nil {
		misc.MemberInfoMap[u.User.ID].Discrim = ""
		misc.MemberInfoMap[u.User.ID].OutsideServer = true
		misc.MemberInfoWrite(misc.MemberInfoMap)
	}
	misc.MapMutex.Unlock()
}

// Checks if specific channel stat should be printed
func isChannelUsable(channel misc.Channel, guild *discordgo.Guild) (misc.Channel, bool) {

	// Checks if channel exists and if it's optin
	for guildIndex := range guild.Channels {
		for roleIndex := range guild.Roles {
			if guild.Roles[roleIndex].Position < misc.OptinUnderPosition &&
				guild.Roles[roleIndex].Position > misc.OptinAbovePosition &&
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
	misc.MapMutex.Lock()
	misc.ChannelStats[channel.ChannelID] = &channel
	misc.MapMutex.Unlock()

	if channel.Exists {
		return channel, true
	}
	return channel, false
}

// Splits the stat messages into blocks
func splitStatMessages (msgs []string, message string) ([]string, string) {
	const maxMsgLength = 1900
	if len(message) > maxMsgLength {
		msgs = append(msgs, message)
		message = ""
	}
	return msgs, message
}

// Posts daily stats
func dailyStats(s *discordgo.Session) {

	var (
		message discordgo.Message
		author  discordgo.User
	)

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string))
			if err != nil {
				return
			}
		}
	}()

	t := time.Now()
	hour := t.Hour()
	minute := t.Minute()

	if hour == 23 && minute == 59 && !dailyFlag {
		_, err := s.ChannelMessageSend(config.BotLogID, fmt.Sprintf("Update for **%v %v, %v**", t.Month(), t.Day(), t.Year()))
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}

		misc.MapMutex.Lock()
		author.ID = s.State.User.ID
		message.Author = &author
		message.Content = config.BotPrefix + "stats"
		misc.MapMutex.Unlock()
		showStats(s, &message)
		dailyFlag = true
	}

	if hour == 0 && minute == 0 && dailyFlag {
		dailyFlag = false
	}
}

// Daily stat update timer
func DailyStatsTimer(s *discordgo.Session, e *discordgo.Ready) {
	for range time.NewTicker(15 * time.Second).C {
		dailyStats(s)
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
		category: "normal",
	})
}