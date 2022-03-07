package commands

import (
	"io/ioutil"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/r-anime/ZeroTsu/config"

	"github.com/bwmarrin/discordgo"
)

type SafeTime struct {
	sync.RWMutex
	Time time.Time
}

var Today = &SafeTime{Time: time.Now()}

// dailyEvents is Daily events
func dailyEvents() {
	t := time.Now()

	Today.Lock()
	if int(Today.Time.Weekday()) == int(t.Weekday()) {
		Today.Unlock()
		return
	}
	Today.Time = t
	Today.Unlock()

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

		guildIDInt, err := strconv.ParseInt(guildID, 10, 64)
		if err != nil {
			continue
		}

		// Wait some milliseconds so it doesn't hit the rate limit easily
		time.Sleep(time.Millisecond * 300)

		// Sends daily schedule if need be
		DailySchedule(config.Mgr.SessionForGuild(guildIDInt), guildID)
	}
}

func DailyStatsTimer(_ *discordgo.Session, _ *discordgo.Ready) {
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

	for range time.NewTicker(10 * time.Minute).C {
		dailyEvents()
	}
}
