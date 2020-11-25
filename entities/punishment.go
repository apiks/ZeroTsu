package entities

import (
	"sync"
	"time"
)

// Punishment holds a punishment timestamp
type Punishment struct {
	sync.RWMutex

	Punishment     string    `json:"punishment"`
	PunishmentType string    `json:"type"`
	Timestamp      time.Time `json:"timestamp"`
}

func NewPunishment(punishment string, punishmentType string, timestamp time.Time) Punishment {
	return Punishment{Punishment: punishment, PunishmentType: punishmentType, Timestamp: timestamp}
}

func (p Punishment) SetPunishment(punishment string) Punishment {
	p.Punishment = punishment
	return p
}

func (p Punishment) GetPunishment() string {
	if p == (Punishment{}) {
		return ""
	}
	return p.Punishment
}

func (p Punishment) SetPunishmentType(punishmentType string) Punishment {
	p.PunishmentType = punishmentType
	return p
}

func (p Punishment) GetPunishmentType() string {
	if p == (Punishment{}) {
		return ""
	}
	return p.PunishmentType
}

func (p Punishment) SetTimestamp(timestamp time.Time) Punishment {
	p.Timestamp = timestamp
	return p
}

func (p Punishment) GetTimestamp() time.Time {
	if p == (Punishment{}) {
		return time.Time{}
	}
	return p.Timestamp
}
