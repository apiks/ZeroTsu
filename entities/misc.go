package entities

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	Mutex sync.RWMutex

	SharedInfo *sharedInfo

	sharedFileNames = [...]string{"remindMes.json", "animeSubs.json"}
	AnimeSchedule   = make(map[int][]*ShowAirTime)
)

// Loads global shared DBs
func LoadSharedDB() {
	// Creates missing "database" and "shared" folder if they don't exist
	if _, err := os.Stat("database"); os.IsNotExist(err) {
		err := os.Mkdir("database", 0777)
		if err != nil {
			log.Println(err)
			return
		}
	}
	if _, err := os.Stat("database/shared"); os.IsNotExist(err) {
		err := os.Mkdir("database/shared", 0777)
		if err != nil {
			log.Println(err)
			return
		}
	}

	files, err := IOReadDir("database/shared")
	if err != nil {
		log.Panicln(err)
	}

	SharedInfo = newSharedInfo(make(map[string]*RemindMeSlice), make(map[string][]*ShowSub))

	for _, file := range files {
		LoadSharedDBFile(file)
	}
}

func LoadSharedDBFile(file string) {
	// Reads all the info from the file and puts them in infoByte as bytes
	infoByte, err := ioutil.ReadFile(fmt.Sprintf("database/shared/%s", file))
	if err != nil {
		log.Println(err)
		return
	}

	// Takes the data and puts it into the appropriate field
	switch file {
	case "remindMes.json":
		_ = json.Unmarshal(infoByte, &SharedInfo.RemindMes)
	case "animeSubs.json":
		_ = json.Unmarshal(infoByte, &SharedInfo.AnimeSubs)
	}
}

// Writes to memberInfo.json
func WriteMemberInfo(info map[string]*UserInfo, guildID string) error {

	// Turns info slice into byte ready to be pushed to file
	MarshaledStruct, err := json.MarshalIndent(info, "", "    ")
	if err != nil {
		return err
	}

	// Writes to file
	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%s/memberInfo.json", guildID), MarshaledStruct, 0644)
	if err != nil {
		return err
	}
	return nil
}

// Writes emoji stats to emojiStats.json
func EmojiStatsWrite(emojiStats map[string]*Emoji, guildID string) error {

	// Turns that map into bytes to be ready to written to file
	marshaledStruct, err := json.MarshalIndent(emojiStats, "", "    ")
	if err != nil {
		return err
	}

	// Writes to file
	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/emojiStats.json", guildID), marshaledStruct, 0644)
	if err != nil {
		return err
	}

	return err
}

// Writes channel stats to channelStats.json
func ChannelStatsWrite(channelStats map[string]*Channel, guildID string) (bool, error) {

	// Turns that map into bytes to be ready to written to file
	marshaledStruct, err := json.MarshalIndent(channelStats, "", "    ")
	if err != nil {
		return false, err
	}

	// Writes to file
	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/channelStats.json", guildID), marshaledStruct, 0644)
	if err != nil {
		return false, err
	}

	return false, err
}

// Writes Username Change stats to userChangeStats.json
func UserChangeStatsWrite(userStats map[string]int, guildID string) (bool, error) {

	// Turns that map into bytes to be ready to written to file
	marshaledStruct, err := json.MarshalIndent(userStats, "", "    ")
	if err != nil {
		return false, err
	}

	// Writes to file
	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/userChangeStats.json", guildID), marshaledStruct, 0644)
	if err != nil {
		return false, err
	}

	return false, err
}

// Writes RemindMe notes to remindMes.json
func RemindMeWrite(remindMe map[string]*RemindMeSlice) error {

	// Checks if the user has hit the db limit
	for _, remindMeSlice := range remindMe {
		if remindMeSlice == nil {
			continue
		}

		if remindMeSlice.GetPremium() && len(remindMeSlice.GetRemindMeSlice()) > 299 {
			return fmt.Errorf("Error: You have reached the RemindMe limit (300) for this premium account.")
		} else if !remindMeSlice.GetPremium() && len(remindMeSlice.GetRemindMeSlice()) > 49 {
			return fmt.Errorf("Error: You have reached the RemindMe limit (50) for this account. Please remove some or increase it to 300 by upgrading to a premium user at <https://patreon.com/apiks>")
		}
	}

	// Turns that slice into bytes to be ready to written to file
	marshaledStruct, err := json.MarshalIndent(remindMe, "", "    ")
	if err != nil {
		return err
	}

	// Writes to file
	err = ioutil.WriteFile("database/shared/remindMes.json", marshaledStruct, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Writes anime notfication subscription to animeSubs.json
func AnimeSubsWrite(animeSubs map[string][]*ShowSub) error {

	// Turns that slice into bytes to be ready to written to file
	marshaledStruct, err := json.MarshalIndent(animeSubs, "", "    ")
	if err != nil {
		return err
	}

	// Writes to file
	err = ioutil.WriteFile("database/shared/animeSubs.json", marshaledStruct, 0644)
	if err != nil {
		return err
	}

	return nil
}

//// Writes vote info to voteInfo.json
//func VoteInfoWrite(info map[string]*VoteInfo, guildID string) error {
//
//	if GuildMap[guildID].GuildSettings.Premium && len(GuildMap[guildID].VoteInfoMap) > 199 {
//		return fmt.Errorf("Error: You have reached the vote limit (200) for this premium server.")
//	} else if !GuildMap[guildID].GuildSettings.Premium && len(GuildMap[guildID].VoteInfoMap) > 49 {
//		return fmt.Errorf("Error: You have reached the vote limit (50) for this server. Please wait for some to be removed or increase them to 200 by upgrading to a premium server at <https://patreon.com/apiks>")
//	}
//
//	// Turns info slice into byte ready to be pushed to file
//	MarshaledStruct, err := json.MarshalIndent(info, "", "    ")
//	if err != nil {
//		return err
//	}
//
//	//Writes to file
//	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/voteInfo.json", guildID), MarshaledStruct, 0644)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//// Writes temp cha info to tempCha.json
//func TempChaWrite(info map[string]*TempChaInfo, guildID string) error {
//
//	if GuildMap[guildID].GuildSettings.Premium && len(GuildMap[guildID].TempChaMap) > 199 {
//		return fmt.Errorf("Error: You have reached the temporary channel limit (200) for this premium server.")
//	} else if !GuildMap[guildID].GuildSettings.Premium && len(GuildMap[guildID].TempChaMap) > 49 {
//		return fmt.Errorf("Error: You have reached the temporary channel limit (50) for this server. Please wait for some to be removed or increase them to 200 by upgrading to a premium server at <https://patreon.com/apiks>")
//	}
//
//	// Turns info map into byte ready to be pushed to file
//	MarshaledStruct, err := json.MarshalIndent(info, "", "    ")
//	if err != nil {
//		return err
//	}
//
//	// Writes to file
//	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/tempCha.json", guildID), MarshaledStruct, 0644)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

//// Writes react channel join info to ReactJoin.json
//func ReactJoinWrite(info map[string]*ReactJoin, guildID string) error {
//
//	if GuildMap[guildID].GuildSettings.Premium && len(GuildMap[guildID].ReactJoinMap) > 399 {
//		return fmt.Errorf("Error: You have reached the react autorole limit (400) for this premium server.")
//	} else if !GuildMap[guildID].GuildSettings.Premium && len(GuildMap[guildID].ReactJoinMap) > 99 {
//		return fmt.Errorf("Error: You have reached the react autorole limit (100) for this server. Please remove some or increase them to 400 by upgrading to a premium server at <https://patreon.com/apiks>")
//	}
//
//	// Turns info slice into byte ready to be pushed to file
//	marshaledStruct, err := json.MarshalIndent(info, "", "    ")
//	if err != nil {
//		return err
//	}
//
//	// Writes to file
//	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/reactJoin.json", guildID), marshaledStruct, 0644)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//// Writes Raffles to raffles.json
//func RafflesWrite(raffle []*Raffle, guildID string) error {
//	Mutex.Lock()
//	defer Mutex.Unlock()
//
//	if GuildMap[guildID].GuildSettings.Premium && len(GuildMap[guildID].Raffles) > 199 {
//		return fmt.Errorf("Error: You have reached the raffle limit (200) for this premium server.")
//	} else if !GuildMap[guildID].GuildSettings.Premium && len(GuildMap[guildID].Raffles) > 49 {
//		return fmt.Errorf("Error: You have reached the raffle limit (50) for this server. Please remove some or increase them to 200 by upgrading to a premium server at <https://patreon.com/apiks>")
//	}
//
//	// Turns that slice into bytes to be ready to written to file
//	marshaledStruct, err := json.MarshalIndent(raffle, "", "    ")
//	if err != nil {
//		return err
//	}
//
//	// Writes to file
//	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/raffles.json", guildID), marshaledStruct, 0644)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//// Writes Waifus to waifus.json
//func WaifusWrite(waifu []*Waifu, guildID string) error {
//
//	if GuildMap[guildID].GuildSettings.Premium && len(GuildMap[guildID].Waifus) > 399 {
//		return fmt.Errorf("Error: You have reached the waifu limit (400) for this premium server.")
//	} else if !GuildMap[guildID].GuildSettings.Premium && len(GuildMap[guildID].Waifus) > 49 {
//		return fmt.Errorf("Error: You have reached the waifu limit (50) for this server. Please remove some or increase them to 400 by upgrading to a premium server at <https://patreon.com/apiks>")
//	}
//
//	// Turns that slice into bytes to be ready to written to file
//	marshaledStruct, err := json.MarshalIndent(waifu, "", "    ")
//	if err != nil {
//		return err
//	}
//
//	// Writes to file
//	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/waifus.json", guildID), marshaledStruct, 0644)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//// Writes WaifuTrades to waifutrades.json
//func WaifuTradesWrite(trade []*WaifuTrade, guildID string) error {
//
//	if GuildMap[guildID].GuildSettings.Premium && len(GuildMap[guildID].WaifuTrades) > 499 {
//		return fmt.Errorf("Error: This premium server has reached the waifu trade limit (500).")
//	} else if !GuildMap[guildID].GuildSettings.Premium && len(GuildMap[guildID].WaifuTrades) > 149 {
//		return fmt.Errorf("Error: This server has reached the waifu trade limit (150). Please contact the bot creator or increase the limit to 500 by upgrading to a premium server at <https://patreon.com/apiks>")
//	}
//
//	// Turns that slice into bytes to be ready to written to file
//	marshaledStruct, err := json.MarshalIndent(trade, "", "    ")
//	if err != nil {
//		return err
//	}
//
//	// Writes to file
//	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/waifuTrades.json", guildID), marshaledStruct, 0644)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

// Writes to punishedUsers.json from []PunishedUsers
func PunishedUsersWrite(punishedUsers []*PunishedUsers, guildID string) error {
	// Turns that slice into bytes to be ready to written to file
	marshaledStruct, err := json.MarshalIndent(punishedUsers, "", "    ")
	if err != nil {
		return err
	}
	// Writes to file
	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/punishedUsers.json", guildID), marshaledStruct, 0644)
	if err != nil {
		return err
	}

	return nil
}

//// Removes raffle with name string "raffle" from raffles.json
//func RaffleRemove(raffle string, guildID string) error {
//
//	var raffleExists bool
//
//	// Checks if that raffle already exists in the raffles slice and deletes it if so
//	Mutex.Lock()
//	for i := len(GuildMap[guildID].Raffles) - 1; i >= 0; i-- {
//		if strings.ToLower(GuildMap[guildID].Raffles[i].Name) != strings.ToLower(raffle) {
//			continue
//		}
//
//		if i < len(GuildMap[guildID].Raffles)-1 {
//			copy(GuildMap[guildID].Raffles[i:], GuildMap[guildID].Raffles[i+1:])
//		}
//		GuildMap[guildID].Raffles[len(GuildMap[guildID].Raffles)-1] = nil
//		GuildMap[guildID].Raffles = GuildMap[guildID].Raffles[:len(GuildMap[guildID].Raffles)-1]
//		raffleExists = true
//		break
//	}
//
//	if !raffleExists {
//		Mutex.Unlock()
//		return fmt.Errorf("Error: No such raffle exists")
//	}
//
//	// Turns that struct slice into bytes again to be ready to written to file
//	marshaledStruct, err := json.Marshal(GuildMap[guildID].Raffles)
//	if err != nil {
//		Mutex.Unlock()
//		return err
//	}
//	Mutex.Unlock()
//
//	// Writes to file
//	err = ioutil.WriteFile(fmt.Sprintf("%s/%s/raffles.json", DBPath, guildID), marshaledStruct, 0644)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//// Adds string "phrase" to filters.json and memory
//func FiltersWrite(phrase string, guildID string) error {
//
//	if GuildMap[guildID].GuildSettings.Premium && len(GuildMap[guildID].Filters) > 299 {
//		return fmt.Errorf("Error: You have reached the filter limit (300) for this premium server.")
//	} else if !GuildMap[guildID].GuildSettings.Premium && len(GuildMap[guildID].Filters) > 49 {
//		return fmt.Errorf("Error: You have reached the filter limit (50) for this server. Please remove some or increase them to 300 by upgrading to a premium server at <https://patreon.com/apiks>")
//	}
//
//	// Appends the new filtered phrase to a slice of all of the old ones if it doesn't exist
//	Mutex.Lock()
//	for _, filter := range GuildMap[guildID].Filters {
//		if filter.Filter == phrase {
//			Mutex.Unlock()
//			return fmt.Errorf(fmt.Sprintf("Error: `%s` is already on the filter list.", phrase))
//		}
//	}
//
//	// Adds the phrase to the filter list
//	GuildMap[guildID].Filters = append(GuildMap[guildID].Filters, &Filter{phrase})
//
//	// Turns that struct slice into bytes again to be ready to written to file
//	marshaledStruct, err := json.MarshalIndent(GuildMap[guildID].Filters, "", "    ")
//	if err != nil {
//		Mutex.Unlock()
//		return err
//	}
//	Mutex.Unlock()
//
//	// Writes to file
//	err = ioutil.WriteFile(fmt.Sprintf("%s/%s/filters.json", DBPath, guildID), marshaledStruct, 0644)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//// Removes string "phrase" from filters.json and memory
//func FiltersRemove(phrase string, guildID string) error {
//
//	var filterExists bool
//
//	// Deletes the filtered phrase if it finds it exists
//	Mutex.Lock()
//	for i, filter := range GuildMap[guildID].Filters {
//		if filter.Filter == phrase {
//
//			if i < len(GuildMap[guildID].Filters)-1 {
//				copy(GuildMap[guildID].Filters[i:], GuildMap[guildID].Filters[i+1:])
//			}
//			GuildMap[guildID].Filters[len(GuildMap[guildID].Filters)-1] = nil
//			GuildMap[guildID].Filters = GuildMap[guildID].Filters[:len(GuildMap[guildID].Filters)-1]
//
//			filterExists = true
//			break
//		}
//	}
//
//	// Exits func if the filter is not on the list
//	if !filterExists {
//		Mutex.Unlock()
//		return fmt.Errorf(fmt.Sprintf("Error: `%v` is not in the filter list.", phrase))
//	}
//
//	// Turns that struct slice into bytes again to be ready to written to file
//	marshaledStruct, err := json.Marshal(GuildMap[guildID].Filters)
//	if err != nil {
//		Mutex.Unlock()
//		return err
//	}
//	Mutex.Unlock()
//
//	// Writes to file
//	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/filters.json", guildID), marshaledStruct, 0644)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//// Adds a string file extension to extensionList.json and memory
//func ExtensionsWrite(extension string, guildID string) error {
//
//	if GuildMap[guildID].GuildSettings.Premium && len(GuildMap[guildID].ExtensionList) > 199 {
//		return fmt.Errorf("Error: You have reached the file extension filter limit (200) for this premium server.")
//	} else if !GuildMap[guildID].GuildSettings.Premium && len(GuildMap[guildID].ExtensionList) > 49 {
//		return fmt.Errorf("Error: You have reached the file extension filter (50) for this server. Please remove some or increase them to 200 by upgrading to a premium server at <https://patreon.com/apiks>")
//	}
//
//	if strings.HasPrefix(extension, ".") {
//		extension = strings.TrimPrefix(extension, ".")
//	}
//
//	// Appends the new file extension to a slice of all of the old ones if it doesn't already exist
//	Mutex.Lock()
//	for ext := range GuildMap[guildID].ExtensionList {
//		if strings.ToLower(ext) == strings.ToLower(extension) {
//			Mutex.Unlock()
//			return fmt.Errorf(fmt.Sprintf("Error: `%v` is already on the file extension list.", ext))
//		}
//	}
//
//	// Adds the extension to the file extension list with its type (blacklist or whitelist)
//	if GuildMap[guildID].GuildSettings.WhitelistFileFilter {
//		GuildMap[guildID].ExtensionList[strings.ToLower(extension)] = "whitelist"
//	} else {
//		GuildMap[guildID].ExtensionList[strings.ToLower(extension)] = "blacklist"
//	}
//
//	// Turns that struct slice into bytes again to be ready to written to file
//	marshaledStruct, err := json.MarshalIndent(GuildMap[guildID].ExtensionList, "", "    ")
//	if err != nil {
//		Mutex.Unlock()
//		return err
//	}
//	Mutex.Unlock()
//
//	// Writes to file
//	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/extensionList.json", guildID), marshaledStruct, 0644)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

//// Removes a file extension from extensionList.json and memory
//func ExtensionsRemove(extension string, guildID string) error {
//
//	var extensionExists bool
//
//	if strings.HasPrefix(extension, ".") {
//		extension = strings.TrimPrefix(extension, ".")
//	}
//
//	// Deletes the filtered phrase if it finds it exists
//	Mutex.Lock()
//	for ext := range GuildMap[guildID].ExtensionList {
//		if strings.ToLower(ext) == strings.ToLower(extension) {
//			delete(GuildMap[guildID].ExtensionList, extension)
//			extensionExists = true
//			break
//		}
//	}
//
//	// Exits func if the extension is not on the blacklist
//	if !extensionExists {
//		Mutex.Unlock()
//		return fmt.Errorf(fmt.Sprintf("Error: `%v` is not in the file extension list.", extension))
//	}
//
//	// Turns that struct slice into bytes again to be ready to written to file
//	marshaledStruct, err := json.Marshal(GuildMap[guildID].ExtensionList)
//	if err != nil {
//		Mutex.Unlock()
//		return err
//	}
//	Mutex.Unlock()
//
//	// Writes to file
//	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/extensionList.json", guildID), marshaledStruct, 0644)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//// Adds string "phrase" to messReqs.json and memory
//func MessRequirementWrite(phrase string, channel string, filterType string, guildID string) error {
//
//	if GuildMap[guildID].GuildSettings.Premium && len(GuildMap[guildID].MessageRequirements) > 149 {
//		return fmt.Errorf("Error: You have reached the message requirement filter limit (150) for this premium server.")
//	} else if !GuildMap[guildID].GuildSettings.Premium && len(GuildMap[guildID].MessageRequirements) > 49 {
//		return fmt.Errorf("Error: You have reached the message requirement filter limit (50) for this server. Please remove some or increase them to 150 by upgrading to a premium server at <https://patreon.com/apiks>")
//	}
//
//	// Appends the new phrase to a slice of all of the old ones if it doesn't exist
//	Mutex.Lock()
//	for _, requirement := range GuildMap[guildID].MessageRequirements {
//		if requirement.Phrase == phrase {
//			Mutex.Unlock()
//			return fmt.Errorf(fmt.Sprintf("Error: `%v` is already on the message requirement list.", phrase))
//		}
//	}
//
//	// Adds the phrase to the message requirement list
//	id, err := GenerateID(guildID)
//	if err != nil {
//		log.Println(err)
//		Mutex.Unlock()
//		return err
//	}
//
//	GuildMap[guildID].MessageRequirements = append(GuildMap[guildID].MessageRequirements, &MessRequirement{id, phrase, filterType, channel, ""})
//
//	// Turns that struct slice into bytes again to be ready to written to file
//	marshaledStruct, err := json.MarshalIndent(GuildMap[guildID].MessageRequirements, "", "    ")
//	if err != nil {
//		Mutex.Unlock()
//		return err
//	}
//	Mutex.Unlock()
//
//	// Writes to file
//	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/messReqs.json", guildID), marshaledStruct, 0644)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//// Removes string "phrase" from messReqs.json and memory
//func MessRequirementRemove(phrase string, channelID string, guildID string) error {
//
//	var phraseExists bool
//
//	// Deletes the filtered phrase if it finds it exists
//	Mutex.Lock()
//	for i, requirement := range GuildMap[guildID].MessageRequirements {
//		if requirement.Phrase == phrase {
//			if channelID != "" {
//				if requirement.ChannelID != channelID {
//					continue
//				}
//			}
//
//			if i < len(GuildMap[guildID].MessageRequirements)-1 {
//				copy(GuildMap[guildID].MessageRequirements[i:], GuildMap[guildID].MessageRequirements[i+1:])
//			}
//			GuildMap[guildID].MessageRequirements[len(GuildMap[guildID].MessageRequirements)-1] = nil
//			GuildMap[guildID].MessageRequirements = GuildMap[guildID].MessageRequirements[:len(GuildMap[guildID].MessageRequirements)-1]
//
//			phraseExists = true
//			break
//		}
//	}
//
//	// Exits func if the filter is not on the list
//	if !phraseExists {
//		Mutex.Unlock()
//		return fmt.Errorf(fmt.Sprintf("Error: `%s` is not in the message requirement list.", phrase))
//	}
//
//	// Turns that struct slice into bytes again to be ready to written to file
//	marshaledStruct, err := json.Marshal(GuildMap[guildID].MessageRequirements)
//	if err != nil {
//		Mutex.Unlock()
//		return err
//	}
//	Mutex.Unlock()
//
//	// Writes to file
//	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%s/messReqs.json", guildID), marshaledStruct, 0644)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//// Writes spoilerRoles map to spoilerRoles.json
//func SpoilerRolesWrite(SpoilerMapWrite map[string]*discordgo.Role, guildID string) {
//
//	var roleExists bool
//
//	// Appends the new spoiler role to a slice of all of the old ones if it doesn't exist
//	if len(GuildMap[guildID].SpoilerRoles) == 0 {
//		for k := range SpoilerMapWrite {
//			GuildMap[guildID].SpoilerRoles = append(GuildMap[guildID].SpoilerRoles, SpoilerMapWrite[k])
//		}
//	} else {
//		for k := range SpoilerMapWrite {
//			for i := 0; i < len(GuildMap[guildID].SpoilerRoles); i++ {
//				if GuildMap[guildID].SpoilerRoles[i].ID == SpoilerMapWrite[k].ID {
//					roleExists = true
//					break
//
//				} else {
//					roleExists = false
//				}
//			}
//
//			if !roleExists {
//				GuildMap[guildID].SpoilerRoles = append(GuildMap[guildID].SpoilerRoles, SpoilerMapWrite[k])
//			}
//		}
//	}
//
//	// Turns that struct slice into bytes again to be ready to written to file
//	marshaledStruct, err := json.MarshalIndent(GuildMap[guildID].SpoilerRoles, "", "    ")
//	if err != nil {
//		return
//	}
//
//	// Writes to file
//	_ = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/spoilerRoles.json", guildID), marshaledStruct, 0644)
//}
//
//// Deletes a role from spoilerRoles map to spoilerRoles.json
//func SpoilerRolesDelete(roleID string, guildID string) {
//
//	if len(GuildMap[guildID].SpoilerRoles) == 0 {
//		return
//	}
//	for i := 0; i < len(GuildMap[guildID].SpoilerRoles); i++ {
//		if GuildMap[guildID].SpoilerRoles[i].ID == roleID {
//
//			if i < len(GuildMap[guildID].SpoilerRoles)-1 {
//				copy(GuildMap[guildID].SpoilerRoles[i:], GuildMap[guildID].SpoilerRoles[i+1:])
//			}
//			GuildMap[guildID].SpoilerRoles[len(GuildMap[guildID].SpoilerRoles)-1] = nil
//			GuildMap[guildID].SpoilerRoles = GuildMap[guildID].SpoilerRoles[:len(GuildMap[guildID].SpoilerRoles)-1]
//		}
//	}
//
//	// Turns that struct slice into bytes again to be ready to written to file
//	marshaledStruct, err := json.MarshalIndent(GuildMap[guildID].SpoilerRoles, "", "    ")
//	if err != nil {
//		return
//	}
//
//	// Writes to file
//	_ = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%s/spoilerRoles.json", guildID), marshaledStruct, 0644)
//}
//
//// Writes rss info to rssThreads.json
//func RssThreadsWrite(subreddit, author, title, postType, channelID, guildID string, pin bool) error {
//
//	if GuildMap[guildID].GuildSettings.Premium && len(GuildMap[guildID].Feeds) > 399 {
//		return fmt.Errorf("Error: You have reached the RSS thread autopost limit (400) for this server.")
//	} else if !GuildMap[guildID].GuildSettings.Premium && len(GuildMap[guildID].Feeds) > 99 {
//		return fmt.Errorf("Error: You have reached the RSS thread autopost limit (100) for this server. Please remove some or increase them to 400 by upgrading to a premium server at <https://patreon.com/apiks>")
//	}
//
//	// Checks if a thread with these settings exist already
//	for _, thread := range GuildMap[guildID].Feeds {
//		if subreddit == thread.Subreddit && title == thread.Title &&
//			postType == thread.PostType && channelID == thread.ChannelID {
//			return fmt.Errorf("Error: This RSS setting already exists.")
//		}
//	}
//
//	// Appends the thread to the guild's threads
//	id, err := GenerateID(guildID)
//	if err != nil {
//		log.Println(err)
//		return err
//	}
//
//	GuildMap[guildID].Feeds = append(GuildMap[guildID].Feeds, &Feed{id, subreddit, title, author, pin, postType, channelID})
//
//	// Turns that struct slice into bytes ready to written to file
//	marshaledStruct, err := json.MarshalIndent(GuildMap[guildID].Feeds, "", "    ")
//	if err != nil {
//		return err
//	}
//
//	// Writes to file
//	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/rssThreads.json", guildID), marshaledStruct, 0644)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//// Removes a feed from rssThreads.json
//func RssThreadsRemove(subreddit, title, author, postType, channelID, guildID string) error {
//
//	var threadExists bool
//
//	// Deletes the thread if it finds it, else throw error
//	for i := len(GuildMap[guildID].Feeds) - 1; i >= 0; i-- {
//
//		if subreddit == GuildMap[guildID].Feeds[i].Subreddit {
//			if title != "" {
//				if GuildMap[guildID].Feeds[i].Title != title {
//					continue
//				}
//			}
//			if author != "" {
//				if GuildMap[guildID].Feeds[i].Author != author {
//					continue
//				}
//			}
//			if postType != "" {
//				if GuildMap[guildID].Feeds[i].PostType != postType {
//					continue
//				}
//			}
//			if channelID != "" {
//				if GuildMap[guildID].Feeds[i].ChannelID != channelID {
//					continue
//				}
//			}
//
//			if i < len(GuildMap[guildID].Feeds)-1 {
//				copy(GuildMap[guildID].Feeds[i:], GuildMap[guildID].Feeds[i+1:])
//			}
//			GuildMap[guildID].Feeds[len(GuildMap[guildID].Feeds)-1] = nil
//			GuildMap[guildID].Feeds = GuildMap[guildID].Feeds[:len(GuildMap[guildID].Feeds)-1]
//
//			threadExists = true
//		}
//	}
//
//	if !threadExists {
//		return fmt.Errorf("Error: No such Feed exists.")
//	}
//
//	// Turns that struct slice into bytes again to be ready to written to file
//	marshaledStruct, err := json.Marshal(GuildMap[guildID].Feeds)
//	if err != nil {
//		return err
//	}
//
//	// Writes to file
//	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/rssThreads.json", guildID), marshaledStruct, 0644)
//	if err != nil {
//		return err
//	}
//
//	return err
//}
//
//// Writes an rssThread with a date to rssThreadCheck.json
//func RssThreadsTimerWrite(thread *Feed, date time.Time, GUID, guildID string) error {
//
//	// Appends the new item to a slice of all of the old ones if it doesn't exist
//	for _, check := range GuildMap[guildID].FeedChecks {
//		if check.GUID == guildID {
//			return nil
//		}
//	}
//
//	id, err := GenerateID(guildID)
//	if err != nil {
//		log.Println(err)
//		return err
//	}
//
//	GuildMap[guildID].FeedChecks = append(GuildMap[guildID].FeedChecks, &FeedCheck{id, thread, date, GUID})
//
//	// Turns that struct slice into bytes again to be ready to written to file
//	marshaledStruct, err := json.MarshalIndent(GuildMap[guildID].FeedChecks, "", "    ")
//	if err != nil {
//		return err
//	}
//
//	// Writes to file
//	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/rssThreadCheck.json", guildID), marshaledStruct, 0644)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//// Removes a feedCheck from rssThreadCheck.json
//func RssThreadsTimerRemove(thread *Feed, guildID string) error {
//
//	var threadExists bool
//
//	// Deletes the check if it finds it, else throw error
//	for i := len(GuildMap[guildID].FeedChecks) - 1; i >= 0; i-- {
//		if GuildMap[guildID].FeedChecks[i].Feed == thread {
//
//			if i < len(GuildMap[guildID].FeedChecks)-1 {
//				copy(GuildMap[guildID].FeedChecks[i:], GuildMap[guildID].FeedChecks[i+1:])
//			}
//			GuildMap[guildID].FeedChecks[len(GuildMap[guildID].FeedChecks)-1] = nil
//			GuildMap[guildID].FeedChecks = GuildMap[guildID].FeedChecks[:len(GuildMap[guildID].FeedChecks)-1]
//
//			threadExists = true
//			break
//		}
//	}
//
//	if !threadExists {
//		return nil
//	}
//
//	// Turns that struct slice into bytes again to be ready to written to file
//	marshaledStruct, err := json.Marshal(GuildMap[guildID].FeedChecks)
//	if err != nil {
//		return err
//	}
//
//	// Writes to file
//	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/rssThreadCheck.json", guildID), marshaledStruct, 0644)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

// Writes guild settings to guildSettings.json
func GuildSettingsWrite(info *GuildSettings, guildID string) error {

	// Turns info map into byte ready to be pushed to file
	MarshaledStruct, err := json.MarshalIndent(info, "", "    ")
	if err != nil {
		return err
	}

	// Writes to file
	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/guildSettings.json", guildID), MarshaledStruct, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Writes autoposts info to autoposts.json
func AutopostsWrite(info map[string]*Cha, guildID string) error {

	// Turns info map into byte ready to be pushed to file
	MarshaledStruct, err := json.MarshalIndent(info, "", "    ")
	if err != nil {
		return err
	}

	// Writes to file
	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%v/autoposts.json", guildID), MarshaledStruct, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Reads and returns the names of every file in that directory
func IOReadDir(root string) ([]string, error) {
	var files []string
	fileInfo, err := ioutil.ReadDir(root)
	if err != nil {
		return files, err
	}

	for _, file := range fileInfo {
		files = append(files, file.Name())
	}
	return files, nil
}

// Initializes BOT DB files
func InitDB(s *discordgo.Session, guildID string) {
	path := fmt.Sprintf("%s/%s", DBPath, guildID)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0777)
		if err != nil {
			log.Println(err)
			return
		}
		// Send message to support server mod log that a server has been created on the public ZeroTsu
		if s.State.User.ID == "614495694769618944" {
			go func() {
				guild, err := s.State.Guild(guildID)
				if err != nil {
					guild, err = s.Guild(guildID)
					if err != nil {
						_, _ = s.ChannelMessageSend("619899424428130315", fmt.Sprintf("A DB entry has been created for guild: %s", guild.Name))
					}
				}
			}()
		}
	}

	for _, name := range guildFileNames {
		file, err := os.OpenFile(fmt.Sprintf("%s/%s/%s", DBPath, guildID, name), os.O_RDONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Println(err)
			continue
		}
		err = file.Close()
		if err != nil {
			log.Println(err)
			continue
		}
	}

	if _, err := os.Stat("database/shared"); os.IsNotExist(err) {
		os.Mkdir(path, 0777)
	}
	for _, name := range sharedFileNames {
		file, err := os.OpenFile(fmt.Sprintf("database/shared/%s", name), os.O_RDONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Println(err)
			continue
		}
		err = file.Close()
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func SetupGuildSub(guildID string) {
	var shows []*ShowSub

	now := time.Now().UTC()

	// Adds every single show as a guild subscription
	for dayInt, scheduleShows := range AnimeSchedule {
		if scheduleShows == nil {
			continue
		}

		for _, show := range scheduleShows {
			if show == nil {
				continue
			}

			// Checks if the show is from today and whether it has already passed (to avoid notifying the user today if it has passed)
			var hasAiredToday bool
			if int(now.Weekday()) == dayInt {

				// Reset bool
				hasAiredToday = false

				// Parse the air hour and minute
				t, err := time.Parse("3:04 PM", show.GetAirTime())
				if err != nil {
					log.Println(err)
					continue
				}

				// Form the air date for today
				scheduleDate := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), now.Second(), now.Nanosecond(), now.Location())
				scheduleDate = scheduleDate.UTC()

				// Calculates whether the show has already aired today
				difference := now.Sub(scheduleDate.UTC())
				if difference >= 0 {
					hasAiredToday = true
				}
			}

			guildSub := NewShowSub(show.GetName(), false, true)
			if hasAiredToday {
				guildSub.SetNotified(true)
			}

			shows = append(shows, guildSub)
		}
	}

	SharedInfo.GetAnimeSubsMap()[guildID] = shows
}

// Returns if a file really exists
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
