package entities

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/config"
	"io/ioutil"
	"log"
	"sync"
)

// GuildInfo contains all the data a guild can contain
type GuildInfo struct {
	sync.RWMutex

	ID            string
	GuildSettings GuildSettings

	PunishedUsers       punishedUsersSliceSafe
	Filters             filtersSliceSafe
	MessageRequirements messageRequirementsSliceSafe
	SpoilerRoles        spoilerRolesSliceSafe
	Feeds               feedsSliceSafe
	FeedChecks          feedChecksSliceSafe
	Raffles             rafflesSliceSafe
	Waifus              waifusSliceSafe
	WaifuTrades         waifuTradesSliceSafe

	MemberInfoMap   memberInfoMapSafe
	SpoilerMap      spoilerMapSafe
	EmojiStats emojiStatsSafe
	ChannelStats    channelStatsSafe
	UserChangeStats stringIntMapSafe
	VerifiedStats   stringIntMapSafe
	VoteInfoMap     voteInfoMapSafe
	TempChaMap      tempChaMapSafe
	ReactJoinMap    reactJoinMapSafe
	ExtensionList   stringStringMapSafe
	Autoposts autopostsMapSafe
}

type punishedUsersSliceSafe struct {
	sync.RWMutex
	punishedUsers []PunishedUsers
}

type filtersSliceSafe struct {
	sync.RWMutex
	filters []Filter
}

type messageRequirementsSliceSafe struct {
	sync.RWMutex
	messageRequirements []MessRequirement
}

type spoilerRolesSliceSafe struct {
	sync.RWMutex
	spoilerRoles []*discordgo.Role
}

type feedsSliceSafe struct {
	sync.RWMutex
	feeds []Feed
}

type feedChecksSliceSafe struct {
	sync.RWMutex
	feedChecks []FeedCheck
}

type rafflesSliceSafe struct {
	sync.RWMutex
	raffles []*Raffle
}

type waifusSliceSafe struct {
	sync.RWMutex
	waifus []*Waifu
}

type waifuTradesSliceSafe struct {
	sync.RWMutex
	waifuTrades []*WaifuTrade
}

type memberInfoMapSafe struct {
	sync.RWMutex
	memberInfo map[string]UserInfo
}

type spoilerMapSafe struct {
	sync.RWMutex
	spoilerMap map[string]*discordgo.Role
}

type emojiStatsSafe struct {
	sync.RWMutex
	emojiStats map[string]Emoji
}

type channelStatsSafe struct {
	sync.RWMutex
	channelStats map[string]Channel
}

type stringIntMapSafe struct {
	sync.RWMutex
	stringIntMap map[string]int
}

type voteInfoMapSafe struct {
	sync.RWMutex
	voteInfo map[string]*VoteInfo
}

type tempChaMapSafe struct {
	sync.RWMutex
	tempCha map[string]*TempChaInfo
}

type reactJoinMapSafe struct {
	sync.RWMutex
	reactJoin map[string]*ReactJoin
}

type stringStringMapSafe struct {
	sync.RWMutex
	stringStringMap map[string]string
}

type autopostsMapSafe struct {
	sync.RWMutex
	autoposts map[string]Cha
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
	g.PunishedUsers.Lock()
	g.PunishedUsers.punishedUsers[index] = punishedUser
	g.PunishedUsers.Unlock()
	g.Unlock()
}

func (g *GuildInfo) AppendToPunishedUsers(punishedUser PunishedUsers) {
	g.Lock()
	g.PunishedUsers.Lock()
	g.PunishedUsers.punishedUsers = append(g.PunishedUsers.punishedUsers, punishedUser)
	g.PunishedUsers.Unlock()
	g.Unlock()
}

func (g *GuildInfo) RemoveFromPunishedUsers(index int) {
	g.Lock()
	g.PunishedUsers.Lock()
	g.PunishedUsers.punishedUsers = append(g.PunishedUsers.punishedUsers[:index], g.PunishedUsers.punishedUsers[index+1:]...)
	g.PunishedUsers.Unlock()
	g.Unlock()
}

func (g *GuildInfo) SetPunishedUsers(punishedUsers []PunishedUsers) {
	g.Lock()
	g.PunishedUsers.Lock()
	g.PunishedUsers.punishedUsers = punishedUsers
	g.PunishedUsers.Unlock()
	g.Unlock()
}

func (g *GuildInfo) GetPunishedUsers() []PunishedUsers {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	g.PunishedUsers.Lock()
	defer g.PunishedUsers.Unlock()
	sliceCopy := append(g.PunishedUsers.punishedUsers[:0:0], g.PunishedUsers.punishedUsers...)
	return sliceCopy
}

func (g *GuildInfo) AssignToFilters(index int, filter Filter) {
	g.Lock()
	g.Filters.Lock()
	g.Filters.filters[index] = filter
	g.Filters.Unlock()
	g.Unlock()
}

func (g *GuildInfo) AppendToFilters(filter Filter) {
	g.Lock()
	g.Filters.Lock()
	g.Filters.filters = append(g.Filters.filters, filter)
	g.Filters.Unlock()
	g.Unlock()
}

func (g *GuildInfo) RemoveFromFilters(index int) {
	g.Lock()
	g.Filters.Lock()
	g.Filters.filters = append(g.Filters.filters[:index], g.Filters.filters[index+1:]...)
	g.Filters.Unlock()
	g.Unlock()
}

func (g *GuildInfo) SetFilters(filters []Filter) {
	g.Lock()
	g.Filters.Lock()
	g.Filters.filters = filters
	g.Filters.Unlock()
	g.Unlock()
}

func (g *GuildInfo) GetFilters() []Filter {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	g.Filters.Lock()
	defer g.Filters.Unlock()
	sliceCopy := append(g.Filters.filters[:0:0], g.Filters.filters...)
	return sliceCopy
}

func (g *GuildInfo) AssignToMessageRequirements(index int, messageRequirement MessRequirement) {
	g.Lock()
	g.MessageRequirements.Lock()
	g.MessageRequirements.messageRequirements[index] = messageRequirement
	g.MessageRequirements.Unlock()
	g.Unlock()
}

func (g *GuildInfo) AppendToMessageRequirements(messageRequirement MessRequirement) {
	g.Lock()
	g.MessageRequirements.Lock()
	g.MessageRequirements.messageRequirements = append(g.MessageRequirements.messageRequirements, messageRequirement)
	g.MessageRequirements.Unlock()
	g.Unlock()
}

func (g *GuildInfo) RemoveFromMessageRequirements(index int) {
	g.Lock()
	g.MessageRequirements.Lock()
	g.MessageRequirements.messageRequirements = append(g.MessageRequirements.messageRequirements[:index], g.MessageRequirements.messageRequirements[index+1:]...)
	g.MessageRequirements.Unlock()
	g.Unlock()
}

func (g *GuildInfo) SetMessageRequirements(messageRequirements []MessRequirement) {
	g.Lock()
	g.MessageRequirements.Lock()
	g.MessageRequirements.messageRequirements = messageRequirements
	g.MessageRequirements.Unlock()
	g.Unlock()
}

func (g *GuildInfo) GetMessageRequirements() []MessRequirement {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	g.MessageRequirements.Lock()
	defer g.MessageRequirements.Unlock()
	sliceCopy := append(g.MessageRequirements.messageRequirements[:0:0], g.MessageRequirements.messageRequirements...)
	return sliceCopy
}

func (g *GuildInfo) AppendToSpoilerRoles(spoilerRole *discordgo.Role) {
	g.Lock()
	g.SpoilerRoles.Lock()
	g.SpoilerRoles.spoilerRoles = append(g.SpoilerRoles.spoilerRoles, spoilerRole)
	g.SpoilerRoles.Unlock()
	g.Unlock()
}

func (g *GuildInfo) RemoveFromSpoilerRoles(index int) {
	g.Lock()
	g.SpoilerRoles.Lock()
	if index < len(g.SpoilerRoles.spoilerRoles)-1 {
		copy(g.SpoilerRoles.spoilerRoles[index:], g.SpoilerRoles.spoilerRoles[index+1:])
	}
	g.SpoilerRoles.spoilerRoles[len(g.SpoilerRoles.spoilerRoles)-1] = nil
	g.SpoilerRoles.spoilerRoles = g.SpoilerRoles.spoilerRoles[:len(g.SpoilerRoles.spoilerRoles)-1]
	g.SpoilerRoles.Unlock()
	g.Unlock()
}

func (g *GuildInfo) SetSpoilerRoles(spoilerRoles []*discordgo.Role) {
	g.Lock()
	g.SpoilerRoles.Lock()
	g.SpoilerRoles.spoilerRoles = spoilerRoles
	g.SpoilerRoles.Unlock()
	g.Unlock()
}

func (g *GuildInfo) GetSpoilerRoles() []*discordgo.Role {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	g.SpoilerRoles.Lock()
	defer g.SpoilerRoles.Unlock()
	sliceCopy := append(g.SpoilerRoles.spoilerRoles[:0:0], g.SpoilerRoles.spoilerRoles...)
	return sliceCopy
}

func (g *GuildInfo) AssignToFeeds(index int, feed Feed) {
	g.Lock()
	g.Feeds.Lock()
	g.Feeds.feeds[index] = feed
	g.Feeds.Unlock()
	g.Unlock()
}

func (g *GuildInfo) AppendToFeeds(feed Feed) {
	g.Lock()
	g.Feeds.Lock()
	g.Feeds.feeds = append(g.Feeds.feeds, feed)
	g.Feeds.Unlock()
	g.Unlock()
}

func (g *GuildInfo) RemoveFromFeeds(index int) {
	g.Lock()
	g.Feeds.Lock()
	g.Feeds.feeds = append(g.Feeds.feeds[:index], g.Feeds.feeds[index+1:]...)
	g.Feeds.Unlock()
	g.Unlock()
}

func (g *GuildInfo) SetFeeds(feeds []Feed) {
	g.Lock()
	g.Feeds.Lock()
	g.Feeds.feeds = feeds
	g.Feeds.Unlock()
	g.Unlock()
}

func (g *GuildInfo) GetFeeds() []Feed {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	g.Feeds.Lock()
	defer g.Feeds.Unlock()
	sliceCopy := append(g.Feeds.feeds[:0:0], g.Feeds.feeds...)
	return sliceCopy
}

func (g *GuildInfo) AssignToFeedChecks(index int, feedCheck FeedCheck) {
	g.Lock()
	g.FeedChecks.Lock()
	g.FeedChecks.feedChecks[index] = feedCheck
	g.FeedChecks.Unlock()
	g.Unlock()
}

func (g *GuildInfo) AppendToFeedChecks(feedCheck FeedCheck) {
	g.Lock()
	g.FeedChecks.Lock()
	g.FeedChecks.feedChecks = append(g.FeedChecks.feedChecks, feedCheck)
	g.FeedChecks.Unlock()
	g.Unlock()
}

func (g *GuildInfo) RemoveFromFeedChecks(index int) {
	g.Lock()
	g.FeedChecks.Lock()
	g.FeedChecks.feedChecks = append(g.FeedChecks.feedChecks[:index], g.FeedChecks.feedChecks[index+1:]...)
	g.FeedChecks.Unlock()
	g.Unlock()
}

func (g *GuildInfo) SetFeedChecks(feedChecks []FeedCheck) {
	g.Lock()
	g.FeedChecks.Lock()
	g.FeedChecks.feedChecks = feedChecks
	g.FeedChecks.Unlock()
	g.Unlock()
}

func (g *GuildInfo) GetFeedChecks() []FeedCheck {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	g.FeedChecks.Lock()
	defer g.FeedChecks.Unlock()
	sliceCopy := append(g.FeedChecks.feedChecks[:0:0], g.FeedChecks.feedChecks...)
	return sliceCopy
}

func (g *GuildInfo) AppendToRaffles(raffle *Raffle) {
	g.Lock()
	g.Raffles.Lock()
	g.Raffles.raffles = append(g.Raffles.raffles, raffle)
	g.Raffles.Unlock()
	g.Unlock()
}

func (g *GuildInfo) RemoveFromRaffles(index int) {
	g.Lock()
	g.Raffles.Lock()
	if index < len(g.Raffles.raffles)-1 {
		copy(g.Raffles.raffles[index:], g.Raffles.raffles[index+1:])
	}
	g.Raffles.raffles[len(g.Raffles.raffles)-1] = nil
	g.Raffles.raffles = g.Raffles.raffles[:len(g.Raffles.raffles)-1]
	g.Raffles.Unlock()
	g.Unlock()
}

func (g *GuildInfo) SetRaffles(raffles []*Raffle) {
	g.Lock()
	g.Raffles.Lock()
	g.Raffles.raffles = raffles
	g.Raffles.Unlock()
	g.Unlock()
}

func (g *GuildInfo) GetRaffles() []*Raffle {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	g.Raffles.Lock()
	defer g.Raffles.Unlock()
	sliceCopy := append(g.Raffles.raffles[:0:0], g.Raffles.raffles...)
	return sliceCopy
}

func (g *GuildInfo) AppendToWaifus(waifu Waifu) {
	g.Lock()
	g.Waifus.Lock()
	g.Waifus.waifus = append(g.Waifus.waifus, &waifu)
	g.Waifus.Unlock()
	g.Unlock()
}

func (g *GuildInfo) RemoveFromWaifus(index int) {
	g.Lock()
	g.Waifus.Lock()
	if index < len(g.Waifus.waifus)-1 {
		copy(g.Waifus.waifus[index:], g.Waifus.waifus[index+1:])
	}
	g.Waifus.waifus[len(g.Waifus.waifus)-1] = nil
	g.Waifus.waifus = g.Waifus.waifus[:len(g.Waifus.waifus)-1]
	g.Waifus.Unlock()
	g.Unlock()
}

func (g *GuildInfo) SetWaifus(waifus []*Waifu) {
	g.Lock()
	g.Waifus.Lock()
	g.Waifus.waifus = waifus
	g.Waifus.Unlock()
	g.Unlock()
}

func (g *GuildInfo) GetWaifus() []*Waifu {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	g.Waifus.Lock()
	defer g.Waifus.Unlock()
	sliceCopy := append(g.Waifus.waifus[:0:0], g.Waifus.waifus...)
	return sliceCopy
}

func (g *GuildInfo) AppendToWaifuTrades(waifuTrade *WaifuTrade) {
	g.Lock()
	g.WaifuTrades.Lock()
	g.WaifuTrades.waifuTrades = append(g.WaifuTrades.waifuTrades, waifuTrade)
	g.WaifuTrades.Unlock()
	g.Unlock()
}

func (g *GuildInfo) RemoveFromWaifuTrades(index int) {
	g.Lock()
	g.WaifuTrades.Lock()
	if index < len(g.WaifuTrades.waifuTrades)-1 {
		copy(g.WaifuTrades.waifuTrades[index:], g.WaifuTrades.waifuTrades[index+1:])
	}
	g.WaifuTrades.waifuTrades[len(g.WaifuTrades.waifuTrades)-1] = nil
	g.WaifuTrades.waifuTrades = g.WaifuTrades.waifuTrades[:len(g.WaifuTrades.waifuTrades)-1]
	g.WaifuTrades.Unlock()
	g.Unlock()
}

func (g *GuildInfo) SetWaifuTrades(waifuTrades []*WaifuTrade) {
	g.Lock()
	g.WaifuTrades.Lock()
	g.WaifuTrades.waifuTrades = waifuTrades
	g.WaifuTrades.Unlock()
	g.Unlock()
}

func (g *GuildInfo) GetWaifuTrades() []*WaifuTrade {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	g.WaifuTrades.Lock()
	defer g.WaifuTrades.Unlock()
	sliceCopy := append(g.WaifuTrades.waifuTrades[:0:0], g.WaifuTrades.waifuTrades...)
	return sliceCopy
}

func (g *GuildInfo) AssignToMemberInfoMap(key string, user UserInfo) {
	g.Lock()
	g.MemberInfoMap.Lock()
	g.MemberInfoMap.memberInfo[key] = user
	g.MemberInfoMap.Unlock()
	g.Unlock()
}

func (g *GuildInfo) RemoveFromMemberInfoMap(key string) {
	g.Lock()
	g.MemberInfoMap.Lock()
	delete(g.MemberInfoMap.memberInfo, key)
	g.MemberInfoMap.Unlock()
	g.Unlock()
}

func (g *GuildInfo) SetMemberInfoMap(memberInfo map[string]UserInfo) {
	g.Lock()
	g.MemberInfoMap.Lock()
	g.MemberInfoMap.memberInfo = memberInfo
	g.MemberInfoMap.Unlock()
	g.Unlock()
}

func (g *GuildInfo) GetMemberInfoMap() map[string]UserInfo {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	g.MemberInfoMap.RLock()
	defer g.MemberInfoMap.RUnlock()
	mapCopy := make(map[string]UserInfo)
	for k,v := range g.MemberInfoMap.memberInfo {
		mapCopy[k] = v
	}
	return mapCopy
}

func (g *GuildInfo) SetSpoilerMap(spoilerMap map[string]*discordgo.Role) {
	g.Lock()
	g.SpoilerMap.Lock()
	g.SpoilerMap.spoilerMap = spoilerMap
	g.SpoilerMap.Unlock()
	g.Unlock()
}

func (g *GuildInfo) GetSpoilerMap() map[string]*discordgo.Role {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	g.SpoilerMap.Lock()
	defer g.SpoilerMap.Unlock()
	mapCopy := make(map[string]*discordgo.Role)
	for k,v := range g.SpoilerMap.spoilerMap {
		mapCopy[k] = v
	}
	return mapCopy
}

func (g *GuildInfo) AssignToEmojiStats(key string, emoji Emoji) {
	g.Lock()
	g.EmojiStats.Lock()
	g.EmojiStats.emojiStats[key] = emoji
	g.EmojiStats.Unlock()
	g.Unlock()
}

func (g *GuildInfo) RemoveFromEmojiStats(key string) {
	g.Lock()
	g.EmojiStats.Lock()
	delete(g.EmojiStats.emojiStats, key)
	g.EmojiStats.Unlock()
	g.Unlock()
}

func (g *GuildInfo) SetEmojiStats(emojiStats map[string]Emoji) {
	g.Lock()
	g.EmojiStats.Lock()
	g.EmojiStats.emojiStats = emojiStats
	g.EmojiStats.Unlock()
	g.Unlock()
}

func (g *GuildInfo) GetEmojiStats() map[string]Emoji {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	g.EmojiStats.Lock()
	defer g.EmojiStats.Unlock()
	mapCopy := make(map[string]Emoji)
	for k,v := range g.EmojiStats.emojiStats {
		mapCopy[k] = v
	}
	return mapCopy
}

func (g *GuildInfo) AssignToChannelStats(key string, channel Channel) {
	g.Lock()
	g.ChannelStats.Lock()
	g.ChannelStats.channelStats[key] = channel
	g.ChannelStats.Unlock()
	g.Unlock()
}

func (g *GuildInfo) RemoveFromChannelStats(key string) {
	g.Lock()
	g.ChannelStats.Lock()
	delete(g.ChannelStats.channelStats, key)
	g.ChannelStats.Unlock()
	g.Unlock()
}

func (g *GuildInfo) SetChannelStats(channelStats map[string]Channel) {
	g.Lock()
	g.ChannelStats.Lock()
	g.ChannelStats.channelStats = channelStats
	g.ChannelStats.Unlock()
	g.Unlock()
}

func (g *GuildInfo) GetChannelStats() map[string]Channel {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	g.ChannelStats.Lock()
	defer 	g.ChannelStats.Unlock()
	mapCopy := make(map[string]Channel)
	for k,v := range g.ChannelStats.channelStats {
		mapCopy[k] = v
	}
	return mapCopy
}

func (g *GuildInfo) AssignToUserChangeStats(key string, amount int) {
	g.Lock()
	g.UserChangeStats.Lock()
	g.UserChangeStats.stringIntMap[key] = amount
	g.UserChangeStats.Unlock()
	g.Unlock()
}

func (g *GuildInfo) AddToUserChangeStats(key string, amount int) {
	g.Lock()
	g.UserChangeStats.Lock()
	g.UserChangeStats.stringIntMap[key] += amount
	g.UserChangeStats.Unlock()
	g.Unlock()
}

func (g *GuildInfo) RemoveFromUserChangeStats(key string) {
	g.Lock()
	g.UserChangeStats.Lock()
	delete(g.UserChangeStats.stringIntMap, key)
	g.UserChangeStats.Unlock()
	g.Unlock()
}

func (g *GuildInfo) SetUserChangeStats(userChangeStats map[string]int) {
	g.Lock()
	g.UserChangeStats.Lock()
	g.UserChangeStats.stringIntMap = userChangeStats
	g.UserChangeStats.Unlock()
	g.Unlock()
}

func (g *GuildInfo) GetUserChangeStats() map[string]int {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	g.UserChangeStats.Lock()
	defer 	g.UserChangeStats.Unlock()
	mapCopy := make(map[string]int)
	for k,v := range g.UserChangeStats.stringIntMap {
		mapCopy[k] = v
	}
	return mapCopy
}

func (g *GuildInfo) AssignToVerifiedStats(key string, amount int) {
	g.Lock()
	g.VerifiedStats.Lock()
	g.VerifiedStats.stringIntMap[key] = amount
	g.VerifiedStats.Unlock()
	g.Unlock()
}

func (g *GuildInfo) AddToVerifiedStats(key string, amount int) {
	g.Lock()
	g.VerifiedStats.Lock()
	g.VerifiedStats.stringIntMap[key] += amount
	g.VerifiedStats.Unlock()
	g.Unlock()
}

func (g *GuildInfo) RemoveFromVerifiedStats(key string) {
	g.Lock()
	g.VerifiedStats.Lock()
	delete(g.VerifiedStats.stringIntMap, key)
	g.VerifiedStats.Unlock()
	g.Unlock()
}

func (g *GuildInfo) SetVerifiedStats(verifiedStats map[string]int) {
	g.Lock()
	g.VerifiedStats.Lock()
	g.VerifiedStats.stringIntMap = verifiedStats
	g.VerifiedStats.Unlock()
	g.Unlock()
}

func (g *GuildInfo) GetVerifiedStats() map[string]int {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	g.VerifiedStats.Lock()
	defer g.VerifiedStats.Unlock()
	mapCopy := make(map[string]int)
	for k,v := range g.VerifiedStats.stringIntMap {
		mapCopy[k] = v
	}
	return mapCopy
}

func (g *GuildInfo) SetVoteInfoMap(voteInfo map[string]*VoteInfo) {
	g.Lock()
	g.VoteInfoMap.Lock()
	g.VoteInfoMap.voteInfo = voteInfo
	g.VoteInfoMap.Unlock()
	g.Unlock()
}

func (g *GuildInfo) GetVoteInfoMap() map[string]*VoteInfo {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	g.VoteInfoMap.Lock()
	defer 	g.VoteInfoMap.Unlock()
	mapCopy := make(map[string]*VoteInfo)
	for k,v := range g.VoteInfoMap.voteInfo {
		mapCopy[k] = v
	}
	return mapCopy
}

func (g *GuildInfo) SetTempChaMap(tempChaMap map[string]*TempChaInfo) {
	g.Lock()
	g.TempChaMap.Lock()
	g.TempChaMap.tempCha = tempChaMap
	g.TempChaMap.Unlock()
	g.Unlock()
}

func (g *GuildInfo) GetTempChaMap() map[string]*TempChaInfo {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	g.TempChaMap.Lock()
	defer g.TempChaMap.Unlock()
	mapCopy := make(map[string]*TempChaInfo)
	for k,v := range g.TempChaMap.tempCha {
		mapCopy[k] = v
	}
	return mapCopy
}

func (g *GuildInfo) AssignToReactJoinMap(key string, reactJoin *ReactJoin) {
	g.Lock()
	g.ReactJoinMap.Lock()
	g.ReactJoinMap.reactJoin[key] = reactJoin
	g.ReactJoinMap.Unlock()
	g.Unlock()
}

func (g *GuildInfo) RemoveFromReactJoinMap(key string) {
	g.Lock()
	g.ReactJoinMap.Lock()
	delete(g.ReactJoinMap.reactJoin, key)
	g.ReactJoinMap.Unlock()
	g.Unlock()
}

func (g *GuildInfo) SetReactJoinMap(reactJoinMap map[string]*ReactJoin) {
	g.Lock()
	g.ReactJoinMap.Lock()
	g.ReactJoinMap.reactJoin = reactJoinMap
	g.ReactJoinMap.Unlock()
	g.Unlock()
}

func (g *GuildInfo) GetReactJoinMap() map[string]*ReactJoin {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	g.ReactJoinMap.Lock()
	defer g.ReactJoinMap.Unlock()
	mapCopy := make(map[string]*ReactJoin)
	for k,v := range g.ReactJoinMap.reactJoin {
		mapCopy[k] = v
	}
	return mapCopy
}

func (g *GuildInfo) SetExtensionList(extensionList map[string]string) {
	g.Lock()
	g.ExtensionList.Lock()
	g.ExtensionList.stringStringMap = extensionList
	g.ExtensionList.Unlock()
	g.Unlock()
}

func (g *GuildInfo) GetExtensionList() map[string]string {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	g.ExtensionList.Lock()
	defer g.ExtensionList.Unlock()
	mapCopy := make(map[string]string)
	for k,v := range g.ExtensionList.stringStringMap {
		mapCopy[k] = v
	}
	return mapCopy
}

func (g *GuildInfo) AssignToAutoposts(key string, autopost Cha) {
	g.Lock()
	g.Autoposts.Lock()
	g.Autoposts.autoposts[key] = autopost
	g.Autoposts.Unlock()
	g.Unlock()
}

func (g *GuildInfo) RemoveFromAutoposts(key string) {
	g.Lock()
	g.Autoposts.Lock()
	delete(g.Autoposts.autoposts, key)
	g.Autoposts.Unlock()
	g.Unlock()
}

func (g *GuildInfo) SetAutoposts(autoposts map[string]Cha) {
	g.Lock()
	g.Autoposts.Lock()
	g.Autoposts.autoposts = autoposts
	g.Autoposts.Unlock()
	g.Unlock()
}

func (g *GuildInfo) GetAutoposts() map[string]Cha {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	g.Autoposts.Lock()
	defer g.Autoposts.Unlock()
	mapCopy := make(map[string]Cha)
	for k,v := range g.Autoposts.autoposts {
		mapCopy[k] = v
	}
	return mapCopy
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
		return json.Unmarshal(fileData, &g.PunishedUsers.punishedUsers)
	case "filters.json":
		return json.Unmarshal(fileData, &g.Filters.filters)
	case "messReqs.json":
		return json.Unmarshal(fileData, &g.MessageRequirements.messageRequirements)
	case "spoilerRoles.json":
		err = json.Unmarshal(fileData, &g.SpoilerRoles.spoilerRoles)
		if err != nil {
			return err
		}
		// Fills spoilerMap with roles from the spoilerRoles.json file if latter is not empty
		g.SpoilerMap.Lock()
		g.SpoilerRoles.Lock()
		for i := 0; i < len(g.SpoilerRoles.spoilerRoles); i++ {
			g.SpoilerMap.spoilerMap[g.SpoilerRoles.spoilerRoles[i].ID] = g.SpoilerRoles.spoilerRoles[i]
		}
		g.SpoilerRoles.Unlock()
		g.SpoilerMap.Unlock()
		return nil
	case "rssThreads.json":
		_ = json.Unmarshal(fileData, &g.Feeds.feeds)
		return nil
	case "rssThreadCheck.json":
		return json.Unmarshal(fileData, &g.FeedChecks.feedChecks)
	case "raffles.json":
		return json.Unmarshal(fileData, &g.Raffles.raffles)
	case "waifus.json":
		return json.Unmarshal(fileData, &g.Waifus.waifus)
	case "waifuTrades.json":
		return json.Unmarshal(fileData, &g.WaifuTrades.waifuTrades)
	case "memberInfo.json":
		return json.Unmarshal(fileData, &g.MemberInfoMap.memberInfo)
	case "emojiStats.json":
		return json.Unmarshal(fileData, &g.EmojiStats.emojiStats)
	case "channelStats.json":
		return json.Unmarshal(fileData, &g.ChannelStats.channelStats)
	case "userChangeStats.json":
		return json.Unmarshal(fileData, &g.UserChangeStats.stringIntMap)
	case "verifiedStats.json":
		if config.Website != "" {
			return json.Unmarshal(fileData, &g.VerifiedStats.stringIntMap)
		}
	case "voteInfo.json":
		return json.Unmarshal(fileData, &g.VoteInfoMap.voteInfo)
	case "tempCha.json":
		return json.Unmarshal(fileData, &g.TempChaMap.tempCha)
	case "reactJoin.json":
		return json.Unmarshal(fileData, &g.ReactJoinMap.reactJoin)
	case "extensionList.json":
		return json.Unmarshal(fileData, &g.ExtensionList.stringStringMap)
	case "autoposts.json":
		return json.Unmarshal(fileData, &g.Autoposts.autoposts)
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
