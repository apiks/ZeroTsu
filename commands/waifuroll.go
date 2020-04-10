package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"math/rand"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

type waifuOwners struct {
	Name   string `json:"Name"`
	Owners int    `json:"Owners"`
}

// Adds a waifu to the waifu list
func addWaifu(s *discordgo.Session, m *discordgo.Message) {
	var temp entities.Waifu

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildWaifus := db.GetGuildWaifus(m.GuildID)

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"addwaifu [waifu]`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if such a waifu already exists
	for _, waifu := range guildWaifus {
		if waifu.GetName() == strings.ToLower(commandStrings[1]) {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: That waifu already exists.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
	}

	// Adds the waifu to the slice and writes to storage
	temp = temp.SetName(commandStrings[1])
	err := db.SetGuildWaifu(m.GuildID, temp)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Success! Added waifu `"+commandStrings[1]+"` to waifu list.")
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Removes a waifu from the waifu list
func removeWaifu(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"removewaifu [waifu]`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if such a waifu already exists and removes it if so
	err := db.SetGuildWaifu(m.GuildID, entities.NewWaifu(strings.ToLower(commandStrings[1])), true)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed waifu `"+commandStrings[1]+"` from waifu list.")
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
	}
}

// Shows the current list of waifus
func viewWaifus(s *discordgo.Session, m *discordgo.Message) {
	var message string

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildWaifus := db.GetGuildWaifus(m.GuildID)

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"viewwaifus`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	if len(guildWaifus) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no waifus.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Iterates through all the waifus if they exist and adds them to the message string
	for _, waifu := range guildWaifus {
		if message == "" {
			message = "Waifus:\n\n`" + waifu.GetName() + "`"
		} else {
			message += "\n `" + waifu.GetName() + "`"
		}
	}

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
	}
}

// Lets a user roll a waifu from the waifu list and assigns him that waifu
func rollWaifu(s *discordgo.Session, m *discordgo.Message) {
	var (
		waifuLen         int
		randomWaifuIndex int
	)

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildWaifus := db.GetGuildWaifus(m.GuildID)

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"rollwaifu`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	memberInfoUser := db.GetGuildMember(m.GuildID, m.Author.ID)
	if memberInfoUser.GetID() == "" {
		functionality.InitializeUser(m.Author, m.GuildID)
	}
	memberInfoUser = db.GetGuildMember(m.GuildID, m.Author.ID)

	if memberInfoUser.GetWaifu().GetName() != "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: More than one waifu will ruin your laifu.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	waifuLen = len(guildWaifus)
	if waifuLen == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no waifus to roll from.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	randomWaifuIndex = rand.Intn(waifuLen)
	memberInfoUser = memberInfoUser.SetWaifu(*guildWaifus[randomWaifuIndex])
	db.SetGuildMember(m.GuildID, memberInfoUser)

	_, err := s.ChannelMessageSend(m.ChannelID, "Your assigned waifu is "+guildWaifus[randomWaifuIndex].GetName()+"! Congratulations!")
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
	}
}

// Posts what the user's assigned waifu is
func myWaifu(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"mywaifu`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	memberInfoUser := db.GetGuildMember(m.GuildID, m.Author.ID)
	if memberInfoUser.GetID() == "" {
		functionality.InitializeUser(m.Author, m.GuildID)
	}
	memberInfoUser = db.GetGuildMember(m.GuildID, m.Author.ID)

	if memberInfoUser.GetWaifu().GetName() == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: You don't have a waifu. Please roll one with `"+guildSettings.GetPrefix()+"rollwaifu`!")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	waifuName := memberInfoUser.GetWaifu().GetName()

	_, err := s.ChannelMessageSend(m.ChannelID, "Your waifu is "+waifuName+"! I hope you two live a happy life!")
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
	}
}

// Starts a waifu trade with another user
func tradeWaifu(s *discordgo.Session, m *discordgo.Message) {
	var temp entities.WaifuTrade

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildWaifuTrades := db.GetGuildWaifuTrades(m.GuildID)

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"tradewaifu [@user, userID or username#discrim]`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings
	userID, err := common.GetUserID(m, commandStrings)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	memberInfoUser := db.GetGuildMember(m.GuildID, m.Author.ID)
	if memberInfoUser.GetID() == "" {
		functionality.InitializeUser(m.Author, m.GuildID)
	}
	memberInfoUser = db.GetGuildMember(m.GuildID, m.Author.ID)

	if memberInfoUser.GetWaifu().GetName() == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: You don't have a waifu. Please roll one with `"+guildSettings.GetPrefix()+"rollwaifu` before initiating a trade!")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	memberInfoUserTarget := db.GetGuildMember(m.GuildID, m.Author.ID)
	if memberInfoUserTarget.GetID() == "" {
		member, err := s.State.Member(m.GuildID, m.Author.ID)
		if err != nil {
			if member, err = s.GuildMember(m.GuildID, m.Author.ID); err != nil {
				_, err := s.ChannelMessageSend(m.ChannelID, "Error: Cannot find the target user in both the server and internal database. Please have them join the server.")
				if err != nil {
					common.LogError(s, guildSettings.BotLog, err)
					return
				}
				return
			}
		}
		functionality.InitializeUser(member.User, m.GuildID)
	}
	memberInfoUserTarget = db.GetGuildMember(m.GuildID, m.Author.ID)

	if memberInfoUserTarget.GetWaifu().GetName() == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Target user doesn't have a waifu. Please wait for them to roll for one before initiating a trade!")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	temp.SetTradeID(strconv.Itoa(len(guildWaifuTrades)))
	temp.SetAccepteeID(userID)
	temp.SetInitiatorID(m.Author.ID)

	err = db.SetGuildWaifuTrade(m.GuildID, &temp)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Waifu trade with ID %v initiated between <@%v> with waifu %v and <@%v> with waifu %v. <@%v> if you agree to "+
		"exchange your waifu for his, please type `%vaccepttrade %v`! To cancel the trade by either side type `%vcanceltrade %v`", temp.GetTradeID(), temp.GetInitiatorID(), memberInfoUser.GetWaifu().GetName(), userID, memberInfoUserTarget.GetWaifu().GetName(), userID, guildSettings.GetPrefix(), temp.GetTradeID(), guildSettings.GetPrefix(), temp.GetTradeID()))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Accepts a proposed waifu trade with another user
func acceptTrade(s *discordgo.Session, m *discordgo.Message) {
	var flag bool

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildWaifuTrades := db.GetGuildWaifuTrades(m.GuildID)

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"accepttrade [TradeID]`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	if len(guildWaifuTrades) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no ongoing trades.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	for _, trade := range guildWaifuTrades {
		if trade == nil {
			continue
		}

		if trade.GetTradeID() == commandStrings[1] {
			flag = true
			if trade.GetAccepteeID() != m.Author.ID {
				_, err := s.ChannelMessageSend(m.ChannelID, "Error: You don't have permissions to accept this trade.")
				if err != nil {
					common.LogError(s, guildSettings.BotLog, err)
					return
				}
				return
			}

			// Checks whether the two users are in memberinfo or server
			memberInfoUser := db.GetGuildMember(m.GuildID, m.Author.ID)
			if memberInfoUser.GetID() == "" {
				functionality.InitializeUser(m.Author, m.GuildID)
			}
			memberInfoUser = db.GetGuildMember(m.GuildID, m.Author.ID)

			memberInfoUserTarget := db.GetGuildMember(m.GuildID, m.Author.ID)
			if memberInfoUserTarget.GetID() == "" {
				member, err := s.State.Member(m.GuildID, m.Author.ID)
				if err != nil {
					if member, err = s.GuildMember(m.GuildID, m.Author.ID); err != nil {
						_, err := s.ChannelMessageSend(m.ChannelID, "Error: Cannot find the target user in both the server and internal database. Please have them join the server.")
						if err != nil {
							common.LogError(s, guildSettings.BotLog, err)
							return
						}
						return
					}
				}
				functionality.InitializeUser(member.User, m.GuildID)
			}
			memberInfoUserTarget = db.GetGuildMember(m.GuildID, m.Author.ID)

			// Trades waifus by switching them around and removes the trade
			accepteeWaifu := memberInfoUser.GetWaifu()
			memberInfoUser = memberInfoUser.SetWaifu(memberInfoUserTarget.GetWaifu())
			memberInfoUserTarget = memberInfoUserTarget.SetWaifu(accepteeWaifu)

			db.SetGuildMember(m.GuildID, memberInfoUser)
			db.SetGuildMember(m.GuildID, memberInfoUserTarget)
			err := db.SetGuildWaifuTrade(m.GuildID, trade, true)
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				break
			}
			break
		}
	}

	if !flag {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such trade exists.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Success! Traded waifus in trade with ID "+commandStrings[1])
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
	}
}

// Cancels a proposed waifu trade with another user
func cancelTrade(s *discordgo.Session, m *discordgo.Message) {
	var flag bool

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildWaifuTrades := db.GetGuildWaifuTrades(m.GuildID)

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"canceltrade [TradeID]`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	if len(guildWaifuTrades) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no ongoing trades.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	for _, trade := range guildWaifuTrades {
		if trade == nil {
			continue
		}

		if trade.GetTradeID() == commandStrings[1] {
			flag = true

			if trade.GetAccepteeID() != m.Author.ID &&
				trade.GetInitiatorID() != m.Author.ID {
				_, err := s.ChannelMessageSend(m.ChannelID, "Error: You are not someone who has permissions to cancel this trade.")
				if err != nil {
					common.LogError(s, guildSettings.BotLog, err)
					return
				}
				return
			}

			err := db.SetGuildWaifuTrade(m.GuildID, trade, true)
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				break
			}
			break
		}
	}

	if !flag {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such trade exists.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Success! Cancelled trade with ID "+commandStrings[1])
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Prints how many users each waifu has
func showOwners(s *discordgo.Session, m *discordgo.Message) {
	var (
		waifuNum int
		message  string
		messages []string
		owners   []waifuOwners
		owner    waifuOwners
	)

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildWaifus := db.GetGuildWaifus(m.GuildID)
	guildMemberInfo := db.GetGuildMemberInfo(m.GuildID)

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"owners`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if there are any waifus
	if len(guildWaifus) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no waifus.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Iterates through each waifu and member, increasing the waifuNum each time it detects a user with that waifu, and saves it to the messsage
	for _, waifu := range guildWaifus {
		for _, member := range guildMemberInfo {
			if member.GetWaifu().GetName() == waifu.GetName() {
				waifuNum++
			}
		}
		owner.Name = waifu.GetName()
		owner.Owners = waifuNum
		owners = append(owners, owner)
		waifuNum = 0
	}

	// Sorts by number of owners and add to message string
	sort.Sort(byOwnerFrequency(owners))
	for _, trueOwner := range owners {
		message += "\n" + trueOwner.Name + " has " + strconv.Itoa(trueOwner.Owners) + " owners"
	}

	// Splits the message string if over 1900 characters
	messages = common.SplitLongMessage(message)

	// Sends the message
	for _, message := range messages {
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
	}
}

// Sort functions for waifu owners use by owner number
type byOwnerFrequency []waifuOwners

func (e byOwnerFrequency) Len() int {
	return len(e)
}
func (e byOwnerFrequency) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
func (e byOwnerFrequency) Less(i, j int) bool {
	return e[j].Owners < e[i].Owners
}

func init() {
	Add(&Command{
		Execute:    addWaifu,
		Trigger:    "addwaifu",
		Aliases:    []string{"addwife"},
		Desc:       "Adds a waifu to the waifu list [WAIFU]",
		Permission: functionality.Mod,
		Module:     "waifus",
	})
	Add(&Command{
		Execute:    removeWaifu,
		Trigger:    "removewaifu",
		Aliases:    []string{"removewife", "deletewaifu", "deletewife"},
		Desc:       "Removes a waifu from the waifu list [WAIFU]",
		Permission: functionality.Mod,
		Module:     "waifus",
	})
	Add(&Command{
		Execute:    viewWaifus,
		Trigger:    "viewwaifus",
		Aliases:    []string{"showwaifus", "vwaifus", "waifulist", "listwaifu", "waifus"},
		Desc:       "Shows the current list of waifus [WAIFU]",
		Permission: functionality.Mod,
		Module:     "waifus",
	})
	Add(&Command{
		Execute: rollWaifu,
		Trigger: "rollwaifu",
		Aliases: []string{"rollwife", "wiferoll", "waifuroll"},
		Desc:    "Rolls a random waifu [WAIFU]",
		Module:  "waifus",
	})
	Add(&Command{
		Execute: myWaifu,
		Trigger: "waifu",
		Aliases: []string{"mywaifu", "mywife", "showwaifu", "waifushow", "viewwaifu", "viewwife"},
		Desc:    "Shows what your assigned waifu is [WAIFU]",
		Module:  "waifus",
	})
	Add(&Command{
		Execute: tradeWaifu,
		Trigger: "tradewaifu",
		Aliases: []string{"tradewife", "sellwaifu", "starttrade", "tradestart"},
		Desc:    "Trades two waifus between two users if both agree [WAIFU]",
		Module:  "waifus",
	})
	Add(&Command{
		Execute: acceptTrade,
		Trigger: "accepttrade",
		Aliases: []string{"tradeaccept", "buywaifu"},
		Desc:    "Accepts a proposed waifu trade [WAIFU]",
		Module:  "waifus",
	})
	Add(&Command{
		Execute: cancelTrade,
		Trigger: "canceltrade",
		Aliases: []string{"tradecancel", "stoptrade", "tradestop"},
		Desc:    "Cancels a proposed waifu trade [WAIFU]",
		Module:  "waifus",
	})
	Add(&Command{
		Execute:    showOwners,
		Trigger:    "owners",
		Aliases:    []string{"showowners", "viewowners", "tradestop"},
		Desc:       "Prints all waifus and how many owners they have [WAIFU]",
		Permission: functionality.Mod,
		Module:     "waifus",
	})
}
