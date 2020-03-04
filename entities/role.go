package entities

import "sync"

type Role struct {
	sync.RWMutex

	Name     string `json:"Name"`
	ID       string `json:"ID"`
	Position int    `json:"Position"`
}

func NewRole(name string, ID string, position int) Role {
	return Role{Name: name, ID: ID, Position: position}
}

func (r Role) SetName(name string) Role {
	r.Name = name
	return r
}

func (r Role) GetName() string {
	if r == (Role{}) {
		return ""
	}
	return r.Name
}

func (r Role) SetID(id string) Role {
	r.ID = id
	return r
}

func (r Role) GetID() string {
	if r == (Role{}) {
		return ""
	}
	return r.ID
}

func (r Role) SetPosition(position int) Role {
	r.Position = position
	return r
}

func (r Role) GetPosition() int {
	if r == (Role{}) {
		return 0
	}
	return r.Position
}
