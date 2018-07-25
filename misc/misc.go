package misc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
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

//Variables for various things
var (
	OptinAbovePosition int
	OptinUnderPosition int
	SpoilerPerms       = discordgo.PermissionSendMessages + discordgo.PermissionReadMessages + discordgo.PermissionReadMessageHistory
	SpoilerMap         = make(map[string]*discordgo.Role)

	roleDeleted = false

	//Variables for filter words
	ReadFilters  []FilterStruct
	FilterExists bool

	//Variables for spoiler roles map
	ReadSpoilerRoles []discordgo.Role
	roleExists       bool

	//Variables for threads
	ReadRssThreads      []RssThreadStruct
	ReadRssThreadsCheck []RssThreadCheckStruct
	ThreadExists        bool
)

//Initialize the FilterStruct type
type FilterStruct struct {
	Filter string `json:"Filter"`
}

//Initialize the RssThread type
type RssThreadStruct struct {
	Thread  string `json:"Thread"`
	Channel string `json:"Channel"`
	Author  string `json:"Author"`
}

//Initialize the RssThreadCheck type
type RssThreadCheckStruct struct {
	Thread string    `json:"Thread"`
	Date   time.Time `json:"Date"`
}

// HasPermissions sees if a user has elevated permissions.
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

//Sorts roles alphabetically
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

// Adds string "word" to filters.json and memory
func FiltersWrite(word string) {

	//Creates a struct in which we'll keep the word
	wordStruct := FilterStruct{word}

	FilterExists = false

	//Appends the new filtered word to a slice of all of the old ones if it doesn't exist
	if len(ReadFilters) != 0 {
		for i := 0; i < len(ReadFilters); i++ {
			if ReadFilters[i].Filter == wordStruct.Filter {

				FilterExists = true
				break
			}
		}
	}

	if FilterExists == false {

		ReadFilters = append(ReadFilters, wordStruct)
	}

	//Turns that struct slice into bytes again to be ready to written to file
	MarshaledStruct, err := json.MarshalIndent(ReadFilters, "", "    ")
	if err != nil {

		fmt.Println(err)
	}

	//Writes to file
	err = ioutil.WriteFile("database/filters.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}
}

// Removes string "word" from filters.json and memory
func FiltersRemove(word string) {

	//Puts the filtered word into lowercase
	word = strings.ToLower(word)

	//Creates a struct in which we'll keep the word
	wordStruct := FilterStruct{word}

	FilterExists = false

	//Deletes the filtered word if it finds it exists
	if len(ReadFilters) != 0 {
		for i := 0; i < len(ReadFilters); i++ {
			if ReadFilters[i].Filter == wordStruct.Filter {

				FilterExists = true

				if FilterExists == true {

					ReadFilters = append(ReadFilters[:i], ReadFilters[i+1:]...)
				}
			}
		}
	}

	//Turns that struct slice into bytes again to be ready to written to file
	MarshaledStruct, err := json.Marshal(ReadFilters)
	if err != nil {

		fmt.Println(err)
	}

	//Writes to file
	err = ioutil.WriteFile("database/filters.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}
}

// Reads filters from filters.json
func FiltersRead() {

	//Reads all the filtered words from the filters.json file and puts them in filtersByte as bytes
	filtersByte, _ := ioutil.ReadFile("database/filters.json")

	//Takes the filtered words from filter.json from byte and puts them into the FilterStruct struct slice
	json.Unmarshal(filtersByte, &ReadFilters)
}

// Writes spoilerRoles map to spoilerRoles.json
func SpoilerRolesWrite(SpoilerMap map[string]*discordgo.Role) {

	//Reads all the spoilerRoles the spoilerRoles.json file and puts them in spoilerRolesByte as bytes
	spoilerRolesByte, _ := ioutil.ReadFile("database/spoilerRoles.json")

	//Makes a new variable in which we'll keep all of the spoiler roles
	var spoilerRolesWrite []discordgo.Role

	//Takes the spoiler roles from spoilerRoles.json from byte and puts them into the spoilerRolesWrite variable
	json.Unmarshal(spoilerRolesByte, &spoilerRolesWrite)

	//Appends the new spoiler role to a slice of all of the old ones if it doesn't exist
	if len(spoilerRolesWrite) == 0 {
		for k, _ := range SpoilerMap {

			spoilerRolesWrite = append(spoilerRolesWrite, *SpoilerMap[k])
		}
	} else {
		for k, _ := range SpoilerMap {
			for i := 0; i < len(spoilerRolesWrite); i++ {
				if spoilerRolesWrite[i].ID == SpoilerMap[k].ID {

					roleExists = true

					break
				} else {

					roleExists = false
				}
			}

			if roleExists == false {

				spoilerRolesWrite = append(spoilerRolesWrite, *SpoilerMap[k])
			}
		}
	}

	//Turns that struct slice into bytes again to be ready to written to file
	MarshaledStruct, err := json.MarshalIndent(spoilerRolesWrite, "", "    ")
	if err != nil {

		fmt.Println(err)
	}

	//Writes to file
	err = ioutil.WriteFile("database/spoilerRoles.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}
}

//Deletes a role from spoilerRoles map to spoilerRoles.json
func SpoilerRolesDelete(roleID string) {

	//Reads all the spoilerRoles the spoilerRoles.json file and puts them in spoilerRolesByte as bytes
	spoilerRolesByte, _ := ioutil.ReadFile("database/spoilerRoles.json")

	//Makes a new variable in which we'll keep all of the spoiler roles
	var spoilerRolesWrite []discordgo.Role

	//Takes the spoiler roles from spoilerRoles.json from byte and puts them into the spoilerRolesWrite variable
	json.Unmarshal(spoilerRolesByte, &spoilerRolesWrite)

	if len(spoilerRolesWrite) != 0 {
		for i := 0; i < len(spoilerRolesWrite); i++ {
			if spoilerRolesWrite[i].ID == roleID {

				spoilerRolesWrite = append(spoilerRolesWrite[:i], spoilerRolesWrite[i+1:]...)
			}
		}
	}

	//Turns that struct slice into bytes again to be ready to written to file
	MarshaledStruct, err := json.MarshalIndent(spoilerRolesWrite, "", "    ")
	if err != nil {

		fmt.Println(err)
	}

	//Writes to file
	err = ioutil.WriteFile("database/spoilerRoles.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}
}

//Reads filters from spoilerRoles.json
func SpoilerRolesRead() {

	//Reads all the spoiler roles from the spoilerRoles.json file and puts them in spoilerRolesByte as bytes
	spoilerRolesByte, _ := ioutil.ReadFile("database/spoilerRoles.json")

	//Takes the spoiler roles from spoilerRoles.json from byte and puts them into the ReadSpoilerRoles struct slice
	json.Unmarshal(spoilerRolesByte, &ReadSpoilerRoles)

	//Resets spoilerMap map to be ready to fill
	SpoilerMap = nil

	if SpoilerMap == nil {

		SpoilerMap = make(map[string]*discordgo.Role)
	}

	//Fills spoilerMap with roles from the spoilerRoles.json file if latter is not empty
	if len(ReadSpoilerRoles) != 0 {
		for i := 0; i < len(ReadSpoilerRoles); i++ {

			SpoilerMap[ReadSpoilerRoles[i].ID] = &ReadSpoilerRoles[i]
		}
	}
}

//Writes string "thread" to rssThreadsCheck.json
func RssThreadsWrite(thread string, channel string, author string) {

	//Creates a struct in which we'll keep the thread
	threadStruct := RssThreadStruct{thread, channel, author}

	//Reads all the rss threads from the rssThreads.json file and puts them in rssThreadsByte as bytes
	rssThreadsByte, _ := ioutil.ReadFile("database/rssThreads.json")

	//Makes a new variable in which we'll keep all of the threads
	var rssThreadsWrite []RssThreadStruct

	//Takes the threads from rssThreads.json from byte and puts them into the RssThreadStruct struct slice
	json.Unmarshal(rssThreadsByte, &rssThreadsWrite)

	ThreadExists = false

	//Appends the new thread to a slice of all of the old ones if it doesn't exist
	if len(rssThreadsWrite) != 0 {
		for i := 0; i < len(rssThreadsWrite); i++ {
			if rssThreadsWrite[i].Thread == threadStruct.Thread {

				ThreadExists = true
				break
			}
		}
	}

	if ThreadExists == false {

		rssThreadsWrite = append(rssThreadsWrite, threadStruct)
	}

	//Turns that struct slice into bytes again to be ready to written to file
	MarshaledStruct, err := json.MarshalIndent(rssThreadsWrite, "", "    ")
	if err != nil {

		fmt.Println(err)
	}

	//Writes to file
	err = ioutil.WriteFile("database/rssThreads.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}
}

//Removes string "thread" to rssThreads.json
func RssThreadsRemove(thread string, channel string, author string) {

	//Puts the thread string into lowercase
	thread = strings.ToLower(thread)

	//Creates a struct in which we'll keep the thread
	threadStruct := RssThreadStruct{thread, channel, author}

	//Reads all the rss threads from the rssThreads.json file and puts them in rssThreadsByte as bytes
	rssThreadsByte, _ := ioutil.ReadFile("database/rssThreads.json")

	//Makes a new variable in which we'll keep all of the threads
	var rssThreadsWrite []RssThreadStruct

	//Takes the set threads from rssThreads.json from byte and puts them into the RssThreadStruct struct slice
	json.Unmarshal(rssThreadsByte, &rssThreadsWrite)

	threadExists := false

	//Deletes the thread if it finds it exists
	if len(rssThreadsWrite) != 0 {
		for i := 0; i < len(rssThreadsWrite); i++ {

			if rssThreadsWrite[i].Thread == threadStruct.Thread {

				threadExists = true

				if threadExists == true {

					rssThreadsWrite = append(rssThreadsWrite[:i], rssThreadsWrite[i+1:]...)
				}
			}
		}
	}

	//Turns that struct slice into bytes again to be ready to written to file
	MarshaledStruct, err := json.Marshal(rssThreadsWrite)
	if err != nil {

		fmt.Println(err)
	}

	//Writes to file
	err = ioutil.WriteFile("database/rssThreads.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}
}

//Reads threads from rssThreads.json
func RssThreadsRead() {

	//Reads all the rss threads from the rssThreads.json file and puts them in rssThreadsByte as bytes
	rssThreadsByte, _ := ioutil.ReadFile("database/rssThreads.json")

	//Takes the set threads from rssThreads.json from byte and puts them into the RssThreadStruct struct slice
	json.Unmarshal(rssThreadsByte, &ReadRssThreads)
}

//Writes string "thread" to rssThreadCheck.json
func RssThreadsCheckWrite(thread string, date time.Time) {

	//Creates a struct in which we'll keep the thread
	threadCheckStruct := RssThreadCheckStruct{thread, date}

	//Reads all the rss threads from the rssThreadCheck.json file and puts them in rssThreadsCheckByte as bytes
	rssThreadsCheckByte, _ := ioutil.ReadFile("database/rssThreadCheck.json")

	//Makes a new variable in which we'll keep all of the threads
	var rssThreadsCheckWrite []RssThreadCheckStruct

	//Takes the threads from rssThreadCheck.json from byte and puts them into the RssThreadStruct struct slice
	json.Unmarshal(rssThreadsCheckByte, &rssThreadsCheckWrite)

	ThreadExists = false

	//Appends the new thread to a slice of all of the old ones if it doesn't exist
	if len(rssThreadsCheckWrite) != 0 {
		for i := 0; i < len(rssThreadsCheckWrite); i++ {
			if rssThreadsCheckWrite[i].Thread == threadCheckStruct.Thread {

				ThreadExists = true
				break
			}
		}
	}

	if ThreadExists == false {

		rssThreadsCheckWrite = append(rssThreadsCheckWrite, threadCheckStruct)
	}

	//Turns that struct slice into bytes again to be ready to written to file
	MarshaledStruct, err := json.MarshalIndent(rssThreadsCheckWrite, "", "    ")
	if err != nil {

		fmt.Println(err)
	}

	//Writes to file
	err = ioutil.WriteFile("database/rssThreadCheck.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}
}

//Removes string "thread" to rssThreadCheck.json
func RssThreadsCheckRemove(thread string, date time.Time) {

	//Puts the thread string into lowercase
	thread = strings.ToLower(thread)

	//Creates a struct in which we'll keep the thread
	threadCheckStruct := RssThreadCheckStruct{thread, date}

	//Reads all the rss threads from the rssThreadCheck.json file and puts them in rssThreadsCheckByte as bytes
	rssThreadsCheckByte, _ := ioutil.ReadFile("database/rssThreadCheck.json")

	//Makes a new variable in which we'll keep all of the threads
	var rssThreadsCheckWrite []RssThreadCheckStruct

	//Takes the set threads from rssThreadCheck.json from byte and puts them into the RssThreadCheckStruct struct slice
	json.Unmarshal(rssThreadsCheckByte, &rssThreadsCheckWrite)

	threadExists := false

	//Deletes the thread if it finds it exists
	if len(rssThreadsCheckWrite) != 0 {
		for i := 0; i < len(rssThreadsCheckWrite); i++ {
			if rssThreadsCheckWrite[i].Thread == threadCheckStruct.Thread {

				threadExists = true

				if threadExists == true {

					rssThreadsCheckWrite = append(rssThreadsCheckWrite[:i], rssThreadsCheckWrite[i+1:]...)
				}
			}
		}
	}

	//Turns that struct slice into bytes again to be ready to written to file
	MarshaledStruct, err := json.Marshal(rssThreadsCheckWrite)
	if err != nil {

		fmt.Println(err)
	}

	//Writes to file
	err = ioutil.WriteFile("database/rssThreadCheck.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}
}

//Reads threads from rssThreadCheck.json
func RssThreadsCheckRead() {

	//Reads all the rss threads from the rssThreadCheck.json file and puts them in rssThreadsCheckByte as bytes
	rssThreadsCheckByte, _ := ioutil.ReadFile("database/rssThreadCheck.json")

	//Takes the set threads from rssThreads.json from byte and puts them into the RssThreadCheckStruct struct slice
	json.Unmarshal(rssThreadsCheckByte, &ReadRssThreadsCheck)
}

//Every time a role is deleted it deletes it from SpoilerMap
func ListenForDeletedRoleHandler(s *discordgo.Session, g *discordgo.GuildRoleDelete) {

	if g.GuildID == config.ServerID {

		if SpoilerMap[g.RoleID] != nil {

			roleDeleted = true
		}

		if roleDeleted == true {

			mutex := &sync.Mutex{}

			mutex.Lock()
			delete(SpoilerMap, g.RoleID)
			mutex.Unlock()

			SpoilerRolesDelete(g.RoleID)
		}
	}
}

// ResolveTimeFromString resolves a time (usually for unbanning) from a given string formatted #w#d#h#m.
// This returns current time + delay.
// If no time is added to the offset, then this returns true for permanent.
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

func GetUserID(s *discordgo.Session, m *discordgo.Message, messageSlice []string) string {

	// Pulls the userID from the second parameter
	userID := messageSlice[1]

	// Trims fluff if it was a mention. Otherwise check if it's a correct user ID
	if strings.Contains(messageSlice[1], "<@") {

		userID = strings.TrimPrefix(userID, "<@")
		userID = strings.TrimSuffix(userID, ">")
	} else {

		_, err := strconv.ParseInt(userID, 10, 64)
		if len(userID) != 18 || err != nil {

			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid user.")
			if err != nil {
				fmt.Println("Error:", err)
			}
			return ""
		}
	}

	return userID
}