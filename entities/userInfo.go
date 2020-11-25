package entities

import (
	"sync"
)

// UserInfo is the in memory storage of each user's information
type UserInfo struct {
	sync.RWMutex

	ID               string       `json:"id"`
	Discrim          string       `json:"discrim"`
	Username         string       `json:"username"`
	Nickname         string       `json:"nickname,omitempty"`
	PastUsernames    []string     `json:"pastUsernames,omitempty"`
	PastNicknames    []string     `json:"pastNicknames,omitempty"`
	Warnings         []string     `json:"warnings,omitempty"`
	Mutes            []string     `json:"mutes,omitempty"`
	Kicks            []string     `json:"kicks,omitempty"`
	Bans             []string     `json:"bans,omitempty"`
	JoinDate         string       `json:"joinDate"`
	RedditUsername   string       `json:"redditUser,omitempty"`
	VerifiedDate     string       `json:"verifiedDate,omitempty"`
	UnmuteDate       string       `json:"unmuteDate,omitempty"`
	UnbanDate        string       `json:"unbanDate,omitempty"`
	Timestamps       []Punishment `json:"timestamps,omitempty"`
	Waifu            Waifu        `json:"waifu,omitempty"`
	SuspectedSpambot bool
}

func NewUserInfo(ID string, discrim string, username string, nickname string, pastUsernames []string, pastNicknames []string, warnings []string, mutes []string, kicks []string, bans []string, joinDate string, redditUsername string, verifiedDate string, unmuteDate string, unbanDate string, timestamps []Punishment, waifu Waifu, suspectedSpambot bool) UserInfo {
	return UserInfo{ID: ID, Discrim: discrim, Username: username, Nickname: nickname, PastUsernames: pastUsernames, PastNicknames: pastNicknames, Warnings: warnings, Mutes: mutes, Kicks: kicks, Bans: bans, JoinDate: joinDate, RedditUsername: redditUsername, VerifiedDate: verifiedDate, UnmuteDate: unmuteDate, UnbanDate: unbanDate, Timestamps: timestamps, Waifu: waifu, SuspectedSpambot: suspectedSpambot}
}

func (u UserInfo) SetID(id string) UserInfo {
	u.ID = id
	return u
}

func (u UserInfo) GetID() string {
	if u.ID == "" {
		return ""
	}
	return u.ID
}

func (u UserInfo) SetDiscrim(discrim string) UserInfo {
	u.Discrim = discrim
	return u
}

func (u UserInfo) GetDiscrim() string {
	if u.Discrim == "" {
		return ""
	}
	return u.Discrim
}

func (u UserInfo) SetUsername(username string) UserInfo {
	u.Username = username
	return u
}

func (u UserInfo) GetUsername() string {
	if u.Username == "" {
		return ""
	}
	return u.Username
}

func (u UserInfo) SetNickname(nickname string) UserInfo {
	u.Lock()
	defer u.Unlock()
	u.Nickname = nickname
	return u
}

func (u UserInfo) GetNickname() string {
	if u.Nickname == "" {
		return ""
	}
	return u.Nickname
}

func (u UserInfo) AppendToPastUsernames(pastUsername string) UserInfo {
	u.PastUsernames = append(u.PastUsernames, pastUsername)
	return u
}

func (u UserInfo) SetPastUsernames(pastUsernames []string) UserInfo {
	u.PastUsernames = pastUsernames
	return u
}

func (u UserInfo) GetPastUsernames() []string {
	if u.PastUsernames == nil {
		return nil
	}
	return u.PastUsernames
}

func (u UserInfo) AppendToPastNicknames(pastNickname string) UserInfo {
	u.PastNicknames = append(u.PastNicknames, pastNickname)
	return u
}

func (u UserInfo) SetPastNicknames(pastNicknames []string) UserInfo {
	u.PastNicknames = pastNicknames
	return u
}

func (u UserInfo) GetPastNicknames() []string {
	if u.PastNicknames == nil {
		return nil
	}
	return u.PastNicknames
}

func (u UserInfo) AppendToWarnings(warning string) UserInfo {
	u.Warnings = append(u.Warnings, warning)
	return u
}

func (u UserInfo) RemoveFromWarnings(index int) UserInfo {
	u.Warnings = append(u.Warnings[:index], u.Warnings[index+1:]...)
	return u
}

func (u UserInfo) SetWarnings(warnings []string) UserInfo {
	u.Warnings = warnings
	return u
}

func (u UserInfo) GetWarnings() []string {
	if u.Warnings == nil {
		return nil
	}
	return u.Warnings
}

func (u UserInfo) AppendToMutes(mute string) UserInfo {
	u.Mutes = append(u.Mutes, mute)
	return u
}

func (u UserInfo) RemoveFromMutes(index int) UserInfo {
	u.Mutes = append(u.Mutes[:index], u.Mutes[index+1:]...)
	return u
}

func (u UserInfo) SetMutes(mutes []string) UserInfo {
	u.Mutes = mutes
	return u
}

func (u UserInfo) GetMutes() []string {
	if u.Mutes == nil {
		return nil
	}
	return u.Mutes
}

func (u UserInfo) AppendToKicks(kick string) UserInfo {
	u.Kicks = append(u.Kicks, kick)
	return u
}

func (u UserInfo) RemoveFromKicks(index int) UserInfo {
	u.Kicks = append(u.Kicks[:index], u.Kicks[index+1:]...)
	return u
}

func (u UserInfo) SetKicks(kicks []string) UserInfo {
	u.Kicks = kicks
	return u
}

func (u UserInfo) GetKicks() []string {
	if u.Kicks == nil {
		return nil
	}
	return u.Kicks
}

func (u UserInfo) AppendToBans(ban string) UserInfo {
	u.Bans = append(u.Bans, ban)
	return u
}

func (u UserInfo) RemoveFromBans(index int) UserInfo {
	u.Bans = append(u.Bans[:index], u.Bans[index+1:]...)
	return u
}

func (u UserInfo) SetBans(bans []string) UserInfo {
	u.Bans = bans
	return u
}

func (u UserInfo) GetBans() []string {
	if u.Bans == nil {
		return nil
	}
	return u.Bans
}

func (u UserInfo) SetJoinDate(joinDate string) UserInfo {
	u.JoinDate = joinDate
	return u
}

func (u UserInfo) GetJoinDate() string {
	if u.JoinDate == "" {
		return ""
	}
	return u.JoinDate
}

func (u UserInfo) SetRedditUsername(redditUsername string) UserInfo {
	u.RedditUsername = redditUsername
	return u
}

func (u UserInfo) GetRedditUsername() string {
	if u.RedditUsername == "" {
		return ""
	}
	return u.RedditUsername
}

func (u UserInfo) SetVerifiedDate(verifiedDate string) UserInfo {
	u.VerifiedDate = verifiedDate
	return u
}

func (u UserInfo) GetVerifiedDate() string {
	if u.VerifiedDate == "" {
		return ""
	}
	return u.VerifiedDate
}

func (u UserInfo) SetUnmuteDate(UnmuteDate string) UserInfo {
	u.UnmuteDate = UnmuteDate
	return u
}

func (u UserInfo) GetUnmuteDate() string {
	if u.UnmuteDate == "" {
		return ""
	}
	return u.UnmuteDate
}

func (u UserInfo) SetUnbanDate(UnbanDate string) UserInfo {
	u.UnbanDate = UnbanDate
	return u
}

func (u UserInfo) GetUnbanDate() string {
	if u.UnbanDate == "" {
		return ""
	}
	return u.UnbanDate
}

func (u UserInfo) AppendToTimestamps(timestamp Punishment) UserInfo {
	u.Timestamps = append(u.Timestamps, timestamp)
	return u
}

func (u UserInfo) RemoveFromTimestamps(index int) UserInfo {
	u.Timestamps = append(u.Timestamps[:index], u.Timestamps[index+1:]...)
	return u
}

func (u UserInfo) SetTimestamps(timestamps []Punishment) UserInfo {
	u.Timestamps = timestamps
	return u
}

func (u UserInfo) GetTimestamps() []Punishment {
	if u.Timestamps == nil {
		return nil
	}
	return u.Timestamps
}

func (u UserInfo) SetWaifu(waifu Waifu) UserInfo {
	u.Waifu = waifu
	return u
}

func (u UserInfo) GetWaifu() Waifu {
	if u.Waifu == (Waifu{}) {
		return Waifu{}
	}
	return u.Waifu
}

func (u UserInfo) SetSuspectedSpambot(suspectedSpambot bool) UserInfo {
	u.SuspectedSpambot = suspectedSpambot
	return u
}

func (u UserInfo) GetSuspectedSpambot() bool {
	if u.SuspectedSpambot == false {
		return false
	}
	return u.SuspectedSpambot
}
