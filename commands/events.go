package commands

import (
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type SafeTime struct {
	sync.RWMutex
	Time time.Time
}

var Today = &SafeTime{Time: time.Now()}

// dailyEvents is Daily events
func dailyEvents(s *discordgo.Session) {
	t := time.Now()

	Today.RLock()
	if Today.Time.Day() == t.Day() {
		Today.RUnlock()
		return
	}
	Today.RUnlock()

	// Update daily anime schedule
	UpdateAnimeSchedule()
	ResetSubscriptions()

	folders, err := ioutil.ReadDir("database/guilds")
	if err != nil {
		log.Panicln(err)
		return
	}

	// Sleeps until anime schedule is (reasonably) definitely updated
	time.Sleep(10 * time.Second)

	for _, f := range folders {
		if !f.IsDir() {
			continue
		}
		guildID := f.Name()

		// Wait some milliseconds so it doesn't hit the rate limit easily
		time.Sleep(time.Millisecond * 300)

		// Sends daily schedule if need be
		DailySchedule(s, guildID)
	}

	Today.Lock()
	Today.Time = t
	Today.Unlock()
}

// Daily stats and schedule update timer
func DailyStatsTimer(s *discordgo.Session, _ *discordgo.Ready) {
	for range time.NewTicker(5 * time.Minute).C {
		dailyEvents(s)
	}
}
