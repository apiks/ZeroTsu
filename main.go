package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
	"github.com/r-anime/ZeroTsu/commands"
)

// Initializes and starts Bot and website
func main() {

	err := config.ReadConfig()
	if err != nil {
		panic(err)
	}

	Start()

	// Web Server
	//http.HandleFunc("/", verification.IndexHandler)
	//http.Handle("/verification/", http.StripPrefix("/verification/", http.FileServer(http.Dir("verification"))))
	//err = http.ListenAndServe(":3000", nil)
	//if err != nil {
	//	panic(err)
	//}

	<-make(chan struct{})
	return
}

// Starts Bot and its Handlers
func Start() {
	goBot, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		panic(err)
	}

	// Reads all spoiler roles created with create command from spoilerRoles.json
	misc.SpoilerRolesRead()

	// Reads filters.json from storage at bot start
	misc.FiltersRead()

	// Reads memberInfo.json from storage at bot start
	misc.MemberInfoRead()

	// Reads bannedUsers.json from storage at bot start
	misc.BannedUsersRead()

	// Reads ongoing votes from VoteInfo.json at bot start
	commands.VoteInfoRead()

	// Reads set react joins from reactChannelJoin.json
	commands.ReactInfoRead()

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

	// Reads user gain stats from UserGainStats.json
	misc.UserChangeStatsRead()

	// Periodic events and status
	goBot.AddHandler(misc.StatusReady)

	// Listens for a role deletion
	goBot.AddHandler(misc.ListenForDeletedRoleHandler)

	// Phrase Filter
	goBot.AddHandler(commands.FilterHandler)

	// React Filter
	goBot.AddHandler(commands.FilterReactsHandler)

	// Deletes non-whitelisted attachments
	goBot.AddHandler(commands.MessageAttachmentsHandler)

	// Abstraction of a command handler
	goBot.AddHandler(commands.HandleCommand)

	//Converter
	//goBot.AddHandler(commands.ConverterHandler)

	// React Channel Join Handler
	goBot.AddHandler(commands.ReactJoinHandler)

	// React Channel Remove Handler
	goBot.AddHandler(commands.ReactRemoveHandler)

	// Channel Vote Timer
	goBot.AddHandler(commands.ChannelVoteTimer)

	// MemberInfo
	//goBot.AddHandler(misc.OnMemberJoinGuild)
	//goBot.AddHandler(misc.OnMemberUpdate)

	// Verified Role and Cookie Map Expiry Deletion Handler
	//goBot.AddHandler(verification.VerifiedRoleAdd)
	//goBot.AddHandler(verification.VerifiedAlready)

	// Emoji Tracker
	goBot.AddHandler(commands.OnMessageEmoji)
	goBot.AddHandler(commands.OnMessageEmojiReact)
	goBot.AddHandler(commands.OnMessageEmojiUnreact)

	// Channel Stats
	goBot.AddHandler(commands.OnMessageChannel)

	// Twenty Minute Timer
	goBot.AddHandler(misc.TwentyMinTimer)

	// Voice Role Event Handler
	//goBot.AddHandler(misc.VoiceRoleHandler)

	// User stats
	goBot.AddHandler(commands.OnMemberJoin)
	goBot.AddHandler(commands.OnMemberRemoval)

	// Spam filter
	//goBot.AddHandler(commands.SpamFilter)
	//goBot.AddHandler(commands.SpamFilterTimer)

	err = goBot.Open()
	if err != nil {
		panic(err)
	}

	fmt.Println("Bot is running!")
}