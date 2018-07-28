package misc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"unicode"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
)

// File for misc. functions, commands and variables.

const UserAgent = "windows:apiksTEST:v1.0 (by /u/thechosenapiks)"

var (
	OptinAbovePosition int
	OptinUnderPosition int
	SpoilerPerms       = discordgo.PermissionSendMessages + discordgo.PermissionReadMessages + discordgo.PermissionReadMessageHistory
	SpoilerMap         = make(map[string]*discordgo.Role)

	roleDeleted = false

	ReadFilters  []FilterStruct

	ReadSpoilerRoles []discordgo.Role
	roleExists       bool

	ReadRssThreads      []RssThreadStruct
	ReadRssThreadsCheck []RssThreadCheckStruct
)

type FilterStruct struct {
	Filter string `json:"Filter"`
}

type RssThreadStruct struct {
	Thread  string `json:"Thread"`
	Channel string `json:"Channel"`
	Author  string `json:"Author"`
}

type RssThreadCheckStruct struct {
	Thread string    `json:"Thread"`
	Date   time.Time `json:"Date"`
}

// HasPermissions sees if a user has elevated permissions. By Kagumi
func HasPermissions(m *discordgo.Member) bool {
	for _, r := range m.Roles {
		for _, goodRole := range config.CommandRoles {
			if r == goodRole {
				return true
			}
		}
	}
	return false
}

// Sorts roles alphabetically
type SortRoleByAlphabet []*discordgo.Role

func (r SortRoleByAlphabet) Len() int {
	return len(r)
}

func (r SortRoleByAlphabet) Less(i, j int) bool {

	iRunes := []rune(r[i].Name)
	jRunes := []rune(r[j].Name)

	max := len(iRunes)
	if max > len(jRunes) {
		max = len(jRunes)
	}

	for idx := 0; idx < max; idx++ {
		ir := iRunes[idx]
		jr := jRunes[idx]

		lir := unicode.ToLower(ir)
		ljr := unicode.ToLower(jr)

		if lir != ljr {
			return lir < ljr
		}

		// the lowercase runes are the same, so compare the original
		if ir != jr {
			return ir < jr
		}
	}

	return false

}

func (r SortRoleByAlphabet) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

//Sorts channels alphabetically
type SortChannelByAlphabet []*discordgo.Channel

func (r SortChannelByAlphabet) Len() int {
	return len(r)
}

func (r SortChannelByAlphabet) Less(i, j int) bool {

	iRunes := []rune(r[i].Name)
	jRunes := []rune(r[j].Name)

	max := len(iRunes)
	if max > len(jRunes) {
		max = len(jRunes)
	}

	for idx := 0; idx < max; idx++ {
		ir := iRunes[idx]
		jr := jRunes[idx]

		lir := unicode.ToLower(ir)
		ljr := unicode.ToLower(jr)

		if lir != ljr {
			return lir < ljr
		}

		// the lowercase runes are the same, so compare the original
		if ir != jr {
			return ir < jr
		}
	}

	return false

}

func (r SortChannelByAlphabet) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

type UserAgentTransport struct {
	http.RoundTripper
}

func (c *UserAgentTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("User-Agent", UserAgent)
	return c.RoundTripper.RoundTrip(r)
}

// Adds string "phrase" to filters.json and memory and returns bool
func FiltersWrite(phrase string) bool {

	var filterExists bool

	// Creates a struct in which we'll keep the phrase
	phraseStruct := FilterStruct{phrase}

	// Appends the new filtered phrase to a slice of all of the old ones if it doesn't exist
	if len(ReadFilters) != 0 {
		for i := 0; i < len(ReadFilters); i++ {
			if ReadFilters[i].Filter == phraseStruct.Filter {

				filterExists = true
				break
			}
		}
	}

	if filterExists == false {

		ReadFilters = append(ReadFilters, phraseStruct)
	}

	// Turns that struct slice into bytes again to be ready to written to file
	MarshaledStruct, err := json.MarshalIndent(ReadFilters, "", "    ")
	if err != nil {

		fmt.Println(err)
	}

	// Writes to file
	err = ioutil.WriteFile("database/filters.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}

	if filterExists == true {

		return true
	} else {

		return false
	}
}

// Removes string "phrase" from filters.json and memory and returns bool
func FiltersRemove(phrase string) bool {

	var filterExists bool

	// Creates a struct in which we'll keep the phrase
	phraseStruct := FilterStruct{phrase}

	// Deletes the filtered phrase if it finds it exists
	if len(ReadFilters) != 0 {
		for i := 0; i < len(ReadFilters); i++ {
			if ReadFilters[i].Filter == phraseStruct.Filter {

				filterExists = true

				if filterExists == true {

					ReadFilters = append(ReadFilters[:i], ReadFilters[i+1:]...)
				}
			}
		}
	}

	// Turns that struct slice into bytes again to be ready to written to file
	MarshaledStruct, err := json.Marshal(ReadFilters)
	if err != nil {

		fmt.Println(err)
	}

	// Writes to file
	err = ioutil.WriteFile("database/filters.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}

	if filterExists == true {

		return true
	} else {

		return false
	}
}

// Reads filters from filters.json
func FiltersRead() {

	// Reads all the filtered words from the filters.json file and puts them in filtersByte as bytes
	filtersByte, err := ioutil.ReadFile("database/filters.json")
	if err != nil {

		fmt.Println("Error:", err)
	}

	// Takes the filtered words from filter.json from byte and puts them into the FilterStruct struct slice
	err = json.Unmarshal(filtersByte, &ReadFilters)
	if err != nil {

		fmt.Println("Error:", err)
	}
}

// Writes spoilerRoles map to spoilerRoles.json
func SpoilerRolesWrite(SpoilerMap map[string]*discordgo.Role) {

	// Appends the new spoiler role to a slice of all of the old ones if it doesn't exist
	if len(ReadSpoilerRoles) == 0 {
		for k, _ := range SpoilerMap {

			ReadSpoilerRoles = append(ReadSpoilerRoles, *SpoilerMap[k])
		}
	} else {
		for k, _ := range SpoilerMap {
			for i := 0; i < len(ReadSpoilerRoles); i++ {
				if ReadSpoilerRoles[i].ID == SpoilerMap[k].ID {

					roleExists = true

					break
				} else {

					roleExists = false
				}
			}

			if roleExists == false {

				ReadSpoilerRoles = append(ReadSpoilerRoles, *SpoilerMap[k])
			}
		}
	}

	// Turns that struct slice into bytes again to be ready to written to file
	MarshaledStruct, err := json.MarshalIndent(ReadSpoilerRoles, "", "    ")
	if err != nil {

		fmt.Println(err)
	}

	// Writes to file
	err = ioutil.WriteFile("database/spoilerRoles.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}
}

// Deletes a role from spoilerRoles map to spoilerRoles.json
func SpoilerRolesDelete(roleID string) {

	if len(ReadSpoilerRoles) != 0 {
		for i := 0; i < len(ReadSpoilerRoles); i++ {
			if ReadSpoilerRoles[i].ID == roleID {

				ReadSpoilerRoles = append(ReadSpoilerRoles[:i], ReadSpoilerRoles[i+1:]...)
			}
		}
	}

	// Turns that struct slice into bytes again to be ready to written to file
	MarshaledStruct, err := json.MarshalIndent(ReadSpoilerRoles, "", "    ")
	if err != nil {

		fmt.Println(err)
	}

	// Writes to file
	err = ioutil.WriteFile("database/spoilerRoles.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}
}

// Reads filters from spoilerRoles.json
func SpoilerRolesRead() {

	// Reads all the spoiler roles from the spoilerRoles.json file and puts them in spoilerRolesByte as bytes
	spoilerRolesByte, err := ioutil.ReadFile("database/spoilerRoles.json")
	if err != nil {

		fmt.Println("Error:", err)
	}

	// Takes the spoiler roles from spoilerRoles.json from byte and puts them into the ReadSpoilerRoles struct slice
	err = json.Unmarshal(spoilerRolesByte, &ReadSpoilerRoles)
	if err != nil {

		fmt.Println("Error:", err)
	}

	// Resets spoilerMap map to be ready to fill
	SpoilerMap = nil

	if SpoilerMap == nil {

		SpoilerMap = make(map[string]*discordgo.Role)
	}

	// Fills spoilerMap with roles from the spoilerRoles.json file if latter is not empty
	if len(ReadSpoilerRoles) != 0 {
		for i := 0; i < len(ReadSpoilerRoles); i++ {

			SpoilerMap[ReadSpoilerRoles[i].ID] = &ReadSpoilerRoles[i]
		}
	}
}

// Every time a role is deleted it deletes it from SpoilerMap
func ListenForDeletedRoleHandler(s *discordgo.Session, g *discordgo.GuildRoleDelete) {

	if g.GuildID == config.ServerID {

		if SpoilerMap[g.RoleID] != nil {

			roleDeleted = true
		}

		if roleDeleted == true {

			MapMutex.Lock()
			delete(SpoilerMap, g.RoleID)
			MapMutex.Unlock()

			SpoilerRolesDelete(g.RoleID)
		}
	}
}

// Writes string "thread" to rssThreadsCheck.json
func RssThreadsWrite(thread string, channel string, author string) bool {

	// Creates a struct in which we'll keep the thread
	threadStruct := RssThreadStruct{thread, channel, author}

	threadExists := false

	// Appends the new thread to a slice of all of the old ones if it doesn't exist
	if len(ReadRssThreads) != 0 {
		for i := 0; i < len(ReadRssThreads); i++ {
			if ReadRssThreads[i].Thread == threadStruct.Thread {

				threadExists = true
				break
			}
		}
	}

	if threadExists == false {

		ReadRssThreads = append(ReadRssThreads, threadStruct)
	}

	// Turns that struct slice into bytes again to be ready to written to file
	MarshaledStruct, err := json.MarshalIndent(ReadRssThreads, "", "    ")
	if err != nil {

		fmt.Println(err)
	}

	// Writes to file
	err = ioutil.WriteFile("database/rssThreads.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}

	if threadExists == true {

		return true
	} else {

		return false
	}
}

// Removes string "thread" from rssThreads.json
func RssThreadsRemove(thread string, channel string, author string) bool {

	// Puts the thread string into lowercase
	thread = strings.ToLower(thread)

	// Creates a struct in which we'll keep the thread
	threadStruct := RssThreadStruct{thread, channel, author}

	threadExists := false

	// Deletes the thread if it finds it exists
	if len(ReadRssThreads) != 0 {
		for i := 0; i < len(ReadRssThreads); i++ {

			if ReadRssThreads[i].Thread == threadStruct.Thread {

				threadExists = true

				if threadExists == true {

					ReadRssThreads = append(ReadRssThreads[:i], ReadRssThreads[i+1:]...)
				}
			}
		}
	}

	// Turns that struct slice into bytes again to be ready to written to file
	MarshaledStruct, err := json.Marshal(ReadRssThreads)
	if err != nil {

		fmt.Println(err)
	}

	// Writes to file
	err = ioutil.WriteFile("database/rssThreads.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}

	if threadExists == true {

		return true
	} else {

		return false
	}
}

// Reads threads from rssThreads.json
func RssThreadsRead() {

	// Reads all the rss threads from the rssThreads.json file and puts them in rssThreadsByte as bytes
	rssThreadsByte, err := ioutil.ReadFile("database/rssThreads.json")
	if err != nil {

		fmt.Println("Error:", err)
	}

	// Takes the set threads from rssThreads.json from byte and puts them into the RssThreadStruct struct slice
	err = json.Unmarshal(rssThreadsByte, &ReadRssThreads)
	if err != nil {

		fmt.Println("Error:", err)
	}
}

// Writes string "thread" to rssThreadCheck.json
func RssThreadsTimerWrite(thread string, date time.Time) {

	// Creates a struct in which we'll keep the thread
	threadCheckStruct := RssThreadCheckStruct{thread, date}

	threadExists := false

	//Appends the new thread to a slice of all of the old ones if it doesn't exist
	if len(ReadRssThreadsCheck) != 0 {
		for i := 0; i < len(ReadRssThreadsCheck); i++ {
			if ReadRssThreadsCheck[i].Thread == threadCheckStruct.Thread {

				threadExists = true
				break
			}
		}
	}

	if threadExists == false {

		ReadRssThreadsCheck = append(ReadRssThreadsCheck, threadCheckStruct)
	}

	// Turns that struct slice into bytes again to be ready to written to file
	MarshaledStruct, err := json.MarshalIndent(ReadRssThreadsCheck, "", "    ")
	if err != nil {

		fmt.Println(err)
	}

	// Writes to file
	err = ioutil.WriteFile("database/rssThreadCheck.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}
}

// Removes string "thread" to rssThreadCheck.json
func RssThreadsTimerRemove(thread string, date time.Time) {

	// Puts the thread string into lowercase
	thread = strings.ToLower(thread)

	// Creates a struct in which we'll keep the thread
	threadCheckStruct := RssThreadCheckStruct{thread, date}

	threadExists := false

	// Deletes the thread if it finds it exists
	if len(ReadRssThreadsCheck) != 0 {
		for i := 0; i < len(ReadRssThreadsCheck); i++ {
			if ReadRssThreadsCheck[i].Thread == threadCheckStruct.Thread {

				threadExists = true

				if threadExists == true {

					ReadRssThreadsCheck = append(ReadRssThreadsCheck[:i], ReadRssThreadsCheck[i+1:]...)
				}
			}
		}
	}

	// Turns that struct slice into bytes again to be ready to written to file
	MarshaledStruct, err := json.Marshal(ReadRssThreads)
	if err != nil {

		fmt.Println(err)
	}

	// Writes to file
	err = ioutil.WriteFile("database/rssThreadCheck.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}
}

// Reads threads from rssThreadCheck.json
func RssThreadsTimerRead() {

	// Reads all the rss threads from the rssThreadCheck.json file and puts them in rssThreadsCheckByte as bytes
	rssThreadsCheckByte, err := ioutil.ReadFile("database/rssThreadCheck.json")
	if err != nil {

		fmt.Println("Error:", err)
	}

	// Takes the set threads from rssThreads.json from byte and puts them into the RssThreadCheckStruct struct slice
	err = json.Unmarshal(rssThreadsCheckByte, &ReadRssThreadsCheck)
	if err != nil {

		fmt.Println("Error:", err)
	}
}

// ResolveTimeFromString resolves a time (usually for unbanning) from a given string formatted #w#d#h#m.
// This returns current time + delay.
// If no time is added to the offset, then this returns true for permanent.
// By Kagumi.
func ResolveTimeFromString(given string) (ret time.Time, perma bool) {

	ret = time.Now()
	comp := ret
	matcher, _ := regexp.Compile(`\d+|[wdhmWDHM]+`)
	groups := matcher.FindAllString(given, -1)
	for i, v := range groups {
		val, err := strconv.Atoi(v)
		if err != nil {
			continue
		}
		switch strings.ToLower(groups[i+1]) {
		case "w":
			ret = ret.AddDate(0, 0, val*7)
		case "d":
			ret = ret.AddDate(0, 0, val)
		case "h":
			ret = ret.Add(time.Hour * time.Duration(val))
		case "m":
			ret = ret.Add(time.Minute * time.Duration(val))
		}
	}
	if ret.Equal(comp) {
		perma = true
	}
	return
}

// Resolves a userID from a userID or Mention
func GetUserID(s *discordgo.Session, m *discordgo.Message, messageSlice []string) string {

	// Pulls the userID from the second parameter
	userID := messageSlice[1]

	// Trims fluff if it was a mention. Otherwise check if it's a correct user ID
	if strings.Contains(messageSlice[1], "<@") {

		userID = strings.TrimPrefix(userID, "<@")
		userID = strings.TrimSuffix(userID, ">")
	} else {

		_, err := strconv.ParseInt(userID, 10, 64)
		if len(userID) < 17 || err != nil {

			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid user.")
			if err != nil {
				fmt.Println("Error:", err)
			}
			return ""
		}
	}

	return userID
}

// Mentions channel by *discordgo.Channel. By Kagumi
func ChMention(ch *discordgo.Channel) string {
	return fmt.Sprintf("<#%s>", ch.ID)
}

// Mentions channel by channel ID. By Kagumi
func ChMentionID(channelID string) string {
	return fmt.Sprintf("<#%s>", channelID)
}

// Sends error message to channel command is in. If that throws an error send error message to bot log channel
func CommandErrorHandler(s *discordgo.Session, m *discordgo.Message, err error) {

	_, err = s.ChannelMessageSend(m.ChannelID, err.Error())
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
		if err != nil {

			return
		}
		return
	}
}