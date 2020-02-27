package entities

import "sync"

type VoiceCha struct {
	sync.RWMutex

	Name  string  `json:"Name"`
	ID    string  `json:"ID"`
	Roles []Role `json:"Roles"`
}

func (v VoiceCha) SetName(name string) VoiceCha {
	v.Name = name
	return v
}

func (v VoiceCha) GetName() string {
	if v.Name == "" {
		return ""
	}
	return v.Name
}

func (v VoiceCha) SetID(id string) VoiceCha {
	v.ID = id
	return v
}

func (v VoiceCha) GetID() string {
	if v.ID == "" {
		return ""
	}
	return v.ID
}

func (v VoiceCha) AppendToRoles(role Role) VoiceCha {
	v.Roles = append(v.Roles, role)
	return v
}

func (v VoiceCha) RemoveFromRoles(index int) VoiceCha {
	v.Roles = append(v.Roles[:index], v.Roles[index+1:]...)
	return v
}

func (v VoiceCha) SetRoles(roles []Role) VoiceCha {
	v.Roles = roles
	return v
}

func (v VoiceCha) GetRoles() []Role {
	if v.Roles == nil {
		return nil
	}
	return v.Roles
}
