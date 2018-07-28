package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/commands"
	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Starts Bot and its Handlers
func Start() {
	goBot, err := discordgo.New("Bot " + config.Token)
	if err != nil {

		fmt.Println(err.Error())
		return
	}

	// Reads spoiler roles database at bot start
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

	// Periodic events and status
	goBot.AddHandler(misc.StatusReady)

	// Listens for a role deletion
	goBot.AddHandler(misc.ListenForDeletedRoleHandler)

	// Phrase Filter Handler
	goBot.AddHandler(commands.FilterHandler)

	// React Filter Handler
	goBot.AddHandler(commands.FilterReactsHandler)

	// Deletes non-whitelisted attachments Handler
	goBot.AddHandler(commands.MessageAttachmentsHandler)

	// Abstraction of a command handler
	goBot.AddHandler(commands.HandleCommand)

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

	err = goBot.Open()
	if err != nil {

		fmt.Println(err.Error())
		return
	}

	fmt.Println("Bot is running!")
}
