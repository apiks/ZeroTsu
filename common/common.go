package common

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/mmcdole/gofeed"
	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"

	"github.com/bwmarrin/discordgo"
	"github.com/vartanbeno/go-reddit/reddit"
)

// File for misc. functions, commands and variables.

const UserAgent = "script:github.com/apiks/zerotsu:v3.5 (by /u/thechosenapiks)"

var StartTime time.Time

// SortRoleByAlphabet sorts roles alphabetically
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

// Sorts channels alphabetically
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
	http.Transport
	http.RoundTripper
}

func (c *UserAgentTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("User-Agent", UserAgent)
	return c.RoundTripper.RoundTrip(r)
}

// ResolveTimeFromString resolves a time (usually for unbanning) from a given string formatted #w#d#h#m.
// This returns current time + delay.
// If no time is added to the offset, then this returns true for permanent.
// By Kagumi. Modified by Apiks
func ResolveTimeFromString(given string) (ret time.Time, perma bool, err error) {
	ret = time.Now()
	comp := ret

	// Match numbers followed by letters (w, d, h, m)
	matcher := regexp.MustCompile(`(\d+)([wdhmWDHM])`)
	groups := matcher.FindAllStringSubmatch(given, -1)

	// If no valid matches, return an error
	if len(groups) == 0 {
		err = fmt.Errorf("invalid time format: %s", given)
		return
	}

	for _, match := range groups {
		if len(match) < 3 {
			continue
		}

		val, convErr := strconv.Atoi(match[1])
		if convErr != nil {
			continue
		}

		switch strings.ToLower(match[2]) {
		case "w":
			ret = ret.AddDate(0, 0, val*7)
		case "d":
			ret = ret.AddDate(0, 0, val)
		case "h":
			ret = ret.Add(time.Hour * time.Duration(val))
		case "m":
			ret = ret.Add(time.Minute * time.Duration(val))
		default:
			err = fmt.Errorf("unrecognized time unit: %s", match[2])
			return
		}
	}

	// If no changes were made, mark as permanent
	if ret.Equal(comp) {
		perma = true
	}

	return
}

// GetUserID resolves a userID from a userID, Mention or username#discrim
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

// ChannelParser parses a string for a channel and returns its ID and name
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

// RoleParser parses a string for a role and returns its ID and name
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

func Uptime() time.Duration {
	return time.Since(StartTime)
}

func WeekStart(year, week int) time.Time {
	// Start from the middle of the year:
	t := time.Date(year, 7, 1, 0, 0, 0, 0, time.UTC)

	// Roll back to Monday:
	if wd := t.Weekday(); wd == time.Sunday {
		t = t.AddDate(0, 0, -6)
	} else {
		t = t.AddDate(0, 0, -int(wd)+1)
	}

	// Difference in weeks:
	_, w := t.ISOWeek()
	t = t.AddDate(0, 0, (week-w)*7)

	return t
}

// GetRedditRSSFeed parses a reddit uri for an RSS feed and returns it
func GetRedditRSSFeed(uri string, retryTimes int) (*gofeed.Feed, int, error) {
	var (
		fp         = gofeed.NewParser()
		httpClient = http.Client{
			Transport: &http.Transport{
				TLSNextProto: map[string]func(authority string, c *tls.Conn) http.RoundTripper{},
			},
			Timeout: 10 * time.Second,
		}
		credentials = reddit.Credentials{ID: config.RedditID, Secret: config.RedditSecret, Username: config.RedditUsername, Password: config.RedditPassword}
	)
	client, err := reddit.NewClient(&httpClient, &credentials)
	if err != nil {
		return nil, 500, err
	}

	req, err := client.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, 500, err
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := client.Do(context.TODO(), req, nil)
	if err != nil {
		if resp != nil {
			return nil, resp.StatusCode, err
		}
		return nil, 500, err
	}
	defer resp.Body.Close()

	// Use ratelimit headers to retry x times after x time
	if resp.StatusCode == 429 && retryTimes > 0 {
		retryAfter := resp.Header.Get("x-ratelimit-reset")
		retryAfterInt, err := strconv.Atoi(retryAfter)
		if err != nil {
			retryAfter = resp.Header.Get("retry-after")
			retryAfterInt, err = strconv.Atoi(retryAfter)
			if err != nil {
				return nil, 500, err
			}
		}
		if retryAfterInt == 0 {
			return nil, resp.StatusCode, err
		}
		time.Sleep(time.Duration(retryAfterInt) * time.Second)
		return GetRedditRSSFeed(uri, retryTimes-1)
	}

	if resp.StatusCode != 200 {
		return nil, resp.StatusCode, errors.New(resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 500, err
	}
	feed, err := fp.ParseString(string(body))
	if err != nil {
		return nil, 500, err
	}

	return feed, resp.StatusCode, nil
}
