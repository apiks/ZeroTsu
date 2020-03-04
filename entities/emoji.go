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

func NewEmoji(ID string, name string, messageUsage int, uniqueMessageUsage int, reactions int) Emoji {
	return Emoji{ID: ID, Name: name, MessageUsage: messageUsage, UniqueMessageUsage: uniqueMessageUsage, Reactions: reactions}
}

func (e Emoji) SetID(id string) Emoji {
	e.ID = id
	return e
}

func (e Emoji) GetID() string {
	if e == (Emoji{}) {
		return ""
	}
	return e.ID
}

func (e Emoji) SetName(name string) Emoji {
	e.Name = name
	return e
}

func (e Emoji) GetName() string {
	if e == (Emoji{}) {
		return ""
	}
	return e.Name
}

func (e Emoji) AddMessageUsage(amount int) Emoji {
	e.MessageUsage += amount
	return e
}

func (e Emoji) SetMessageUsage(amount int) Emoji {
	e.MessageUsage = amount
	return e
}

func (e Emoji) GetMessageUsage() int {
	if e == (Emoji{}) {
		return 0
	}
	return e.MessageUsage
}

func (e Emoji) AddUniqueMessageUsage(amount int) Emoji {
	e.UniqueMessageUsage += amount
	return e
}

func (e Emoji) SetUniqueMessageUsage(amount int) Emoji {
	e.UniqueMessageUsage = amount
	return e
}

func (e Emoji) GetUniqueMessageUsage() int {
	if e == (Emoji{}) {
		return 0
	}
	return e.UniqueMessageUsage
}

func (e Emoji) AddSetReactions(amount int) Emoji {
	e.Reactions += amount
	return e
}

func (e Emoji) SetReactions(amount int) Emoji {
	e.Reactions = amount
	return e
}

func (e Emoji) GetReactions() int {
	if e == (Emoji{}) {
		return 0
	}
	return e.Reactions
}
