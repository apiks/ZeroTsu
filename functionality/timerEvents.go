package functionality

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mmcdole/gofeed"

	"github.com/r-anime/ZeroTsu/config"
)

var redditFeedBlock bool

// Write Events
func WriteEvents(s *discordgo.Session, e *discordgo.Ready) {
	for range time.NewTicker(30 * time.Minute).C {

		t := time.Now()

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

		for _, guild := range e.Guilds {
			HandleNewGuild(s, guild.ID)

			// Writes emoji stats to disk
			err := EmojiStatsWrite(GuildMap[guild.ID].EmojiStats, guild.ID)
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

			// Clears up spoilerRoles.json
			err = cleanSpoilerRoles(s, guild.ID)
			if err != nil {
				log.Println(err)
			}

			// Updates BOT nickname
			DynamicNicknameChange(s, guild.ID)

			// Fetches all server roles
			roles, err := s.GuildRoles(guild.ID)
			if err != nil {
				log.Println(err)
			}

			// Updates optin role stat
			for channel := range GuildMap[guild.ID].ChannelStats {
				if GuildMap[guild.ID].ChannelStats[channel].RoleCount == nil {
					GuildMap[guild.ID].ChannelStats[channel].RoleCount = make(map[string]int)
				}
				if GuildMap[guild.ID].ChannelStats[channel].Optin {
					GuildMap[guild.ID].ChannelStats[channel].RoleCount[t.Format(DateFormat)] = GetRoleUserAmount(guild, roles, GuildMap[guild.ID].ChannelStats[channel].Name)
				}
			}
		}
		MapMutex.Unlock()

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
	MapMutex.Lock()
	defer MapMutex.Unlock()

	// Checks if there are punishedUsers in this guild
	if GuildMap[guildID].PunishedUsers == nil || len(GuildMap[guildID].PunishedUsers) == 0 {
		return
	}

	// Fetches all server bans so it can check if the memberInfo User is banned there (whether he's been manually unbanned for example)
	bans, err := s.GuildBans(guildID)
	if err != nil {
		return
	}

	// Unbans/Unmutes users
	for i := len(GuildMap[guildID].PunishedUsers) - 1; i >= 0; i-- {
		fieldRemoved := unbanHandler(s, guildID, i, bans)
		if fieldRemoved {
			continue
		}
		unmuteHandler(s, guildID, i)
	}
}

func unbanHandler(s *discordgo.Session, guildID string, i int, bans []*discordgo.GuildBan) bool {
	t := time.Now()
	zeroTimeValue := time.Time{}

	if GuildMap[guildID].PunishedUsers[i].UnbanDate == zeroTimeValue {
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
		err := s.GuildBanDelete(guildID, GuildMap[guildID].PunishedUsers[i].ID)
		if err != nil {
			guildSettings := GuildMap[guildID].GetGuildSettings()
			LogError(s, guildSettings.BotLog, err)
			return false
		}
	}

	// Removes unban date entirely
	memberInfoUser, ok := GuildMap[guildID].MemberInfoMap[GuildMap[guildID].PunishedUsers[i].ID]
	if ok {
		GuildMap[guildID].MemberInfoMap[GuildMap[guildID].PunishedUsers[i].ID].UnbanDate = ""
	}

	// Removes the unbanDate from punishedUsers.json
	if GuildMap[guildID].PunishedUsers[i].UnmuteDate != zeroTimeValue {
		temp := PunishedUsers{
			ID:         GuildMap[guildID].PunishedUsers[i].ID,
			User:       GuildMap[guildID].PunishedUsers[i].User,
			UnmuteDate: GuildMap[guildID].PunishedUsers[i].UnmuteDate,
		}
		GuildMap[guildID].PunishedUsers[i] = &temp
	} else {
		GuildMap[guildID].PunishedUsers = append(GuildMap[guildID].PunishedUsers[:i], GuildMap[guildID].PunishedUsers[i+1:]...)
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

func unmuteHandler(s *discordgo.Session, guildID string, i int) {
	t := time.Now()
	zeroTimeValue := time.Time{}

	if GuildMap[guildID].PunishedUsers[i].UnmuteDate == zeroTimeValue {
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
	if GuildMap[guildID].PunishedUsers[i].UnbanDate != zeroTimeValue {
		temp := PunishedUsers{
			ID:        GuildMap[guildID].PunishedUsers[i].ID,
			User:      GuildMap[guildID].PunishedUsers[i].User,
			UnbanDate: GuildMap[guildID].PunishedUsers[i].UnbanDate,
		}
		GuildMap[guildID].PunishedUsers[i] = &temp
	} else {
		GuildMap[guildID].PunishedUsers = append(GuildMap[guildID].PunishedUsers[:i], GuildMap[guildID].PunishedUsers[i+1:]...)
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
		}
	}()

	var writeFlag bool
	t := time.Now()

	MapMutex.Lock()
	defer MapMutex.Unlock()

	for userID, remindMeSlice := range SharedInfo.RemindMes {
		if remindMeSlice == nil || remindMeSlice.RemindMeSlice == nil {
			continue
		}
		if len(remindMeSlice.RemindMeSlice) == 0 {
			delete(SharedInfo.RemindMes, userID)
			writeFlag = true
			continue
		}

		temp := remindMeSlice.RemindMeSlice[:0]
		for i, remindMe := range remindMeSlice.RemindMeSlice {
			// Checks if it's time to send message/ping the user
			difference := t.Sub(remindMeSlice.RemindMeSlice[i].Date)
			if difference <= 0 {
				temp = append(temp, remindMe)
				continue
			}

			// Sends message to user DMs if possible
			// Else sends the message in the channel the command was made in with a ping
			dm, err := s.UserChannelCreate(userID)
			if err == nil {
				_, err = s.ChannelMessageSend(dm.ID, "RemindMe: "+remindMeSlice.RemindMeSlice[i].Message)
			}
			if err != nil && guildID != "" {
				// Checks if the user is in the server and then pings him if true
				_, err := s.GuildMember(guildID, userID)
				if err == nil {
					_, err := s.ChannelMessageSend(remindMeSlice.RemindMeSlice[i].CommandChannel, fmt.Sprintf("<@%s> Remindme: %s", userID, remindMeSlice.RemindMeSlice[i].Message))
					if err != nil {
						temp = append(temp, remindMe)
						continue
					}
				}
			}

			// Sets write Flag
			writeFlag = true
		}
		if len(temp) == 0 {
			delete(SharedInfo.RemindMes, userID)
			writeFlag = true
			continue
		}
		remindMeSlice.RemindMeSlice = temp
	}

	if !writeFlag {
		return
	}

	err := RemindMeWrite(SharedInfo.RemindMes)
	if err != nil && guildID != "" {
		guildSettings := GuildMap[guildID].GetGuildSettings()
		LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Pulls reddit feeds and prints them
func feedHandler(s *discordgo.Session, guildID string) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in remindMeHandler")
		}
	}()

	// Blocks handling of new feed threads if there are some currently being sent
	if redditFeedBlock {
		return
	}

	var (
		pinnedItems   = make(map[*gofeed.Item]bool)
		subMap        = make(map[string]*gofeed.Feed)
		threadsToPost = make(map[*RssThread][]*gofeed.Item)
	)

	t := time.Now()
	hours := time.Hour * 1440

	// Sets up the feed parser
	fp := gofeed.NewParser()
	fp.Client = &http.Client{Transport: &UserAgentTransport{http.DefaultTransport}, Timeout: time.Second * 10}

	// Checks if there are any feeds for this guild
	MapMutex.Lock()
	if GuildMap[guildID].Feeds == nil || len(GuildMap[guildID].Feeds) == 0 {
		MapMutex.Unlock()
		return
	}

	// Save current threads as a copy so mapMutex isn't taken all the time when checking the feeds
	rssThreads := GuildMap[guildID].Feeds
	rssThreadChecks := GuildMap[guildID].RssThreadChecks

	// Removes a thread if more than 60 days have passed from the rss thread checks. This is to keep DB manageable
	for p := 0; p < len(rssThreadChecks); p++ {
		// Calculates if it's time to remove
		dateRemoval := rssThreadChecks[p].Date.Add(hours)
		difference := t.Sub(dateRemoval)

		// Removes the fact that the thread had been posted already if it's time
		if difference <= 0 {
			continue
		}
		err := RssThreadsTimerRemove(rssThreadChecks[p].Thread, guildID)
		if err != nil {
			log.Println(err)
			continue
		}
	}
	// Updates rssThreadChecks var after the removal
	rssThreadChecks = GuildMap[guildID].RssThreadChecks
	MapMutex.Unlock()

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
		feed := subMap[fmt.Sprintf("%s:%s", thread.Subreddit, thread.PostType)]

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
			MapMutex.Lock()
			err := RssThreadsTimerWrite(thread, t, item.GUID, guildID)
			if err != nil {
				MapMutex.Unlock()
				log.Println(err)
				continue
			}
			// Updates rssThreadChecks var after the write
			rssThreadChecks = GuildMap[guildID].RssThreadChecks
			MapMutex.Unlock()

			// Adds the item to the threads to send map
			threadsToPost[thread] = append(threadsToPost[thread], item)
		}
	}

	// Sends the threads concurrently in slow mode
	go func() {
		redditFeedBlock = true
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
		redditFeedBlock = false
	}()
}
