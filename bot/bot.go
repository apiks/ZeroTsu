package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/commands"
	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

var (
	BotID string
	goBot *discordgo.Session
)

//Starts Bot and its Handlers
func Start() {
	goBot, err := discordgo.New("Bot " + config.Token)

	if err != nil {

		fmt.Println(err.Error())
		return
	}
	u, err := goBot.User("@me")

	if err != nil {

		fmt.Println(err.Error())
	}

	//Saves bot ID
	BotID = u.ID

	//Reads spoiler roles database at bot start
	misc.SpoilerRolesRead()

	// Reads filters.json from storage at bot start
	misc.FiltersRead()

	// Reads memberInfo.json from storage at bot start
	misc.MemberInfoRead()

	// Reads bannedUsers.json from storage at bot start
	misc.BannedUsersRead()

	// Reads ongoing votes from VoteInfo.json
	commands.VoteInfoRead()

	// Reads set react joins from reactChannelJoin.json
	commands.ReactInfoRead()

	//Reads all the rss threads from rssThreads.json
	misc.RssThreadsRead()

	//Updates Playing Status
	goBot.AddHandler(misc.StatusReady)

	//Listens for a role deletion
	goBot.AddHandler(misc.ListenForDeletedRoleHandler)

	//Phrase Filter
	goBot.AddHandler(commands.FilterHandler)

	//React Filter
	goBot.AddHandler(commands.FilterReactsHandler)

	//Deletes non-whitelisted attachments
	goBot.AddHandler(commands.MessageAttachmentsHandler)

	// Abstraction of a command handler
	goBot.AddHandler(commands.HandleCommand)

	//MemberInfo
	//goBot.AddHandler(misc.OnMemberJoinGuild)
	//goBot.AddHandler(misc.OnMemberUpdate)

	//Whois Command
	//goBot.AddHandler(commands.WhoisHandler)

	//Unban Command
	//goBot.AddHandler(commands.UnbanHandler)

	//React Channel Join Command
	goBot.AddHandler(commands.ReactJoinHandler)

	//React Channel Remove Command
	goBot.AddHandler(commands.ReactRemoveHandler)

	//RSS Thread Check
	goBot.AddHandler(misc.RssThreadReady)

	//Channel Vote Timer
	goBot.AddHandler(commands.ChannelVoteTimer)

	//Verified Role and Cookie Map Expiry Deletion Handler
	//goBot.AddHandler(verification.VerifiedRoleAdd)
	//goBot.AddHandler(verification.VerifiedAlready)

	err = goBot.Open()

	if err != nil {

		fmt.Println(err.Error())
		return
	}

	fmt.Println("Bot is running!")
}
