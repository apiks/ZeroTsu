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

	//Adds Ping Command (Converted)
	// goBot.AddHandler(commands.PingHandler)

	//Adds Channel Creation Command
	goBot.AddHandler(commands.CreateChannelHandler)

	//Adds Help Command
	goBot.AddHandler(commands.HelpHandler)

	//Updates Playing Status
	goBot.AddHandler(misc.StatusReady)

	//Listens for a role deletion
	goBot.AddHandler(misc.ListenForDeletedRoleHandler)

	//Sorts spoiler roles alphabetically between opt-in dummy roles
	goBot.AddHandler(commands.SortRolesHandler)

	//Sorts a category alphabetically
	goBot.AddHandler(commands.SortCategoryHandler)

	//Join Channel Command
	goBot.AddHandler(commands.JoinChannelHandler)

	//Leave Channel Command
	goBot.AddHandler(commands.LeaveChannelHandler)

	//Word Filter
	goBot.AddHandler(commands.FilterHandler)

	//React Filter
	goBot.AddHandler(commands.FilterReacts)

	//Say Command (Converted)
	//goBot.AddHandler(commands.SayHandler)

	//Deletes non-whitelisted attachments
	goBot.AddHandler(commands.MessageAttachmentsHandler)

	//Avatar Command
	goBot.AddHandler(commands.AvatarHandler)

	//Channel Lock/Unlock Command
	goBot.AddHandler(commands.ChannelLockHandler)

	//About Command (Converted)
	// goBot.AddHandler(commands.AboutHandler)

	// Abstraction of a command handler
	goBot.AddHandler(commands.HandleCommand)

	//MemberInfo
	//goBot.AddHandler(misc.OnMemberJoinGuild)
	//goBot.AddHandler(misc.OnMemberUpdate)

	//Whois Command
	//goBot.AddHandler(commands.WhoisHandler)

	//Kick Command
	//goBot.AddHandler(commands.KickHandler)

	//Ban Command
	//goBot.AddHandler(commands.BanHandler)

	//Unban Command
	//goBot.AddHandler(commands.UnbanHandler)

	//AddWarning Command
	//goBot.AddHandler(commands.AddWarningHandler)

	//IssueWarning Command
	//goBot.AddHandler(commands.IssueWarningHandler)

	//Converter
	//goBot.AddHandler(commands.ConverterHandler)

	//React Channel Set Command
	goBot.AddHandler(commands.SetReactChannelHandler)

	//React Channel Join Command
	goBot.AddHandler(commands.ReactJoinHandler)

	//React Channel Remove Command
	goBot.AddHandler(commands.ReactRemoveHandler)

	//React Channel Join View Command
	goBot.AddHandler(commands.ViewSetReactJoinsHandler)

	//React Channel Join Remove Command
	goBot.AddHandler(commands.RemoveReactJoinHandler)

	//RSS Parse Command
	goBot.AddHandler(commands.RSSHandler)

	//RSS Thread Check
	goBot.AddHandler(misc.RssThreadReady)

	//Channel Vote Command
	goBot.AddHandler(commands.ChannelVoteHandler)

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
