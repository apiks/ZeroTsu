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

func NewPunishment(punishment string, punishmentType string, timestamp time.Time) *Punishment {
	return &Punishment{Punishment: punishment, PunishmentType: punishmentType, Timestamp: timestamp}
}

func (p *Punishment) SetPunishment(punishment string) {
	p.Lock()
	p.Punishment = punishment
	p.Unlock()
}

func (p *Punishment) GetPunishment() string {
	p.RLock()
	defer p.RUnlock()
	if p == nil {
		return ""
	}
	return p.Punishment
}

func (p *Punishment) SetPunishmentType(punishmentType string) {
	p.Lock()
	p.PunishmentType = punishmentType
	p.Unlock()
}

func (p *Punishment) GetPunishmentType() string {
	p.RLock()
	defer p.RUnlock()
	if p == nil {
		return ""
	}
	return p.PunishmentType
}

func (p *Punishment) SetTimestamp(timestamp time.Time) {
	p.Lock()
	p.Timestamp = timestamp
	p.Unlock()
}

func (p *Punishment) GetTimestamp() time.Time {
	p.RLock()
	defer p.RUnlock()
	if p == nil {
		return time.Time{}
	}
	return p.Timestamp
}