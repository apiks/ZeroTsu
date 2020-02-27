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

func NewChannel(channelID string, name string, messages map[string]int, roleCount map[string]int, optin bool, exists bool) *Channel {
	return &Channel{ChannelID: channelID, Name: name, Messages: messages, RoleCount: roleCount, Optin: optin, Exists: exists}
}

func (c *Channel) SetChannelID(channelID string) {
	c.Lock()
	c.ChannelID = channelID
	c.Unlock()
}

func (c *Channel) GetChannelID() string {
	c.RLock()
	defer c.RUnlock()
	if c == nil {
		return ""
	}
	return c.ChannelID
}

func (c *Channel) SetName(name string) {
	c.Lock()
	c.Name = name
	c.Unlock()
}

func (c *Channel) GetName() string {
	c.RLock()
	defer c.RUnlock()
	if c == nil {
		return ""
	}
	return c.Name
}

func (c *Channel) SetMessagesMap(messages map[string]int) {
	c.Lock()
	c.Messages = messages
	c.Unlock()
}

func (c *Channel) GetMessagesMap() map[string]int {
	c.RLock()
	defer c.RUnlock()
	if c == nil {
		return nil
	}
	return c.Messages
}

func (c *Channel) AddMessages(date string, messages int) {
	c.Lock()
	if c.RoleCount == nil {
		c.RoleCount = make(map[string]int)
	}
	c.Messages[date] += messages
	c.Unlock()
}

func (c *Channel) SetMessages(date string, messages int) {
	c.Lock()
	if c.RoleCount == nil {
		c.RoleCount = make(map[string]int)
	}
	c.Messages[date] = messages
	c.Unlock()
}

func (c *Channel) GetMessages(date string) int {
	c.RLock()
	defer c.RUnlock()
	if c == nil {
		return 0
	}
	return c.Messages[date]
}

func (c *Channel) SetRoleCountMap(roleCount map[string]int) {
	c.Lock()
	c.RoleCount = roleCount
	c.Unlock()
}

func (c *Channel) GetRoleCountMap() map[string]int {
	c.RLock()
	defer c.RUnlock()
	if c == nil {
		return nil
	}
	return c.RoleCount
}

func (c *Channel) SetRoleCount(date string, roleCount int) {
	c.Lock()
	if c.RoleCount == nil {
		c.RoleCount = make(map[string]int)
	}
	c.RoleCount[date] = roleCount
	c.Unlock()
}

func (c *Channel) GetRoleCount(date string) int {
	c.RLock()
	defer c.RUnlock()
	if c == nil {
		return 0
	}
	return c.RoleCount[date]
}

func (c *Channel) SetOptin(optin bool) {
	c.Lock()
	c.Optin = optin
	c.Unlock()
}

func (c *Channel) GetOptin() bool {
	c.RLock()
	defer c.RUnlock()
	if c == nil {
		return false
	}
	return c.Optin
}

func (c *Channel) SetExists(exists bool) {
	c.Lock()
	c.Exists = exists
	c.Unlock()
}

func (c *Channel) GetExists() bool {
	c.RLock()
	defer c.RUnlock()
	if c == nil {
		return false
	}
	return c.Exists
}