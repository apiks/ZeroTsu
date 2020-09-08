package entities

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"log"
	"sync"
)

// GuildInfo contains all the data a guild can contain
type GuildInfo struct {
	sync.RWMutex

	ID            string
	GuildSettings GuildSettings

	PunishedUsers       []PunishedUsers
	Filters             []Filter
	MessageRequirements []MessRequirement
	SpoilerRoles        []*discordgo.Role
	Feeds               []Feed
	FeedChecks          []FeedCheck
	Raffles             []*Raffle
	Waifus              []*Waifu
	WaifuTrades         []*WaifuTrade

	MemberInfoMap   map[string]UserInfo
	SpoilerMap      map[string]*discordgo.Role
	EmojiStats      map[string]Emoji
	ChannelStats    map[string]Channel
	UserChangeStats map[string]int
	VoteInfoMap     map[string]*VoteInfo
	TempChaMap      map[string]*TempChaInfo
	ReactJoinMap    map[string]*ReactJoin
	ExtensionList   map[string]string
	Autoposts       map[string]Cha
}

func (g *GuildInfo) SetID(id string) {
	g.Lock()
	g.ID = id
	g.Unlock()
}

func (g *GuildInfo) GetID() string {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return ""
	}
	return g.ID
}

func (g *GuildInfo) SetGuildSettings(guildSettings GuildSettings) {
	g.Lock()
	g.GuildSettings = guildSettings
	g.Unlock()
}

func (g *GuildInfo) GetGuildSettings() GuildSettings {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return GuildSettings{}
	}
	return g.GuildSettings
}

func (g *GuildInfo) AssignToPunishedUsers(index int, punishedUser PunishedUsers) {
	g.Lock()
	g.PunishedUsers[index] = punishedUser
	g.Unlock()
}

func (g *GuildInfo) AppendToPunishedUsers(punishedUser PunishedUsers) {
	g.Lock()
	g.PunishedUsers = append(g.PunishedUsers, punishedUser)
	g.Unlock()
}

func (g *GuildInfo) RemoveFromPunishedUsers(index int) {
	g.Lock()
	g.PunishedUsers = append(g.PunishedUsers[:index], g.PunishedUsers[index+1:]...)
	g.Unlock()
}

func (g *GuildInfo) SetPunishedUsers(punishedUsers []PunishedUsers) {
	g.Lock()
	g.PunishedUsers = punishedUsers
	g.Unlock()
}

func (g *GuildInfo) GetPunishedUsers() []PunishedUsers {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.PunishedUsers
}

func (g *GuildInfo) AssignToFilters(index int, filter Filter) {
	g.Lock()
	g.Filters[index] = filter
	g.Unlock()
}

func (g *GuildInfo) AppendToFilters(filter Filter) {
	g.Lock()
	g.Filters = append(g.Filters, filter)
	g.Unlock()
}

func (g *GuildInfo) RemoveFromFilters(index int) {
	g.Lock()
	g.Filters = append(g.Filters[:index], g.Filters[index+1:]...)
	g.Unlock()
}

func (g *GuildInfo) SetFilters(filters []Filter) {
	g.Lock()
	g.Filters = filters
	g.Unlock()
}

func (g *GuildInfo) GetFilters() []Filter {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.Filters
}

func (g *GuildInfo) AssignToMessageRequirements(index int, messageRequirement MessRequirement) {
	g.Lock()
	g.MessageRequirements[index] = messageRequirement
	g.Unlock()
}

func (g *GuildInfo) AppendToMessageRequirements(messageRequirement MessRequirement) {
	g.Lock()
	g.MessageRequirements = append(g.MessageRequirements, messageRequirement)
	g.Unlock()
}

func (g *GuildInfo) RemoveFromMessageRequirements(index int) {
	g.Lock()
	g.MessageRequirements = append(g.MessageRequirements[:index], g.MessageRequirements[index+1:]...)
	g.Unlock()
}

func (g *GuildInfo) SetMessageRequirements(messageRequirements []MessRequirement) {
	g.Lock()
	g.MessageRequirements = messageRequirements
	g.Unlock()
}

func (g *GuildInfo) GetMessageRequirements() []MessRequirement {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.MessageRequirements
}

func (g *GuildInfo) AppendToSpoilerRoles(spoilerRole *discordgo.Role) {
	g.Lock()
	g.SpoilerRoles = append(g.SpoilerRoles, spoilerRole)
	g.Unlock()
}

func (g *GuildInfo) RemoveFromSpoilerRoles(index int) {
	g.Lock()
	if index < len(g.SpoilerRoles)-1 {
		copy(g.SpoilerRoles[index:], g.SpoilerRoles[index+1:])
	}
	g.SpoilerRoles[len(g.SpoilerRoles)-1] = nil
	g.SpoilerRoles = g.SpoilerRoles[:len(g.SpoilerRoles)-1]
	g.Unlock()
}

func (g *GuildInfo) SetSpoilerRoles(spoilerRoles []*discordgo.Role) {
	g.Lock()
	g.SpoilerRoles = spoilerRoles
	g.Unlock()
}

func (g *GuildInfo) GetSpoilerRoles() []*discordgo.Role {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.SpoilerRoles
}

func (g *GuildInfo) AssignToFeeds(index int, feed Feed) {
	g.Lock()
	g.Feeds[index] = feed
	g.Unlock()
}

func (g *GuildInfo) AppendToFeeds(feed Feed) {
	g.Lock()
	g.Feeds = append(g.Feeds, feed)
	g.Unlock()
}

func (g *GuildInfo) RemoveFromFeeds(index int) {
	g.Lock()
	g.Feeds = append(g.Feeds[:index], g.Feeds[index+1:]...)
	g.Unlock()
}

func (g *GuildInfo) SetFeeds(feeds []Feed) {
	g.Lock()
	g.Feeds = feeds
	g.Unlock()
}

func (g *GuildInfo) GetFeeds() []Feed {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.Feeds
}

func (g *GuildInfo) AssignToFeedChecks(index int, feedCheck FeedCheck) {
	g.Lock()
	g.FeedChecks[index] = feedCheck
	g.Unlock()
}

func (g *GuildInfo) AppendToFeedChecks(feedCheck FeedCheck) {
	g.Lock()
	g.FeedChecks = append(g.FeedChecks, feedCheck)
	g.Unlock()
}

func (g *GuildInfo) RemoveFromFeedChecks(index int) {
	g.Lock()
	g.FeedChecks = append(g.FeedChecks[:index], g.FeedChecks[index+1:]...)
	g.Unlock()
}

func (g *GuildInfo) SetFeedChecks(feedChecks []FeedCheck) {
	g.Lock()
	g.FeedChecks = feedChecks
	g.Unlock()
}

func (g *GuildInfo) GetFeedChecks() []FeedCheck {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.FeedChecks
}

func (g *GuildInfo) AppendToRaffles(raffle *Raffle) {
	g.Lock()
	g.Raffles = append(g.Raffles, raffle)
	g.Unlock()
}

func (g *GuildInfo) RemoveFromRaffles(index int) {
	g.Lock()
	if index < len(g.Raffles)-1 {
		copy(g.Raffles[index:], g.Raffles[index+1:])
	}
	g.Raffles[len(g.Raffles)-1] = nil
	g.Raffles = g.Raffles[:len(g.Raffles)-1]
	g.Unlock()
}

func (g *GuildInfo) SetRaffles(raffles []*Raffle) {
	g.Lock()
	g.Raffles = raffles
	g.Unlock()
}

func (g *GuildInfo) GetRaffles() []*Raffle {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.Raffles
}

func (g *GuildInfo) AppendToWaifus(waifu Waifu) {
	g.Lock()
	g.Waifus = append(g.Waifus, &waifu)
	g.Unlock()
}

func (g *GuildInfo) RemoveFromWaifus(index int) {
	g.Lock()
	if index < len(g.Waifus)-1 {
		copy(g.Waifus[index:], g.Waifus[index+1:])
	}
	g.Waifus[len(g.Waifus)-1] = nil
	g.Waifus = g.Waifus[:len(g.Waifus)-1]
	g.Unlock()
}

func (g *GuildInfo) SetWaifus(waifus []*Waifu) {
	g.Lock()
	g.Waifus = waifus
	g.Unlock()
}

func (g *GuildInfo) GetWaifus() []*Waifu {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.Waifus
}

func (g *GuildInfo) AppendToWaifuTrades(waifuTrade *WaifuTrade) {
	g.Lock()
	g.WaifuTrades = append(g.WaifuTrades, waifuTrade)
	g.Unlock()
}

func (g *GuildInfo) RemoveFromWaifuTrades(index int) {
	g.Lock()
	if index < len(g.WaifuTrades)-1 {
		copy(g.WaifuTrades[index:], g.WaifuTrades[index+1:])
	}
	g.WaifuTrades[len(g.WaifuTrades)-1] = nil
	g.WaifuTrades = g.WaifuTrades[:len(g.WaifuTrades)-1]
	g.Unlock()
}

func (g *GuildInfo) SetWaifuTrades(waifuTrades []*WaifuTrade) {
	g.Lock()
	g.WaifuTrades = waifuTrades
	g.Unlock()
}

func (g *GuildInfo) GetWaifuTrades() []*WaifuTrade {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.WaifuTrades
}

func (g *GuildInfo) AssignToMemberInfoMap(key string, user UserInfo) {
	g.Lock()
	g.MemberInfoMap[key] = user
	g.Unlock()
}

func (g *GuildInfo) RemoveFromMemberInfoMap(key string) {
	g.Lock()
	delete(g.MemberInfoMap, key)
	g.Unlock()
}

func (g *GuildInfo) SetMemberInfoMap(memberInfo map[string]UserInfo) {
	g.Lock()
	g.MemberInfoMap = memberInfo
	g.Unlock()
}

func (g *GuildInfo) GetMemberInfoMap() map[string]UserInfo {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.MemberInfoMap
}

func (g *GuildInfo) SetSpoilerMap(spoilerMap map[string]*discordgo.Role) {
	g.Lock()
	g.SpoilerMap = spoilerMap
	g.Unlock()
}

func (g *GuildInfo) GetSpoilerMap() map[string]*discordgo.Role {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.SpoilerMap
}

func (g *GuildInfo) AssignToEmojiStats(key string, emoji Emoji) {
	g.Lock()
	g.EmojiStats[key] = emoji
	g.Unlock()
}

func (g *GuildInfo) RemoveFromEmojiStats(key string) {
	g.Lock()
	delete(g.EmojiStats, key)
	g.Unlock()
}

func (g *GuildInfo) SetEmojiStats(emojiStats map[string]Emoji) {
	g.Lock()
	g.EmojiStats = emojiStats
	g.Unlock()
}

func (g *GuildInfo) GetEmojiStats() map[string]Emoji {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.EmojiStats
}

func (g *GuildInfo) AssignToChannelStats(key string, channel Channel) {
	g.Lock()
	g.ChannelStats[key] = channel
	g.Unlock()
}

func (g *GuildInfo) RemoveFromChannelStats(key string) {
	g.Lock()
	delete(g.ChannelStats, key)
	g.Unlock()
}

func (g *GuildInfo) SetChannelStats(channelStats map[string]Channel) {
	g.Lock()
	g.ChannelStats = channelStats
	g.Unlock()
}

func (g *GuildInfo) GetChannelStats() map[string]Channel {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.ChannelStats
}

func (g *GuildInfo) AssignToUserChangeStats(key string, amount int) {
	g.Lock()
	g.UserChangeStats[key] = amount
	g.Unlock()
}

func (g *GuildInfo) AddToUserChangeStats(key string, amount int) {
	g.Lock()
	g.UserChangeStats[key] += amount
	g.Unlock()
}

func (g *GuildInfo) RemoveFromUserChangeStats(key string) {
	g.Lock()
	delete(g.UserChangeStats, key)
	g.Unlock()
}

func (g *GuildInfo) SetUserChangeStats(userChangeStats map[string]int) {
	g.Lock()
	g.UserChangeStats = userChangeStats
	g.Unlock()
}

func (g *GuildInfo) GetUserChangeStats() map[string]int {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.UserChangeStats
}

func (g *GuildInfo) SetVoteInfoMap(voteInfo map[string]*VoteInfo) {
	g.Lock()
	g.VoteInfoMap = voteInfo
	g.Unlock()
}

func (g *GuildInfo) GetVoteInfoMap() map[string]*VoteInfo {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.VoteInfoMap
}

func (g *GuildInfo) SetTempChaMap(tempChaMap map[string]*TempChaInfo) {
	g.Lock()
	g.TempChaMap = tempChaMap
	g.Unlock()
}

func (g *GuildInfo) GetTempChaMap() map[string]*TempChaInfo {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.TempChaMap
}

func (g *GuildInfo) AssignToReactJoinMap(key string, reactJoin *ReactJoin) {
	g.Lock()
	g.ReactJoinMap[key] = reactJoin
	g.Unlock()
}

func (g *GuildInfo) RemoveFromReactJoinMap(key string) {
	g.Lock()
	delete(g.ReactJoinMap, key)
	g.Unlock()
}

func (g *GuildInfo) SetReactJoinMap(reactJoinMap map[string]*ReactJoin) {
	g.Lock()
	g.ReactJoinMap = reactJoinMap
	g.Unlock()
}

func (g *GuildInfo) GetReactJoinMap() map[string]*ReactJoin {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.ReactJoinMap
}

func (g *GuildInfo) SetExtensionList(extensionList map[string]string) {
	g.Lock()
	g.ExtensionList = extensionList
	g.Unlock()
}

func (g *GuildInfo) GetExtensionList() map[string]string {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.ExtensionList
}

func (g *GuildInfo) AssignToAutoposts(key string, autopost Cha) {
	g.Lock()
	g.Autoposts[key] = autopost
	g.Unlock()
}

func (g *GuildInfo) RemoveFromAutoposts(key string) {
	g.Lock()
	delete(g.Autoposts, key)
	g.Unlock()
}

func (g *GuildInfo) SetAutoposts(autoposts map[string]Cha) {
	g.Lock()
	g.Autoposts = autoposts
	g.Unlock()
}

func (g *GuildInfo) GetAutoposts() map[string]Cha {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.Autoposts
}

// Load loads a guild file into the guild memory
func (g *GuildInfo) Load(file, guildID string) error {
	fileData, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/%s", DBPath, guildID, file))
	if err != nil {
		return err
	}
	if len(fileData) == 0 {
		return nil
	}

	g.Lock()
	defer g.Unlock()

	switch file {
	case "guildSettings.json":
		return json.Unmarshal(fileData, &g.GuildSettings)
	case "punishedUsers.json":
		return json.Unmarshal(fileData, &g.PunishedUsers)
	case "filters.json":
		return json.Unmarshal(fileData, &g.Filters)
	case "messReqs.json":
		return json.Unmarshal(fileData, &g.MessageRequirements)
	case "spoilerRoles.json":
		err = json.Unmarshal(fileData, &g.SpoilerRoles)
		if err != nil {
			return err
		}
		// Fills spoilerMap with roles from the spoilerRoles.json file if latter is not empty
		for i := 0; i < len(g.SpoilerRoles); i++ {
			g.SpoilerMap[g.SpoilerRoles[i].ID] = g.SpoilerRoles[i]
		}
		return nil
	case "rssThreads.json":
		_ = json.Unmarshal(fileData, &g.Feeds)
		return nil
	case "rssThreadCheck.json":
		return json.Unmarshal(fileData, &g.FeedChecks)
	case "raffles.json":
		return json.Unmarshal(fileData, &g.Raffles)
	case "waifus.json":
		return json.Unmarshal(fileData, &g.Waifus)
	case "waifuTrades.json":
		return json.Unmarshal(fileData, &g.WaifuTrades)
	case "memberInfo.json":
		return json.Unmarshal(fileData, &g.MemberInfoMap)
	case "emojiStats.json":
		return json.Unmarshal(fileData, &g.EmojiStats)
	case "channelStats.json":
		return json.Unmarshal(fileData, &g.ChannelStats)
	case "userChangeStats.json":
		return json.Unmarshal(fileData, &g.UserChangeStats)
	case "voteInfo.json":
		return json.Unmarshal(fileData, &g.VoteInfoMap)
	case "tempCha.json":
		return json.Unmarshal(fileData, &g.TempChaMap)
	case "reactJoin.json":
		return json.Unmarshal(fileData, &g.ReactJoinMap)
	case "extensionList.json":
		return json.Unmarshal(fileData, &g.ExtensionList)
	case "autoposts.json":
		return json.Unmarshal(fileData, &g.Autoposts)
	}

	return nil
}

// WriteData writes some kind of guild data to the target guild file
func (g *GuildInfo) WriteData(fileName string, data interface{}) {
	g.RLock()
	marshaledData, err := json.MarshalIndent(&data, "", "    ")
	if err != nil {
		g.RUnlock()
		log.Println(err)
		return
	}
	g.RUnlock()
	if len(marshaledData) == 0 {
		return
	}

	err = ioutil.WriteFile(fmt.Sprintf(DBPath+"/%s/%s.json", g.ID, fileName), marshaledData, 0644)
	if err != nil {
		log.Println(err)
	}
}
