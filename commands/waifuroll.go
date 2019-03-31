package commands

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
	"github.com/r-anime/ZeroTsu/config"
)

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
	for _, waifu := range misc.WaifuSlice {
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
	misc.WaifuSlice = append(misc.WaifuSlice, temp)
	err := misc.WaifusWrite(misc.WaifuSlice)
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
	for i, waifu := range misc.WaifuSlice {
		if waifu.Name == strings.ToLower(commandStrings[1]) {
			misc.WaifuSlice = append(misc.WaifuSlice[:i], misc.WaifuSlice[i+1:]...)
			err := misc.WaifusWrite(misc.WaifuSlice)
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
	if len(misc.WaifuSlice) == 0 {
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
	for _, waifu := range misc.WaifuSlice {
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
	_, ok := misc.MemberInfoMap[m.Author.ID];if !ok {
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
	if misc.MemberInfoMap[m.Author.ID].Waifu.Name != "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: You already have a waifu. Please love her more!")
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

	waifuLen = len(misc.WaifuSlice)
	randomWaifuIndex = rand.Intn(waifuLen)
	waifuRoll = misc.WaifuSlice[randomWaifuIndex]
	misc.MemberInfoMap[m.Author.ID].Waifu = waifuRoll
	misc.MemberInfoWrite(misc.MemberInfoMap)
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
	_, ok := misc.MemberInfoMap[m.Author.ID];if !ok {
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
	if misc.MemberInfoMap[m.Author.ID].Waifu.Name == "" {
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
	_, err := s.ChannelMessageSend(m.ChannelID, "Your waifu is " + misc.MemberInfoMap[m.Author.ID].Waifu.Name +"! I hope you two live a happy life!")
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
	_, ok := misc.MemberInfoMap[m.Author.ID];if !ok {
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
	if misc.MemberInfoMap[m.Author.ID].Waifu.Name == "" {
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

	_, ok = misc.MemberInfoMap[userID];if !ok {
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
	if misc.MemberInfoMap[userID].Waifu.Name == "" {
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

	temp.TradeID = strconv.Itoa(len(misc.WaifuTradeSlice))
	temp.AccepteeID = userID
	temp.InitiatorID = m.Author.ID
	misc.WaifuTradeSlice = append(misc.WaifuTradeSlice, temp)
	err = misc.WaifuTradesWrite(misc.WaifuTradeSlice)
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
		"exchange your waifu for his, please type `%vaccepttrade %v`! To cancel the trade by either side type `%vcanceltrade %v`", temp.TradeID, temp.InitiatorID, misc.MemberInfoMap[temp.InitiatorID].Waifu.Name, userID, misc.MemberInfoMap[userID].Waifu.Name, userID, config.BotPrefix, temp.TradeID, config.BotPrefix , temp.TradeID))
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
	if len(misc.WaifuTradeSlice) == 0 {
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
	for i, trade := range misc.WaifuTradeSlice {
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
			accepteeWaifu := misc.MemberInfoMap[m.Author.ID].Waifu
			misc.MemberInfoMap[m.Author.ID].Waifu = misc.MemberInfoMap[trade.InitiatorID].Waifu
			misc.MemberInfoMap[trade.InitiatorID].Waifu = accepteeWaifu
			misc.MemberInfoWrite(misc.MemberInfoMap)

			misc.WaifuTradeSlice = append(misc.WaifuTradeSlice[:i], misc.WaifuTradeSlice[i+1:]...)
			err := misc.WaifuTradesWrite(misc.WaifuTradeSlice)
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
	if len(misc.WaifuTradeSlice) == 0 {
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
	for i, trade := range misc.WaifuTradeSlice {
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
			misc.WaifuTradeSlice = append(misc.WaifuTradeSlice[:i], misc.WaifuTradeSlice[i+1:]...)
			err := misc.WaifuTradesWrite(misc.WaifuTradeSlice)
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

func init() {
	add(&command{
		execute:  addWaifu,
		trigger:  "addwaifu",
		aliases:  []string{"addwife"},
		desc:     "Adds a waifu to the waifu list.",
		elevated: true,
		category: "misc",
	})
	add(&command{
		execute:  removeWaifu,
		trigger:  "removewaifu",
		aliases:  []string{"removewife", "deletewaifu", "deletewife"},
		desc:     "Removes a waifu from the waifu list.",
		elevated: true,
		category: "misc",
	})
	add(&command{
		execute:  viewWaifus,
		trigger:  "viewwaifus",
		aliases:  []string{"showwaifus", "vwaifus", "waifulist", "listwaifu", "waifus"},
		desc:     "Shows the current list of waifus.",
		elevated: true,
		category: "misc",
	})
	add(&command{
		execute:  rollWaifu,
		trigger:  "rollwaifu",
		aliases:  []string{"rollwife", "wiferoll", "waifuroll"},
		desc:     "Rolls a random waifu.",
		elevated: true,
		category: "misc",
	})
	add(&command{
		execute:  myWaifu,
		trigger:  "waifu",
		aliases:  []string{"mywaifu", "mywife", "showwaifu", "waifushow", "viewwaifu", "viewwife"},
		desc:     "Shows what your assigned waifu is.",
		elevated: true,
		category: "misc",
	})
	add(&command{
		execute:  tradeWaifu,
		trigger:  "tradewaifu",
		aliases:  []string{"tradewife", "sellwaifu", "starttrade", "tradestart"},
		desc:     "Trades two waifus between two users if both agree.",
		elevated: true,
		category: "misc",
	})
	add(&command{
		execute:  acceptTrade,
		trigger:  "accepttrade",
		aliases:  []string{"tradeaccept", "buywaifu"},
		desc:     "Accepts a proposed waifu trade.",
		elevated: true,
		category: "misc",
	})
	add(&command{
		execute:  cancelTrade,
		trigger:  "canceltrade",
		aliases:  []string{"tradecancel", "stoptrade", "tradestop"},
		desc:     "Cancels a proposed waifu trade.",
		elevated: true,
		category: "misc",
	})
}