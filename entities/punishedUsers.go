package entities

import (
	"sync"
	"time"
)

// PunishedUsers holds every banned and muted user
type PunishedUsers struct {
	sync.RWMutex

	ID         string    `json:"id"`
	Username   string    `json:"user"`
	UnbanDate  time.Time `json:"unbanDate"`
	UnmuteDate time.Time `json:"unmuteDate"`
}

func NewPunishedUsers(ID string, username string, unbanDate time.Time, unmuteDate time.Time) *PunishedUsers {
	return &PunishedUsers{ID: ID, Username: username, UnbanDate: unbanDate, UnmuteDate: unmuteDate}
}

func (p *PunishedUsers) SetID(id string) {
	p.Lock()
	p.ID = id
	p.Unlock()
}

func (p *PunishedUsers) GetID() string {
	p.RLock()
	defer p.RUnlock()
	if p == nil {
		return ""
	}
	return p.ID
}

func (p *PunishedUsers) SetUsername(username string) {
	p.Lock()
	p.Username = username
	p.Unlock()
}

func (p *PunishedUsers) GetUsername() string {
	p.RLock()
	defer p.RUnlock()
	if p == nil {
		return ""
	}
	return p.Username
}

func (p *PunishedUsers) SetUnbanDate(date time.Time) {
	p.Lock()
	p.UnbanDate = date
	p.Unlock()
}

func (p *PunishedUsers) GetUnbanDate() time.Time {
	p.RLock()
	defer p.RUnlock()
	if p == nil {
		return time.Time{}
	}
	return p.UnbanDate
}

func (p *PunishedUsers) SetUnmuteDate(date time.Time) {
	p.Lock()
	p.UnmuteDate = date
	p.Unlock()
}

func (p *PunishedUsers) GetUnmuteDate() time.Time {
	p.RLock()
	defer p.RUnlock()
	if p == nil {
		return time.Time{}
	}
	return p.UnmuteDate
}