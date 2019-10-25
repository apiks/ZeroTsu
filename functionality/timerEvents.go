package functionality

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mmcdole/gofeed"

	"github.com/r-anime/ZeroTsu/config"
)

var redditFeedBlock bool

// Write Events
func WriteEvents(s *discordgo.Session, e *discordgo.Ready) {
	for range time.NewTicker(30 * time.Minute).C {
		var randomPlayingMsg string
		t := time.Now()
		rand.Seed(t.UnixNano())

		// Updates playing status
		Mutex.RLock()
		if len(config.PlayingMsg) > 1 {
			randomPlayingMsg = config.PlayingMsg[rand.Intn(len(config.PlayingMsg))]
		}
		Mutex.RUnlock()
		if randomPlayingMsg != "" {
			_ = s.UpdateStatus(0, randomPlayingMsg)
		}

		for _, guild := range e.Guilds {

			HandleNewGuild(s, guild.ID)

			// Updates BOT nickname
			DynamicNicknameChange(s, guild.ID)

			// Clears up spoilerRoles.json
			err := cleanSpoilerRoles(s, guild.ID)
			if err != nil {
				log.Println(err)
			}

			// Fetches all server roles
			roles, err := s.GuildRoles(guild.ID)
			if err != nil {
				log.Println(err)
			}

			Mutex.Lock()
			// Updates optin role stat
			if roles != nil {
				for channel := range GuildMap[guild.ID].ChannelStats {
					if GuildMap[guild.ID].ChannelStats[channel].RoleCount == nil {
						GuildMap[guild.ID].ChannelStats[channel].RoleCount = make(map[string]int)
					}
					if GuildMap[guild.ID].ChannelStats[channel].Optin {
						GuildMap[guild.ID].ChannelStats[channel].RoleCount[t.Format(DateFormat)] = GetRoleUserAmount(guild, roles, GuildMap[guild.ID].ChannelStats[channel].Name)
					}
				}
			}

			// Writes emoji stats to disk
			err = EmojiStatsWrite(GuildMap[guild.ID].EmojiStats, guild.ID)
			if err != nil {
				log.Println(err)
			}

			// Writes user gain stats to disk
			_, err = UserChangeStatsWrite(GuildMap[guild.ID].UserChangeStats, guild.ID)
			if err != nil {
				log.Println(err)
			}

			// Writes verified stats to disk
			if config.Website != "" {
				err = VerifiedStatsWrite(GuildMap[guild.ID].VerifiedStats, guild.ID)
				if err != nil {
					log.Println(err)
				}
			}

			// Writes channel stats to disk
			_, err = ChannelStatsWrite(GuildMap[guild.ID].ChannelStats, guild.ID)
			if err != nil {
				log.Println(err)
			}
			Mutex.Unlock()
		}

		// Sends server count to bot list sites if it's the public ZeroTsu
		sendServers(s)
	}
}

// Common Timer Events (every 30 seconds)
func CommonEvents(s *discordgo.Session, e *discordgo.Ready) {
	for range time.NewTicker(30 * time.Second).C {
		for _, guild := range e.Guilds {
			// Handles Unbans and Unmutes
			punishmentHandler(s, guild.ID)

			// Handles RemindMes
			remindMeHandler(s, guild.ID)

			// Handles Reddit Feeds
			feedHandler(s, guild.ID)
		}
	}
}

func punishmentHandler(s *discordgo.Session, guildID string) {

	// Fetches all server bans so it can check if the memberInfo User is banned there (whether he's been manually unbanned for example)
	bans, err := s.GuildBans(guildID)
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	t := time.Now()

	Mutex.Lock()
	defer Mutex.Unlock()

	// Checks if there are punishedUsers in this guild
	if GuildMap[guildID].PunishedUsers == nil || len(GuildMap[guildID].PunishedUsers) == 0 {
		return
	}

	// Unbans/Unmutes users
	wg.Add(len(GuildMap[guildID].PunishedUsers))
	for i := len(GuildMap[guildID].PunishedUsers) - 1; i >= 0; i-- {
		go func(i int) {
			defer wg.Done()
			fieldRemoved := unbanHandler(s, guildID, i, bans, &t)
			if fieldRemoved {
				return
			}
			unmuteHandler(s, guildID, i, &t)
		}(i)
	}

	wg.Wait()
}

func unbanHandler(s *discordgo.Session, guildID string, i int, bans []*discordgo.GuildBan, t *time.Time) bool {
	if GuildMap[guildID].PunishedUsers[i].UnbanDate == ZeroTimeValue {
		return false
	}
	banDifference := t.Sub(GuildMap[guildID].PunishedUsers[i].UnbanDate)
	if banDifference <= 0 {
		return false
	}

	// Set flag for whether user is a banned user in guild and db
	banFlag := false
	for _, ban := range bans {
		if ban.User.ID == GuildMap[guildID].PunishedUsers[i].ID {
			banFlag = true
			break
		}
	}

	// Unbans User if possible and needed
	if banFlag {
		_ = s.GuildBanDelete(guildID, GuildMap[guildID].PunishedUsers[i].ID)
	}

	// Removes unban date entirely
	memberInfoUser, ok := GuildMap[guildID].MemberInfoMap[GuildMap[guildID].PunishedUsers[i].ID]
	if ok {
		GuildMap[guildID].MemberInfoMap[GuildMap[guildID].PunishedUsers[i].ID].UnbanDate = ""
	}

	// Removes the unbanDate from punishedUsers.json
	if GuildMap[guildID].PunishedUsers[i].UnmuteDate != ZeroTimeValue {
		temp := &PunishedUsers{
			ID:         GuildMap[guildID].PunishedUsers[i].ID,
			User:       GuildMap[guildID].PunishedUsers[i].User,
			UnmuteDate: GuildMap[guildID].PunishedUsers[i].UnmuteDate,
		}
		GuildMap[guildID].PunishedUsers[i] = temp
	} else {
		if i < len(GuildMap[guildID].PunishedUsers)-1 {
			copy(GuildMap[guildID].PunishedUsers[i:], GuildMap[guildID].PunishedUsers[i+1:])
		}
		GuildMap[guildID].PunishedUsers[len(GuildMap[guildID].PunishedUsers)-1] = nil
		GuildMap[guildID].PunishedUsers = GuildMap[guildID].PunishedUsers[:len(GuildMap[guildID].PunishedUsers)-1]
	}

	// Sends an embed message to bot-log
	if banFlag && ok {
		guildSettings := GuildMap[guildID].GetGuildSettings()
		if guildSettings.BotLog != nil {
			if guildSettings.BotLog.ID != "" {
				_ = UnbanEmbed(s, memberInfoUser, "", GuildMap[guildID].GuildConfig.BotLog.ID)
			}
		}
	}

	if ok {
		_ = WriteMemberInfo(GuildMap[guildID].MemberInfoMap, guildID)
	}
	_ = PunishedUsersWrite(GuildMap[guildID].PunishedUsers, guildID)

	return true
}

func unmuteHandler(s *discordgo.Session, guildID string, i int, t *time.Time) {
	if GuildMap[guildID].PunishedUsers[i].UnmuteDate == ZeroTimeValue {
		return
	}
	muteDifference := t.Sub(GuildMap[guildID].PunishedUsers[i].UnmuteDate)
	if muteDifference <= 0 {
		return
	}

	// Set flag for whether user has the muted role and unmutes them
	muteFlag := false
	mem, err := s.State.Member(guildID, GuildMap[guildID].PunishedUsers[i].ID)
	if err != nil {
		mem, _ = s.GuildMember(guildID, GuildMap[guildID].PunishedUsers[i].ID)
	}
	if mem != nil {
		for _, roleID := range mem.Roles {
			if GuildMap[guildID].GuildConfig.MutedRole != nil {
				if roleID == GuildMap[guildID].GuildConfig.MutedRole.ID {
					err := s.GuildMemberRoleRemove(guildID, GuildMap[guildID].PunishedUsers[i].ID, roleID)
					if err == nil {
						muteFlag = true
					}
					break
				}
			} else {
				deb, _ := s.GuildRoles(guildID)
				for _, role := range deb {
					if strings.ToLower(role.Name) == "muted" || strings.ToLower(role.Name) == "t-mute" {
						err := s.GuildMemberRoleRemove(guildID, GuildMap[guildID].PunishedUsers[i].ID, role.ID)
						if err == nil {
							muteFlag = true
						}
						break
					}
				}
			}
			if muteFlag {
				break
			}
		}
	}

	// Removes unmute date entirely
	memberInfoUser, ok := GuildMap[guildID].MemberInfoMap[GuildMap[guildID].PunishedUsers[i].ID]
	if ok {
		memberInfoUser.UnmuteDate = ""
	}

	// Removes the unmuteDate from punishedUsers.json
	if GuildMap[guildID].PunishedUsers[i].UnbanDate != ZeroTimeValue {
		temp := &PunishedUsers{
			ID:        GuildMap[guildID].PunishedUsers[i].ID,
			User:      GuildMap[guildID].PunishedUsers[i].User,
			UnbanDate: GuildMap[guildID].PunishedUsers[i].UnbanDate,
		}
		GuildMap[guildID].PunishedUsers[i] = temp
	} else {
		if i < len(GuildMap[guildID].PunishedUsers)-1 {
			copy(GuildMap[guildID].PunishedUsers[i:], GuildMap[guildID].PunishedUsers[i+1:])
		}
		GuildMap[guildID].PunishedUsers[len(GuildMap[guildID].PunishedUsers)-1] = nil
		GuildMap[guildID].PunishedUsers = GuildMap[guildID].PunishedUsers[:len(GuildMap[guildID].PunishedUsers)-1]
	}

	// Sends an embed message to bot-log
	if muteFlag && ok {
		guildSettings := GuildMap[guildID].GetGuildSettings()
		if guildSettings.BotLog != nil {
			if guildSettings.BotLog.ID != "" {
				_ = UnmuteEmbed(s, memberInfoUser, "", GuildMap[guildID].GuildConfig.BotLog.ID)
			}
		}
	}

	if ok {
		_ = WriteMemberInfo(GuildMap[guildID].MemberInfoMap, guildID)
	}
	_ = PunishedUsersWrite(GuildMap[guildID].PunishedUsers, guildID)
}

// Sends remindMe message if it is time, either as a DM or ping
func remindMeHandler(s *discordgo.Session, guildID string) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in remindMeHandler")
			log.Println("stacktrace from panic: \n" + string(debug.Stack()))
		}
	}()

	var writeFlag bool
	var wg sync.WaitGroup
	t := time.Now()

	Mutex.Lock()
	if SharedInfo.RemindMes == nil {
		Mutex.Unlock()
		return
	}

	for userID, remindMeSlice := range SharedInfo.RemindMes {
		if remindMeSlice == nil ||
			remindMeSlice.RemindMeSlice == nil ||
			len(remindMeSlice.RemindMeSlice) == 0 {
			continue
		}

		for i := len(remindMeSlice.RemindMeSlice)-1; i >= 0; i-- {
			if remindMeSlice.RemindMeSlice == nil || remindMeSlice.RemindMeSlice[i] == nil {
				continue
			}

			// Checks if it's time to send message/ping the user
			difference := t.Sub(remindMeSlice.RemindMeSlice[i].Date)
			if difference <= 0 {
				continue
			}

			// Sends message to user DMs if possible
			// Else sends the message in the channel the command was made in with a ping
			msgDM := fmt.Sprintf("RemindMe: %s", remindMeSlice.RemindMeSlice[i].Message)
			msgChannel := fmt.Sprintf("<@%s> Remindme: %s", userID, remindMeSlice.RemindMeSlice[i].Message)
			cmdChannel := remindMeSlice.RemindMeSlice[i].CommandChannel
			Mutex.Unlock()

			wg.Add(1)
			go func() {
				defer wg.Done()

				dm, err := s.UserChannelCreate(userID)
				if err == nil {
					_, err = s.ChannelMessageSend(dm.ID, msgDM)
				}
				if err != nil && guildID != "" {
					// Checks if the user is in the server and then pings him if true
					_, err := s.GuildMember(guildID, userID)
					if err == nil {
						_, _ = s.ChannelMessageSend(cmdChannel, msgChannel)
					}
				}
			}()

			// Sets write Flag
			writeFlag = true

			Mutex.Lock()
			SharedInfo.RemindMes[userID].RemindMeSlice[i] = SharedInfo.RemindMes[userID].RemindMeSlice[len(SharedInfo.RemindMes[userID].RemindMeSlice)-1]
			SharedInfo.RemindMes[userID].RemindMeSlice[len(SharedInfo.RemindMes[userID].RemindMeSlice)-1] = nil
			SharedInfo.RemindMes[userID].RemindMeSlice = SharedInfo.RemindMes[userID].RemindMeSlice[:len(SharedInfo.RemindMes[userID].RemindMeSlice)-1]
		}
	}

	if !writeFlag {
		Mutex.Unlock()
		return
	}

	err := RemindMeWrite(SharedInfo.RemindMes)
	if err != nil && guildID != "" {
		guildSettings := GuildMap[guildID].GetGuildSettings()
		Mutex.Unlock()
		LogError(s, guildSettings.BotLog, err)
		return
	}
	Mutex.Unlock()

	wg.Wait()
}

// Pulls reddit feeds and prints them
func feedHandler(s *discordgo.Session, guildID string) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in feedHandler")
			log.Println("stacktrace from panic: \n" + string(debug.Stack()))
		}
	}()

	// Blocks handling of new feed threads if there are some currently being sent
	Mutex.RLock()
	if redditFeedBlock {
		Mutex.RUnlock()
		return
	}
	Mutex.RUnlock()

	var (
		pinnedItems   = make(map[*gofeed.Item]bool)
		subMap        = make(map[string]*gofeed.Feed)
		threadsToPost = make(map[*RssThread][]*gofeed.Item)

		rssThreadChecksFlag bool
	)

	t := time.Now()
	hours := time.Hour * 1440

	// Sets up the feed parser
	fp := gofeed.NewParser()
	fp.Client = &http.Client{Transport: &UserAgentTransport{http.DefaultTransport}, Timeout: time.Second * 10}

	// Checks if there are any feeds for this guild
	Mutex.RLock()
	if GuildMap[guildID].Feeds == nil || len(GuildMap[guildID].Feeds) == 0 {
		Mutex.RUnlock()
		return
	}

	// Save current threads as a copy so mapMutex isn't taken all the time when checking the feeds
	rssThreads := GuildMap[guildID].Feeds
	rssThreadChecks := GuildMap[guildID].RssThreadChecks
	Mutex.RUnlock()

	// Removes a thread if more than 60 days have passed from the rss thread checks. This is to keep DB manageable
	for p := 0; p < len(rssThreadChecks); p++ {
		// Calculates if it's time to remove
		dateRemoval := rssThreadChecks[p].Date.Add(hours)
		difference := t.Sub(dateRemoval)

		// Removes the fact that the thread had been posted already if it's time
		if difference <= 0 {
			continue
		}

		Mutex.Lock()
		err := RssThreadsTimerRemove(rssThreadChecks[p].Thread, guildID)
		if err != nil {
			Mutex.Unlock()
			log.Println(err)
			continue
		}
		Mutex.Unlock()

		rssThreadChecksFlag = true
	}

	// Updates rssThreadChecks var after the removal
	if rssThreadChecksFlag {
		Mutex.RLock()
		rssThreadChecks = GuildMap[guildID].RssThreadChecks
		Mutex.RUnlock()
	}

	// Save all feeds early to save performance
	for _, thread := range rssThreads {
		if _, ok := subMap[thread.Subreddit]; ok {
			continue
		}
		// Parse feed
		feed, err := fp.ParseURL(fmt.Sprintf("https://www.reddit.com/r/%s/%s/.rss", thread.Subreddit, thread.PostType))
		if err != nil {
			if _, ok := err.(*gofeed.HTTPError); ok {
				if err.(*gofeed.HTTPError).StatusCode == 429 {
					return
				}
			}
			continue
		}
		subMap[fmt.Sprintf("%s:%s", thread.Subreddit, thread.PostType)] = feed
	}

	t = time.Now()
	for _, thread := range rssThreads {

		// Get the necessary feed from the subMap
		feed, ok := subMap[fmt.Sprintf("%s:%s", thread.Subreddit, thread.PostType)]
		if !ok {
			continue
		}

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
				if strings.ToLower(item.Author.Name) != fmt.Sprintf("/u/%s", thread.Author) {
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
			Mutex.Lock()
			err := RssThreadsTimerWrite(thread, t, item.GUID, guildID)
			if err != nil {
				Mutex.Unlock()
				log.Println(err)
				continue
			}
			// Updates rssThreadChecks var after the write
			rssThreadChecks = GuildMap[guildID].RssThreadChecks
			Mutex.Unlock()

			// Adds the item to the threads to send map
			threadsToPost[thread] = append(threadsToPost[thread], item)
		}
	}

	// Sends the threads concurrently in slow mode
	go func() {
		Mutex.Lock()
		redditFeedBlock = true
		Mutex.Unlock()
		for thread, items := range threadsToPost {
			for _, item := range items {
				// Sends the feed item
				time.Sleep(time.Second * 4)
				message, err := feedEmbed(s, thread, item)
				if err != nil {
					continue
				}

				// Pins/unpins the feed items if necessary
				if !thread.Pin {
					continue
				}
				if _, ok := pinnedItems[item]; ok {
					continue
				}

				pins, err := s.ChannelMessagesPinned(message.ChannelID)
				if err != nil {
					continue
				}
				// Unpins if necessary
				for _, pin := range pins {

					// Checks for whether the pin is one that should be unpinned
					if pin.Author.ID != s.State.User.ID {
						continue
					}
					if len(pin.Embeds) == 0 {
						continue
					}
					if pin.Embeds[0].Author == nil {
						continue
					}
					if !strings.HasPrefix(strings.ToLower(pin.Embeds[0].Author.URL), fmt.Sprintf("https://www.reddit.com/r/%s/comments/", thread.Subreddit)) {
						continue
					}

					_ = s.ChannelMessageUnpin(pin.ChannelID, pin.ID)
				}
				// Pins
				_ = s.ChannelMessagePin(message.ChannelID, message.ID)
				pinnedItems[item] = true
			}
			delete(threadsToPost, thread)
		}
		Mutex.Lock()
		redditFeedBlock = false
		Mutex.Unlock()
	}()
}
