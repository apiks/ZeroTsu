package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/commands"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/entities"
	"github.com/r-anime/ZeroTsu/events"
	"log"
	"time"
)

// Initializes and starts Bot
func main() {
	// Initialize Config values
	err := config.ReadConfig()
	if err != nil {
		panic(err)
	}
	err = config.ReadConfigSecrets()
	if err != nil {
		panic(err)
	}

	// Load animeschedule data
	commands.UpdateAnimeSchedule()

	// Load all guild and shared info
	entities.Mutex.Lock()
	entities.LoadSharedDB()
	entities.Mutex.Unlock()
	commands.ResetSubscriptions()

	Start()

	<-make(chan struct{})
	return
}

// Starts BOT and its Handlers
func Start() {
	log.Println("Starting BOT...")
	goBot, err := discordgo.New(fmt.Sprintf("Bot %s", config.Token))
	if err != nil {
		log.Fatal(err)
	}
	goBot.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll - discordgo.IntentsGuildPresences)
	goBot.StateEnabled = false

	// Guild join and leave listener
	goBot.AddHandler(events.GuildCreate)
	goBot.AddHandler(events.GuildDelete)

	// Periodic events and status
	goBot.AddHandler(events.StatusReady)
	goBot.AddHandler(events.CommonEvents)
	goBot.AddHandler(events.WriteEvents)

	// React Channel Join Handler
	goBot.AddHandler(commands.ReactJoinHandler)

	// React Channel Remove Handler
	goBot.AddHandler(commands.ReactRemoveHandler)

	//// Channel Stats
	//goBot.AddHandler(commands.OnMessageChannel)
	goBot.AddHandler(commands.DailyStatsTimer)

	// Voice Role Event Handler
	goBot.AddHandler(events.VoiceRoleHandler)

	// Bot fluff
	goBot.AddHandler(events.OnBotPing)

	// Abstraction of a command handler
	goBot.AddHandler(commands.HandleCommand)

	// Raffle react handler
	goBot.AddHandler(commands.RaffleReactJoin)
	goBot.AddHandler(commands.RaffleReactLeave)

	// Anime subscription handler
	goBot.AddHandler(commands.AnimeSubsTimer)

	// Anime schedule timer
	goBot.AddHandler(commands.ScheduleTimer)

	err = goBot.Open()
	if err != nil {
		panic("Critical error: BOT cannot start: " + err.Error())
	}

	// Start tracking uptime from here
	common.StartTime = time.Now()
}
