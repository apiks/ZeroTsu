package main

import (
	"log"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/mux"

	"github.com/r-anime/ZeroTsu/commands"
	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/functionality"
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
	functionality.Mutex.Lock()
	functionality.LoadSharedDB()
	functionality.LoadGuilds()
	functionality.Mutex.Unlock()

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
	goBot.AddHandler(functionality.GuildCreate)
	goBot.AddHandler(functionality.GuildDelete)

	// Updates schedule command print message on load
	commands.UpdateAnimeSchedule()
	commands.ResetSubscriptions()

	// Cleans up duplicate usernames and nicknames (Run once per cleanup, keep off unless needed)
	//misc.DuplicateUsernamesAndNicknamesCleanup()

	// Fixes users whose usernames/discrims are different from the ones in memberinfo.json. Keep off unless needed
	//goBot.AddHandlerOnce(misc.UsernameCleanup)

	// Periodic events and status
	goBot.AddHandler(functionality.StatusReady)
	goBot.AddHandler(functionality.CommonEvents)

	// Listens for a role deletion
	goBot.AddHandler(functionality.ListenForDeletedRoleHandler)

	// Phrase Filter
	goBot.AddHandler(commands.FilterHandler)

	// Message Edit Filter
	goBot.AddHandler(commands.FilterEditHandler)

	// React Filter
	goBot.AddHandler(commands.FilterReactsHandler)

	// Deletes non-whitelisted attachments
	goBot.AddHandler(commands.MessageAttachmentsHandler)

	//Converter
	//goBot.AddHandler(commands.ConverterHandler)

	// React Channel Join Handler
	goBot.AddHandler(commands.ReactJoinHandler)

	// React Channel Remove Handler
	goBot.AddHandler(commands.ReactRemoveHandler)

	// Channel Vote Timer
	goBot.AddHandler(commands.ChannelVoteTimer)

	// MemberInfo
	goBot.AddHandler(functionality.OnMemberJoinGuild)
	goBot.AddHandler(functionality.OnMemberUpdate)
	goBot.AddHandler(functionality.OnPresenceUpdate)

	// Verified Role and Cookie Map Expiry Deletion Handler
	goBot.AddHandler(web.VerifiedRoleAdd)
	goBot.AddHandler(web.VerifiedAlready)

	// Emoji ChannelStats
	goBot.AddHandler(commands.OnMessageEmoji)
	goBot.AddHandler(commands.OnMessageEmojiReact)
	goBot.AddHandler(commands.OnMessageEmojiUnreact)

	// Channel Stats
	goBot.AddHandler(commands.OnMessageChannel)
	goBot.AddHandler(commands.DailyStatsTimer)

	// Periodic Write Events
	goBot.AddHandler(functionality.WriteEvents)

	// Voice Role Event Handler
	goBot.AddHandler(functionality.VoiceRoleHandler)

	// User stats
	goBot.AddHandler(commands.OnMemberJoin)
	goBot.AddHandler(commands.OnMemberRemoval)

	// Spam filter
	//goBot.AddHandler(commands.SpamFilter)
	//goBot.AddHandler(commands.SpamFilterTimer)

	// Bot fluff
	goBot.AddHandler(functionality.OnBotPing)

	// Manual ban handler
	goBot.AddHandler(functionality.OnGuildBan)

	// Abstraction of a command handler
	goBot.AddHandler(functionality.HandleCommand)

	// Raffle react handler
	goBot.AddHandler(commands.RaffleReactJoin)
	goBot.AddHandler(commands.RaffleReactLeave)

	// Logs each user that joins the server
	goBot.AddHandler(functionality.GuildJoin)

	// Auto spambot ban for r/anime
	if config.ServerID == "267799767843602452" {
		goBot.AddHandler(functionality.SpambotJoin)
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
	functionality.StartTime = time.Now()

	log.Println("BOT is running!")
}
