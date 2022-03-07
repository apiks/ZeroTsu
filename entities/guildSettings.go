package entities

import "sync"

// GuildSettings contains the guild-specific settings and toggled modules
type GuildSettings struct {
	sync.RWMutex

	Prefix       string     `json:"Prefix"`
	BotLog       Cha        `json:"BotLogID"`
	CommandRoles []Role     `json:"CommandRoles"`
	MutedRole    Role       `json:"MutedRole"`
	VoiceChas    []VoiceCha `json:"VoiceChas"`
	ModOnly      bool       `json:"ModOnly"`
	Donghua      bool       `json:"Donghua"`
	ReactsModule bool       `json:"ReactsModule"`
	PingMessage  string     `json:"PingMessage"`
	Premium      bool       `json:"Premium"`
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

func (g GuildSettings) SetDonghua(donghua bool) GuildSettings {
	g.Donghua = donghua
	return g
}

func (g GuildSettings) GetDonghua() bool {
	if g.Donghua == false {
		return false
	}
	return g.Donghua
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
