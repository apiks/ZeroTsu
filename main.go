package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/commands"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/entities"
	"github.com/r-anime/ZeroTsu/events"
	"github.com/sasha-s/go-deadlock"
	"github.com/servusdei2018/shards"
)

// Initializes and starts Bot
func main() {
	deadlock.Opts.DeadlockTimeout = 0

	// Initialize Config values
	err := config.ReadConfig()
	if err != nil {
		panic(err)
	}
	err = config.ReadConfigSecrets()
	if err != nil {
		panic(err)
	}

	commands.UpdateAnimeSchedule()

	entities.InitMongoDB("mongodb://localhost:27017")
	entities.EnsureAnimeSubsIndexes()
	entities.EnsureRemindersIndexes()
	entities.EnsureGuildsIndexes()
	entities.EnsureFeedsIndexes()
	entities.EnsureAutopostIndexes()
	entities.EnsureGuildSettingsIndexes()
	entities.EnsureRaffleIndexes()
	entities.EnsureReactJoinIndexes()

	// Enable to migrate from JSON database to MongoDB
	//entities.MigrateGuilds()
	//entities.MigrateReminders()
	//entities.MigrateAnimeSubs()

	commands.ResetSubscriptions()

	Start()

	fmt.Println("[SUCCESS] Bot is now running.  Press CTRL-C to exit.")
	commands.WebhooksMapHandler()
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	entities.CloseMongoDB()

	// Cleanly close down the Manager.
	fmt.Println("[INFO] Stopping shard manager...")
	config.Mgr.Shutdown()
	fmt.Println("[SUCCESS] Shard manager stopped. Bot is shut down.")
}

// Start starts the BOT and its handlers
func Start() {
	var err error

	log.Println("Starting BOT...")
	config.Mgr, err = shards.New(fmt.Sprintf("Bot %s", config.Token))
	if err != nil {
		fmt.Println("[ERROR] Error creating manager,", err)
		return
	}

	config.Mgr.RegisterIntent(discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged + discordgo.IntentGuildMembers))

	// Guild join and leave listener
	config.Mgr.AddHandler(events.GuildCreate)
	config.Mgr.AddHandler(events.GuildDelete)

	// Slash Commands
	config.Mgr.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commands.SlashCommandsHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	// Periodic events and status
	config.Mgr.AddHandler(events.StatusReady)
	config.Mgr.AddHandler(events.CommonEvents)
	config.Mgr.AddHandler(events.WriteEvents)

	// React Channel Join Handler
	config.Mgr.AddHandler(commands.ReactJoinHandler)

	// React Channel Remove Handler
	config.Mgr.AddHandler(commands.ReactRemoveHandler)

	// Voice Role Event Handler
	config.Mgr.AddHandler(events.VoiceRoleHandler)

	// Bot fluff
	config.Mgr.AddHandler(events.OnBotPing)

	// Abstraction of a command handler
	config.Mgr.AddHandler(commands.HandleCommand)

	// Raffle react handler
	config.Mgr.AddHandler(commands.RaffleReactJoinHandler)
	config.Mgr.AddHandler(commands.RaffleReactLeaveHandler)

	// Anime schedule timer
	config.Mgr.AddHandler(commands.ScheduleTimer)

	// Daily Timer
	config.Mgr.AddHandler(commands.DailyStatsTimer)

	// Anime subscription handler
	config.Mgr.AddHandler(commands.AnimeSubsTimer)
	config.Mgr.AddHandler(commands.AnimeSubsWebhookTimer)
	config.Mgr.AddHandler(commands.AnimeSubsWebhooksMapTimer)

	err = config.Mgr.Start()
	if err != nil {
		panic("Critical error: BOT cannot start: " + err.Error())
	}

	// Start tracking uptime from here
	common.StartTime = time.Now()

	// Register Slash Commands
	for _, v := range commands.SlashCommands {
		err := config.Mgr.ApplicationCommandCreate("", v)
		if err != nil {
			log.Panicf("Cannot create '%s' command: %v", v.Name, err)
		}
	}
	log.Println("Slash command registration is done.")
}
