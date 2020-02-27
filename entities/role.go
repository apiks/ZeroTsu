package entities

import "sync"

type Role struct {
	sync.RWMutex

	Name     string `json:"Name"`
	ID       string `json:"ID"`
	Position int    `json:"Position"`
}

func NewRole(name string, ID string, position int) *Role {
	return &Role{Name: name, ID: ID, Position: position}
}

func (r *Role) SetName(name string) {
	r.Lock()
	r.Name = name
	r.Unlock()
}

func (r *Role) GetName() string {
	r.RLock()
	defer r.RUnlock()
	if r == nil {
		return ""
	}
	return r.Name
}

func (r *Role) SetID(id string) {
	r.Lock()
	r.ID = id
	r.Unlock()
}

func (r *Role) GetID() string {
	r.RLock()
	defer r.RUnlock()
	if r == nil {
		return ""
	}
	return r.ID
}

func (r *Role) SetPosition(position int) {
	r.Lock()
	r.Position = position
	r.Unlock()
}

func (r *Role) GetPosition() int {
	r.RLock()
	defer r.RUnlock()
	if r == nil {
		return 0
	}
	return r.Position
}