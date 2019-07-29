package misc

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
)

// File for misc. functions, commands and variables.

const (
	UserAgent  			= "script:github.com/r-anime/zerotsu:v1.0.0 (by /u/thechosenapiks, /u/geo1088)"
	DateFormat 			= "2006-01-02"
	ToleranceLevel 		= 10
	PixelSampleSize 	= 10000
	CorrectPixelLimit	= 9000
)

var (
	OptinAbovePosition int
	OptinUnderPosition int
	SpoilerPerms       = discordgo.PermissionSendMessages + discordgo.PermissionReadMessages + discordgo.PermissionReadMessageHistory

	ImageSlice		[]io.Reader

	startTime 		time.Time
)

type Filter struct {
	Filter 	string	`json:"Filter"`
}

type MessRequirement struct {
	Phrase 		string	`json:"Phrase"`
	Type 		string	`json:"Type"`
	Channel		string	`json:"Channel"`
	LastUserID	string
}

type RssThread struct {
	Thread  string `json:"Thread"`
	Channel string `json:"Channel"`
	Author  string `json:"Author"`
}

type RssThreadCheck struct {
	Thread string    `json:"Thread"`
	Date   time.Time `json:"Date"`
	ChannelID string `json:"ChannelID"`
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

type RemindMeSlice struct {
	RemindMeSlice []RemindMe
}

type RemindMe struct {
	Message			string
	Date			time.Time
	CommandChannel	string
	RemindID		int
}

type Raffle struct {
	Name			string		`json:"Name"`
	ParticipantIDs	[]string	`json:"ParticipantIDs"`
	ReactMessageID	string		`json:"ReactMessageID"`
}

type Waifu struct {
	Name			string				`json:"Name"`
}

type WaifuTrade struct {
	TradeID			string				`json:"TradeID"`
	InitiatorID		string				`json:"InitiatorID"`
	AccepteeID		string				`json:"AccepteeID"`
}

type Coordinates struct {
	X	int		`json:"X"`
	Y	int		`json:"Y"`
}

//// HasPermissions sees if a user has elevated permissions in a given server
//func HasPermissions(m *discordgo.Member, guildID string) bool {
//	for _, r := range m.Roles {
//		for _, goodRole := range GuildMap[guildID].GuildConfig.CommandRoles {
//			if r == goodRole {
//				return true
//			}
//		}
//	}
//	return false
//}

// HasPermissions sees if a user has elevated permissions in a given server
func HasPermissions(m *discordgo.Member, guildID string) bool {
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

// Every time a role is deleted it deletes it from SpoilerMap
func ListenForDeletedRoleHandler(s *discordgo.Session, g *discordgo.GuildRoleDelete) {

	if g.GuildID != config.ServerID {
		return
	}

	MapMutex.Lock()
	if GuildMap[g.GuildID].SpoilerMap[g.RoleID] != nil {
		MapMutex.Unlock()
		return
	}

	delete(GuildMap[g.GuildID].SpoilerMap, g.RoleID)

	SpoilerRolesDelete(g.RoleID, g.GuildID)
	MapMutex.Unlock()
}

// ResolveTimeFromString resolves a time (usually for unbanning) from a given string formatted #w#d#h#m.
// This returns current time + delay.
// If no time is added to the offset, then this returns true for permanent.
// By Kagumi. Modified by Apiks
func ResolveTimeFromString(given string) (ret time.Time, perma bool, err error) {

	ret = time.Now()
	comp := ret
	matcher, _ := regexp.Compile(`\d+|[wdhmWDHM]+`)
	groups := matcher.FindAllString(given, -1)
	if len(groups)%2 != 0 {
		err = fmt.Errorf("Error: invalid date given.")
		return
	}
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

// Resolves a userID from a userID, Mention or username#discrim
func GetUserID(s *discordgo.Session, m *discordgo.Message, messageSlice []string) (string, error) {

	var err 	error

	if len(messageSlice) < 2 {
		err = fmt.Errorf("Error: No @user, userID or username#discrim detected.")
		return "", err
	}

	// Pulls the userID from the second parameter
	userID := messageSlice[1]

	// Handles "me" string on whois
	if strings.ToLower(userID) == "me" {
		userID = m.Author.ID
	}
	// Handles userID if it was in reddit username format
	if strings.Contains(userID, "/u/") {
		userID = strings.TrimPrefix(userID, "/u/")
		MapMutex.Lock()
		for _, user := range GuildMap[m.GuildID].MemberInfoMap {
			if strings.ToLower(user.RedditUsername) == userID {
				userID = user.ID
				break
			}
		}
		MapMutex.Unlock()
	}
	if strings.Contains(userID, "u/") {
		userID = strings.TrimPrefix(userID, "u/")
		MapMutex.Lock()
		for _, user := range GuildMap[m.GuildID].MemberInfoMap {
			if strings.ToLower(user.RedditUsername) == userID {
				userID = user.ID
				break
			}
		}
		MapMutex.Unlock()
	}
	// Handles userID if it was username#discrim format
	if strings.Contains(userID, "#") {
		splitUser := strings.SplitN(userID, "#", 2)
		if len(splitUser) < 2 {
			err = fmt.Errorf("Error: Invalid user. You're trying to username#discrim with spaces in the username." +
				" This command does not support that. Please use an ID.")
			return userID, err
		}
		MapMutex.Lock()
		for _, user := range GuildMap[m.GuildID].MemberInfoMap {
			if strings.ToLower(user.Username) == splitUser[0] && user.Discrim == splitUser[1] {
				userID = user.ID
				break
			}
		}
		MapMutex.Unlock()
	}

	// Trims fluff if it was a mention. Otherwise check if it's a correct user ID
	if strings.Contains(messageSlice[1], "<@") {
		userID = strings.TrimPrefix(userID, "<@")
		userID = strings.TrimPrefix(userID, "!")
		userID = strings.TrimSuffix(userID, ">")
	}
	_, err = strconv.ParseInt(userID, 10, 64)
	if len(userID) < 17 || err != nil {
		err = fmt.Errorf("Error: Invalid user.")
		return userID, err
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

// Returns a string that shows where the error occured exactly
func ErrorLocation(err error) string {
	_, file, line, _ := runtime.Caller(1)
	errorLocation := fmt.Sprintf("Error is in file [%v] near line %v", file, line)
	return errorLocation
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

// Puts banned users in bannedUsersSlice on bot startup from memberInfo
func GetBannedUsers() {
	var (
		bannedUserInfo BannedUsers
		flag bool
	)

	MapMutex.Lock()
	for guildID, _ := range GuildMap {
		for _, user := range GuildMap[guildID].MemberInfoMap {
			for _, ban := range GuildMap[guildID].BannedUsers {
				if user.ID == ban.ID {
					flag = true
					break
				}
			}
			if flag {
				flag = false
				continue
			}
			if user.UnbanDate == "_Never_" ||
				user.UnbanDate == "" ||
				user.UnbanDate == "No ban" {
				continue
			}
			date, err := time.Parse(time.RFC3339, user.UnbanDate)
			if err != nil {
				date, err = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", user.UnbanDate)
				if err != nil {
					date, err = time.Parse("2006-01-02 15:04:05", user.UnbanDate)
					if err != nil {
						fmt.Println("in getBannedUsers date err")
						fmt.Println(err)
						continue
					}
				}
			}
			bannedUserInfo.ID = user.ID
			bannedUserInfo.User = user.Username
			bannedUserInfo.UnbanDate = date
			GuildMap[guildID].BannedUsers = append(GuildMap[guildID].BannedUsers, bannedUserInfo)
		}
		BannedUsersWrite(GuildMap[guildID].BannedUsers, guildID)
	}
	MapMutex.Unlock()
}

//// Writes to bannedUsers.json from bannedUsersSlice
//func BannedUsersWrite(bannedUsers []BannedUsers) {
//	// Turns that slice into bytes to be ready to written to file
//	marshaledStruct, err := json.MarshalIndent(bannedUsers, "", "    ")
//	if err != nil {
//		return
//	}
//	// Writes to file
//	err = ioutil.WriteFile("database/bannedUsers.json", marshaledStruct, 0644)
//	if err != nil {
//		return
//	}
//	return
//}

// Checks if a message contains a channel or user mentions and changes them to a non-mention if true
func MentionParser(s *discordgo.Session, m string, guildID string) string {

	var (
		mentions				string
		userID					string
		userMentionCheck		[]string
		channelMentionCheck		[]string
	)

	if strings.Contains(m, "<@") {

		// Checks for both <@! and <@ mentions
		mentionRegex := regexp.MustCompile(`(?m)<@!?\d+>`)
		userMentionCheck = mentionRegex.FindAllString(m, -1)
		if userMentionCheck != nil {
			for i := range userMentionCheck {
				userID = strings.TrimPrefix(userMentionCheck[i], "<@")
				userID = strings.TrimPrefix(userID, "!")
				userID = strings.TrimSuffix(userID, ">")

				// Checks first in memberInfo. Only checks serverside if it doesn't exist. Saves performance
				MapMutex.Lock()
				if len(GuildMap[guildID].MemberInfoMap) != 0 {
					if _, ok := GuildMap[guildID].MemberInfoMap[userID]; ok {
						mentions += " " + strings.ToLower(GuildMap[guildID].MemberInfoMap[userID].Nickname)
						MapMutex.Unlock()
						continue
					}
				}
				MapMutex.Unlock()

				// If user wasn't found in memberInfo with that username+discrim combo then fetch manually from Discord and then replace mentions with nick
				user, err := s.State.Member(config.ServerID, userID)
				if err != nil {
					user, _ = s.GuildMember(config.ServerID, userID)
				}
				if user != nil {
					m = strings.Replace(m, userMentionCheck[i], fmt.Sprintf("@%v", user.Nick), -1)
				}
			}
		}
	}

	// Checks for channel and replaces mention with channel name
	if strings.Contains(m, "#") {
		channelMentionRegex := regexp.MustCompile(`(?m)(<#\d+>)`)
		channelMentionCheck = channelMentionRegex.FindAllString(m, -1)
		if channelMentionCheck != nil {
			for i := range channelMentionCheck {
				channelID := strings.TrimPrefix(channelMentionCheck[i], "<#")
				channelID = strings.TrimSuffix(channelID, ">")

				// Fetches channel so we can parse its string name
				cha, err := s.Channel(channelID)
				if err != nil {
					continue
				}
				if cha != nil {
					m = strings.Replace(m, channelMentionCheck[i], fmt.Sprintf("#%v", cha.Name), -1)
				}
			}
		}
	}

	return m
}

// Parses a string for a channel and returns its ID and name
func ChannelParser(s *discordgo.Session, channel string) (string, string) {
	var (
		channelID 	string
		channelName string
		flag		bool
	)

	// If it's a channel ping remove <# and > from it to get the channel ID
	if strings.Contains(channel, "#") {
		channelID = strings.TrimPrefix(channel,"<#")
		channelID = strings.TrimSuffix(channelID,">")
	}

	// Check if it's an ID by length and save the ID if so
	_, err := strconv.Atoi(channel)
	if len(channel) >= 17 && err == nil {
		channelID = channel
	}

	// Find the channelID if it doesn't exists via channel name, else find the channel name
	channels, err := s.GuildChannels(config.ServerID)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + ErrorLocation(err))
		if err != nil {
			return channelID, channelName
		}
		return channelID, channelName
	}
	for _, cha := range channels {
		if channelID == "" {
			if strings.ToLower(cha.Name) == strings.ToLower(channel) {
				channelID = cha.ID
				channelName = cha.Name
				flag = true
				break
			}
		}
		if cha.ID == channelID {
			channelName = cha.Name
			flag = true
			break
		}
	}

	if !flag {
		return "", ""
	}

	return channelID, channelName
}

// Parses a string for a category and returns its ID and name
func CategoryParser(s *discordgo.Session, category string) (string, string) {
	var (
		categoryID 		string
		categoryName 	string
		flag			bool
	)

	// Check if it's an ID by length and save the ID if so
	_, err := strconv.Atoi(category)
	if len(category) >= 17 && err == nil {
		categoryID = category
	}

	// Find the categoryID if it doesn't exists via category name, else find the category name
	channels, err := s.GuildChannels(config.ServerID)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + ErrorLocation(err))
		if err != nil {
			return categoryID, categoryID
		}
		return categoryID, categoryName
	}
	for _, cha := range channels {
		if categoryID == "" {
			if cha.Type != discordgo.ChannelTypeGuildCategory {
				continue
			}
			if strings.ToLower(cha.Name) == strings.ToLower(category) {
				categoryID = cha.ID
				categoryName = cha.Name
				flag = true
				break
			}
		}
		if cha.ID == categoryID {
			categoryName = cha.Name
			flag = true
			break
		}
	}

	if !flag {
		return "", ""
	}

	return categoryID, categoryName
}

// Snowflake creation date calculator
func CreationTime(ID string) (t time.Time, err error) {
	i, err := strconv.ParseInt(ID, 10, 64)
	if err != nil {
		return
	}
	timestamp := (i >> 22) + 1420070400000
	t = time.Unix(timestamp/1000, 0)
	return
}

func Uptime() time.Duration {
	return time.Since(startTime)
}
