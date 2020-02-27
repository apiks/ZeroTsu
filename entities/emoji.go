package entities

import "sync"

type Emoji struct {
	sync.RWMutex

	ID                 string `json:"id"`
	Name               string `json:"name"`
	MessageUsage       int    `json:"messageUsage"`
	UniqueMessageUsage int    `json:"uniqueMessages"`
	Reactions          int    `json:"reactions"`
}

func NewEmoji(ID string, name string, messageUsage int, uniqueMessageUsage int, reactions int) *Emoji {
	return &Emoji{ID: ID, Name: name, MessageUsage: messageUsage, UniqueMessageUsage: uniqueMessageUsage, Reactions: reactions}
}

func (e *Emoji) SetID(id string) {
	e.Lock()
	e.ID = id
	e.Unlock()
}

func (e *Emoji) GetID() string {
	e.RLock()
	defer e.RUnlock()
	if e == nil {
		return ""
	}
	return e.ID
}

func (e *Emoji) SetName(name string) {
	e.Lock()
	e.Name = name
	e.Unlock()
}

func (e *Emoji) GetName() string {
	e.RLock()
	defer e.RUnlock()
	if e == nil {
		return ""
	}
	return e.Name
}

func (e *Emoji) AddMessageUsage(amount int) {
	e.Lock()
	e.MessageUsage += amount
	e.Unlock()
}

func (e *Emoji) SetMessageUsage(amount int) {
	e.Lock()
	e.MessageUsage = amount
	e.Unlock()
}

func (e *Emoji) GetMessageUsage() int {
	e.RLock()
	defer e.RUnlock()
	if e == nil {
		return 0
	}
	return e.MessageUsage
}

func (e *Emoji) AddUniqueMessageUsage(amount int) {
	e.Lock()
	e.UniqueMessageUsage += amount
	e.Unlock()
}

func (e *Emoji) SetUniqueMessageUsage(amount int) {
	e.Lock()
	e.UniqueMessageUsage = amount
	e.Unlock()
}

func (e *Emoji) GetUniqueMessageUsage() int {
	e.RLock()
	defer e.RUnlock()
	if e == nil {
		return 0
	}
	return e.UniqueMessageUsage
}

func (e *Emoji) AddSetReactions(amount int) {
	e.Lock()
	e.Reactions += amount
	e.Unlock()
}

func (e *Emoji) SetReactions(amount int) {
	e.Lock()
	e.Reactions = amount
	e.Unlock()
}

func (e *Emoji) GetReactions() int {
	e.RLock()
	defer e.RUnlock()
	if e == nil {
		return 0
	}
	return e.Reactions
}