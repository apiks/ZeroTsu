package entities

import "sync"

type Channel struct {
	sync.RWMutex

	ChannelID string
	Name      string
	Messages  map[string]int
	RoleCount map[string]int `json:",omitempty"`
	Optin     bool
	Exists    bool
}

func (c Channel) SetChannelID(channelID string) Channel{
	c.ChannelID = channelID
	return c
}

func (c Channel) GetChannelID() string {
	if c.ChannelID == "" {
		return ""
	}
	return c.ChannelID
}

func (c Channel) SetName(name string) Channel {
	c.Name = name
	return c
}

func (c Channel) GetName() string {
	if c.Name == "" {
		return ""
	}
	return c.Name
}

func (c Channel) SetMessagesMap(messages map[string]int) Channel {
	c.Messages = messages
	return c
}

func (c Channel) GetMessagesMap() map[string]int {
	if c.Messages == nil {
		return nil
	}
	return c.Messages
}

func (c Channel) AddMessages(date string, messages int) Channel {
	if c.RoleCount == nil {
		c.RoleCount = make(map[string]int)
	}
	c.Messages[date] += messages
	return c
}

func (c Channel) SetMessages(date string, messages int) Channel {
	if c.RoleCount == nil {
		c.RoleCount = make(map[string]int)
	}
	c.Messages[date] = messages
	return c
}

func (c Channel) GetMessages(date string) int {
	if c.Messages == nil {
		return 0
	}
	return c.Messages[date]
}

func (c Channel) SetRoleCountMap(roleCount map[string]int) Channel {
	c.RoleCount = roleCount
	return c
}

func (c Channel) GetRoleCountMap() map[string]int {
	if c.RoleCount == nil {
		return nil
	}
	return c.RoleCount
}

func (c Channel) SetRoleCount(date string, roleCount int) Channel {
	if c.RoleCount == nil {
		c.RoleCount = make(map[string]int)
	}
	c.RoleCount[date] = roleCount
	return c
}

func (c Channel) GetRoleCount(date string) int {
	if c.RoleCount == nil {
		return 0
	}
	return c.RoleCount[date]
}

func (c Channel) SetOptin(optin bool) Channel {
	c.Optin = optin
	return c
}

func (c Channel) GetOptin() bool {
	if c.Optin == false {
		return false
	}
	return c.Optin
}

func (c Channel) SetExists(exists bool) Channel{
	c.Exists = exists
	return c
}

func (c Channel) GetExists() bool {
	if c.Exists == false {
		return false
	}
	return c.Exists
}