package misc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
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

const (
	UserAgent  = "windows:apiksTEST:v1.0 (by /u/thechosenapiks)"
	DateFormat = "2006-01-02"
)

var (
	OptinAbovePosition int
	OptinUnderPosition int
	SpoilerPerms       = discordgo.PermissionSendMessages + discordgo.PermissionReadMessages + discordgo.PermissionReadMessageHistory
	SpoilerMap         = make(map[string]*discordgo.Role)

	ReadFilters  []FilterStruct

	ReadSpoilerRoles []discordgo.Role

	ReadRssThreads      []RssThreadStruct
	ReadRssThreadsCheck []RssThreadCheckStruct

	EmojiStats = make(map[string]*Emoji)
	ChannelStats = make(map[string]*Channel)
	UserStats = make(map[string]int)
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

type Emoji struct {
	ID          	   string `json:"id"`
	Name               string `json:"name"`
	MessageUsage       int    `json:"messageUsage"`
	UniqueMessageUsage int    `json:"uniqueMessages"`
	Reactions          int    `json:"reactions"`
}

type Channel struct {
	ChannelID 	  string
	Name 		  string
	Messages  	  map[string]int
	RoleCount 	  map[string]int `json:",omitempty"`
	Optin     	  bool
	Exists    	  bool
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

// Adds string "phrase" to filters.json and memory
func FiltersWrite(phrase string) (bool, error) {

	var (
		phraseStruct = 	FilterStruct{phrase}
		err 			error
	)

	// Appends the new filtered phrase to a slice of all of the old ones if it doesn't exist
	for i := 0; i < len(ReadFilters); i++ {
		if ReadFilters[i].Filter == phraseStruct.Filter {
			return true, err
		}
	}

	ReadFilters = append(ReadFilters, phraseStruct)

	// Turns that struct slice into bytes again to be ready to written to file
	marshaledStruct, err := json.MarshalIndent(ReadFilters, "", "    ")
	if err != nil {
		return false, err
	}

	// Writes to file
	err = ioutil.WriteFile("database/filters.json", marshaledStruct, 0644)
	if err != nil {
		return false, err
	}

	return false, err
}

// Removes string "phrase" from filters.json and memory
func FiltersRemove(phrase string) (bool, error) {

	var (
		filterExists 	bool
		phraseStruct = 	FilterStruct{phrase}
		err          	error
	)

	// Deletes the filtered phrase if it finds it exists
	for i := 0; i < len(ReadFilters); i++ {
		if ReadFilters[i].Filter == phraseStruct.Filter {
			filterExists = true
			ReadFilters = append(ReadFilters[:i], ReadFilters[i+1:]...)
		}
	}

	if filterExists == false {
		return false, err
	}

	// Turns that struct slice into bytes again to be ready to written to file
	marshaledStruct, err := json.Marshal(ReadFilters)
	if err != nil {
		return true, err
	}

	// Writes to file
	err = ioutil.WriteFile("database/filters.json", marshaledStruct, 0644)
	if err != nil {
		return true, err
	}

	return true, err
}

// Reads filters from filters.json
func FiltersRead() {

	// Reads all the filtered words from the filters.json file and puts them in filtersByte as bytes
	filtersByte, err := ioutil.ReadFile("database/filters.json")
	if err != nil {
		return
	}

	// Takes the filtered words from filter.json from byte and puts them into the FilterStruct struct slice
	_ = json.Unmarshal(filtersByte, &ReadFilters)
}

// Writes spoilerRoles map to spoilerRoles.json
func SpoilerRolesWrite(SpoilerMap map[string]*discordgo.Role) {

	var (
		roleExists  bool
	)

	// Appends the new spoiler role to a slice of all of the old ones if it doesn't exist
	if len(ReadSpoilerRoles) == 0 {
		for k := range SpoilerMap {

			ReadSpoilerRoles = append(ReadSpoilerRoles, *SpoilerMap[k])
		}
	} else {
		for k := range SpoilerMap {
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
	marshaledStruct, err := json.MarshalIndent(ReadSpoilerRoles, "", "    ")
	if err != nil {
		return
	}

	// Writes to file
	_ = ioutil.WriteFile("database/spoilerRoles.json", marshaledStruct, 0644)
}

// Deletes a role from spoilerRoles map to spoilerRoles.json
func SpoilerRolesDelete(roleID string) {

	if len(ReadSpoilerRoles) == 0 {
		return
	}
	for i := 0; i < len(ReadSpoilerRoles); i++ {
		if ReadSpoilerRoles[i].ID == roleID {
			ReadSpoilerRoles = append(ReadSpoilerRoles[:i], ReadSpoilerRoles[i+1:]...)
		}
	}

	// Turns that struct slice into bytes again to be ready to written to file
	marshaledStruct, err := json.MarshalIndent(ReadSpoilerRoles, "", "    ")
	if err != nil {
		return
	}

	// Writes to file
	_ = ioutil.WriteFile("database/spoilerRoles.json", marshaledStruct, 0644)
}

// Reads spoiler roles from spoilerRoles.json
func SpoilerRolesRead() {

	// Reads all the spoiler roles from the spoilerRoles.json file and puts them in spoilerRolesByte as bytes
	spoilerRolesByte, err := ioutil.ReadFile("database/spoilerRoles.json")
	if err != nil {
		return
	}

	// Takes the spoiler roles from spoilerRoles.json from byte and puts them into the ReadSpoilerRoles struct slice
	err = json.Unmarshal(spoilerRolesByte, &ReadSpoilerRoles)
	if err != nil {
		return
	}

	// Fills spoilerMap with roles from the spoilerRoles.json file if latter is not empty
	for i := 0; i < len(ReadSpoilerRoles); i++ {
		SpoilerMap[ReadSpoilerRoles[i].ID] = &ReadSpoilerRoles[i]
	}
}

// Every time a role is deleted it deletes it from SpoilerMap
func ListenForDeletedRoleHandler(s *discordgo.Session, g *discordgo.GuildRoleDelete) {

	if g.GuildID != config.ServerID {
		return
	}
	if SpoilerMap[g.RoleID] != nil {
		return
	}

	MapMutex.Lock()
	delete(SpoilerMap, g.RoleID)
	MapMutex.Unlock()

	SpoilerRolesDelete(g.RoleID)
}

// Writes string "thread" to rssThreadsCheck.json
func RssThreadsWrite(thread string, channel string, author string) (bool, error) {

	thread = strings.ToLower(thread)

	var (
		threadStruct = 	RssThreadStruct{thread, channel, author}
		err				error
	)

	// Appends the new thread to a slice of all of the old ones if it doesn't exist
	for i := 0; i < len(ReadRssThreads); i++ {
		if ReadRssThreads[i].Thread == threadStruct.Thread {
			return true, err
		}
	}

	ReadRssThreads = append(ReadRssThreads, threadStruct)

	// Turns that struct slice into bytes again to be ready to written to file
	marshaledStruct, err := json.MarshalIndent(ReadRssThreads, "", "    ")
	if err != nil {
		return false, err
	}

	// Writes to file
	err = ioutil.WriteFile("database/rssThreads.json", marshaledStruct, 0644)
	if err != nil {
		return false, err
	}

	return false, err
}

// Removes string "thread" from rssThreads.json
func RssThreadsRemove(thread string, channel string, author string) (bool, error) {

	thread = strings.ToLower(thread)

	var (
		threadExists = false
		threadStruct = RssThreadStruct{thread, channel, author}
		err          error
	)

	// Deletes the thread if it finds it exists
	for i := 0; i < len(ReadRssThreads); i++ {
		if ReadRssThreads[i].Thread == threadStruct.Thread {
			threadExists = true
			ReadRssThreads = append(ReadRssThreads[:i], ReadRssThreads[i+1:]...)
		}
	}

	if threadExists == false {
		return false, err
	}

	// Turns that struct slice into bytes again to be ready to written to file
	marshaledStruct, err := json.Marshal(ReadRssThreads)
	if err != nil {
		return true, err
	}

	// Writes to file
	err = ioutil.WriteFile("database/rssThreads.json", marshaledStruct, 0644)
	if err != nil {
		return true, err
	}

	return true, err
}

// Reads threads from rssThreads.json
func RssThreadsRead() {

	// Reads all the rss threads from the rssThreads.json file and puts them in rssThreadsByte as bytes
	rssThreadsByte, err := ioutil.ReadFile("database/rssThreads.json")
	if err != nil {
		return
	}

	// Takes the set threads from rssThreads.json from byte and puts them into the RssThreadStruct struct slice
	err = json.Unmarshal(rssThreadsByte, &ReadRssThreads)
	if err != nil {
		return
	}
}

// Writes string "thread" to rssThreadCheck.json. Returns bool depending on success or not
func RssThreadsTimerWrite(thread string, date time.Time) bool {

	thread = strings.ToLower(thread)

	var threadCheckStruct= RssThreadCheckStruct{thread, date}

	// Appends the new thread to a slice of all of the old ones if it doesn't exist
	for p := 0; p < len(ReadRssThreadsCheck); p++ {
		if ReadRssThreadsCheck[p].Thread == threadCheckStruct.Thread {
			return false
		}
	}

	ReadRssThreadsCheck = append(ReadRssThreadsCheck, threadCheckStruct)

	// Turns that struct slice into bytes again to be ready to written to file
	marshaledStruct, err := json.MarshalIndent(ReadRssThreadsCheck, "", "    ")
	if err != nil {
		return false
	}

	// Writes to file
	err = ioutil.WriteFile("database/rssThreadCheck.json", marshaledStruct, 0644)
	if err != nil {
		return false
	}

	return true
}

// Removes string "thread" to rssThreadCheck.json
func RssThreadsTimerRemove(thread string, date time.Time) {

	thread = strings.ToLower(thread)

	var (
		threadExists= false
		threadCheckStruct= RssThreadCheckStruct{thread, date}
	)

	// Deletes the thread if it finds it exists
	for i := 0; i < len(ReadRssThreadsCheck); i++ {
		if ReadRssThreadsCheck[i].Thread == threadCheckStruct.Thread {
			threadExists = true
			ReadRssThreadsCheck = append(ReadRssThreadsCheck[:i], ReadRssThreadsCheck[i+1:]...)
			break
		}
	}
	if threadExists == false {
		return
	}

	// Turns that struct slice into bytes again to be ready to written to file
	marshaledStruct, err := json.Marshal(ReadRssThreadsCheck)
	if err != nil {
		return
	}

	// Writes to file
	err = ioutil.WriteFile("database/rssThreadCheck.json", marshaledStruct, 0644)
	if err != nil {
		return
	}
}

// Reads threads from rssThreadCheck.json
func RssThreadsTimerRead() {

	// Reads all the rss threads from the rssThreadCheck.json file and puts them in rssThreadsCheckByte as bytes
	rssThreadsCheckByte, err := ioutil.ReadFile("database/rssThreadCheck.json")
	if err != nil {
		return
	}

	// Takes the set threads from rssThreads.json from byte and puts them into the RssThreadCheckStruct struct slice
	err = json.Unmarshal(rssThreadsCheckByte, &ReadRssThreadsCheck)
	if err != nil {
		return
	}
}

// Writes emoji stats to emojiStats.json
func EmojiStatsWrite(emojiStats map[string]*Emoji) (bool, error) {

	// Turns that map into bytes to be ready to written to file
	marshaledStruct, err := json.MarshalIndent(emojiStats, "", "    ")
	if err != nil {
		return false, err
	}

	// Writes to file
	err = ioutil.WriteFile("database/emojiStats.json", marshaledStruct, 0644)
	if err != nil {
		return false, err
	}

	return false, err
}

// Reads emoji stats from emojiStats.json
func EmojiStatsRead() {

	// Reads the emoji stats and puts them in emojiStatsByte as bytes
	emojiStatsByte, _ := ioutil.ReadFile("database/emojiStats.json")

	// Takes the bytes and puts them into the EmojiStats map
	MapMutex.Lock()
	_ = json.Unmarshal(emojiStatsByte, &EmojiStats)
	MapMutex.Unlock()
}

// Writes channel stats to channelStats.json
func ChannelStatsWrite(channelStats map[string]*Channel) (bool, error) {

	// Turns that map into bytes to be ready to written to file
	marshaledStruct, err := json.MarshalIndent(channelStats, "", "    ")
	if err != nil {
		return false, err
	}

	// Writes to file
	err = ioutil.WriteFile("database/channelStats.json", marshaledStruct, 0644)
	if err != nil {
		return false, err
	}

	return false, err
}

// Reads channel stats from channelStats.json
func ChannelStatsRead() {

	// Reads the channel stats and puts them in channelStatsByte as bytes
	channelStatsByte, _ := ioutil.ReadFile("database/channelStats.json")

	// Takes the bytes and puts them into the ChannelStats map
	MapMutex.Lock()
	_ = json.Unmarshal(channelStatsByte, &ChannelStats)
	MapMutex.Unlock()
}

// Writes User Change stats to userChangeStats.json
func UserChangeStatsWrite(userStats map[string]int) (bool, error) {

	// Turns that map into bytes to be ready to written to file
	marshaledStruct, err := json.MarshalIndent(UserStats, "", "    ")
	if err != nil {
		return false, err
	}

	// Writes to file
	err = ioutil.WriteFile("database/userChangeStats.json", marshaledStruct, 0644)
	if err != nil {
		return false, err
	}

	return false, err
}

// Reads User Change stats from userChangeStats.json
func UserChangeStatsRead() {

	// Reads the channel stats and puts them in userChangeStatsByte as bytes
	userChangeStatsByte, _ := ioutil.ReadFile("database/userChangeStats.json")

	// Takes the bytes and puts them into the userStats map
	MapMutex.Lock()
	_ = json.Unmarshal(userChangeStatsByte, &UserStats)
	MapMutex.Unlock()
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
func GetUserID(s *discordgo.Session, m *discordgo.Message, messageSlice []string) (string, error) {

	var err error

	// Pulls the userID from the second parameter
	userID := messageSlice[1]

	// Trims fluff if it was a mention. Otherwise check if it's a correct user ID
	if strings.Contains(messageSlice[1], "<@") {

		userID = strings.TrimPrefix(userID, "<@")
		userID = strings.TrimPrefix(userID, "!")
		userID = strings.TrimSuffix(userID, ">")
	} else {

		_, err := strconv.ParseInt(userID, 10, 64)
		if len(userID) < 17 || err != nil {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid user.")
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
				if err != nil {
					return "", err
				}
				return "", err
			}
			return "", err
		}
	}

	return userID, err
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
		_, _ = s.ChannelMessageSend(config.BotLogID, err.Error())
	}
}

// SplitLongMessage takes a message and splits it if it's longer than 1900. By Kagumi
func SplitLongMessage(message string) (split []string) {
	const maxLength = 1950
	if len(message) > maxLength {
		partitions := len(message) / maxLength
		if math.Mod(float64(len(message)), maxLength) > 0 {
			partitions++
		}
		split = make([]string, partitions)
		for i := 0; i < partitions; i++ {
			if i == partitions-1 {
				split[i] = message[i*maxLength:]
				break
			}
			split[i] = message[i*maxLength : (i+1)*maxLength]
		}
	} else {
		split = make([]string, 1)
		split[0] = message
	}
	return
}

// Finds out how many users have the role and returns that number
func GetRoleUserAmount(guild *discordgo.Guild, roles []*discordgo.Role, roleName string) int {

	var (
		users int
		roleID string
	)

	// Finds and saves the requested role's ID
	for roleIndex := range roles {
		if roles[roleIndex].Name == roleName {
			roleID = roles[roleIndex].ID
			break
		}
	}
	// If a user has the requested role, add +1 to users var
	for userID := range guild.Members {
		for roleIndex := range guild.Members[userID].Roles {
			if guild.Members[userID].Roles[roleIndex] == roleID {
				users++
				break
			}
		}
	}
	return users
}