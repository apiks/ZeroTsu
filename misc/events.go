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

			// Goes through punishedUsers.json if it's not empty and removes punishments if needed
			if len(GuildMap[guild.ID].PunishedUsers) == 0 {
				MapMutex.Unlock()
				continue
			}

			// Fetches all server bans so it can check if the memberInfo User is banned there (whether he's been manually unbanned for example)
			bans, err := s.GuildBans(guild.ID)
			if err != nil {
				MapMutex.Unlock()
				_, _ = s.ChannelMessageSend(GuildMap[guild.ID].GuildConfig.BotLog.ID, err.Error()+"\n"+ErrorLocation(err))
				continue
			}

			for i := len(GuildMap[guild.ID].PunishedUsers) - 1; i >= 0; i-- {
				fieldRemoved := unbanHandler(s, guild.ID, i, bans)
				if fieldRemoved {
					continue
				}
				unmuteHandler(s, guild.ID, i)
			}
			MapMutex.Unlock()
		}
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

	// Set flag for whether user is a banned user
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
			_, _ = s.ChannelMessageSend(GuildMap[guildID].GuildConfig.BotLog.ID, err.Error()+"\n"+ErrorLocation(err))
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
		temp := PunishedUsers {
			ID:         GuildMap[guildID].PunishedUsers[i].ID,
			User:       GuildMap[guildID].PunishedUsers[i].User,
			UnmuteDate: GuildMap[guildID].PunishedUsers[i].UnmuteDate,
		}
		GuildMap[guildID].PunishedUsers[i] = temp
	} else {
		GuildMap[guildID].PunishedUsers = append(GuildMap[guildID].PunishedUsers[:i], GuildMap[guildID].PunishedUsers[i+1:]...)
	}


	// Sends an embed message to bot-log
	if banFlag && ok {
		err := UnbanEmbed(s, memberInfoUser, "", GuildMap[guildID].GuildConfig.BotLog.ID)
		if err != nil {
			log.Println(err)
		}
	}

	if ok {
		WriteMemberInfo(GuildMap[guildID].MemberInfoMap, guildID)
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
		GuildMap[guildID].MemberInfoMap[memberInfoUser.ID].UnmuteDate = ""
	}

	// Removes the unmuteDate from punishedUsers.json
	if GuildMap[guildID].PunishedUsers[i].UnbanDate != zeroTimeValue {
		temp := PunishedUsers {
			ID:         GuildMap[guildID].PunishedUsers[i].ID,
			User:       GuildMap[guildID].PunishedUsers[i].User,
			UnbanDate: GuildMap[guildID].PunishedUsers[i].UnbanDate,
		}
		GuildMap[guildID].PunishedUsers[i] = temp
	} else {
		GuildMap[guildID].PunishedUsers = append(GuildMap[guildID].PunishedUsers[:i], GuildMap[guildID].PunishedUsers[i+1:]...)
	}

	// Sends an embed message to bot-log
	if muteFlag && ok {
		err = UnmuteEmbed(s, memberInfoUser, "", GuildMap[guildID].GuildConfig.BotLog.ID)
		if err != nil {
			log.Println(err)
		}
	}

	if ok {
		WriteMemberInfo(GuildMap[guildID].MemberInfoMap, guildID)
	}
	_ = PunishedUsersWrite(GuildMap[guildID].PunishedUsers, guildID)
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

func UnmuteEmbed(s *discordgo.Session, user *UserInfo, mod string, botLog string) error {

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
		embedMess.Title = fmt.Sprintf("%v#%v has been unmuted.", user.Username, user.Discrim)
	} else {
		embedMess.Title = fmt.Sprintf("%v#%v has been unmuted by %v.", user.Username, user.Discrim, mod)
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
		if _, ok := subMap[thread.Subreddit]; ok {
			continue
		}
		// Parse feed
		feed, err := fp.ParseURL(fmt.Sprintf("http://www.reddit.com/r/%v/%v/.rss", thread.Subreddit, thread.PostType))
		if err != nil {
			return
		}
		subMap[fmt.Sprintf("%v:%v", thread.Subreddit, thread.PostType)] = feed
	}

	threadsToPost := make(map[RssThread][]*gofeed.Item)

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
				time.Sleep(time.Second * 3)
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
					if !strings.HasPrefix(strings.ToLower(pin.Embeds[0].Author.URL), fmt.Sprintf("https://www.reddit.com/r/%v/comments/", thread.Subreddit)) {
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

func feedEmbed(s *discordgo.Session, thread RssThread, item *gofeed.Item) (*discordgo.Message, error) {
	var (
		embedImage = &discordgo.MessageEmbedImage{}
		imageLink  = "https://"
		footerText = fmt.Sprintf("r/%v - %v", thread.Subreddit, thread.PostType)
	)

	// Append custom user author to footer if he exists in thread
	if thread.Author != "" {
		footerText += fmt.Sprintf(" - u/%v", thread.Author)
	}

	// Parse image if it exists between a preset number of allowed domains
	imageStrings := strings.Split(item.Content, "[link]")
	if len(imageStrings) > 1 {
		imageStrings = strings.Split(imageStrings[0], "https://")
		imageLink += strings.Split(imageStrings[len(imageStrings)-1], "\"")[0]
	}
	if strings.HasPrefix(imageLink, "https://i.redd.it/") ||
		strings.HasPrefix(imageLink, "https://i.imgur.com/") ||
		strings.HasPrefix(imageLink, "https://i.gyazo.com/") {
		if strings.Contains(imageLink, ".jpg") ||
			strings.Contains(imageLink, ".jpeg") ||
			strings.Contains(imageLink, ".png") ||
			strings.Contains(imageLink, ".webp") ||
			strings.Contains(imageLink, ".gifv") ||
			strings.Contains(imageLink, ".gif") {
			embedImage.URL = imageLink
		}
	}

	embed := &discordgo.MessageEmbed {
		Author: &discordgo.MessageEmbedAuthor {
			URL:          item.Link,
			Name:         item.Title,
			IconURL:      "https://images-eu.ssl-images-amazon.com/images/I/418PuxYS63L.png",
			ProxyIconURL: "",
		},
		Description: item.Description,
		Timestamp:   item.Published,
		Color:       16758465,
		Footer:      &discordgo.MessageEmbedFooter {
			Text:	 footerText,
		},
		Image:       embedImage,
	}

	// Creates the complex message to send
	data := &discordgo.MessageSend {
		Content: item.Link,
		Embed:   embed,
	}

	// Sends the message
	message, err := s.ChannelMessageSendComplex(thread.ChannelID, data)
	if err != nil {
		return nil, err
	}


	return message, nil
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

	for _, user := range GuildMap[e.GuildID].PunishedUsers {
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
	MapMutex.Unlock()

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

		temp    PunishedUsers
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
	for index, val := range GuildMap[u.GuildID].PunishedUsers {
		if val.ID == u.User.ID {
			GuildMap[u.GuildID].PunishedUsers = append(GuildMap[u.GuildID].PunishedUsers[:index], GuildMap[u.GuildID].PunishedUsers[index+1:]...)
		}
	}
	GuildMap[u.GuildID].PunishedUsers = append(GuildMap[u.GuildID].PunishedUsers, temp)
	_ = PunishedUsersWrite(GuildMap[u.GuildID].PunishedUsers, u.GuildID)

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
