package entities

import (
	"sync"
)

type Cha struct {
	sync.RWMutex

	Name string `json:"Name"`
	ID   string `json:"ID"`
}

func NewCha(name string, ID string) *Cha {
	return &Cha{Name: name, ID: ID}
}

func (c *Cha) SetName(name string) {
	c.Lock()
	c.Name = name
	c.Unlock()
}

func (c *Cha) GetName() string {
	c.RLock()
	defer c.RUnlock()
	if c == nil {
		return ""
	}
	return c.Name
}

func (c *Cha) SetID(id string) {
	c.Lock()
	c.ID = id
	c.Unlock()
}

func (c *Cha) GetID() string {
	c.RLock()
	defer c.RUnlock()
	if c == nil {
		return ""
	}
	return c.ID
}

