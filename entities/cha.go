package entities

import (
	"sync"
)

type Cha struct {
	sync.RWMutex

	Name string `json:"Name"`
	ID   string `json:"ID"`
}

func NewCha(name string, ID string) Cha {
	return Cha{Name: name, ID: ID}
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

