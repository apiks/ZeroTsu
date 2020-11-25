package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"github.com/r-anime/ZeroTsu/functionality"
	"io"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/bwmarrin/discordgo"
)

// File for misc. functions, commands and variables.

const (
	UserAgent       = "script:github.com/apiks/zerotsu:v1.1.0 (by /u/thechosenapiks)"
	ShortDateFormat = "2006-01-02"
	LongDateFormat  = "2006-01-02 15:04:05.999999999 -0700 MST"
)

var (
	FullSpoilerPerms = discordgo.PermissionSendMessages + discordgo.PermissionReadMessages + discordgo.PermissionReadMessageHistory
	ReadSpoilerPerms = discordgo.PermissionReadMessages + discordgo.PermissionReadMessageHistory

	StartTime time.Time

	NilTime = time.Time{}
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
func ListenForDeletedRoleHandler(_ *discordgo.Session, g *discordgo.GuildRoleDelete) {
	entities.HandleNewGuild(g.GuildID)
	db.SetGuildSpoilerRole(g.GuildID, &discordgo.Role{ID: g.RoleID}, true)
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
		return "", fmt.Errorf("Error: No @user, userID or username#discrim detected.")
	}

	// Pulls the userArgument from the second parameter
	userArgument := strings.ToLower(messageSlice[1])

	// Handles "me" parameter on whois
	if strings.ToLower(userArgument) == "me" {
		userArgument = m.Author.ID
	}

	// Handles userArgument if it was in reddit username format
	if strings.Contains(userArgument, "/u/") || strings.Contains(userArgument, "u/") {
		var exists bool
		userArgument = strings.TrimPrefix(strings.TrimPrefix(userArgument, "/u/"), "u/")

		guildMemberInfo := db.GetGuildMemberInfo(m.GuildID)
		for _, user := range guildMemberInfo {
			if strings.ToLower(user.GetRedditUsername()) == userArgument {
				userArgument = user.GetID()
				exists = true
				break
			}
		}

		if !exists {
			return userArgument, fmt.Errorf("Error: This reddit user is not in the internal database. Try using an ID.")
		}
	}

	// Handles userArgument if it was username#discrim format
	if strings.Contains(userArgument, "#") {
		var exists bool
		splitUser := strings.SplitN(userArgument, "#", 2)
		if len(splitUser) < 2 {
			return userArgument, fmt.Errorf("Error: Invalid user. You're trying to username#discrim with spaces in the username." +
				" This command does not support that. Please use a valid ID instead.")
		}

		guildMemberInfo := db.GetGuildMemberInfo(m.GuildID)
		for _, user := range guildMemberInfo {
			if strings.ToLower(user.GetUsername()) == splitUser[0] && user.GetDiscrim() == splitUser[1] {
				userArgument = user.GetID()
				exists = true
				break
			}
		}

		if !exists {
			return userArgument, fmt.Errorf("Error: This username#discrim value is not in the internal database. Try using an ID.")
		}
	}

	// Trims fluff if it was a mention. Otherwise check if it's a correct user ID
	if strings.Contains(messageSlice[1], "<@") {
		userArgument = strings.TrimPrefix(userArgument, "<@")
		userArgument = strings.TrimPrefix(userArgument, "!")
		userArgument = strings.TrimSuffix(userArgument, ">")
	}

	_, err := strconv.ParseInt(userArgument, 10, 64)
	if len(userArgument) < 17 || err != nil {
		return userArgument, fmt.Errorf("Error: Cannot parse user.")
	}

	return userArgument, nil
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
func CommandErrorHandler(s *discordgo.Session, m *discordgo.Message, botLog entities.Cha, err error) {
	_, err = s.ChannelMessageSend(m.ChannelID, err.Error())
	if err != nil {
		if botLog == (entities.Cha{}) || botLog.GetID() == "" {
			return
		}
		if _, ok := err.(*discordgo.RESTError); ok && err.(*discordgo.RESTError).Response.Status == "500: Internal Server Error" {
			return
		}

		_, _ = s.ChannelMessageSend(botLog.GetID(), err.Error())
	}
}

// Logs the error in the guild BotLog
func LogError(s *discordgo.Session, botLog entities.Cha, err error) {
	if botLog == (entities.Cha{}) || botLog.GetID() == "" {
		return
	}

	if restErr, ok := err.(*discordgo.RESTError); ok {
		if restErr.Message.Message == "500: Internal Server Error" {
			return
		}
	}

	_, _ = s.ChannelMessageSend(botLog.GetID(), err.Error())
}

// SplitLongMessage takes a message and splits it if it's longer than 1900
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
	// If a user has the requested role, Add +1 to users var
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
			if len(userMentionCheck) == 0 {
				return m
			}

			mem := db.GetGuildMember(guildID, userID)

			for _, mention := range userMentionCheck {
				if len(entities.Guilds.DB[guildID].GetMemberInfoMap()) != 0 {
					continue
				}

				userID = strings.TrimPrefix(mention, "<@")
				userID = strings.TrimPrefix(userID, "!")
				userID = strings.TrimSuffix(userID, ">")

				// Checks first in memberInfo. Only checks serverside if it doesn't exist. Saves performance
				if mem.GetID() != "" && mem.GetNickname() != "" {
					mentions += " " + strings.ToLower(mem.GetNickname())
					continue
				}

				// If user wasn't found in memberInfo then fetch manually from Discord and add to memberInfo
				user, err := s.State.Member(guildID, userID)
				if err != nil {
					user, err = s.GuildMember(guildID, userID)
					if err != nil {
						continue
					}
				}
				functionality.InitializeUser(user.User, guildID)

				m = strings.Replace(m, mention, fmt.Sprintf("@%s", user.Nick), -1)
			}
		}
	}

	// Checks for channel and replaces mention with channel name
	if strings.Contains(m, "#") {
		channelMentionRegex := regexp.MustCompile(`(?m)(<#\d+>)`)
		channelMentionCheck = channelMentionRegex.FindAllString(m, -1)
		if channelMentionCheck != nil {
			if len(channelMentionCheck) == 0 {
				return m
			}

			for _, mention := range channelMentionCheck {
				channelID := strings.TrimPrefix(mention, "<#")
				channelID = strings.TrimSuffix(channelID, ">")

				// Fetches channel so we can parse its string name
				cha, err := s.State.Channel(channelID)
				if err != nil {
					cha, err = s.Channel(channelID)
					if err != nil {
						continue
					}
				}
				m = strings.Replace(m, mention, fmt.Sprintf("#%s", cha.Name), -1)
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
		guildSettings := db.GetGuildSettings(guildID)
		LogError(s, guildSettings.BotLog, err)
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
		guildSettings := db.GetGuildSettings(guildID)
		LogError(s, guildSettings.BotLog, err)
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

	// Get role if its a tag
	if strings.HasPrefix(role, "<@") {
		roleID = strings.TrimPrefix(role, "<@&")
		roleID = strings.TrimPrefix(roleID, "<@\u200B&")
		roleID = strings.TrimSuffix(roleID, ">")
	}

	// Find the roleID if it doesn't exists via role name, else find the role name
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		guildSettings := db.GetGuildSettings(guildID)
		LogError(s, guildSettings.BotLog, err)
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

// OptInsExist checks whether optin roles exist
func OptInsExist(s *discordgo.Session, guildID string) bool {
	var (
		optInUnderExists bool
		optInAboveExists bool
	)

	guildSettings := db.GetGuildSettings(guildID)

	// Saves guild roles
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		return false
	}

	// Checks if optins exist
	if guildSettings.GetOptInUnder() != (entities.Role{}) {
		if guildSettings.GetOptInUnder().GetID() != "" {
			for _, role := range roles {
				if role.ID == guildSettings.GetOptInUnder().GetID() {
					optInUnderExists = true
					break
				}
			}
		}
	}

	if guildSettings.GetOptInAbove() != (entities.Role{}) {
		if guildSettings.GetOptInAbove().GetID() != "" {
			for _, role := range roles {
				if role.ID == guildSettings.GetOptInAbove().GetID() {
					optInAboveExists = true
					break
				}
			}
		}
	}

	if optInUnderExists && optInAboveExists {
		return true
	}

	return false
}

// Checks if optins exist and creates them if they don't
func OptInsHandler(s *discordgo.Session, channelID, guildID string) error {
	var (
		optInUnderExists bool
		optInAboveExists bool
	)

	// Saves guild roles
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		return err
	}

	guildSettings := db.GetGuildSettings(guildID)

	// Checks if optins exist
	if guildSettings.GetOptInUnder() != (entities.Role{}) {
		if guildSettings.GetOptInUnder().GetID() != "" {
			for _, role := range roles {
				if role.ID == guildSettings.GetOptInUnder().GetID() {
					optInUnderExists = true
					break
				}
			}
		}
	}

	if guildSettings.GetOptInAbove() != (entities.Role{}) {
		if guildSettings.GetOptInAbove().GetID() != "" {
			for _, role := range roles {
				if role.ID == guildSettings.GetOptInAbove().GetID() {
					optInAboveExists = true
					break
				}
			}
		}
	}

	if optInUnderExists && optInAboveExists {
		return nil
	}

	// Handles opt-in-under
	if !optInUnderExists {
		var optIn entities.Role

		_, err := s.ChannelMessageSend(channelID, "Necessary opt-in-under role not detected. Trying to create it.")
		if err != nil {
			if guildSettings.BotLog != (entities.Cha{}) {
				if guildSettings.BotLog.GetID() != "" {
					_, _ = s.ChannelMessageSend(guildSettings.BotLog.GetID(), "Necessary opt-in-under role not detected. Trying to create it.")
				}
			}
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
		optIn = optIn.SetID(role.ID)
		optIn = optIn.SetName(role.Name)
		optIn = optIn.SetPosition(5)

		// Saves the new opt-in guild data
		guildSettings = guildSettings.SetOptInUnder(optIn)
	}
	// Handles opt-in-above
	if !optInAboveExists {
		var optIn entities.Role

		_, err := s.ChannelMessageSend(channelID, "Necessary opt-in-above role not detected. Trying to create it.")
		if err != nil {
			if guildSettings.BotLog != (entities.Cha{}) {
				if guildSettings.BotLog.GetID() != "" {
					_, _ = s.ChannelMessageSend(guildSettings.BotLog.GetID(), "Necessary opt-in-above role not detected. Trying to create it.")
				}
			}
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
		optIn = optIn.SetID(role.ID)
		optIn = optIn.SetName(role.Name)
		optIn = optIn.SetPosition(2)

		// Saves the new opt-in guild data
		guildSettings = guildSettings.SetOptInAbove(optIn)
	}

	// Reorders the optin roles with space inbetween them
	deb, err := s.GuildRoles(guildID)
	if err != nil {
		return err
	}

	for i, role := range deb {
		if role.ID == guildSettings.GetOptInUnder().GetID() {
			deb[i].Position = guildSettings.GetOptInUnder().GetPosition()
		}
		if role.ID == guildSettings.GetOptInAbove().GetID() {
			deb[i].Position = guildSettings.GetOptInAbove().GetPosition()
		}
	}

	_, err = s.GuildRoleReorder(guildID, deb)
	if err != nil {
		return err
	}

	db.SetGuildSettings(guildID, guildSettings)

	return err
}

// Encrypt string to base64 crypto using AES
func Encrypt(key []byte, text string) string {
	// key := []byte(keyText)
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println(err)
		return ""
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Println(err)
		return ""
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// convert to base64
	return base64.URLEncoding.EncodeToString(ciphertext)
}

// Decrypt from base64 to decrypted string
func Decrypt(key []byte, cryptoText string) (string, bool) {
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println(err)
		return "", false
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		log.Println("ciphertext too short")
		return "", false
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)

	return fmt.Sprintf("%s", ciphertext), true
}
