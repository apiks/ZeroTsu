package entities

import "sync"

type MessRequirement struct {
	sync.RWMutex

	Phrase          string `json:"Phrase"`
	RequirementType string `json:"Type"`
	ChannelID       string `json:"Channel"`
	LastUserID      string
}

func NewMessRequirement(phrase string, requirementType string, channelID string, lastUserID string) *MessRequirement {
	return &MessRequirement{Phrase: phrase, RequirementType: requirementType, ChannelID: channelID, LastUserID: lastUserID}
}

func (m *MessRequirement) SetPhrase(phrase string) {
	m.Lock()
	m.Phrase = phrase
	m.Unlock()
}

func (m *MessRequirement) GetPhrase() string {
	m.RLock()
	defer m.RUnlock()
	if m == nil {
		return ""
	}
	return m.Phrase
}

func (m *MessRequirement) SetRequirementType(requirementType string) {
	m.Lock()
	m.RequirementType = requirementType
	m.Unlock()
}

func (m *MessRequirement) GetRequirementType() string {
	m.RLock()
	defer m.RUnlock()
	if m == nil {
		return ""
	}
	return m.RequirementType
}

func (m *MessRequirement) SetChannelID(channelID string) {
	m.Lock()
	m.ChannelID = channelID
	m.Unlock()
}

func (m *MessRequirement) GetChannelID() string {
	m.RLock()
	defer m.RUnlock()
	if m == nil {
		return ""
	}
	return m.ChannelID
}

func (m *MessRequirement) SetLastUserID(lastUserID string) {
	m.Lock()
	m.LastUserID = lastUserID
	m.Unlock()
}

func (m *MessRequirement) GetLastUserID() string {
	m.RLock()
	defer m.RUnlock()
	if m == nil {
		return ""
	}
	return m.LastUserID
}