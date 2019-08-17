package misc

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mmcdole/gofeed"

	"github.com/r-anime/ZeroTsu/config"
)

var darlingTrigger int

// Periodic events such as Unbanning and RSS timer every 30 sec
func StatusReady(s *discordgo.Session, e *discordgo.Ready) {

	var banFlag bool

	for _, guild := range e.Guilds {
		// Initialize guild if missing
		initDB(guild.ID)
		writeAll(guild.ID)
	}

	// Clean up SpoilerRoles.json in each guild
	MapMutex.Lock()
	for _, guild := range e.Guilds {
		err := cleanSpoilerRoles(s, guild.ID)
		if err != nil {
			_, _ = s.ChannelMessageSend(GuildMap[guild.ID].GuildConfig.BotLog.ID, err.Error()+"\n"+ErrorLocation(err))
		}
	}
	MapMutex.Unlock()

	_ = s.UpdateStatus(0, config.PlayingMsg)

	for range time.NewTicker(30 * time.Second).C {

		// Checks whether it has to post rss thread and handle remindMes
		for _, guild := range e.Guilds {
			RSSParser(s, guild.ID)
			MapMutex.Lock()
			remindMeHandler(s, guild.ID)

			// Goes through bannedUsers.json if it's not empty and unbans if needed
			if len(GuildMap[guild.ID].BannedUsers) != 0 {
				t := time.Now()
				for index, user := range GuildMap[guild.ID].BannedUsers {
					difference := t.Sub(user.UnbanDate)
					if difference > 0 {
						banFlag = false

						// Checks if user is in MemberInfo and saves him if true
						memberInfoUser, ok := GuildMap[guild.ID].MemberInfoMap[user.ID]
						if !ok {
							continue
						}
						// Fetches all server bans so it can check if the memberInfoUser is banned there (whether he's been manually unbanned for example)
						bans, err := s.GuildBans(guild.ID)
						if err != nil {
							_, err = s.ChannelMessageSend(GuildMap[guild.ID].GuildConfig.BotLog.ID, err.Error()+"\n"+ErrorLocation(err))
							if err != nil {
								continue
							}
							continue
						}
						for _, ban := range bans {
							if ban.User.ID == memberInfoUser.ID {
								banFlag = true
								break
							}
						}
						if banFlag {
							// Unbans memberInfoUser if possible
							err = s.GuildBanDelete(guild.ID, memberInfoUser.ID)
							if err != nil {
								_, err = s.ChannelMessageSend(GuildMap[guild.ID].GuildConfig.BotLog.ID, err.Error()+"\n"+ErrorLocation(err))
								if err != nil {
									continue
								}
								continue
							}
						}

						// Removes unban date entirely
						GuildMap[guild.ID].MemberInfoMap[memberInfoUser.ID].UnbanDate = ""

						// Removes the memberInfoUser ban from bannedUsers.json
						GuildMap[guild.ID].BannedUsers = append(GuildMap[guild.ID].BannedUsers[:index], GuildMap[guild.ID].BannedUsers[index+1:]...)

						// Writes to memberInfo.json and bannedUsers.json
						WriteMemberInfo(GuildMap[guild.ID].MemberInfoMap, guild.ID)
						BannedUsersWrite(GuildMap[guild.ID].BannedUsers, guild.ID)

						// Sends an embed message to bot-log
						_ = UnbanEmbed(s, memberInfoUser, "", GuildMap[guild.ID].GuildConfig.BotLog.ID)
						break
					}
				}
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
	embedMess.Color = 0x00ff00

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

		MapMutex.Lock()
		for _, guild := range e.Guilds {

			guildBotLog := GuildMap[guild.ID].GuildConfig.BotLog.ID

			// Writes emoji stats to disk
			_, err := EmojiStatsWrite(GuildMap[guild.ID].EmojiStats, guild.ID)
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				if err != nil {
					continue
				}
				continue
			}

			// Writes user gain stats to disk
			_, err = UserChangeStatsWrite(GuildMap[guild.ID].UserChangeStats, guild.ID)
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				if err != nil {
					continue
				}
				continue
			}

			// Writes verified stats to disk
			if config.Website != "" {
				err = VerifiedStatsWrite(GuildMap[guild.ID].VerifiedStats, guild.ID)
				if err != nil {
					_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
					if err != nil {
						continue
					}
					continue
				}
			}

			// Writes memberInfo to disk
			WriteMemberInfo(GuildMap[guild.ID].MemberInfoMap, guild.ID)

			// Fetches all server roles
			roles, err := s.GuildRoles(guild.ID)
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				if err != nil {
					continue
				}
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
		}
		MapMutex.Unlock()
	}
}

// Pulls the rss thread and prints it
func RSSParser(s *discordgo.Session, guildID string) {

	// Checks if there are any rss settings for this guild
	MapMutex.Lock()
	if len(GuildMap[guildID].RssThreads) == 0 {
		MapMutex.Unlock()
		return
	}
	// Save current threads as a copy so mapMutex isn't taken all the time when checking the feeds
	rssThreads := GuildMap[guildID].RssThreads
	rssThreadChecks := GuildMap[guildID].RssThreadChecks
	bogLogID := GuildMap[guildID].GuildConfig.BotLog.ID

	t := time.Now()
	hours := time.Hour * 1440

	// Removes a thread if more than 60 days have passed from the rss thread checks. This is to keep DB manageable
	for p := 0; p < len(rssThreadChecks); p++ {
		// Calculates if it's time to remove
		dateRemoval := rssThreadChecks[p].Date.Add(hours)
		difference := t.Sub(dateRemoval)

		// Removes the fact that the thread had been posted already if it's time
		if difference > 0 {
			err := RssThreadsTimerRemove(rssThreadChecks[p].Thread, rssThreadChecks[p].Date, guildID)
			if err != nil {
				_, err = s.ChannelMessageSend(bogLogID, err.Error()+"\n"+ErrorLocation(err))
				if err != nil {
					MapMutex.Unlock()
					return
				}
				MapMutex.Unlock()
				return
			}
		}
	}
	// Updates rssThreadChecks var after the removal
	rssThreadChecks = GuildMap[guildID].RssThreadChecks
	MapMutex.Unlock()

	// Sets up the feed parser
	fp := gofeed.NewParser()
	fp.Client = &http.Client{Transport: &UserAgentTransport{http.DefaultTransport}, Timeout: time.Minute * 1}

	// Save all feeds early to save performance
	var subMap = make(map[string]*gofeed.Feed)
	for _, thread := range rssThreads {
		if _, ok := subMap[thread.Subreddit]; !ok {
			// Wait a bit between each parse
			time.Sleep(250 * time.Millisecond)
			// Parse feed
			feed, err := fp.ParseURL(fmt.Sprintf("http://www.reddit.com/r/%v/%v/.rss", thread.Subreddit, thread.PostType))
			if err != nil {
				return
			}
			subMap[fmt.Sprintf("%v:%v", thread.Subreddit, thread.PostType)] = feed
		}
	}

	for _, thread := range rssThreads {

		// Get the necessary feed from the subMap
		feed := subMap[fmt.Sprintf("%v:%v", thread.Subreddit, thread.PostType)]

		t = time.Now()

		// Iterates through each feed item to see if it finds something from storage that should be posted
		for _, item := range feed.Items {

			// Check if this item exists in rssThreadChecks and skips the item if it does
			var skip = false
			for _, check := range rssThreadChecks {
				if check.GUID == item.GUID {
					skip = true
					break
				}
			}
			if skip {
				continue
			}

			// Save lowercase feed item title
			var itemTitleLrcase  string
			itemTitleLrcase = strings.ToLower(item.Title)

			// Check if author is same and skip if not true
			if thread.Author != "" && item.Author != nil {
				if strings.ToLower(item.Author.Name) != thread.Author {
					continue
				}
			}

			// Check if the feed item title starts with the set thread title
			if thread.Title != "" {
				if !strings.HasPrefix(itemTitleLrcase, thread.Title) {
					continue
				}
			}

			// Writes that thread has been posted
			MapMutex.Lock()
			err := RssThreadsTimerWrite(thread, t, item.GUID, guildID)
			if err != nil {
				MapMutex.Unlock()
				_, _ = s.ChannelMessageSend(bogLogID, err.Error()+"\n"+ErrorLocation(err))
				continue
			}
			// Updates rssThreadChecks var after the write
			rssThreadChecks = GuildMap[guildID].RssThreadChecks
			MapMutex.Unlock()

			// Sends feed item to chat
			message, err := s.ChannelMessageSend(thread.ChannelID, item.Link)
			if err != nil {
				_, _ = s.ChannelMessageSend(bogLogID, err.Error()+"\n"+ErrorLocation(err))
				continue
			}

			// Pins/unpins the feed items if necessary
			if !thread.Pin {
				continue
			}

			pins, err := s.ChannelMessagesPinned(message.ChannelID)
			if err != nil {
				_, _ = s.ChannelMessageSend(bogLogID, err.Error()+"\n"+ErrorLocation(err))
				continue
			}
			// Unpins if necessary
			if len(pins) != 0 {
				for _, pin := range pins {
					if pin.Author.ID == s.State.User.ID {
						if strings.HasPrefix(strings.ToLower(pin.Content), fmt.Sprintf("https://www.reddit.com/r/%v/comments/", thread.Subreddit)) ||
							strings.HasPrefix(strings.ToLower(pin.Content), fmt.Sprintf("http://www.reddit.com/r/%v/comments/", thread.Subreddit)) {
							err = s.ChannelMessageUnpin(pin.ChannelID, pin.ID)
							if err != nil {
								_, _ = s.ChannelMessageSend(bogLogID, err.Error()+"\n"+ErrorLocation(err))
								continue
							}
						}
					}
				}
			}
			// Pins
			err = s.ChannelMessagePin(message.ChannelID, message.ID)
			if err != nil {
				_, _ = s.ChannelMessageSend(bogLogID, err.Error()+"\n"+ErrorLocation(err))
			}
		}
	}
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
	if len(GuildMap[v.GuildID].GuildConfig.VoiceChas) == 0 {
		MapMutex.Unlock()
		return
	} else {
		voiceChannels = GuildMap[v.GuildID].GuildConfig.VoiceChas
	}
	guildBotLog := GuildMap[v.GuildID].GuildConfig.BotLog.ID
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
					_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
					if err != nil {
						return
					}
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
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				if err != nil {
					return
				}
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

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && m.Author.ID == "128312718779219968" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Professor!")
		if err != nil {

			MapMutex.Lock()
			guildBotLog := GuildMap[m.GuildID].GuildConfig.BotLog.ID
			MapMutex.Unlock()

			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && m.Author.ID == "66207186417627136" {

		MapMutex.Lock()
		guildBotLog := GuildMap[m.GuildID].GuildConfig.BotLog.ID
		MapMutex.Unlock()

		randomNum := rand.Intn(5)
		if randomNum == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Bug hunter!")
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
			return
		}
		if randomNum == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Player!")
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
			return
		}
		if randomNum == 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Big brain!")
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
			return
		}
		if randomNum == 3 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Poster expert!")
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
			return
		}
		if randomNum == 4 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Idiot!")
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
			return
		}
		return
	}

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && m.Author.ID == "365245718866427904" {

		MapMutex.Lock()
		guildBotLog := GuildMap[m.GuildID].GuildConfig.BotLog.ID
		MapMutex.Unlock()

		randomNum := rand.Intn(5)
		if randomNum == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Begone ethot.")
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
			return
		}
		if randomNum == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Humph!")
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
			return
		}
		if randomNum == 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Wannabe ethot.")
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
			return
		}
		if randomNum == 3 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Not even worth my time.")
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
			return
		}
		if randomNum == 4 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Okay, maybe you're not that bad.")
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
			return
		}
		return
	}

	if (m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID)) && darlingTrigger > 10 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Daaarling~")
		if err != nil {

			MapMutex.Lock()
			guildBotLog := GuildMap[m.GuildID].GuildConfig.BotLog.ID
			MapMutex.Unlock()

			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		darlingTrigger = 0
		return
	}

	if m.Content == fmt.Sprintf("<@%v>", s.State.User.ID) || m.Content == fmt.Sprintf("<@!%v>", s.State.User.ID) {
		_, err := s.ChannelMessageSend(m.ChannelID, "Baka!")
		if err != nil {

			MapMutex.Lock()
			guildBotLog := GuildMap[m.GuildID].GuildConfig.BotLog.ID
			MapMutex.Unlock()

			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
			if err != nil {
				return
			}
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
	guildBotLog := GuildMap[e.GuildID].GuildConfig.BotLog.ID

	for i := 0; i < len(GuildMap[e.GuildID].BannedUsers); i++ {
		if GuildMap[e.GuildID].BannedUsers[i].ID == e.User.ID {
			MapMutex.Unlock()
			return
		}
	}
	MapMutex.Unlock()
	_, err := s.ChannelMessageSend(guildBotLog, fmt.Sprintf("%v#%v was manually permabanned. ID: %v", e.User.Username, e.User.Discriminator, e.User.ID))
	if err != nil {
		return
	}
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
	for userID, remindMeSlice := range GuildMap[guildID].RemindMes {
		for index, remindMeObject := range remindMeSlice.RemindMeSlice {

			// Checks if it's time to send message/ping the user
			difference := t.Sub(remindMeObject.Date)
			if difference > 0 {

				// Sends message to user DMs if possible
				dm, err := s.UserChannelCreate(userID)
				_, err = s.ChannelMessageSend(dm.ID, "RemindMe: "+remindMeObject.Message)
				// Else sends the message in the channel the command was made in with a ping
				if err != nil {
					// Checks if the user is in the server before pinging him
					_, err := s.GuildMember(guildID, userID)
					if err == nil {
						pingMessage := fmt.Sprintf("<@%v> Remindme: %v", userID, remindMeObject.Message)
						_, err = s.ChannelMessageSend(remindMeObject.CommandChannel, pingMessage)
						if err != nil {
							_, err = s.ChannelMessageSend(GuildMap[guildID].GuildConfig.BotLog.ID, err.Error()+"\n"+ErrorLocation(err))
							if err != nil {
								return
							}
							return
						}
					}
				}

				// Removes the RemindMe object from the RemindMe slice and writes to disk
				if len(remindMeSlice.RemindMeSlice) == 1 {
					delete(GuildMap[guildID].RemindMes, userID)
				} else {
					remindMeSlice.RemindMeSlice = append(remindMeSlice.RemindMeSlice[:index], remindMeSlice.RemindMeSlice[index+1:]...)
					GuildMap[guildID].RemindMes[userID].RemindMeSlice = remindMeSlice.RemindMeSlice
				}
				_, err = RemindMeWrite(GuildMap[guildID].RemindMes, guildID)
				if err != nil {
					_, err = s.ChannelMessageSend(GuildMap[guildID].GuildConfig.BotLog.ID, err.Error()+"\n"+ErrorLocation(err))
					if err != nil {
						return
					}
					return
				}
				break
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
		_, err = s.ChannelMessageSend("566233292026937345", fmt.Sprintf("User joined the server: %v\nAccount age: %v", u.User.Mention(), creationDate.String()))
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
	BannedUsersWrite(GuildMap[u.GuildID].BannedUsers, u.GuildID)

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
