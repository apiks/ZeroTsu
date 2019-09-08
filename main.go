package main

import (
	"log"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/mux"

	"github.com/r-anime/ZeroTsu/commands"
	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
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
	misc.LoadGuilds()
	misc.LoadSharedDB()

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

	// Guild join and leave
	goBot.AddHandler(misc.GuildCreate)
	goBot.AddHandler(misc.GuildDelete)

	// Reads all banned users from memberInfo on bot start
	misc.GetBannedUsers()

	// Updates schedule command print message on load
	commands.UpdateAnimeSchedule()
	commands.ResetSubscriptions()

	//misc.ReadImages()

	// Cleans up duplicate usernames and nicknames (Run once per cleanup, keep off unless needed)
	//misc.DuplicateUsernamesAndNicknamesCleanup()

	// Fixes users whose usernames/discrims are different from the ones in memberinfo.json. Keep off unless needed
	//goBot.AddHandlerOnce(misc.UsernameCleanup)

	// Periodic events and status
	goBot.AddHandler(misc.StatusReady)

	// Listens for a role deletion
	goBot.AddHandler(misc.ListenForDeletedRoleHandler)

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
	goBot.AddHandler(misc.OnMemberJoinGuild)
	goBot.AddHandler(misc.OnMemberUpdate)
	goBot.AddHandler(misc.OnPresenceUpdate)

	// Verified Role and Cookie Map Expiry Deletion Handler
	goBot.AddHandler(web.VerifiedRoleAdd)
	goBot.AddHandler(web.VerifiedAlready)

	// Emoji ChannelStats
	goBot.AddHandler(commands.OnMessageEmoji)
	goBot.AddHandler(commands.OnMessageEmojiReact)
	goBot.AddHandler(commands.OnMessageEmojiUnreact)

	// Channel ChannelStats
	goBot.AddHandler(commands.OnMessageChannel)
	goBot.AddHandler(commands.DailyStatsTimer)

	// Twenty Minute Timer
	goBot.AddHandler(misc.TwentyMinTimer)

	// Voice Role Event Handler
	goBot.AddHandler(misc.VoiceRoleHandler)

	// User stats
	goBot.AddHandler(commands.OnMemberJoin)
	goBot.AddHandler(commands.OnMemberRemoval)

	// Spam filter
	goBot.AddHandler(commands.SpamFilter)
	goBot.AddHandler(commands.SpamFilterTimer)

	// Bot fluff
	goBot.AddHandler(misc.OnBotPing)

	// Manual ban handler
	goBot.AddHandler(misc.OnGuildBan)

	// Abstraction of a command handler
	goBot.AddHandler(commands.HandleCommand)

	// Raffle react handler
	goBot.AddHandler(commands.RaffleReactJoin)
	goBot.AddHandler(commands.RaffleReactLeave)

	// Logs each user that joins the server
	goBot.AddHandler(misc.GuildJoin)

	// Auto spambot ban
	goBot.AddHandler(misc.SpambotJoin)

	// Anime subscription handler
	goBot.AddHandler(commands.AnimeSubsTimer)

	// Anime schedule timer
	goBot.AddHandler(commands.ScheduleTimer)

	err = goBot.Open()
	if err != nil {
		panic("Critical error: BOT cannot start.")
	}

	// Start tracking uptime from here
	misc.StartTime = time.Now()

	log.Println("BOT is running!")
}
