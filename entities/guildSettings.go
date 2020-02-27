package entities

import "sync"

// GuildSettings contains the guild-specific settings and toggled modules
type GuildSettings struct {
	sync.RWMutex

	Prefix              string      `json:"Prefix"`
	BotLog              *Cha        `json:"BotLogID"`
	CommandRoles        []*Role     `json:"CommandRoles"`
	OptInUnder          *Role       `json:"OptInUnder"`
	OptInAbove          *Role       `json:"OptInAbove"`
	MutedRole           *Role       `json:"MutedRole"`
	VoiceChas           []*VoiceCha `json:"VoiceChas"`
	VoteModule          bool        `json:"VoteModule"`
	ModOnly             bool        `json:"ModOnly"`
	VoteChannelCategory *Cha        `json:"VoteChannelCategory"`
	WaifuModule         bool        `json:"WaifuModule"`
	WhitelistFileFilter bool        `json:"WhitelistFileFilter"`
	ReactsModule        bool        `json:"ReactsModule"`
	PingMessage         string      `json:"PingMessage"`
	Premium             bool        `json:"Premium"`
}

func NewGuildSettings(prefix string, botLog *Cha, commandRoles []*Role, optInUnder *Role, optInAbove *Role, mutedRole *Role, voiceChas []*VoiceCha, voteModule bool, modOnly bool, voteChannelCategory *Cha, waifuModule bool, whitelistFileFilter bool, reactsModule bool, pingMessage string, premium bool) *GuildSettings {
	return &GuildSettings{Prefix: prefix, BotLog: botLog, CommandRoles: commandRoles, OptInUnder: optInUnder, OptInAbove: optInAbove, MutedRole: mutedRole, VoiceChas: voiceChas, VoteModule: voteModule, ModOnly: modOnly, VoteChannelCategory: voteChannelCategory, WaifuModule: waifuModule, WhitelistFileFilter: whitelistFileFilter, ReactsModule: reactsModule, PingMessage: pingMessage, Premium: premium}
}

func (g *GuildSettings) SetPrefix(prefix string) {
	g.Lock()
	g.Prefix = prefix
	g.Unlock()
}

func (g *GuildSettings) GetPrefix() string {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return ""
	}
	return g.Prefix
}

func (g *GuildSettings) SetBotLog(cha *Cha) {
	g.Lock()
	g.BotLog = cha
	g.Unlock()
}

func (g *GuildSettings) GetBotLog() *Cha {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.BotLog
}

func (g *GuildSettings) AppendToCommandRoles(commandRole *Role) {
	g.Lock()
	g.CommandRoles = append(g.CommandRoles, commandRole)
	g.Unlock()
}

func (g *GuildSettings) RemoveFromCommandRoles(index int) {
	g.Lock()
	if index < len(g.CommandRoles)-1 {
		copy(g.CommandRoles[index:], g.CommandRoles[index+1:])
	}
	g.CommandRoles[len(g.CommandRoles)-1] = nil
	g.CommandRoles = g.CommandRoles[:len(g.CommandRoles)-1]
	g.Unlock()
}

func (g *GuildSettings) SetCommandRoles(roles []*Role) {
	g.Lock()
	g.CommandRoles = roles
	g.Unlock()
}

func (g *GuildSettings) GetCommandRoles() []*Role {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.CommandRoles
}

func (g *GuildSettings) SetOptInUnder(role *Role) {
	g.Lock()
	g.OptInUnder = role
	g.Unlock()
}

func (g *GuildSettings) GetOptInUnder() *Role {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.OptInUnder
}

func (g *GuildSettings) SetOptInAbove(role *Role) {
	g.Lock()
	g.OptInAbove = role
	g.Unlock()
}

func (g *GuildSettings) GetOptInAbove() *Role {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.OptInAbove
}

func (g *GuildSettings) SetMutedRole(role *Role) {
	g.Lock()
	g.MutedRole = role
	g.Unlock()
}

func (g *GuildSettings) GetMutedRole() *Role {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.MutedRole
}

func (g *GuildSettings) AppendToVoiceChas(voiceCha *VoiceCha) {
	g.Lock()
	g.VoiceChas = append(g.VoiceChas, voiceCha)
	g.Unlock()
}

func (g *GuildSettings) RemoveFromVoiceChas(index int) {
	g.Lock()
	if index < len(g.VoiceChas)-1 {
		copy(g.VoiceChas[index:], g.VoiceChas[index+1:])
	}
	g.VoiceChas[len(g.VoiceChas)-1] = nil
	g.VoiceChas = g.VoiceChas[:len(g.VoiceChas)-1]
	g.Unlock()
}

func (g *GuildSettings) SetVoiceChas(voiceChas []*VoiceCha) {
	g.Lock()
	g.VoiceChas = voiceChas
	g.Unlock()
}

func (g *GuildSettings) GetVoiceChas() []*VoiceCha {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.VoiceChas
}

func (g *GuildSettings) SetVoteModule(voteModule bool) {
	g.Lock()
	g.VoteModule = voteModule
	g.Unlock()
}

func (g *GuildSettings) GetVoteModule() bool {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return false
	}
	return g.VoteModule
}

func (g *GuildSettings) SetModOnly(modOnly bool) {
	g.Lock()
	g.ModOnly = modOnly
	g.Unlock()
}

func (g *GuildSettings) GetModOnly() bool {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return false
	}
	return g.ModOnly
}

func (g *GuildSettings) SetVoteChannelCategory(cha *Cha) {
	g.Lock()
	g.VoteChannelCategory = cha
	g.Unlock()
}

func (g *GuildSettings) GetVoteChannelCategory() *Cha {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return nil
	}
	return g.VoteChannelCategory
}

func (g *GuildSettings) SetWaifuModule(waifuModule bool) {
	g.Lock()
	g.WaifuModule = waifuModule
	g.Unlock()
}

func (g *GuildSettings) GetWaifuModule() bool {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return false
	}
	return g.WaifuModule
}

func (g *GuildSettings) SetWhitelistFileFilter(whitelistFileFilter bool) {
	g.Lock()
	g.WhitelistFileFilter = whitelistFileFilter
	g.Unlock()
}

func (g *GuildSettings) GetWhitelistFileFilter() bool {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return false
	}
	return g.WhitelistFileFilter
}

func (g *GuildSettings) SetReactsModule(reactsModule bool) {
	g.Lock()
	g.ReactsModule = reactsModule
	g.Unlock()
}

func (g *GuildSettings) GetReactsModule() bool {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return false
	}
	return g.ReactsModule
}

func (g *GuildSettings) SetPingMessage(pingMessage string) {
	g.Lock()
	g.PingMessage = pingMessage
	g.Unlock()
}

func (g *GuildSettings) GetPingMessage() string {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return ""
	}
	return g.PingMessage
}

func (g *GuildSettings) SetPremium(premium bool) {
	g.Lock()
	g.Premium = premium
	g.Unlock()
}

func (g *GuildSettings) GetPremium() bool {
	g.RLock()
	defer g.RUnlock()
	if g == nil {
		return false
	}
	return g.Premium
}