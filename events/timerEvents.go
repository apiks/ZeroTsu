package events

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"time"

	"github.com/r-anime/ZeroTsu/cache"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"github.com/r-anime/ZeroTsu/functionality"
	"github.com/sasha-s/go-deadlock"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
)

var (
	DailyScheduleWebhooksMap = &safeWebhooksMap{WebhooksMap: make(map[string]*discordgo.Webhook)}

	DailyScheduleWebhooksMapBlock Block
	remindMesFeedBlock            Block
)

type Block struct {
	deadlock.RWMutex
	Block bool
}

type safeWebhooksMap struct {
	deadlock.RWMutex
	WebhooksMap map[string]*discordgo.Webhook
}

func UpdateDailyScheduleWebhooks() {
	DailyScheduleWebhooksMapBlock.Lock()
	if DailyScheduleWebhooksMapBlock.Block {
		DailyScheduleWebhooksMapBlock.Unlock()
		return
	}
	DailyScheduleWebhooksMapBlock.Block = true
	DailyScheduleWebhooksMapBlock.Unlock()

	defer func() {
		DailyScheduleWebhooksMapBlock.Lock()
		DailyScheduleWebhooksMapBlock.Block = false
		DailyScheduleWebhooksMapBlock.Unlock()
	}()

	// Store all of the valid guilds' valid webhooks in a map
	tempWebhooksMap := make(map[string]*discordgo.Webhook)

	// Fetch only guild anime subscriptions from cache
	animeSubsMap := cache.AnimeSubs.GetGuild()

	for guildID, subs := range animeSubsMap {
		if len(subs) == 0 {
			continue
		}

		// Get Discord session for the guild
		guildIDInt, err := strconv.ParseInt(guildID, 10, 64)
		if err != nil {
			continue
		}
		s := config.Mgr.SessionForGuild(guildIDInt)

		// Ensure bot is in the target guild
		_, err = s.State.Guild(guildID)
		if err != nil {
			continue
		}

		// Check if bot has required permissions for the daily schedule channel
		newepisodes := db.GetGuildAutopost(guildID, "dailyschedule")
		if newepisodes == (entities.Cha{}) {
			continue
		}

		perms, err := s.State.UserChannelPermissions(s.State.User.ID, newepisodes.GetID())
		if err != nil || perms&discordgo.PermissionManageWebhooks != discordgo.PermissionManageWebhooks || perms&discordgo.PermissionViewChannel != discordgo.PermissionViewChannel || perms&discordgo.PermissionSendMessages != discordgo.PermissionSendMessages {
			continue
		}

		// Get existing webhook
		ws, err := s.ChannelWebhooks(newepisodes.GetID())
		if err != nil {
			continue
		}
		for _, w := range ws {
			if w.User.ID == s.State.User.ID && w.ChannelID == newepisodes.GetID() {
				tempWebhooksMap[guildID] = w
				break
			}
		}

		// If a valid webhook exists, continue
		if _, ok := tempWebhooksMap[guildID]; ok {
			continue
		}

		// Create a new webhook if none exists
		avatar, err := s.UserAvatarDecode(s.State.User)
		if err != nil {
			continue
		}
		out := new(bytes.Buffer)
		err = png.Encode(out, avatar)
		if err != nil {
			continue
		}
		base64Img := base64.StdEncoding.EncodeToString(out.Bytes())

		wh, err := s.WebhookCreate(newepisodes.GetID(), s.State.User.Username, fmt.Sprintf("data:image/png;base64,%s", base64Img))
		if err != nil {
			continue
		}
		tempWebhooksMap[guildID] = wh
	}

	// Update the global webhooks map
	DailyScheduleWebhooksMap.Lock()
	DailyScheduleWebhooksMap.WebhooksMap = tempWebhooksMap
	DailyScheduleWebhooksMap.Unlock()
}

func WriteEvents(s *discordgo.Session, _ *discordgo.Ready) {
	var randomPlayingMsg string

	for range time.NewTicker(30 * time.Minute).C {
		// Updates playing status
		entities.Mutex.RLock()
		if len(config.PlayingMsg) > 1 {
			randomPlayingMsg = config.PlayingMsg[rand.Intn(len(config.PlayingMsg))]
		}
		entities.Mutex.RUnlock()
		if randomPlayingMsg != "" {
			_ = s.UpdateGameStatus(0, randomPlayingMsg)
		}

		// Sends server count to bot list sites if it's the public ZeroTsu
		functionality.SendServers(strconv.Itoa(config.Mgr.GuildCount()), s)
	}
}

func CommonEvents(_ *discordgo.Session, _ *discordgo.Ready) {
	for range time.NewTicker(1 * time.Minute).C {
		guildIds, err := entities.LoadAllGuildIDs()
		if err != nil {
			log.Printf("Error fetching guild IDs: %v", err)
		}

		// Handles RemindMes
		remindMeHandler(config.Mgr.SessionForDM())

		// Handles Reddit Feeds
		FeedWebhookHandler(guildIds)
		FeedHandler(guildIds)

		// Force garbage collection after processing
		runtime.GC()
	}
}

// remindMeHandler handles sending remindMe messages when called if it's time.
func remindMeHandler(s *discordgo.Session) {
	remindMesFeedBlock.Lock()
	if remindMesFeedBlock.Block {
		remindMesFeedBlock.Unlock()
		return
	}
	remindMesFeedBlock.Block = true
	remindMesFeedBlock.Unlock()

	defer func() {
		remindMesFeedBlock.Lock()
		remindMesFeedBlock.Block = false
		remindMesFeedBlock.Unlock()
	}()

	// Fetch only reminders that are due
	reminders := db.GetDueReminders()
	if len(reminders) == 0 {
		return
	}

	// Process reminders in batches to reduce memory pressure
	const batchSize = 50
	userIDs := make([]string, 0, len(reminders))
	for userID := range reminders {
		userIDs = append(userIDs, userID)
	}

	for i := 0; i < len(userIDs); i += batchSize {
		end := i + batchSize
		if end > len(userIDs) {
			end = len(userIDs)
		}

		batch := userIDs[i:end]
		for _, userID := range batch {
			remindMeSlice := reminders[userID]
			if remindMeSlice == nil || len(remindMeSlice.GetRemindMeSlice()) == 0 {
				continue
			}

			var newReminders []*entities.RemindMe
			now := time.Now()

			for _, remindMe := range remindMeSlice.GetRemindMeSlice() {
				if remindMe == nil {
					continue
				}

				// Check if it's time to send the reminder
				if now.After(remindMe.GetDate()) {
					// Send the reminder
					dm, err := s.UserChannelCreate(userID)
					if err != nil {
						log.Printf("Error creating DM channel for %s: %v\n", userID, err)
						continue
					}
					_, err = s.ChannelMessageSend(dm.ID, fmt.Sprintf("RemindMe: %s", remindMe.GetMessage()))
					if err != nil {
						log.Printf("Error sending reminder to %s: %v\n", userID, err)
						continue
					}

					// Remove reminder from MongoDB **right after sending**
					db.RemoveReminder(userID, remindMe.GetRemindID())
				} else {
					// Keep future reminders
					newReminders = append(newReminders, remindMe)
				}
			}

			// Update the reminders if any were removed
			if len(newReminders) != len(remindMeSlice.GetRemindMeSlice()) {
				remindMeSlice.SetRemindMeSlice(newReminders)
				db.SetReminder(userID, nil, remindMeSlice.Guild, remindMeSlice.Premium)
			}
		}

		// Force GC after each batch
		runtime.GC()
	}
}
