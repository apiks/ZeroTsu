package misc

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mmcdole/gofeed"

	"github.com/r-anime/ZeroTsu/config"
)

var (
	darlingTrigger int
	redditFeedBlock bool
)

// Periodic events such as Unbanning and RSS timer every 30 sec
func StatusReady(s *discordgo.Session, e *discordgo.Ready) {

	var banFlag bool

	for _, guild := range e.Guilds {

		// Initialize guild if missing
		MapMutex.Lock()
		InitDB(s, guild.ID)
		writeAll(guild.ID)

		// Clean up SpoilerRoles.json in each guild
		err := cleanSpoilerRoles(s, guild.ID)
		if err != nil {
			_, _ = s.ChannelMessageSend(GuildMap[guild.ID].GuildConfig.BotLog.ID, err.Error())
		}

		DynamicNicknameChange(s, guild.ID)
		MapMutex.Unlock()
	}

	// Updates playing status
	MapMutex.Lock()
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

	for range time.NewTicker(55 * time.Second).C {

		// Checks whether it has to post Reddit feeds and handle remindMes and handle bans
		for _, guild := range e.Guilds {
			RSSParser(s, guild.ID)
			MapMutex.Lock()
			remindMeHandler(s, guild.ID)

			// Goes through bannedUsers.json if it's not empty and unbans if needed
			if len(GuildMap[guild.ID].BannedUsers) == 0 {
				MapMutex.Unlock()
				continue
			}

			t := time.Now()
			var writeFlag bool
			for i := len(GuildMap[guild.ID].BannedUsers) - 1; i >= 0; i-- {

				difference := t.Sub(GuildMap[guild.ID].BannedUsers[i].UnbanDate)
				if difference <= 0 {
					continue
				}

				banFlag = false

				// Checks if user is in MemberInfo and saves him if true
				memberInfoUser, ok := GuildMap[guild.ID].MemberInfoMap[GuildMap[guild.ID].BannedUsers[i].ID]
				if !ok {
					continue
				}
				// Fetches all server bans so it can check if the memberInfoUser is banned there (whether he's been manually unbanned for example)
				bans, err := s.GuildBans(guild.ID)
				if err != nil {
					_, _ = s.ChannelMessageSend(GuildMap[guild.ID].GuildConfig.BotLog.ID, err.Error()+"\n"+ErrorLocation(err))
					continue
				}

				// Set flag for whether user is a banned user, and then check for that flag so you can continue from the upper loop if error
				for _, ban := range bans {
					if ban.User.ID == memberInfoUser.ID {
						banFlag = true
						break
					}
				}
				// Unbans memberInfoUser if possible
				if banFlag {
					err = s.GuildBanDelete(guild.ID, memberInfoUser.ID)
					if err != nil {
						_, _ = s.ChannelMessageSend(GuildMap[guild.ID].GuildConfig.BotLog.ID, err.Error()+"\n"+ErrorLocation(err))
						continue
					}
				}

				// Removes unban date entirely
				GuildMap[guild.ID].MemberInfoMap[memberInfoUser.ID].UnbanDate = ""

				// Removes the memberInfoUser ban from bannedUsers.json
				GuildMap[guild.ID].BannedUsers = append(GuildMap[guild.ID].BannedUsers[:i], GuildMap[guild.ID].BannedUsers[i+1:]...)

				// Writes to memberInfo.json and bannedUsers.json
				writeFlag = true

				// Sends an embed message to bot-log
				if !banFlag {
					_ = UnbanEmbed(s, memberInfoUser, "", GuildMap[guild.ID].GuildConfig.BotLog.ID)
				}
			}
			if writeFlag {
				WriteMemberInfo(GuildMap[guild.ID].MemberInfoMap, guild.ID)
				_ = BannedUsersWrite(GuildMap[guild.ID].BannedUsers, guild.ID)
			}
			MapMutex.Unlock()
		}
	}
}

func UnbanEmbed(s *discordgo.Session, user *UserInfo, mod string, botLog string) error {

	var (
		embedMess discordgo.MessageEmbed
		embed     []*discordgo.MessageEmbedField
	)

	// Sets timestamp of unban
	t := time.Now()
	now := t.Format(time.RFC3339)
	embedMess.Timestamp = now

	// Set embed color
	embedMess.Color = 16758465

	if mod == "" {
		embedMess.Title = fmt.Sprintf("%v#%v has been unbanned.", user.Username, user.Discrim)
	} else {
		embedMess.Title = fmt.Sprintf("%v#%v has been unbanned by %v.", user.Username, user.Discrim, mod)
	}

	// Adds everything together
	embedMess.Fields = embed

	// Sends embed in bot-log
	_, err := s.ChannelMessageSendEmbed(botLog, &embedMess)
	if err != nil {
		return err
	}
	return err
}

// Periodic 20min events
func TwentyMinTimer(s *discordgo.Session, e *discordgo.Ready) {
	for range time.NewTicker(20 * time.Minute).C {

		// Updates playing status
		MapMutex.Lock()
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

		MapMutex.Lock()
		for _, guild := range e.Guilds {

			if _, ok := GuildMap[guild.ID]; !ok {
				InitDB(s, guild.ID)
				LoadGuilds()
			}

			guildBotLog := GuildMap[guild.ID].GuildConfig.BotLog.ID

			// Writes emoji stats to disk
			_, err := EmojiStatsWrite(GuildMap[guild.ID].EmojiStats, guild.ID)
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				continue
			}

			// Writes user gain stats to disk
			_, err = UserChangeStatsWrite(GuildMap[guild.ID].UserChangeStats, guild.ID)
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				continue
			}

			// Writes verified stats to disk
			if config.Website != "" {
				err = VerifiedStatsWrite(GuildMap[guild.ID].VerifiedStats, guild.ID)
				if err != nil {
					_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
					continue
				}
			}

			// Writes memberInfo to disk
			WriteMemberInfo(GuildMap[guild.ID].MemberInfoMap, guild.ID)

			// Writes channel stats to disk
			_, err = ChannelStatsWrite(GuildMap[guild.ID].ChannelStats, guild.ID)
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				if err != nil {
					continue
				}
				continue
			}

			// Clears up spoilerRoles.json
			err = cleanSpoilerRoles(s, guild.ID)
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				if err != nil {
					continue
				}
				continue
			}

			// Updates BOT nickname
			DynamicNicknameChange(s, guild.ID)

			// Fetches all server roles
			roles, err := s.GuildRoles(guild.ID)
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				continue
			}
			// Updates optin role stat
			t := time.Now()
			for chas := range GuildMap[guild.ID].ChannelStats {
				if GuildMap[guild.ID].ChannelStats[chas].RoleCount == nil {
					GuildMap[guild.ID].ChannelStats[chas].RoleCount = make(map[string]int)
				}
				if GuildMap[guild.ID].ChannelStats[chas].Optin {
					GuildMap[guild.ID].ChannelStats[chas].RoleCount[t.Format(DateFormat)] = GetRoleUserAmount(guild, roles, GuildMap[guild.ID].ChannelStats[chas].Name)
				}
			}
		}
		MapMutex.Unlock()

		// Sends server count to bot list sites if it's the public ZeroTsu
		sendServers(s)
	}
}

// Pulls the reddit feed threads and print them
func RSSParser(s *discordgo.Session, guildID string) {

	// Stops handling of new feed threads if there are some currently being sent
	if redditFeedBlock {
		return
	}

	var pinnedItems = make(map[*gofeed.Item]bool)

	// Checks if there are any rss settings for this guild
	MapMutex.Lock()
	if len(GuildMap[guildID].RssThreads) == 0 {
		MapMutex.Unlock()
		return
	}
	// Save current threads as a copy so mapMutex isn't taken all the time when checking the feeds
	rssThreads := GuildMap[guildID].RssThreads
	rssThreadChecks := GuildMap[guildID].RssThreadChecks
	botLogID := GuildMap[guildID].GuildConfig.BotLog.ID

	t := time.Now()
	hours := time.Hour * 1440

	// Removes a thread if more than 60 days have passed from the rss thread checks. This is to keep DB manageable
	for p := 0; p < len(rssThreadChecks); p++ {
		// Calculates if it's time to remove
		dateRemoval := rssThreadChecks[p].Date.Add(hours)
		difference := t.Sub(dateRemoval)

		// Removes the fact that the thread had been posted already if it's time
		if difference <= 0 {
			continue
		}
		err := RssThreadsTimerRemove(rssThreadChecks[p].Thread, rssThreadChecks[p].Date, guildID)
		if err != nil {
			_, _ = s.ChannelMessageSend(botLogID, err.Error()+"\n"+ErrorLocation(err))
			continue
		}
	}
	// Updates rssThreadChecks var after the removal
	rssThreadChecks = GuildMap[guildID].RssThreadChecks
	MapMutex.Unlock()

	// Sets up the feed parser
	fp := gofeed.NewParser()
	fp.Client = &http.Client{Transport: &UserAgentTransport{http.DefaultTransport}, Timeout: time.Second * 10}

	// Save all feeds early to save performance
	var subMap = make(map[string]*gofeed.Feed)
	for _, thread := range rssThreads {
		if _, ok := subMap[thread.Subreddit]; !ok {
			// Parse feed
			feed, err := fp.ParseURL(fmt.Sprintf("http://www.reddit.com/r/%v/%v/.rss", thread.Subreddit, thread.PostType))
			if err != nil {
				return
			}
			subMap[fmt.Sprintf("%v:%v", thread.Subreddit, thread.PostType)] = feed
		}
	}

	threadsToPost := make(map[*gofeed.Item]*RssThread)

	for _, thread := range rssThreads {

		// Get the necessary feed from the subMap
		feed := subMap[fmt.Sprintf("%v:%v", thread.Subreddit, thread.PostType)]

		t = time.Now()

		// Iterates through each feed item to see if it finds something from storage that should be posted
		for _, item := range feed.Items {

			// Check if this item exists in rssThreadChecks and skips the item if it does
			var skip = false
			for _, check := range rssThreadChecks {
				if check.GUID == item.GUID &&
					check.Thread.ChannelID == thread.ChannelID {
					skip = true
					break
				}
			}
			if skip {
				continue
			}

			// Check if author is same and skip if not true
			if thread.Author != "" && item.Author != nil {
				if strings.ToLower(item.Author.Name) != fmt.Sprintf("/u/%v", thread.Author) {
					continue
				}
			}

			// Check if the feed item title starts with the set thread title
			if thread.Title != "" {
				if !strings.HasPrefix(strings.ToLower(item.Title), thread.Title) {
					continue
				}
			}

			// Writes that thread has been posted
			MapMutex.Lock()
			err := RssThreadsTimerWrite(thread, t, item.GUID, guildID)
			if err != nil {
				MapMutex.Unlock()
				_, _ = s.ChannelMessageSend(botLogID, err.Error()+"\n"+ErrorLocation(err))
				continue
			}
			// Updates rssThreadChecks var after the write
			rssThreadChecks = GuildMap[guildID].RssThreadChecks
			MapMutex.Unlock()

			// Adds the thread to the threads to send map
			if _, ok := threadsToPost[item]; !ok {
				threadsToPost[item] = &thread
			}
		}
	}

	// Sends the threads concurrently in slow mode
	go func() {
		redditFeedBlock = true
		for item, thread := range threadsToPost {
			// Sends the feed item
			time.Sleep(time.Second * 4)
			message, err := s.ChannelMessageSend(thread.ChannelID, item.Link)
			if err != nil {
				continue
			}

			// Pins/unpins the feed items if necessary
			if !thread.Pin {
				delete(threadsToPost, item)
				continue
			}
			if _, ok := pinnedItems[item]; ok {
				delete(threadsToPost, item)
				continue
			}

			pins, err := s.ChannelMessagesPinned(message.ChannelID)
			if err != nil {
				delete(threadsToPost, item)
				continue
			}
			// Unpins if necessary
			for _, pin := range pins {

				// Checks for whether the pin is one that should be unpinned
				if pin.Author.ID != s.State.User.ID {
					continue
				}
				if !strings.HasPrefix(strings.ToLower(pin.Content), fmt.Sprintf("https://www.reddit.com/r/%v/comments/", thread.Subreddit)) ||
					!strings.HasPrefix(strings.ToLower(pin.Content), fmt.Sprintf("http://www.reddit.com/r/%v/comments/", thread.Subreddit)) {
					continue
				}

				err = s.ChannelMessageUnpin(pin.ChannelID, pin.ID)
				if err != nil {
					continue
				}
			}
			// Pins
			_ = s.ChannelMessagePin(message.ChannelID, message.ID)
			pinnedItems[item] = true
			delete(threadsToPost, item)
		}
		redditFeedBlock = false
	}()
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

	var voiceChannels []VoiceCha
	var noRemovalRoles []Role
	var dontRemove bool

	MapMutex.Lock()
	if _, ok := GuildMap[v.GuildID]; !ok {
		InitDB(s, v.GuildID)
		LoadGuilds()
	}

	if len(GuildMap[v.GuildID].GuildConfig.VoiceChas) == 0 {
		MapMutex.Unlock()
		return
	}

	voiceChannels = GuildMap[v.GuildID].GuildConfig.VoiceChas
	MapMutex.Unlock()

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

	var (
		guildPrefix = "."
		guildBotLog string
	)

	if m.GuildID != "" {
		MapMutex.Lock()
		if _, ok := GuildMap[m.GuildID]; !ok {
			InitDB(s, m.GuildID)
			LoadGuilds()
		}
		guildPrefix = GuildMap[m.GuildID].GuildConfig.Prefix
		guildBotLog = GuildMap[m.GuildID].GuildConfig.BotLog.ID
		MapMutex.Unlock()
	}

	if strings.ToLower(m.Content) == fmt.Sprintf("<@%v> good bot", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v> good bot", s.State.User.ID) {
		_, err := s.ChannelMessageSend(m.ChannelID, "Thank you ❤")
		if err != nil {
			_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
			return
		}
		return
	}

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && m.Author.ID == "128312718779219968" {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Professor!\n\nPrefix: `%v`", guildPrefix))
		if err != nil {
			_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
			return
		}
		return
	}

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && m.Author.ID == "66207186417627136" {

		randomNum := rand.Intn(5)
		if randomNum == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Bug hunter!\n\nPrefix: `%v`", guildPrefix))
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				return
			}
			return
		}
		if randomNum == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Player!\n\nPrefix: `%v`", guildPrefix))
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				return
			}
			return
		}
		if randomNum == 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Big brain!\n\nPrefix: `%v`", guildPrefix))
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				return
			}
			return
		}
		if randomNum == 3 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Poster expert!\n\nPrefix: `%v`", guildPrefix))
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				return
			}
			return
		}
		if randomNum == 4 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Idiot!\n\nPrefix: `%v`", guildPrefix))
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				return
			}
			return
		}
		return
	}

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && m.Author.ID == "365245718866427904" {

		randomNum := rand.Intn(5)
		if randomNum == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Begone ethot.\n\nPrefix: `%v`", guildPrefix))
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				return
			}
			return
		}
		if randomNum == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Humph!\n\nPrefix: `%v`", guildPrefix))
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				return
			}
			return
		}
		if randomNum == 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Wannabe ethot!\n\nPrefix: `%v`", guildPrefix))
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				return
			}
			return
		}
		if randomNum == 3 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Not even worth my time.\n\nPrefix: `%v`", guildPrefix))
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				return
			}
			return
		}
		if randomNum == 4 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Okay, maybe you're not that bad.\n\nPrefix: `%v`", guildPrefix))
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				return
			}
			return
		}
		return
	}

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && m.Author.ID == "315201054377771009" {

		randomNum := rand.Intn(5)
		if randomNum == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("https://cdn.discordapp.com/attachments/618463738504151086/619090216329674800/uiz31mhq12k11.gif\n\nPrefix: `%v`", guildPrefix))
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				return
			}
			return
		}
		if randomNum == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Onii-chan no ecchi!\n\nPrefix: `%v`", guildPrefix))
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				return
			}
			return
		}
		if randomNum == 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Kusuguttai Neiru-kun.\n\nPrefix: `%v`", guildPrefix))
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				return
			}
			return
		}
		if randomNum == 3 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Liking lolis isn't a crime, but I'll still visit you in prison.\n\nPrefix: `%v`", guildPrefix))
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				return
			}
			return
		}
		if randomNum == 4 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Iris told me you wanted her to meow at you while she was still young.\n\nPrefix: `%v`", guildPrefix))
			if err != nil {
				_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				return
			}
			return
		}
		return
	}

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && darlingTrigger > 10 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Daaarling~\n\nPrefix: `%v`", guildPrefix))
		if err != nil {
			_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
			return
		}
		darlingTrigger = 0
		return
	}

	if m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID) {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Baka!\n\nPrefix: `%v`", guildPrefix))
		if err != nil {
			_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
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
	if _, ok := GuildMap[e.GuildID]; !ok {
		InitDB(s, e.GuildID)
		LoadGuilds()
	}

	guildBotLog := GuildMap[e.GuildID].GuildConfig.BotLog.ID

	for _, user := range GuildMap[e.GuildID].BannedUsers {
		if user.ID == e.User.ID {
			MapMutex.Unlock()
			return
		}
	}
	MapMutex.Unlock()

	_, _ = s.ChannelMessageSend(guildBotLog, fmt.Sprintf("%v#%v was manually permabanned. ID: %v", e.User.Username, e.User.Discriminator, e.User.ID))
}

// Sends remindMe message if it is time, either as a DM or ping
func remindMeHandler(s *discordgo.Session, guildID string) {
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in remindMeHandler")
		}
	}()

	t := time.Now()
	for userID, remindMeSlice := range SharedInfo.RemindMes {
		for i := len(remindMeSlice.RemindMeSlice) - 1; i >= 0; i-- {

			// Checks if it's time to send message/ping the user
			difference := t.Sub(remindMeSlice.RemindMeSlice[i].Date)
			if difference <= 0 {
				continue
			}

			// Sends message to user DMs if possible
			dm, err := s.UserChannelCreate(userID)
			_, err = s.ChannelMessageSend(dm.ID, "RemindMe: "+remindMeSlice.RemindMeSlice[i].Message)
			// Else sends the message in the channel the command was made in with a ping
			if err != nil && guildID != "" {
				// Checks if the user is in the server before pinging him
				_, err := s.GuildMember(guildID, userID)
				if err == nil {
					pingMessage := fmt.Sprintf("<@%v> Remindme: %v", userID, remindMeSlice.RemindMeSlice[i].Message)
					_, err := s.ChannelMessageSend(remindMeSlice.RemindMeSlice[i].CommandChannel, pingMessage)
					if err != nil {
						return
					}
				}
			}

			// Removes the RemindMe from the RemindMe slice and writes to disk
			remindMeSlice.RemindMeSlice = append(remindMeSlice.RemindMeSlice[:i], remindMeSlice.RemindMeSlice[i+1:]...)
			SharedInfo.RemindMes[userID].RemindMeSlice = remindMeSlice.RemindMeSlice
			err = RemindMeWrite(SharedInfo.RemindMes)
			if err != nil && guildID != "" {
				_, err = s.ChannelMessageSend(GuildMap[guildID].GuildConfig.BotLog.ID, err.Error()+"\n"+ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
		}
	}
}

// Sends a message to a channel to log whenever a user joins. Intended use was to catch spambots for r/anime
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

	if u.GuildID != "267799767843602452" {
		return
	}

	MapMutex.Lock()
	if _, ok := GuildMap[u.GuildID]; !ok {
		InitDB(s, u.GuildID)
		LoadGuilds()
	}
	MapMutex.Unlock()

	creationDate, err := CreationTime(u.User.ID)
	if err != nil {

		MapMutex.Lock()
		guildBotLog := GuildMap[u.GuildID].GuildConfig.BotLog.ID
		MapMutex.Unlock()

		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
		if err != nil {
			return
		}
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

		temp    BannedUsers
		tempMem UserInfo

		dmMessage string
	)

	MapMutex.Lock()
	if _, ok := GuildMap[u.GuildID]; !ok {
		InitDB(s, u.GuildID)
		LoadGuilds()
	}

	guildBotLog := GuildMap[u.GuildID].GuildConfig.BotLog.ID
	MapMutex.Unlock()

	// Fetches date of account creation and checks if it's younger than 14 days
	creationDate, err := CreationTime(u.User.ID)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
		if err != nil {
			return
		}
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

	// Adds the spambot ban to bannedUsersSlice so it doesn't trigger the OnGuildBan func
	temp.ID = u.User.ID
	temp.User = u.User.Username
	temp.UnbanDate = time.Date(9999, 9, 9, 9, 9, 9, 9, time.Local)
	for index, val := range GuildMap[u.GuildID].BannedUsers {
		if val.ID == u.User.ID {
			GuildMap[u.GuildID].BannedUsers = append(GuildMap[u.GuildID].BannedUsers[:index], GuildMap[u.GuildID].BannedUsers[index+1:]...)
		}
	}
	GuildMap[u.GuildID].BannedUsers = append(GuildMap[u.GuildID].BannedUsers, temp)
	_ = BannedUsersWrite(GuildMap[u.GuildID].BannedUsers, u.GuildID)

	// Adds a bool to memberInfo that it's a suspected spambot account in case they try to reverify
	tempMem = *GuildMap[u.GuildID].MemberInfoMap[u.User.ID]
	tempMem.SuspectedSpambot = true
	GuildMap[u.GuildID].MemberInfoMap[u.User.ID] = &tempMem
	WriteMemberInfo(GuildMap[u.GuildID].MemberInfoMap, u.GuildID)
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
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	// Botlog message
	_, _ = s.ChannelMessageSend(guildBotLog, fmt.Sprintf("Suspected spambot was banned. User: <@!%v>", u.User.ID))
}

// Cleans spoilerroles.json
func cleanSpoilerRoles(s *discordgo.Session, guildID string) error {

	var shouldDelete bool

	guildBotLog := GuildMap[guildID].GuildConfig.BotLog.ID

	// Pulls all of the server roles
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
		if err != nil {
			return err
		}
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
	InitDB(s, g.Guild.ID)
	LoadGuilds()
	MapMutex.Unlock()

	log.Println(fmt.Sprintf("Joined guild %v", g.Guild.Name))
}

// Logs BOT leaving a server
func GuildDelete(s *discordgo.Session, g *discordgo.GuildDelete) {
	if g.Name == "" {
		return
	}
	log.Println(fmt.Sprintf("Left guild %v", g.Guild.Name))
}

// Send number of servers via post request
func sendServers(s *discordgo.Session) {

	if s.State.User.ID != "614495694769618944" {
		return
	}

	guildCountStr := strconv.Itoa(len(s.State.Guilds))
	client := &http.Client{}

	// Discord Bots
	discordBotsGuildCount(client, guildCountStr)

	// Discord Boats
	discordBoatsGuildCount(client, guildCountStr)

	// Bots on Discord
	discordBotsOnDiscordGuildCount(client, guildCountStr)
}

// Sends guild count to discordbots.org
func discordBotsGuildCount(client *http.Client, guildCount string) {

	data := url.Values{
		"server_count": {guildCount},
	}
	req, err := http.NewRequest("POST", "https://discordbots.org/api/bots/614495694769618944/stats", bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	req.Header.Add("Authorization", config.DiscordBotsSecret)
	_, err = client.Do(req)
	if err != nil {
		log.Println(err)
	}
}

// Sends guild count to discord.boats
func discordBoatsGuildCount(client *http.Client, guildCount string) {
	data := url.Values{
		"server_count": {guildCount},
	}
	req, err := http.NewRequest("POST", "https://discord.boats/api/bot/614495694769618944", bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	req.Header.Add("Authorization", config.DiscordBoatsSecret)
	_, err = client.Do(req)
	if err != nil {
		log.Println(err)
	}
}

// Sends guild count to bots.ondiscord.xyz
func discordBotsOnDiscordGuildCount(client *http.Client, guildCount string) {
	data := url.Values{
		"guildCount": {guildCount},
	}
	req, err := http.NewRequest("POST", "https://bots.ondiscord.xyz/bot-api/bots/614495694769618944/guilds", bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	req.Header.Add("Authorization", config.BotsOnDiscordSecret)
	_, err = client.Do(req)
	if err != nil {
		log.Println(err)
	}
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
		if me.Nick != fmt.Sprintf("%v %v", targetPrefix, s.State.User.Username) && me.Nick != "" {
			return
		}
	}

	err = s.GuildMemberNickname(guildID, "@me", fmt.Sprintf("%v %v", guildPrefix, s.State.User.Username))
	if err != nil {
		if _, ok := err.(*discordgo.RESTError); ok {
			if err.(*discordgo.RESTError).Response.Status == "400 Bad Request" {
				_ = s.GuildMemberNickname(guildID, "@me", fmt.Sprintf("%v", s.State.User.Username))
			}
		}
	}
}

// Updates the member counter map
func updateUserCounter(s *discordgo.Session, guildID string) {

	// Fetch guild
	guild, err := s.State.Guild(guildID)
	if err != nil {
		guild, err = s.Guild(guildID)
		if err != nil {
			log.Println(err)
			return
		}
	}

	// Check if member is already in UserCounter map and add him if he isn't
	// Skips BOTs
	for _, member := range guild.Members {
		if member.User.Bot {
			continue
		}
		if _, ok := UserCounter[member.User.ID]; ok {
			continue
		}
		UserCounter[member.User.ID] = true
	}
}
