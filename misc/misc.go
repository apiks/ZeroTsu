package misc

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/bwmarrin/discordgo"
)

// File for misc. functions, commands and variables.

const (
	UserAgent         = "script:github.com/r-anime/zerotsu:v1.0.0 (by /u/thechosenapiks, /u/geo1088)"
	DateFormat        = "2006-01-02"
)

var (
	SpoilerPerms = discordgo.PermissionSendMessages + discordgo.PermissionReadMessages + discordgo.PermissionReadMessageHistory

	StartTime time.Time
)

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

	if _, ok := GuildMap[g.GuildID]; !ok {
		InitDB(g.GuildID)
		LoadGuilds()
	}

	MapMutex.Lock()
	if GuildMap[g.GuildID].SpoilerMap[g.RoleID] == nil {
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
func GetUserID(m *discordgo.Message, messageSlice []string) (string, error) {

	if len(messageSlice) < 2 {
		return "", fmt.Errorf("Error: No @user, userID or username#discrim detected")
	}

	// Pulls the userID from the second parameter
	userID := messageSlice[1]

	// Handles "me" string on whois
	if strings.ToLower(userID) == "me" {
		userID = m.Author.ID
	}
	// Handles userID if it was in reddit username format
	if strings.Contains(userID, "/u/") {
		exists := false
		userID = strings.TrimPrefix(userID, "/u/")
		MapMutex.Lock()
		for _, user := range GuildMap[m.GuildID].MemberInfoMap {
			if strings.ToLower(user.RedditUsername) == userID {
				userID = user.ID
				exists = true
				break
			}
		}
		MapMutex.Unlock()

		if !exists {
			return userID, fmt.Errorf("Error: This reddit user is not in the internal database. Cannot whois")
		}
	}
	if strings.Contains(userID, "u/") {
		exists := false
		userID = strings.TrimPrefix(userID, "u/")
		MapMutex.Lock()
		for _, user := range GuildMap[m.GuildID].MemberInfoMap {
			if strings.ToLower(user.RedditUsername) == userID {
				userID = user.ID
				exists = true
				break
			}
		}
		MapMutex.Unlock()

		if !exists {
			return userID, fmt.Errorf("Error: This reddit user is not in the internal database. Cannot whois")
		}
	}
	// Handles userID if it was username#discrim format
	if strings.Contains(userID, "#") {
		splitUser := strings.SplitN(userID, "#", 2)
		if len(splitUser) < 2 {
			return userID, fmt.Errorf("Error: Invalid user. You're trying to username#discrim with spaces in the username." +
				" This command does not support that. Please use an ID")
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
	_, err := strconv.ParseInt(userID, 10, 64)
	if len(userID) < 17 || err != nil {
		return userID, fmt.Errorf("Error: Cannot parse user")
	}
	return userID, nil
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
func CommandErrorHandler(s *discordgo.Session, m *discordgo.Message, err error, botLogID string) {
	_, err = s.ChannelMessageSend(m.ChannelID, err.Error())
	if err != nil {
		if botLogID == "" {
			return
		}
		_, _ = s.ChannelMessageSend(botLogID, err.Error())
	}
}

// SplitLongMessage takes a message and splits it if it's longer than 1900. By Kagumi
func SplitLongMessage(message string) (split []string) {
	const maxLength = 1900
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

// Returns a string that shows where the error occurred exactly
func ErrorLocation(err error) string {
	_, file, line, _ := runtime.Caller(1)
	errorLocation := fmt.Sprintf("Error is in file [%v] near line %v", file, line)
	return errorLocation
}

// Finds out how many users have the role and returns that number
func GetRoleUserAmount(guild *discordgo.Guild, roles []*discordgo.Role, roleName string) int {

	var (
		users  int
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
		flag           bool
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
						log.Println("in getBannedUsers date err")
						log.Println(err)
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

// Checks if a message contains a channel or user mentions and changes them to a non-mention if true
func MentionParser(s *discordgo.Session, m string, guildID string) string {

	var (
		mentions            string
		userID              string
		userMentionCheck    []string
		channelMentionCheck []string
	)

	if strings.Contains(m, "<@") {

		// Checks for both <@! and <@ mentions
		mentionRegex := regexp.MustCompile(`(?m)<@!?\d+>`)
		userMentionCheck = mentionRegex.FindAllString(m, -1)
		if userMentionCheck != nil {
			MapMutex.Lock()
			for i := range userMentionCheck {
				userID = strings.TrimPrefix(userMentionCheck[i], "<@")
				userID = strings.TrimPrefix(userID, "!")
				userID = strings.TrimSuffix(userID, ">")

				// Checks first in memberInfo. Only checks serverside if it doesn't exist. Saves performance
				if len(GuildMap[guildID].MemberInfoMap) != 0 {
					if _, ok := GuildMap[guildID].MemberInfoMap[userID]; ok {
						mentions += " " + strings.ToLower(GuildMap[guildID].MemberInfoMap[userID].Nickname)
						continue
					}
				}

				// If user wasn't found in memberInfo with that username+discrim combo then fetch manually from Discord and then replace mentions with nick
				user, err := s.State.Member(guildID, userID)
				if err != nil {
					user, _ = s.GuildMember(guildID, userID)
				}
				if user != nil {
					m = strings.Replace(m, userMentionCheck[i], fmt.Sprintf("@%v", user.Nick), -1)
				}
			}
			MapMutex.Unlock()
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
func ChannelParser(s *discordgo.Session, channel string, guildID string) (string, string) {
	var (
		channelID   string
		channelName string
		flag        bool
	)

	// If it's a channel ping remove <# and > from it to get the channel ID
	if strings.Contains(channel, "#") {
		channelID = strings.TrimPrefix(channel, "<#")
		channelID = strings.TrimSuffix(channelID, ">")
	}

	// Check if it's an ID by length and save the ID if so
	_, err := strconv.Atoi(channel)
	if len(channel) >= 17 && err == nil {
		channelID = channel
	}

	// Find the channelID if it doesn't exists via channel name, else find the channel name
	channels, err := s.GuildChannels(guildID)
	if err != nil {

		MapMutex.Lock()
		guildBotLog := GuildMap[guildID].GuildConfig.BotLog.ID
		MapMutex.Unlock()

		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
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
func CategoryParser(s *discordgo.Session, category string, guildID string) (string, string) {
	var (
		categoryID   string
		categoryName string
		flag         bool
	)

	// Check if it's an ID by length and save the ID if so
	_, err := strconv.Atoi(category)
	if len(category) >= 17 && err == nil {
		categoryID = category
	}

	// Find the categoryID if it doesn't exists via category name, else find the category name
	channels, err := s.GuildChannels(guildID)
	if err != nil {

		MapMutex.Lock()
		guildBotLog := GuildMap[guildID].GuildConfig.BotLog.ID
		MapMutex.Unlock()

		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
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

// Parses a string for a role and returns its ID and name
func RoleParser(s *discordgo.Session, role string, guildID string) (string, string) {
	var (
		roleID   string
		roleName string
		flag     bool
	)

	// Check if it's an ID by length and save the ID if so
	_, err := strconv.Atoi(role)
	if len(role) >= 17 && err == nil {
		roleID = role
	}

	// Find the roleID if it doesn't exists via role name, else find the role name
	roles, err := s.GuildRoles(guildID)
	if err != nil {

		MapMutex.Lock()
		guildBotLog := GuildMap[guildID].GuildConfig.BotLog.ID
		MapMutex.Unlock()

		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+ErrorLocation(err))
		if err != nil {
			return roleID, roleName
		}
		return roleID, roleName
	}
	for _, roleIteration := range roles {
		if roleID == "" {
			if strings.ToLower(roleIteration.Name) == strings.ToLower(role) {
				roleID = roleIteration.ID
				roleName = roleIteration.Name
				flag = true
				break
			}
		}
		if roleIteration.ID == roleID {
			roleName = roleIteration.Name
			flag = true
			break
		}
	}

	if !flag {
		return "", ""
	}

	return roleID, roleName
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
	return time.Since(StartTime)
}

// Checks if optins exist and creates them if they don't
func OptInsHandler(s *discordgo.Session, channelID, guildID string) error {

	var (
		optInUnderExists bool
		optInAboveExists bool
		err              error
	)

	// Saves guild roles
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		return err
	}

	MapMutex.Lock()
	guildBotLog := GuildMap[guildID].GuildConfig.BotLog.ID

	// Checks if optins exist
	if GuildMap[guildID].GuildConfig.OptInUnder.ID != "" {
		for _, role := range roles {
			if role.ID == GuildMap[guildID].GuildConfig.OptInUnder.ID {
				optInUnderExists = true
				break
			}
		}
	}
	if GuildMap[guildID].GuildConfig.OptInAbove.ID != "" {
		for _, role := range roles {
			if role.ID == GuildMap[guildID].GuildConfig.OptInAbove.ID {
				optInAboveExists = true
				break
			}
		}
	}

	MapMutex.Unlock()

	if optInUnderExists && optInAboveExists {
		return err
	}

	// Handles opt-in-under
	if !optInUnderExists {
		var optIn OptinRole

		_, err := s.ChannelMessageSend(channelID, "Necessary opt-in-under role not detected. Trying to create it.")
		if err != nil {
			_, _ = s.ChannelMessageSend(guildBotLog, "Necessary opt-in-under role not detected. Trying to create it.")
		}

		// Creates opt-in-under role
		role, err := s.GuildRoleCreate(guildID)
		if err != nil {
			return fmt.Errorf("Error: Could not create necessary opt-in roles. Please make sure I have role creation permissions.")
		}
		// Edits the new role
		role, err = s.GuildRoleEdit(guildID, role.ID, "opt-in-under/DO-NOT-DELETE", 65280, false, 0, false)
		if err != nil {
			return fmt.Errorf("Error: Could not edit the new opt-in role. Please make sure I have the necessary role edit permissions.")
		}

		// Sets values
		optIn.ID = role.ID
		optIn.Name = role.Name
		optIn.Position = 5

		// Saves the new opt-in guild data
		MapMutex.Lock()
		GuildMap[guildID].GuildConfig.OptInUnder = optIn
		MapMutex.Unlock()
	}
	// Handles opt-in-above
	if !optInAboveExists {
		var optIn OptinRole

		_, err := s.ChannelMessageSend(channelID, "Necessary opt-in-above role not detected. Trying to create it.")
		if err != nil {
			_, _ = s.ChannelMessageSend(guildBotLog, "Necessary opt-in-above role not detected. Trying to create it.")
		}

		// Creates opt-in-above role
		role, err := s.GuildRoleCreate(guildID)
		if err != nil {
			return fmt.Errorf("Error: Could not create necessary opt-in roles. Please make sure I have role creation permissions.")
		}
		// Edits the new role
		role, err = s.GuildRoleEdit(guildID, role.ID, "opt-in-above/DO-NOT-DELETE", 65280, false, 0, false)
		if err != nil {
			return fmt.Errorf("Error: Could not edit the new opt-in role. Please make sure I have the necessary role edit permissions.")
		}

		// Sets values
		optIn.ID = role.ID
		optIn.Name = role.Name
		optIn.Position = 2

		// Saves the new opt-in guild data
		MapMutex.Lock()
		GuildMap[guildID].GuildConfig.OptInAbove = optIn
		MapMutex.Unlock()
	}

	// Reorders the optin roles with space inbetween them
	deb, err := s.GuildRoles(guildID)
	if err != nil {
		return err
	}
	MapMutex.Lock()
	for i, role := range deb {
		if role.ID == GuildMap[guildID].GuildConfig.OptInUnder.ID {
			deb[i].Position = GuildMap[guildID].GuildConfig.OptInUnder.Position
		}
		if role.ID == GuildMap[guildID].GuildConfig.OptInAbove.ID {
			deb[i].Position = GuildMap[guildID].GuildConfig.OptInAbove.Position
		}
	}
	_, err = s.GuildRoleReorder(guildID, deb)
	if err != nil {
		MapMutex.Unlock()
		return err
	}
	GuildSettingsWrite(GuildMap[guildID].GuildConfig, guildID)
	MapMutex.Unlock()

	return err
}

// Replaces all instances of spaces in a string with hyphens
func RemoveSpaces(str string) string {
	return strings.Replace(str, " ", "-", -1)
}

// Replaces all instances of hyphens in a string with spaces
func RemoveHyphens(str string) string {
	return strings.Replace(str, "-", " ", -1)
}
