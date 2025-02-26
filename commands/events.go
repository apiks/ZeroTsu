package commands

import (
	"github.com/r-anime/ZeroTsu/entities"
	"log"
	"strconv"
	"time"

	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/events"
	"github.com/sasha-s/go-deadlock"
	"golang.org/x/sync/errgroup"

	"github.com/bwmarrin/discordgo"
)

type SafeTime struct {
	deadlock.RWMutex
	Time time.Time
}

var DailyScheduleEventsBlock events.Block
var Today = &SafeTime{Time: time.Now()}

func dailyScheduleEvents() {
	DailyScheduleEventsBlock.Lock()
	if DailyScheduleEventsBlock.Block {
		DailyScheduleEventsBlock.Unlock()
		return
	}
	DailyScheduleEventsBlock.Block = true
	DailyScheduleEventsBlock.Unlock()

	t := time.Now()
	Today.Lock()
	if int(Today.Time.Weekday()) == int(t.Weekday()) {
		Today.Unlock()
		DailyScheduleEventsBlock.Lock()
		DailyScheduleEventsBlock.Block = false
		DailyScheduleEventsBlock.Unlock()
		return
	}
	Today.Time = t
	Today.Unlock()

	// Update daily anime schedule
	events.UpdateDailyScheduleWebhooks()
	UpdateAnimeSchedule()
	ResetSubscriptions()

	// Fetch all guild IDs from MongoDB
	guildIDs, err := entities.LoadAllGuildIDs()
	if err != nil {
		log.Printf("Error fetching guild IDs: %v\n", err)
		DailyScheduleEventsBlock.Lock()
		DailyScheduleEventsBlock.Block = false
		DailyScheduleEventsBlock.Unlock()
		return
	}

	DailyScheduleEventsBlock.Lock()
	DailyScheduleEventsBlock.Block = false
	DailyScheduleEventsBlock.Unlock()

	var (
		eg            errgroup.Group
		maxGoroutines = 32
		guard         = make(chan struct{}, maxGoroutines)
	)

	for _, guildID := range guildIDs {
		guard <- struct{}{}
		eg.Go(func(guildID string) func() error {
			return func() error {
				guildIDInt, err := strconv.ParseInt(guildID, 10, 64)
				if err != nil {
					<-guard
					return nil
				}

				// Sends daily schedule via webhook
				events.DailyScheduleWebhooksMap.RLock()
				w, exists := events.DailyScheduleWebhooksMap.WebhooksMap[guildID]
				events.DailyScheduleWebhooksMap.RUnlock()
				if !exists {
					<-guard
					return nil
				}

				s := config.Mgr.SessionForGuild(guildIDInt)
				guildSettings := db.GetGuildSettings(guildID)
				content := getDaySchedule(int(time.Now().Weekday()), guildSettings.GetDonghua())
				content += "\n**Full Week:** <https://AnimeSchedule.net>"

				_, err = s.WebhookExecute(w.ID, w.Token, false, &discordgo.WebhookParams{
					Content: content,
				})
				if err != nil {
					common.LogError(s, guildSettings.BotLog, err)
					<-guard
					return nil
				}

				<-guard
				return nil
			}
		}(guildID))
	}

	err = eg.Wait()
	if err != nil {
		log.Println(err)
	}

	for _, guildID := range guildIDs {
		guildIDInt, err := strconv.ParseInt(guildID, 10, 64)
		if err != nil {
			continue
		}

		events.DailyScheduleWebhooksMap.RLock()
		_, exists := events.DailyScheduleWebhooksMap.WebhooksMap[guildID]
		events.DailyScheduleWebhooksMap.RUnlock()
		if exists {
			continue
		}

		// Wait some milliseconds to prevent hitting the rate limit easily
		time.Sleep(time.Millisecond * 300)

		// Sends daily schedule via message
		DailySchedule(config.Mgr.SessionForGuild(guildIDInt), guildID)
	}
}

func DailyStatsTimer(_ *discordgo.Session, e *discordgo.Ready) {
	// Register slash commands per guild.
	// Used for testing purposes since propagation is faster.
	//for _, guild := range e.Guilds {
	//	for _, v := range SlashCommands {
	//		err := config.Mgr.ApplicationCommandCreate(guild.ID, v)
	//		if err != nil {
	//			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
	//		}
	//	}
	//}
	//log.Println("Slash command registration is done.")

	for range time.NewTicker(15 * time.Second).C {
		dailyScheduleEvents()
	}
}
