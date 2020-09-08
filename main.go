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
	//_ "net/http/pprof"
)

// Initializes and starts Bot
func main() {
	//defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()

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
	entities.Guilds.LoadAll()
	commands.ResetSubscriptions()
	//commands.CleanGuild()

	Start()

	//go func() {
	//	log.Println(http.ListenAndServe(":6969", nil))
	//}()

	<-make(chan struct{})
	return
}

// Starts BOT and its Handlers
func Start() {
	goBot, err := discordgo.New(fmt.Sprintf("Bot %s", config.Token))
	if err != nil {
		log.Println(err)
	}

	// Guild join and leave listener
	goBot.AddHandler(events.GuildCreate)
	goBot.AddHandler(events.GuildDelete)

	// Periodic events and status
	goBot.AddHandler(events.StatusReady)
	goBot.AddHandler(events.CommonEvents)

	// Listens for a role deletion
	goBot.AddHandler(common.ListenForDeletedRoleHandler)

	// Phrase Filter
	goBot.AddHandler(commands.FilterHandler)

	// Message Edit Filter
	goBot.AddHandler(commands.FilterEditHandler)

	// React Filter
	goBot.AddHandler(commands.FilterReactsHandler)

	// Deletes non-whitelisted attachments
	goBot.AddHandler(commands.MessageAttachmentsHandler)

	// React Channel Join Handler
	goBot.AddHandler(commands.ReactJoinHandler)

	// React Channel Remove Handler
	goBot.AddHandler(commands.ReactRemoveHandler)

	// Channel Vote Timer
	goBot.AddHandler(commands.ChannelVoteTimer)

	// MemberInfo
	goBot.AddHandler(events.OnMemberJoinGuild)
	goBot.AddHandler(events.OnMemberUpdate)
	//goBot.AddHandler(events.OnPresenceUpdate)

	// Emoji ChannelStats
	goBot.AddHandler(commands.OnMessageEmoji)
	goBot.AddHandler(commands.OnMessageEmojiReact)
	goBot.AddHandler(commands.OnMessageEmojiUnreact)

	// Channel Stats
	goBot.AddHandler(commands.OnMessageChannel)
	goBot.AddHandler(commands.DailyStatsTimer)

	// Periodic Write Events
	goBot.AddHandler(events.WriteEvents)

	// Voice Role Event Handler
	goBot.AddHandler(events.VoiceRoleHandler)

	// Username stats
	goBot.AddHandler(commands.OnMemberJoin)
	goBot.AddHandler(commands.OnMemberRemoval)

	// Spam filter
	//goBot.AddHandler(commands.SpamFilter)
	//goBot.AddHandler(commands.SpamFilterTimer)

	// Bot fluff
	goBot.AddHandler(events.OnBotPing)

	// Manual ban handler
	goBot.AddHandler(events.OnGuildBan)

	// Abstraction of a command handler
	goBot.AddHandler(commands.HandleCommand)

	// Raffle react handler
	goBot.AddHandler(commands.RaffleReactJoin)
	goBot.AddHandler(commands.RaffleReactLeave)

	// Mute command
	goBot.AddHandler(events.GuildJoin)

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
