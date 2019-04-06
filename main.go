package main

import (
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/mux"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/commands"
	"github.com/r-anime/ZeroTsu/web"
	"github.com/r-anime/ZeroTsu/misc"
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

	Start()

	// Web Server
	if config.Website != "" {
		r := mux.NewRouter()
		staticFileHandler := http.StripPrefix("/web/assets", http.FileServer(http.Dir("./web/assets")))
		r.PathPrefix("/web/assets/").Handler(staticFileHandler)
		r.HandleFunc("/", web.HomepageHandler)
		r.HandleFunc("/verification", web.VerificationHandler)
		r.HandleFunc("/verification/", web.VerificationHandler)
		r.HandleFunc("/channelstats", web.StatsPageHandler)
		r.HandleFunc("/channelstats/", web.StatsPageHandler)
		err := http.ListenAndServe(":3000", r)
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
		fmt.Println(err)
	}

	// Reads all spoiler roles created with create command from spoilerRoles.json
	misc.SpoilerRolesRead()

	// Reads filters.json from storage at bot start
	misc.FiltersRead()

	// Reads memberInfo.json from storage at bot start
	misc.MemberInfoRead()

	// Reads all banned users from memberInfo on bot start
	misc.GetBannedUsers()

	// Reads ongoing votes from VoteInfo.json at bot start. Depreciated
	commands.VoteInfoRead()

	// Reads set react joins from reactChannelJoin.json. Disabled in favor of Kaguya by default. Uncomment if not using Kaguya
	//commands.ReactInfoRead()

	// Reads all the rss threads from rssThreads.json
	misc.RssThreadsRead()

	// Reads all the timer rss threads from rssThreadsCheck.json
	misc.RssThreadsTimerRead()

	// Reads all the user created temp channels from userTempCha.json
	commands.TempChaRead()

	// Reads saved emoji stats from emojiStats.json
	misc.EmojiStatsRead()

	// Reads saved channel stats from channelStats.json
	misc.ChannelStatsRead()

	// Reads user gain stats from userGainStats.json
	misc.UserChangeStatsRead()

	// Reads remindMes from remindMe.json
	misc.RemindMeRead()

	// Reads raffles from raffles.json
	misc.RafflesRead()

	// Reads waifus from waifus.json
	if config.Waifus == "true" {
		misc.WaifusRead()

		// Reads waifu trades from waifutrades.json
		misc.WaifuTradesRead()
	}

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

	// React Channel Join Handler. Disabled in favor of Kaguya by default. Uncomment if not using Kaguya
	//goBot.AddHandler(commands.ReactJoinHandler)

	// React Channel Remove Handler. Disabled in favor of Kaguya by default. Uncomment if not using Kaguya
	//goBot.AddHandler(commands.ReactRemoveHandler)

	// Channel Vote Timer
	goBot.AddHandler(commands.ChannelVoteTimer)

	// MemberInfo
	goBot.AddHandler(misc.OnMemberJoinGuild)
	goBot.AddHandler(misc.OnMemberUpdate)
	goBot.AddHandler(misc.OnPresenceUpdate)

	// Verified Role and Cookie Map Expiry Deletion Handler
	goBot.AddHandler(web.VerifiedRoleAdd)
	goBot.AddHandler(web.VerifiedAlready)

	// Emoji Stats
	goBot.AddHandler(commands.OnMessageEmoji)
	goBot.AddHandler(commands.OnMessageEmojiReact)
	goBot.AddHandler(commands.OnMessageEmojiUnreact)

	// Channel Stats
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

	err = goBot.Open()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Bot is running!")
}