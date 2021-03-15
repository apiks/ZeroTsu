package commands

import (
	"github.com/r-anime/ZeroTsu/entities"
	"io/ioutil"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

var Today = time.Now()

// Posts daily stats and update schedule command
func dailyStats(s *discordgo.Session) {
	t := time.Now()

	entities.Mutex.RLock()
	if Today.Day() == t.Day() {
		entities.Mutex.RUnlock()
		return
	}
	entities.Mutex.RUnlock()

	// Update daily anime schedule
	UpdateAnimeSchedule()
	ResetSubscriptions()

	folders, err := ioutil.ReadDir("database/guilds")
	if err != nil {
		log.Panicln(err)
		return
	}

	// Sleeps until anime schedule is definitely updated
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

	entities.Mutex.Lock()
	Today = t
	entities.Mutex.Unlock()
}

// Daily stats and schedule update timer
func DailyStatsTimer(s *discordgo.Session, _ *discordgo.Ready) {
	for range time.NewTicker(5 * time.Minute).C {
		dailyStats(s)
	}
}
