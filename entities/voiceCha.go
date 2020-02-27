package entities

import "sync"

type VoiceCha struct {
	sync.RWMutex

	Name  string  `json:"Name"`
	ID    string  `json:"ID"`
	Roles []*Role `json:"Roles"`
}

func NewVoiceCha(name string, ID string, roles []*Role) *VoiceCha {
	return &VoiceCha{Name: name, ID: ID, Roles: roles}
}

func (v *VoiceCha) SetName(name string) {
	v.Lock()
	v.Name = name
	v.Unlock()
}

func (v *VoiceCha) GetName() string {
	v.RLock()
	defer v.RUnlock()
	if v == nil {
		return ""
	}
	return v.Name
}

func (v *VoiceCha) SetID(id string) {
	v.Lock()
	v.ID = id
	v.Unlock()
}

func (v *VoiceCha) GetID() string {
	v.RLock()
	defer v.RUnlock()
	if v == nil {
		return ""
	}
	return v.ID
}

func (v *VoiceCha) AppendToRoles(role *Role) {
	v.Lock()
	v.Roles = append(v.Roles, role)
	v.Unlock()
}

func (v *VoiceCha) RemoveFromRoles(index int) {
	v.Lock()
	if index < len(v.Roles)-1 {
		copy(v.Roles[index:], v.Roles[index+1:])
	}
	v.Roles[len(v.Roles)-1] = nil
	v.Roles = v.Roles[:len(v.Roles)-1]
	v.Unlock()
}

func (v *VoiceCha) SetRoles(roles []*Role) {
	v.Lock()
	v.Roles = roles
	v.Unlock()
}

func (v *VoiceCha) GetRoles() []*Role {
	v.RLock()
	defer v.RUnlock()
	if v == nil {
		return nil
	}
	return v.Roles
}
