package entities

import (
	"sync"
)

// UserInfo is the in memory storage of each user's information
type UserInfo struct {
	sync.RWMutex

	ID               string        `json:"id"`
	Discrim          string        `json:"discrim"`
	Username         string        `json:"username"`
	Nickname         string        `json:"nickname,omitempty"`
	PastUsernames    []string      `json:"pastUsernames,omitempty"`
	PastNicknames    []string      `json:"pastNicknames,omitempty"`
	Warnings         []string      `json:"warnings,omitempty"`
	Mutes            []string      `json:"mutes,omitempty"`
	Kicks            []string      `json:"kicks,omitempty"`
	Bans             []string      `json:"bans,omitempty"`
	JoinDate         string        `json:"joinDate"`
	RedditUsername   string        `json:"redditUser,omitempty"`
	VerifiedDate     string        `json:"verifiedDate,omitempty"`
	UnmuteDate       string        `json:"unmuteDate,omitempty"`
	UnbanDate        string        `json:"unbanDate,omitempty"`
	Timestamps       []*Punishment `json:"timestamps,omitempty"`
	Waifu            *Waifu        `json:"waifu,omitempty"`
	SuspectedSpambot bool
}

func NewUserInfo(ID string, discrim string, username string, nickname string, pastUsernames []string, pastNicknames []string, warnings []string, mutes []string, kicks []string, bans []string, joinDate string, redditUsername string, verifiedDate string, unmuteDate string, unbanDate string, timestamps []*Punishment, waifu *Waifu, suspectedSpambot bool) *UserInfo {
	return &UserInfo{ID: ID, Discrim: discrim, Username: username, Nickname: nickname, PastUsernames: pastUsernames, PastNicknames: pastNicknames, Warnings: warnings, Mutes: mutes, Kicks: kicks, Bans: bans, JoinDate: joinDate, RedditUsername: redditUsername, VerifiedDate: verifiedDate, UnmuteDate: unmuteDate, UnbanDate: unbanDate, Timestamps: timestamps, Waifu: waifu, SuspectedSpambot: suspectedSpambot}
}

func (u *UserInfo) SetID(id string) {
	u.Lock()
	u.ID = id
	u.Unlock()
}

func (u *UserInfo) GetID() string {
	u.RLock()
	defer u.RUnlock()
	if u == nil {
		return ""
	}
	return u.ID
}

func (u *UserInfo) SetDiscrim(discrim string) {
	u.Lock()
	u.Discrim = discrim
	u.Unlock()
}

func (u *UserInfo) GetDiscrim() string {
	u.RLock()
	defer u.RUnlock()
	if u == nil {
		return ""
	}
	return u.Discrim
}

func (u *UserInfo) SetUsername(username string) {
	u.Lock()
	u.Username = username
	u.Unlock()
}

func (u *UserInfo) GetUsername() string {
	u.RLock()
	defer u.RUnlock()
	if u == nil {
		return ""
	}
	return u.Username
}

func (u *UserInfo) SetNickname(nickname string) {
	u.Lock()
	u.Nickname = nickname
	u.Unlock()
}

func (u *UserInfo) GetNickname() string {
	u.RLock()
	defer u.RUnlock()
	if u == nil {
		return ""
	}
	return u.Nickname
}

func (u *UserInfo) AppendToPastUsernames(pastUsername string) {
	u.Lock()
	u.PastUsernames = append(u.PastUsernames, pastUsername)
	u.Unlock()
}

func (u *UserInfo) SetPastUsernames(pastUsernames []string) {
	u.Lock()
	u.PastUsernames = pastUsernames
	u.Unlock()
}

func (u *UserInfo) GetPastUsernames() []string {
	u.RLock()
	defer u.RUnlock()
	if u == nil {
		return nil
	}
	return u.PastUsernames
}

func (u *UserInfo) AppendToPastNicknames(pastNickname string) {
	u.Lock()
	u.PastNicknames = append(u.PastNicknames, pastNickname)
	u.Unlock()
}

func (u *UserInfo) SetPastNicknames(pastNicknames []string) {
	u.Lock()
	u.PastNicknames = pastNicknames
	u.Unlock()
}

func (u *UserInfo) GetPastNicknames() []string {
	u.RLock()
	defer u.RUnlock()
	if u == nil {
		return nil
	}
	return u.PastNicknames
}

func (u *UserInfo) AppendToWarnings(warning string) {
	u.Lock()
	u.Warnings = append(u.Warnings, warning)
	u.Unlock()
}

func (u *UserInfo) RemoveFromWarnings(index int) {
	u.Lock()
	u.Warnings = append(u.Warnings[:index], u.Warnings[index+1:]...)
	u.Unlock()
}

func (u *UserInfo) SetWarnings(warnings []string) {
	u.Lock()
	u.Warnings = warnings
	u.Unlock()
}

func (u *UserInfo) GetWarnings() []string {
	u.RLock()
	defer u.RUnlock()
	if u == nil {
		return nil
	}
	return u.Warnings
}

func (u *UserInfo) AppendToMutes(mute string) {
	u.Lock()
	u.Mutes = append(u.Mutes, mute)
	u.Unlock()
}

func (u *UserInfo) RemoveFromMutes(index int) {
	u.Lock()
	u.Mutes = append(u.Mutes[:index], u.Mutes[index+1:]...)
	u.Unlock()
}

func (u *UserInfo) SetMutes(mutes []string) {
	u.Lock()
	u.Mutes = mutes
	u.Unlock()
}

func (u *UserInfo) GetMutes() []string {
	u.RLock()
	defer u.RUnlock()
	if u == nil {
		return nil
	}
	return u.Mutes
}

func (u *UserInfo) AppendToKicks(kick string) {
	u.Lock()
	u.Kicks = append(u.Kicks, kick)
	u.Unlock()
}

func (u *UserInfo) RemoveFromKicks(index int) {
	u.Lock()
	u.Kicks = append(u.Kicks[:index], u.Kicks[index+1:]...)
	u.Unlock()
}

func (u *UserInfo) SetKicks(kicks []string) {
	u.Lock()
	u.Kicks = kicks
	u.Unlock()
}

func (u *UserInfo) GetKicks() []string {
	u.RLock()
	defer u.RUnlock()
	if u == nil {
		return nil
	}
	return u.Kicks
}

func (u *UserInfo) AppendToBans(ban string) {
	u.Lock()
	u.Bans = append(u.Bans, ban)
	u.Unlock()
}

func (u *UserInfo) RemoveFromBans(index int) {
	u.Lock()
	u.Bans = append(u.Bans[:index], u.Bans[index+1:]...)
	u.Unlock()
}

func (u *UserInfo) SetBans(bans []string) {
	u.Lock()
	u.Bans = bans
	u.Unlock()
}

func (u *UserInfo) GetBans() []string {
	u.RLock()
	defer u.RUnlock()
	if u == nil {
		return nil
	}
	return u.Bans
}

func (u *UserInfo) SetJoinDate(joinDate string) {
	u.Lock()
	u.JoinDate = joinDate
	u.Unlock()
}

func (u *UserInfo) GetJoinDate() string {
	u.RLock()
	defer u.RUnlock()
	if u == nil {
		return ""
	}
	return u.JoinDate
}

func (u *UserInfo) SetRedditUsername(redditUsername string) {
	u.Lock()
	u.RedditUsername = redditUsername
	u.Unlock()
}

func (u *UserInfo) GetRedditUsername() string {
	u.RLock()
	defer u.RUnlock()
	if u == nil {
		return ""
	}
	return u.RedditUsername
}

func (u *UserInfo) SetVerifiedDate(verifiedDate string) {
	u.Lock()
	u.VerifiedDate = verifiedDate
	u.Unlock()
}

func (u *UserInfo) GetVerifiedDate() string {
	u.RLock()
	defer u.RUnlock()
	if u == nil {
		return ""
	}
	return u.VerifiedDate
}

func (u *UserInfo) SetUnmuteDate(UnmuteDate string) {
	u.Lock()
	u.UnmuteDate = UnmuteDate
	u.Unlock()
}

func (u *UserInfo) GetUnmuteDate() string {
	u.RLock()
	defer u.RUnlock()
	if u == nil {
		return ""
	}
	return u.UnmuteDate
}

func (u *UserInfo) SetUnbanDate(UnbanDate string) {
	u.Lock()
	u.UnbanDate = UnbanDate
	u.Unlock()
}

func (u *UserInfo) GetUnbanDate() string {
	u.RLock()
	defer u.RUnlock()
	if u == nil {
		return ""
	}
	return u.UnbanDate
}

func (u *UserInfo) AppendToTimestamps(timestamp *Punishment) {
	u.Lock()
	u.Timestamps = append(u.Timestamps, timestamp)
	u.Unlock()
}

func (u *UserInfo) RemoveFromTimestamps(index int) {
	u.Lock()
	if index < len(u.Timestamps)-1 {
		copy(u.Timestamps[index:], u.Timestamps[index+1:])
	}
	u.Timestamps[len(u.Timestamps)-1] = nil
	u.Timestamps = u.Timestamps[:len(u.Timestamps)-1]
	u.Unlock()
}

func (u *UserInfo) SetTimestamps(timestamps []*Punishment) {
	u.Lock()
	u.Timestamps = timestamps
	u.Unlock()
}

func (u *UserInfo) GetTimestamps() []*Punishment {
	u.RLock()
	defer u.RUnlock()
	if u == nil {
		return nil
	}

	return u.Timestamps
}

func (u *UserInfo) SetWaifu(waifu *Waifu) {
	u.Lock()
	u.Waifu = waifu
	u.Unlock()
}

func (u *UserInfo) GetWaifu() *Waifu {
	u.RLock()
	defer u.RUnlock()
	if u == nil {
		return nil
	}
	return u.Waifu
}

func (u *UserInfo) SetSuspectedSpambot(suspectedSpambot bool) {
	u.Lock()
	u.SuspectedSpambot = suspectedSpambot
	u.Unlock()
}

func (u *UserInfo) GetSuspectedSpambot() bool {
	u.RLock()
	defer u.RUnlock()
	if u == nil {
		return false
	}
	return u.SuspectedSpambot
}