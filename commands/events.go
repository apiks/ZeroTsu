package commands

import (
	"io/ioutil"
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

	folders, err := ioutil.ReadDir("database/guilds")
	if err != nil {
		DailyScheduleEventsBlock.Lock()
		DailyScheduleEventsBlock.Block = false
		DailyScheduleEventsBlock.Unlock()
		log.Panicln(err)
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

	for _, f := range folders {
		if !f.IsDir() {
			continue
		}
		guildID := f.Name()

		guard <- struct{}{}
		eg.Go(func() error {
			guildIDInt, err := strconv.ParseInt(guildID, 10, 64)
			if err != nil {
				<-guard
				return nil
			}

			// Sends daily schedule via webhook
			events.DailyScheduleWebhooksMap.RLock()
			if _, ok := events.DailyScheduleWebhooksMap.WebhooksMap[guildID]; !ok {
				events.DailyScheduleWebhooksMap.RUnlock()
				<-guard
				return nil
			}
			w := events.DailyScheduleWebhooksMap.WebhooksMap[guildID]
			events.DailyScheduleWebhooksMap.RUnlock()

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
		})
	}

	err = eg.Wait()
	if err != nil {
		log.Println(err)
	}

	for _, f := range folders {
		if !f.IsDir() {
			continue
		}
		guildID := f.Name()

		guildIDInt, err := strconv.ParseInt(guildID, 10, 64)
		if err != nil {
			continue
		}

		events.DailyScheduleWebhooksMap.RLock()
		if _, ok := events.DailyScheduleWebhooksMap.WebhooksMap[guildID]; ok {
			events.DailyScheduleWebhooksMap.RUnlock()
			continue
		}
		events.DailyScheduleWebhooksMap.RUnlock()

		// Wait some milliseconds so it doesn't hit the rate limit easily
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
