package entities

import (
	"github.com/sasha-s/go-deadlock"
)

type Cha struct {
	deadlock.RWMutex

	Name   string `json:"Name"`
	ID     string `json:"ID"`
	RoleID string `json:"RoleID"`
}

func NewCha(name string, ID string, roleID string) Cha {
	return Cha{Name: name, ID: ID, RoleID: roleID}
}

func (c Cha) SetName(name string) Cha {
	c.Name = name
	return c
}

func (c Cha) GetName() string {
	if c == (Cha{}) {
		return ""
	}
	return c.Name
}

func (c Cha) SetID(id string) Cha {
	c.ID = id
	return c
}

func (c Cha) GetID() string {
	if c == (Cha{}) {
		return ""
	}
	return c.ID
}

func (c Cha) SetRoleID(roleId string) Cha {
	c.RoleID = roleId
	return c
}

func (c Cha) GetRoleID() string {
	if c == (Cha{}) {
		return ""
	}
	return c.RoleID
}
