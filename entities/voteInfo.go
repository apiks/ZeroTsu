package entities

import (
	"github.com/bwmarrin/discordgo"
	"sync"
	"time"
)

// VoteInfo is the in memory storage of each vote channel's info
type VoteInfo struct {
	sync.RWMutex

	Date         time.Time          `json:"Date"`
	Channel      string             `json:"Channel"`
	ChannelType  string             `json:"ChannelType"`
	Category     string             `json:"Category,omitempty"`
	Description  string             `json:"Description,omitempty"`
	VotesReq     int                `json:"VotesReq"`
	MessageReact *discordgo.Message `json:"MessageReact"`
	User         *discordgo.User    `json:"User"`
}

func NewVoteInfo(date time.Time, channel string, channelType string, category string, description string, votesReq int, messageReact *discordgo.Message, user *discordgo.User) *VoteInfo {
	return &VoteInfo{Date: date, Channel: channel, ChannelType: channelType, Category: category, Description: description, VotesReq: votesReq, MessageReact: messageReact, User: user}
}

func (v *VoteInfo) SetDate(date time.Time) {
	v.Lock()
	v.Date = date
	v.Unlock()
}

func (v *VoteInfo) GetDate() time.Time {
	v.RLock()
	defer v.RUnlock()
	if v == nil {
		return time.Time{}
	}
	return v.Date
}

func (v *VoteInfo) SetChannel(channel string) {
	v.Lock()
	v.Channel = channel
	v.Unlock()
}

func (v *VoteInfo) GetChannel() string {
	v.RLock()
	defer v.RUnlock()
	if v == nil {
		return ""
	}
	return v.Channel
}

func (v *VoteInfo) SetChannelType(channelType string) {
	v.Lock()
	v.ChannelType = channelType
	v.Unlock()
}

func (v *VoteInfo) GetChannelType() string {
	v.RLock()
	defer v.RUnlock()
	if v == nil {
		return ""
	}
	return v.ChannelType
}

func (v *VoteInfo) SetCategory(category string) {
	v.Lock()
	v.Category = category
	v.Unlock()
}

func (v *VoteInfo) GetCategory() string {
	v.RLock()
	defer v.RUnlock()
	if v == nil {
		return ""
	}
	return v.Category
}

func (v *VoteInfo) SetDescription(description string) {
	v.Lock()
	v.Description = description
	v.Unlock()
}

func (v *VoteInfo) GetDescription() string {
	v.RLock()
	defer v.RUnlock()
	if v == nil {
		return ""
	}
	return v.Description
}

func (v *VoteInfo) SetVotesReq(votesReq int) {
	v.Lock()
	v.VotesReq = votesReq
	v.Unlock()
}

func (v *VoteInfo) GetVotesReq() int {
	v.RLock()
	defer v.RUnlock()
	if v == nil {
		return 0
	}
	return v.VotesReq
}

func (v *VoteInfo) SetMessageReact(messageReact *discordgo.Message) {
	v.Lock()
	v.MessageReact = messageReact
	v.Unlock()
}

func (v *VoteInfo) GetMessageReact() *discordgo.Message {
	v.RLock()
	defer v.RUnlock()
	if v == nil {
		return nil
	}
	return v.MessageReact
}

func (v *VoteInfo) SetUser(user *discordgo.User) {
	v.Lock()
	v.User = user
	v.Unlock()
}

func (v *VoteInfo) GetUser() *discordgo.User {
	v.RLock()
	defer v.RUnlock()
	if v == nil {
		return nil
	}
	return v.User
}
