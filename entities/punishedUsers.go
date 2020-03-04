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

func NewPunishedUsers(ID string, username string, unbanDate time.Time, unmuteDate time.Time) PunishedUsers {
	return PunishedUsers{ID: ID, Username: username, UnbanDate: unbanDate, UnmuteDate: unmuteDate}
}

func (p PunishedUsers) SetID(id string) PunishedUsers {
	p.ID = id
	return p
}

func (p PunishedUsers) GetID() string {
	if p == (PunishedUsers{}) {
		return ""
	}
	return p.ID
}

func (p PunishedUsers) SetUsername(username string) PunishedUsers {
	p.Username = username
	return p
}

func (p PunishedUsers) GetUsername() string {
	if p == (PunishedUsers{}) {
		return ""
	}
	return p.Username
}

func (p PunishedUsers) SetUnbanDate(date time.Time) PunishedUsers {
	p.UnbanDate = date
	return p
}

func (p PunishedUsers) GetUnbanDate() time.Time {
	if p == (PunishedUsers{}) {
		return time.Time{}
	}
	return p.UnbanDate
}

func (p PunishedUsers) SetUnmuteDate(date time.Time) PunishedUsers {
	p.UnmuteDate = date
	return p
}

func (p PunishedUsers) GetUnmuteDate() time.Time {
	if p == (PunishedUsers{}) {
		return time.Time{}
	}
	return p.UnmuteDate
}
