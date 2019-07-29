package commands

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
	"github.com/r-anime/ZeroTsu/config"
)

type waifuOwners struct {
	Name 	string 		`json:"Name"`
	Owners 	int 		`json:"Owners"`

}

// Adds a waifu to the waifu list
func addWaifu(s *discordgo.Session, m *discordgo.Message) {

	var temp misc.Waifu

	commandStrings := strings.SplitN(m.Content, " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "addwaifu [waifu]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Checks if such a waifu already exists
	misc.MapMutex.Lock()
	for _, waifu := range misc.GuildMap[m.GuildID].Waifus {
		if waifu.Name == strings.ToLower(commandStrings[1]) {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: That waifu already exists.")
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
				if err != nil {
					misc.MapMutex.Unlock()
					return
				}
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
	}

	// Adds the waifu to the slice and writes to storage
	temp.Name = strings.ToLower(commandStrings[1])
	misc.GuildMap[m.GuildID].Waifus = append(misc.GuildMap[m.GuildID].Waifus, temp)
	err := misc.WaifusWrite(misc.GuildMap[m.GuildID].Waifus, m.GuildID)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, "Success! Added waifu `" + commandStrings[1] + "` to waifu list.")
	if err != nil {
		_, _ = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
	}
}

// Removes a waifu from the waifu list
func removeWaifu(s *discordgo.Session, m *discordgo.Message) {

	commandStrings := strings.SplitN(m.Content, " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "removewaifu [waifu]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Checks if such a waifu already exists and removes it if so
	misc.MapMutex.Lock()
	for i, waifu := range misc.GuildMap[m.GuildID].Waifus {
		if waifu.Name == strings.ToLower(commandStrings[1]) {
			misc.GuildMap[m.GuildID].Waifus = append(misc.GuildMap[m.GuildID].Waifus[:i], misc.GuildMap[m.GuildID].Waifus[i+1:]...)
			err := misc.WaifusWrite(misc.GuildMap[m.GuildID].Waifus, m.GuildID)
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
				if err != nil {
					misc.MapMutex.Unlock()
					return
				}
				misc.MapMutex.Unlock()
				return
			}
			_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed waifu `" + commandStrings[1] + "` from waifu list.")
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
				if err != nil {
					misc.MapMutex.Unlock()
					return
				}
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
	}
	misc.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such waifu found.")
	if err != nil {
		_, _ = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
	}
}

// Shows the current list of waifus
func viewWaifus(s *discordgo.Session, m *discordgo.Message) {
	var message string

	commandStrings := strings.Split(m.Content, " ")

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "viewwaifus`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	misc.MapMutex.Lock()
	if len(misc.GuildMap[m.GuildID].Waifus) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no waifus.")
		if err != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}

	// Iterates through all the waifus if they exist and adds them to the message string
	for _, waifu := range misc.GuildMap[m.GuildID].Waifus {
		if message == "" {
			message = "Waifus:\n\n`" + waifu.Name + "`"
		} else {
			message += "\n `" + waifu.Name + "`"
		}
	}
	misc.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		_, _ = s.ChannelMessageSend(config.BotLogID, err.Error())
	}
}

// Lets a user roll a waifu from the waifu list and assigns him that waifu
func rollWaifu(s *discordgo.Session, m *discordgo.Message) {

	var (
		waifuLen 			int
		randomWaifuIndex 	int
		waifuRoll 			misc.Waifu
	)

	commandStrings := strings.Split(m.Content, " ")

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "rollwaifu`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	misc.MapMutex.Lock()
	_, ok := misc.GuildMap[m.GuildID].MemberInfoMap[m.Author.ID];if !ok {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: You are not in memberInfo and therefore not allowed to roll a waifu. Please notify a mod.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	if misc.GuildMap[m.GuildID].MemberInfoMap[m.Author.ID].Waifu.Name != "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: More than one waifu will ruin your laifu.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}

	waifuLen = len(misc.GuildMap[m.GuildID].Waifus)
	randomWaifuIndex = rand.Intn(waifuLen)
	waifuRoll = misc.GuildMap[m.GuildID].Waifus[randomWaifuIndex]
	misc.GuildMap[m.GuildID].MemberInfoMap[m.Author.ID].Waifu = waifuRoll
	misc.WriteMemberInfo(misc.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	misc.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, "Your assigned waifu is " + waifuRoll.Name + "! Congratulations!")
	if err != nil {
		_, _ = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
	}
}

// Posts what the user's assigned waifu is
func myWaifu(s *discordgo.Session, m *discordgo.Message) {

	commandStrings := strings.Split(m.Content, " ")

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "mywaifu`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	misc.MapMutex.Lock()
	_, ok := misc.GuildMap[m.GuildID].MemberInfoMap[m.Author.ID];if !ok {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: You are not in memberInfo and therefore not allowed to roll a waifu. Please notify a mod.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	if misc.GuildMap[m.GuildID].MemberInfoMap[m.Author.ID].Waifu.Name == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: You don't have a waifu. Please roll one with `" + config.BotPrefix + "rollwaifu`!")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	_, err := s.ChannelMessageSend(m.ChannelID, "Your waifu is " + misc.GuildMap[m.GuildID].MemberInfoMap[m.Author.ID].Waifu.Name +"! I hope you two live a happy life!")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()
}

// Starts a waifu trade with another user
func tradeWaifu(s *discordgo.Session, m *discordgo.Message) {

	var temp misc.WaifuTrade

	commandStrings := strings.Split(m.Content, " ")

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+config.BotPrefix+"tradewaifu [@user, userID or username#discrim]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings
	userID, err := misc.GetUserID(s, m, commandStrings)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	misc.MapMutex.Lock()
	_, ok := misc.GuildMap[m.GuildID].MemberInfoMap[m.Author.ID];if !ok {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: You are not in memberInfo and therefore not allowed to roll a waifu. Please notify a mod.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	if misc.GuildMap[m.GuildID].MemberInfoMap[m.Author.ID].Waifu.Name == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: You don't have a waifu. Please roll one with `" + config.BotPrefix + "rollwaifu` before initiating a trade!")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}

	_, ok = misc.GuildMap[m.GuildID].MemberInfoMap[userID];if !ok {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Target user is not in memberInfo and therefore not allowed to roll have a waifu. Please notify a mod.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	if misc.GuildMap[m.GuildID].MemberInfoMap[userID].Waifu.Name == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Target user doesn't have a waifu. Please wait for them to roll for one before initiating a trade!")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}

	temp.TradeID = strconv.Itoa(len(misc.GuildMap[m.GuildID].WaifuTrades))
	temp.AccepteeID = userID
	temp.InitiatorID = m.Author.ID
	misc.GuildMap[m.GuildID].WaifuTrades = append(misc.GuildMap[m.GuildID].WaifuTrades, temp)
	err = misc.WaifuTradesWrite(misc.GuildMap[m.GuildID].WaifuTrades, m.GuildID)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Waifu trade with ID %v initiated between <@%v> with waifu %v and <@%v> with waifu %v. <@%v> if you agree to " +
		"exchange your waifu for his, please type `%vaccepttrade %v`! To cancel the trade by either side type `%vcanceltrade %v`", temp.TradeID, temp.InitiatorID, misc.GuildMap[m.GuildID].MemberInfoMap[temp.InitiatorID].Waifu.Name, userID, misc.GuildMap[m.GuildID].MemberInfoMap[userID].Waifu.Name, userID, config.BotPrefix, temp.TradeID, config.BotPrefix , temp.TradeID))
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()
}

// Accepts a proposed waifu trade with another user
func acceptTrade(s *discordgo.Session, m *discordgo.Message) {

	var flag bool

	commandStrings := strings.Split(m.Content, " ")

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+config.BotPrefix+"accepttrade [TradeID]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	misc.MapMutex.Lock()
	if len(misc.GuildMap[m.GuildID].WaifuTrades) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no ongoing trades.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	for i, trade := range misc.GuildMap[m.GuildID].WaifuTrades {
		if trade.TradeID == commandStrings[1] {
			flag = true
			if trade.AccepteeID != m.Author.ID {
				_, err := s.ChannelMessageSend(m.ChannelID, "Error: You are not the one who has permissions to accept this trade.")
				if err != nil {
					_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
					if err != nil {
						misc.MapMutex.Unlock()
						return
					}
					misc.MapMutex.Unlock()
					return
				}
				misc.MapMutex.Unlock()
				return
			}

			// Trades waifus by switching them around and removes the trade
			accepteeWaifu := misc.GuildMap[m.GuildID].MemberInfoMap[m.Author.ID].Waifu
			misc.GuildMap[m.GuildID].MemberInfoMap[m.Author.ID].Waifu = misc.GuildMap[m.GuildID].MemberInfoMap[trade.InitiatorID].Waifu
			misc.GuildMap[m.GuildID].MemberInfoMap[trade.InitiatorID].Waifu = accepteeWaifu
			misc.WriteMemberInfo(misc.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)

			misc.GuildMap[m.GuildID].WaifuTrades = append(misc.GuildMap[m.GuildID].WaifuTrades[:i], misc.GuildMap[m.GuildID].WaifuTrades[i+1:]...)
			err := misc.WaifuTradesWrite(misc.GuildMap[m.GuildID].WaifuTrades, m.GuildID)
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
				if err != nil {
					misc.MapMutex.Unlock()
					return
				}
				misc.MapMutex.Unlock()
				return
			}
			break
		}
	}
	misc.MapMutex.Unlock()

	if !flag {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such trade exists.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Success! Traded waifus in trade with ID " + commandStrings[1])
	if err != nil {
		_, _ = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
	}
}

// Cancels a proposed waifu trade with another user
func cancelTrade(s *discordgo.Session, m *discordgo.Message) {

	var flag bool

	commandStrings := strings.Split(m.Content, " ")

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+config.BotPrefix+"canceltrade [TradeID]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	misc.MapMutex.Lock()
	if len(misc.GuildMap[m.GuildID].WaifuTrades) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no ongoing trades.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	for i, trade := range misc.GuildMap[m.GuildID].WaifuTrades {
		if trade.TradeID == commandStrings[1] {
			flag = true
			if trade.AccepteeID != m.Author.ID &&
				trade.InitiatorID != m.Author.ID {
				_, err := s.ChannelMessageSend(m.ChannelID, "Error: You are not someone who has permissions to cancel this trade.")
				if err != nil {
					_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
					if err != nil {
						misc.MapMutex.Unlock()
						return
					}
					misc.MapMutex.Unlock()
					return
				}
				misc.MapMutex.Unlock()
				return
			}

			// Removes a trade
			misc.GuildMap[m.GuildID].WaifuTrades = append(misc.GuildMap[m.GuildID].WaifuTrades[:i], misc.GuildMap[m.GuildID].WaifuTrades[i+1:]...)
			err := misc.WaifuTradesWrite(misc.GuildMap[m.GuildID].WaifuTrades, m.GuildID)
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
				if err != nil {
					misc.MapMutex.Unlock()
					return
				}
				misc.MapMutex.Unlock()
				return
			}
			break
		}
	}
	misc.MapMutex.Unlock()

	if !flag {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such trade exists.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Success! Cancelled trade with ID " + commandStrings[1])
	if err != nil {
		_, _ = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
	}
}

// Prints how many users each waifu has
func showOwners(s *discordgo.Session, m *discordgo.Message) {
	var (
		waifuNum 	int
		message		string
		messages 	[]string
		owners		[]waifuOwners
		owner		waifuOwners
	)

	commandStrings := strings.Split(m.Content, " ")

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+config.BotPrefix+"owners`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Iterates through each waifu and member, increasing the waifuNum each time it detects a user with that waifu, and saves it to the messsage
	misc.MapMutex.Lock()
	for _, waifu := range misc.GuildMap[m.GuildID].Waifus {
		for _, member := range misc.GuildMap[m.GuildID].MemberInfoMap {
			if member.Waifu.Name == waifu.Name {
				waifuNum++
			}
		}
		owner.Name = waifu.Name
		owner.Owners = waifuNum
		owners = append(owners, owner)
		waifuNum = 0
	}
	misc.MapMutex.Unlock()

	// Sorts by number of owners and add to message string
	sort.Sort(byOwnerFrequency(owners))
	for _, trueOwner := range owners {
		message += "\n" + trueOwner.Name + " has " + strconv.Itoa(trueOwner.Owners) + " owners"
	}

	// Splits the message string if over 1900 characters
	messages = misc.SplitLongMessage(message)

	// Sends the message
	for _, message := range messages {
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			_, _ = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
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
	add(&command{
		execute:  addWaifu,
		trigger:  "addwaifu",
		aliases:  []string{"addwife"},
		desc:     "Adds a waifu to the waifu list.",
		elevated: true,
		category: "waifus",
	})
	add(&command{
		execute:  removeWaifu,
		trigger:  "removewaifu",
		aliases:  []string{"removewife", "deletewaifu", "deletewife"},
		desc:     "Removes a waifu from the waifu list.",
		elevated: true,
		category: "waifus",
	})
	add(&command{
		execute:  viewWaifus,
		trigger:  "viewwaifus",
		aliases:  []string{"showwaifus", "vwaifus", "waifulist", "listwaifu", "waifus"},
		desc:     "Shows the current list of waifus.",
		elevated: true,
		category: "waifus",
	})
	add(&command{
		execute:  rollWaifu,
		trigger:  "rollwaifu",
		aliases:  []string{"rollwife", "wiferoll", "waifuroll"},
		desc:     "Rolls a random waifu.",
		category: "waifus",
	})
	add(&command{
		execute:  myWaifu,
		trigger:  "waifu",
		aliases:  []string{"mywaifu", "mywife", "showwaifu", "waifushow", "viewwaifu", "viewwife"},
		desc:     "Shows what your assigned waifu is.",
		category: "waifus",
	})
	add(&command{
		execute:  tradeWaifu,
		trigger:  "tradewaifu",
		aliases:  []string{"tradewife", "sellwaifu", "starttrade", "tradestart"},
		desc:     "Trades two waifus between two users if both agree.",
		category: "waifus",
	})
	add(&command{
		execute:  acceptTrade,
		trigger:  "accepttrade",
		aliases:  []string{"tradeaccept", "buywaifu"},
		desc:     "Accepts a proposed waifu trade.",
		category: "waifus",
	})
	add(&command{
		execute:  cancelTrade,
		trigger:  "canceltrade",
		aliases:  []string{"tradecancel", "stoptrade", "tradestop"},
		desc:     "Cancels a proposed waifu trade.",
		category: "waifus",
	})
	add(&command{
		execute:  showOwners,
		trigger:  "owners",
		aliases:  []string{"showowners", "viewowners", "tradestop"},
		desc:     "Prints all waifus and how many owners they have.",
		elevated: true,
		category: "waifus",
	})
}