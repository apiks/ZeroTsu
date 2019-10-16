package commands

import (
	"fmt"
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

	var temp functionality.Waifu

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"addwaifu [waifu]`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if such a waifu already exists
	functionality.MapMutex.Lock()
	for _, waifu := range functionality.GuildMap[m.GuildID].Waifus {
		if waifu.Name == strings.ToLower(commandStrings[1]) {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: That waifu already exists.")
			if err != nil {
				functionality.MapMutex.Unlock()
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			functionality.MapMutex.Unlock()
			return
		}
	}

	// Adds the waifu to the slice and writes to storage
	temp.Name = commandStrings[1]
	functionality.GuildMap[m.GuildID].Waifus = append(functionality.GuildMap[m.GuildID].Waifus, &temp)
	err := functionality.WaifusWrite(functionality.GuildMap[m.GuildID].Waifus, m.GuildID)
	if err != nil {
		functionality.MapMutex.Unlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	functionality.MapMutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, "Success! Added waifu `"+commandStrings[1]+"` to waifu list.")
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Removes a waifu from the waifu list
func removeWaifu(s *discordgo.Session, m *discordgo.Message) {

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"removewaifu [waifu]`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if such a waifu already exists and removes it if so
	functionality.MapMutex.Lock()
	for i, waifu := range functionality.GuildMap[m.GuildID].Waifus {
		if strings.ToLower(waifu.Name) == strings.ToLower(commandStrings[1]) {
			functionality.GuildMap[m.GuildID].Waifus = append(functionality.GuildMap[m.GuildID].Waifus[:i], functionality.GuildMap[m.GuildID].Waifus[i+1:]...)
			_ = functionality.WaifusWrite(functionality.GuildMap[m.GuildID].Waifus, m.GuildID)
			_, err := s.ChannelMessageSend(m.ChannelID, "Success! Removed waifu `"+commandStrings[1]+"` from waifu list.")
			if err != nil {
				functionality.MapMutex.Unlock()
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			functionality.MapMutex.Unlock()
			return
		}
	}
	functionality.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such waifu found.")
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Shows the current list of waifus
func viewWaifus(s *discordgo.Session, m *discordgo.Message) {

	var message string

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"viewwaifus`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	functionality.MapMutex.Lock()
	if len(functionality.GuildMap[m.GuildID].Waifus) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no waifus.")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}

	// Iterates through all the waifus if they exist and adds them to the message string
	for _, waifu := range functionality.GuildMap[m.GuildID].Waifus {
		if message == "" {
			message = "Waifus:\n\n`" + waifu.Name + "`"
		} else {
			message += "\n `" + waifu.Name + "`"
		}
	}
	functionality.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Lets a user roll a waifu from the waifu list and assigns him that waifu
func rollWaifu(s *discordgo.Session, m *discordgo.Message) {

	var (
		waifuLen         int
		randomWaifuIndex int
	)

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"rollwaifu`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	functionality.MapMutex.Lock()
	memberInfoUser, ok := functionality.GuildMap[m.GuildID].MemberInfoMap[m.Author.ID]
	if !ok {
		// Fetch user and initialize him
		member, err := s.State.Member(m.GuildID, m.Author.ID)
		if err != nil {
			if member, err = s.GuildMember(m.GuildID, m.Author.ID); err != nil {
				functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
				functionality.MapMutex.Unlock()
				return
			}
		}
		functionality.InitializeMember(member, m.GuildID)
		memberInfoUser = functionality.GuildMap[m.GuildID].MemberInfoMap[m.Author.ID]
	}

	if memberInfoUser.Waifu != nil && memberInfoUser.Waifu.Name == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: More than one waifu will ruin your laifu.")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}

	waifuLen = len(functionality.GuildMap[m.GuildID].Waifus)
	if waifuLen == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no waifus to roll.")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}

	randomWaifuIndex = rand.Intn(waifuLen)
	functionality.GuildMap[m.GuildID].MemberInfoMap[m.Author.ID].Waifu = functionality.GuildMap[m.GuildID].Waifus[randomWaifuIndex]
	_ = functionality.WriteMemberInfo(functionality.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	functionality.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, "Your assigned waifu is "+functionality.GuildMap[m.GuildID].Waifus[randomWaifuIndex].Name+"! Congratulations!")
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Posts what the user's assigned waifu is
func myWaifu(s *discordgo.Session, m *discordgo.Message) {

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"mywaifu`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	functionality.MapMutex.Lock()
	memberInfoUser, ok := functionality.GuildMap[m.GuildID].MemberInfoMap[m.Author.ID]
	if !ok {
		// Fetch user and initialize him
		member, err := s.State.Member(m.GuildID, m.Author.ID)
		if err != nil {
			if member, err = s.GuildMember(m.GuildID, m.Author.ID); err != nil {
				_, err := s.ChannelMessageSend(m.ChannelID, "Error: Cannot find you in both the server and internal database. Please rejoin the server.")
				if err != nil {
					functionality.MapMutex.Unlock()
					functionality.LogError(s, guildSettings.BotLog, err)
					return
				}
				functionality.MapMutex.Unlock()
				return
			}
		}
		functionality.InitializeMember(member, m.GuildID)
		memberInfoUser = functionality.GuildMap[m.GuildID].MemberInfoMap[m.Author.ID]
	}

	if memberInfoUser.Waifu == nil || memberInfoUser.Waifu.Name == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: You don't have a waifu. Please roll one with `"+guildSettings.Prefix+"rollwaifu`!")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			functionality.MapMutex.Unlock()
			return
		}
		functionality.MapMutex.Unlock()
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Your waifu is "+memberInfoUser.Waifu.Name+"! I hope you two live a happy life!")
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		functionality.MapMutex.Unlock()
		return
	}
	functionality.MapMutex.Unlock()
}

// Starts a waifu trade with another user
func tradeWaifu(s *discordgo.Session, m *discordgo.Message) {

	var temp functionality.WaifuTrade

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"tradewaifu [@user, userID or username#discrim]`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings
	userID, err := functionality.GetUserID(m, commandStrings)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	functionality.MapMutex.Lock()
	memberInfoUser, ok := functionality.GuildMap[m.GuildID].MemberInfoMap[m.Author.ID]
	if !ok {
		// Fetch user and initialize him
		member, err := s.State.Member(m.GuildID, m.Author.ID)
		if err != nil {
			if member, err = s.GuildMember(m.GuildID, m.Author.ID); err != nil {
				_, err := s.ChannelMessageSend(m.ChannelID, "Error: Cannot find you in both the server and internal database. Please rejoin the server.")
				if err != nil {
					functionality.MapMutex.Unlock()
					functionality.LogError(s, guildSettings.BotLog, err)
					return
				}
				functionality.MapMutex.Unlock()
				return
			}
		}
		functionality.InitializeMember(member, m.GuildID)
		memberInfoUser = functionality.GuildMap[m.GuildID].MemberInfoMap[m.Author.ID]
	}

	if memberInfoUser.Waifu == nil || memberInfoUser.Waifu.Name == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: You don't have a waifu. Please roll one with `"+guildSettings.Prefix+"rollwaifu` before initiating a trade!")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}

	memberInfoUserTarget, ok2 := functionality.GuildMap[m.GuildID].MemberInfoMap[userID]
	if !ok2 {
		// Fetch user and initialize him
		member, err := s.State.Member(m.GuildID, m.Author.ID)
		if err != nil {
			if member, err = s.GuildMember(m.GuildID, m.Author.ID); err != nil {
				_, err := s.ChannelMessageSend(m.ChannelID, "Error: Cannot find the target user in both the server and internal database. Please have them join the server.")
				if err != nil {
					functionality.MapMutex.Unlock()
					functionality.LogError(s, guildSettings.BotLog, err)
					return
				}
				functionality.MapMutex.Unlock()
				return
			}
		}
		functionality.InitializeMember(member, m.GuildID)
		memberInfoUserTarget = functionality.GuildMap[m.GuildID].MemberInfoMap[userID]
	}

	if memberInfoUserTarget.Waifu == nil || memberInfoUserTarget.Waifu.Name == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Target user doesn't have a waifu. Please wait for them to roll for one before initiating a trade!")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}

	temp.TradeID = strconv.Itoa(len(functionality.GuildMap[m.GuildID].WaifuTrades))
	temp.AccepteeID = userID
	temp.InitiatorID = m.Author.ID
	functionality.GuildMap[m.GuildID].WaifuTrades = append(functionality.GuildMap[m.GuildID].WaifuTrades, &temp)
	err = functionality.WaifuTradesWrite(functionality.GuildMap[m.GuildID].WaifuTrades, m.GuildID)
	if err != nil {
		functionality.MapMutex.Unlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Waifu trade with ID %v initiated between <@%v> with waifu %v and <@%v> with waifu %v. <@%v> if you agree to "+
		"exchange your waifu for his, please type `%vaccepttrade %v`! To cancel the trade by either side type `%vcanceltrade %v`", temp.TradeID, temp.InitiatorID, functionality.GuildMap[m.GuildID].MemberInfoMap[temp.InitiatorID].Waifu.Name, userID, functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Waifu.Name, userID, guildSettings.Prefix, temp.TradeID, guildSettings.Prefix, temp.TradeID))
	if err != nil {
		functionality.MapMutex.Unlock()
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
	functionality.MapMutex.Unlock()
}

// Accepts a proposed waifu trade with another user
func acceptTrade(s *discordgo.Session, m *discordgo.Message) {

	var flag bool

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"accepttrade [TradeID]`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	functionality.MapMutex.Lock()
	if len(functionality.GuildMap[m.GuildID].WaifuTrades) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no ongoing trades.")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}
	for i, trade := range functionality.GuildMap[m.GuildID].WaifuTrades {
		if trade.TradeID == commandStrings[1] {
			flag = true
			if trade.AccepteeID != m.Author.ID {
				_, err := s.ChannelMessageSend(m.ChannelID, "Error: You are not the one who has permissions to accept this trade.")
				if err != nil {
					functionality.MapMutex.Unlock()
					functionality.LogError(s, guildSettings.BotLog, err)
					return
				}
				functionality.MapMutex.Unlock()
				return
			}

			// Checks whether the two users are in memberinfo or server
			if _, ok := functionality.GuildMap[m.GuildID].MemberInfoMap[m.Author.ID]; !ok {
				// Fetch user and initialize him
				member, err := s.State.Member(m.GuildID, m.Author.ID)
				if err != nil {
					if member, err = s.GuildMember(m.GuildID, m.Author.ID); err != nil {
						_, err := s.ChannelMessageSend(m.ChannelID, "Error: Cannot find you in both the server and internal database. Please rejoin the server.")
						if err != nil {
							functionality.MapMutex.Unlock()
							functionality.LogError(s, guildSettings.BotLog, err)
							return
						}
						functionality.MapMutex.Unlock()
						return
					}
				}
				functionality.InitializeMember(member, m.GuildID)
			}
			if _, ok := functionality.GuildMap[m.GuildID].MemberInfoMap[trade.InitiatorID]; !ok {
				// Fetch user and initialize him
				member, err := s.State.Member(m.GuildID, m.Author.ID)
				if err != nil {
					if member, err = s.GuildMember(m.GuildID, m.Author.ID); err != nil {
						_, err := s.ChannelMessageSend(m.ChannelID, "Error: Cannot find the initiator of the trade in both the server and internal database. Please wait until they join the server.")
						if err != nil {
							functionality.MapMutex.Unlock()
							functionality.LogError(s, guildSettings.BotLog, err)
							return
						}
						functionality.MapMutex.Unlock()
						return
					}
				}
				functionality.InitializeMember(member, m.GuildID)
			}

			// Trades waifus by switching them around and removes the trade
			accepteeWaifu := functionality.GuildMap[m.GuildID].MemberInfoMap[m.Author.ID].Waifu
			functionality.GuildMap[m.GuildID].MemberInfoMap[m.Author.ID].Waifu = functionality.GuildMap[m.GuildID].MemberInfoMap[trade.InitiatorID].Waifu
			functionality.GuildMap[m.GuildID].MemberInfoMap[trade.InitiatorID].Waifu = accepteeWaifu
			_ = functionality.WriteMemberInfo(functionality.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)

			functionality.GuildMap[m.GuildID].WaifuTrades = append(functionality.GuildMap[m.GuildID].WaifuTrades[:i], functionality.GuildMap[m.GuildID].WaifuTrades[i+1:]...)
			_ = functionality.WaifuTradesWrite(functionality.GuildMap[m.GuildID].WaifuTrades, m.GuildID)
			break
		}
	}
	functionality.MapMutex.Unlock()

	if !flag {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such trade exists.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Success! Traded waifus in trade with ID "+commandStrings[1])
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Cancels a proposed waifu trade with another user
func cancelTrade(s *discordgo.Session, m *discordgo.Message) {

	var flag bool

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"canceltrade [TradeID]`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	functionality.MapMutex.Lock()
	if len(functionality.GuildMap[m.GuildID].WaifuTrades) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no ongoing trades.")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}
	for i, trade := range functionality.GuildMap[m.GuildID].WaifuTrades {
		if trade.TradeID == commandStrings[1] {
			flag = true
			if trade.AccepteeID != m.Author.ID &&
				trade.InitiatorID != m.Author.ID {
				_, err := s.ChannelMessageSend(m.ChannelID, "Error: You are not someone who has permissions to cancel this trade.")
				if err != nil {
					functionality.MapMutex.Unlock()
					functionality.LogError(s, guildSettings.BotLog, err)
					return
				}
				functionality.MapMutex.Unlock()
				return
			}

			// Removes a trade
			functionality.GuildMap[m.GuildID].WaifuTrades = append(functionality.GuildMap[m.GuildID].WaifuTrades[:i], functionality.GuildMap[m.GuildID].WaifuTrades[i+1:]...)
			_ = functionality.WaifuTradesWrite(functionality.GuildMap[m.GuildID].WaifuTrades, m.GuildID)
			break
		}
	}
	functionality.MapMutex.Unlock()

	if !flag {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such trade exists.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Success! Cancelled trade with ID "+commandStrings[1])
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
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

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"owners`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if there are any waifus
	functionality.MapMutex.Lock()
	if len(functionality.GuildMap[m.GuildID].Waifus) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no waifus.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			functionality.MapMutex.Unlock()
			return
		}
		functionality.MapMutex.Unlock()
		return
	}

	// Iterates through each waifu and member, increasing the waifuNum each time it detects a user with that waifu, and saves it to the messsage
	for _, waifu := range functionality.GuildMap[m.GuildID].Waifus {
		for _, member := range functionality.GuildMap[m.GuildID].MemberInfoMap {
			if member.Waifu.Name == waifu.Name {
				waifuNum++
			}
		}
		owner.Name = waifu.Name
		owner.Owners = waifuNum
		owners = append(owners, owner)
		waifuNum = 0
	}
	functionality.MapMutex.Unlock()

	// Sorts by number of owners and add to message string
	sort.Sort(byOwnerFrequency(owners))
	for _, trueOwner := range owners {
		message += "\n" + trueOwner.Name + " has " + strconv.Itoa(trueOwner.Owners) + " owners"
	}

	// Splits the message string if over 1900 characters
	messages = functionality.SplitLongMessage(message)

	// Sends the message
	for _, message := range messages {
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
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
	functionality.Add(&functionality.Command{
		Execute:    addWaifu,
		Trigger:    "addwaifu",
		Aliases:    []string{"addwife"},
		Desc:       "Adds a waifu to the waifu list [WAIFU]",
		Permission: functionality.Mod,
		Module:     "waifus",
	})
	functionality.Add(&functionality.Command{
		Execute:    removeWaifu,
		Trigger:    "removewaifu",
		Aliases:    []string{"removewife", "deletewaifu", "deletewife"},
		Desc:       "Removes a waifu from the waifu list [WAIFU]",
		Permission: functionality.Mod,
		Module:     "waifus",
	})
	functionality.Add(&functionality.Command{
		Execute:    viewWaifus,
		Trigger:    "viewwaifus",
		Aliases:    []string{"showwaifus", "vwaifus", "waifulist", "listwaifu", "waifus"},
		Desc:       "Shows the current list of waifus [WAIFU]",
		Permission: functionality.Mod,
		Module:     "waifus",
	})
	functionality.Add(&functionality.Command{
		Execute: rollWaifu,
		Trigger: "rollwaifu",
		Aliases: []string{"rollwife", "wiferoll", "waifuroll"},
		Desc:    "Rolls a random waifu [WAIFU]",
		Module:  "waifus",
	})
	functionality.Add(&functionality.Command{
		Execute: myWaifu,
		Trigger: "waifu",
		Aliases: []string{"mywaifu", "mywife", "showwaifu", "waifushow", "viewwaifu", "viewwife"},
		Desc:    "Shows what your assigned waifu is [WAIFU]",
		Module:  "waifus",
	})
	functionality.Add(&functionality.Command{
		Execute: tradeWaifu,
		Trigger: "tradewaifu",
		Aliases: []string{"tradewife", "sellwaifu", "starttrade", "tradestart"},
		Desc:    "Trades two waifus between two users if both agree [WAIFU]",
		Module:  "waifus",
	})
	functionality.Add(&functionality.Command{
		Execute: acceptTrade,
		Trigger: "accepttrade",
		Aliases: []string{"tradeaccept", "buywaifu"},
		Desc:    "Accepts a proposed waifu trade [WAIFU]",
		Module:  "waifus",
	})
	functionality.Add(&functionality.Command{
		Execute: cancelTrade,
		Trigger: "canceltrade",
		Aliases: []string{"tradecancel", "stoptrade", "tradestop"},
		Desc:    "Cancels a proposed waifu trade [WAIFU]",
		Module:  "waifus",
	})
	functionality.Add(&functionality.Command{
		Execute:    showOwners,
		Trigger:    "owners",
		Aliases:    []string{"showowners", "viewowners", "tradestop"},
		Desc:       "Prints all waifus and how many owners they have [WAIFU]",
		Permission: functionality.Mod,
		Module:     "waifus",
	})
}
