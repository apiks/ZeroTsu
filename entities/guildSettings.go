package entities

import "sync"

// GuildSettings contains the guild-specific settings and toggled modules
type GuildSettings struct {
	sync.RWMutex

	Prefix              string     `json:"Prefix"`
	BotLog              Cha        `json:"BotLogID"`
	CommandRoles        []Role     `json:"CommandRoles"`
	OptInUnder          Role       `json:"OptInUnder"`
	OptInAbove          Role       `json:"OptInAbove"`
	MutedRole           Role       `json:"MutedRole"`
	VoiceChas           []VoiceCha `json:"VoiceChas"`
	VoteModule          bool       `json:"VoteModule"`
	ModOnly             bool       `json:"ModOnly"`
	VoteChannelCategory Cha        `json:"VoteChannelCategory"`
	WaifuModule         bool       `json:"WaifuModule"`
	WhitelistFileFilter bool       `json:"WhitelistFileFilter"`
	ReactsModule        bool       `json:"ReactsModule"`
	PingMessage         string     `json:"PingMessage"`
	Premium             bool       `json:"Premium"`
}

func (g GuildSettings) SetPrefix(prefix string) GuildSettings {
	g.Prefix = prefix
	return g
}

func (g GuildSettings) GetPrefix() string {
	if g.Prefix == "" {
		return ""
	}
	return g.Prefix
}

func (g GuildSettings) SetBotLog(cha Cha) GuildSettings {
	g.BotLog = cha
	return g
}

func (g GuildSettings) GetBotLog() Cha {
	if g.BotLog == (Cha{}) {
		return Cha{}
	}
	return g.BotLog
}

func (g GuildSettings) AppendToCommandRoles(commandRole Role) GuildSettings {
	g.CommandRoles = append(g.CommandRoles, commandRole)
	return g
}

func (g GuildSettings) RemoveFromCommandRoles(index int) GuildSettings {
	g.CommandRoles = append(g.CommandRoles[:index], g.CommandRoles[index+1:]...)
	return g
}

func (g GuildSettings) SetCommandRoles(roles []Role) GuildSettings {
	g.CommandRoles = roles
	return g
}

func (g GuildSettings) GetCommandRoles() []Role {
	if g.CommandRoles == nil {
		return nil
	}
	return g.CommandRoles
}

func (g GuildSettings) SetOptInUnder(role Role) GuildSettings {
	g.OptInUnder = role
	return g
}

func (g GuildSettings) GetOptInUnder() Role {
	if g.OptInUnder == (Role{}) {
		return Role{}
	}
	return g.OptInUnder
}

func (g GuildSettings) SetOptInAbove(role Role) GuildSettings {
	g.OptInAbove = role
	return g
}

func (g GuildSettings) GetOptInAbove() Role {
	if g.OptInAbove == (Role{}) {
		return Role{}
	}
	return g.OptInAbove
}

func (g GuildSettings) SetMutedRole(role Role) GuildSettings {
	g.MutedRole = role
	return g
}

func (g GuildSettings) GetMutedRole() Role {
	if g.MutedRole == (Role{}) {
		return Role{}
	}
	return g.MutedRole
}

func (g GuildSettings) AppendToVoiceChas(voiceCha VoiceCha) GuildSettings {
	g.VoiceChas = append(g.VoiceChas, voiceCha)
	return g
}

func (g GuildSettings) RemoveFromVoiceChas(index int) GuildSettings {
	g.VoiceChas = append(g.VoiceChas[:index], g.VoiceChas[index+1:]...)
	return g
}

func (g GuildSettings) SetVoiceChas(voiceChas []VoiceCha) GuildSettings {
	g.VoiceChas = voiceChas
	return g
}

func (g GuildSettings) GetVoiceChas() []VoiceCha {
	if g.VoiceChas == nil {
		return nil
	}
	return g.VoiceChas
}

func (g GuildSettings) SetVoteModule(voteModule bool) GuildSettings {
	g.VoteModule = voteModule
	return g
}

func (g GuildSettings) GetVoteModule() bool {
	if g.VoteModule == false {
		return false
	}
	return g.VoteModule
}

func (g GuildSettings) SetModOnly(modOnly bool) GuildSettings {
	g.ModOnly = modOnly
	return g
}

func (g GuildSettings) GetModOnly() bool {
	if g.ModOnly == false {
		return false
	}
	return g.ModOnly
}

func (g GuildSettings) SetVoteChannelCategory(cha Cha) GuildSettings {
	g.VoteChannelCategory = cha
	return g
}

func (g GuildSettings) GetVoteChannelCategory() Cha {
	if g.VoteChannelCategory == (Cha{}) {
		return Cha{}
	}
	return g.VoteChannelCategory
}

func (g GuildSettings) SetWaifuModule(waifuModule bool) GuildSettings {
	g.WaifuModule = waifuModule
	return g
}

func (g GuildSettings) GetWaifuModule() bool {
	if g.WaifuModule == false {
		return false
	}
	return g.WaifuModule
}

func (g GuildSettings) SetWhitelistFileFilter(whitelistFileFilter bool) GuildSettings {
	g.WhitelistFileFilter = whitelistFileFilter
	return g
}

func (g GuildSettings) GetWhitelistFileFilter() bool {
	if g.WhitelistFileFilter == false {
		return false
	}
	return g.WhitelistFileFilter
}

func (g GuildSettings) SetReactsModule(reactsModule bool) GuildSettings {
	g.ReactsModule = reactsModule
	return g
}

func (g GuildSettings) GetReactsModule() bool {
	if g.ReactsModule == false {
		return false
	}
	return g.ReactsModule
}

func (g GuildSettings) SetPingMessage(pingMessage string) GuildSettings {
	g.PingMessage = pingMessage
	return g
}

func (g GuildSettings) GetPingMessage() string {
	if g.PingMessage == "" {
		return ""
	}
	return g.PingMessage
}

func (g GuildSettings) SetPremium(premium bool) GuildSettings {
	g.Premium = premium
	return g
}

func (g GuildSettings) GetPremium() bool {
	if g.Premium == false {
		return false
	}
	return g.Premium
}
