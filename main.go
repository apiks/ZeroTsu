package main

import (
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/entities"
	"github.com/r-anime/ZeroTsu/events"
	"log"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/mux"

	"github.com/r-anime/ZeroTsu/commands"
	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/web"
)

// Initializes and starts Bot and website
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
	// Load all guild and shared info
	entities.Mutex.Lock()
	entities.LoadSharedDB()
	//entities.LoadGuilds()
	entities.Mutex.Unlock()
	entities.Guilds.LoadAll()

	Start()

	// Web Server
	if config.Website != "" {
		r := mux.NewRouter()
		staticFileHandler := http.StripPrefix("/web/assets", http.FileServer(http.Dir("./web/assets")))
		r.PathPrefix("/web/assets/").Handler(staticFileHandler)
		r.HandleFunc("/", web.HomepageHandler)
		r.HandleFunc("/verification", web.VerificationHandler)
		r.HandleFunc("/verification/", web.VerificationHandler)
		r.HandleFunc("/channelstats", web.ChannelStatsPageHandler)
		r.HandleFunc("/channelstats/", web.ChannelStatsPageHandler)
		r.HandleFunc("/userchangestats", web.UserChangeStatsPageHandler)
		r.HandleFunc("/userchangestats/", web.UserChangeStatsPageHandler)
		err := http.ListenAndServe(":8080", r)
		if err != nil {
			panic(err)
		}
	}

	<-make(chan struct{})
	return
}

// Starts BOT and its Handlers
func Start() {
	goBot, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		log.Println(err)
	}

	// Guild join and leave listener
	goBot.AddHandler(events.GuildCreate)
	goBot.AddHandler(events.GuildDelete)

	// Updates schedule command print message on load
	commands.UpdateAnimeSchedule()
	commands.ResetSubscriptions()

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
	goBot.AddHandler(events.OnPresenceUpdate)

	// Verified Role and Cookie Map Expiry Deletion Handler
	if config.Website != "" {
		goBot.AddHandler(web.VerifiedRoleAdd)
		goBot.AddHandler(web.VerifiedAlready)
	}

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

	// Auto spambot ban for r/anime
	// Logs each user that joins the server
	// Mute command
	goBot.AddHandler(events.GuildJoin)
	if config.ServerID == "267799767843602452" {
		//goBot.AddHandler(functionality.SpambotJoin)
	}

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

	log.Println("BOT is running!")
}
