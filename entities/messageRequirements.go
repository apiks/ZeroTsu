package entities

import "sync"

type MessRequirement struct {
	sync.RWMutex

	Phrase          string `json:"Phrase"`
	RequirementType string `json:"Type"`
	ChannelID       string `json:"Channel"`
	LastUserID      string
}

func NewMessRequirement(phrase string, requirementType string, channelID string, lastUserID string) MessRequirement {
	return MessRequirement{Phrase: phrase, RequirementType: requirementType, ChannelID: channelID, LastUserID: lastUserID}
}

func (m MessRequirement) SetPhrase(phrase string) MessRequirement {
	m.Phrase = phrase
	return m
}

func (m MessRequirement) GetPhrase() string {
	if m == (MessRequirement{}) {
		return ""
	}
	return m.Phrase
}

func (m MessRequirement) SetRequirementType(requirementType string) MessRequirement {
	m.RequirementType = requirementType
	return m
}

func (m MessRequirement) GetRequirementType() string {
	if m == (MessRequirement{}) {
		return ""
	}
	return m.RequirementType
}

func (m MessRequirement) SetChannelID(channelID string) MessRequirement {
	m.ChannelID = channelID
	return m
}

func (m MessRequirement) GetChannelID() string {
	if m == (MessRequirement{}) {
		return ""
	}
	return m.ChannelID
}

func (m MessRequirement) SetLastUserID(lastUserID string) MessRequirement {
	m.LastUserID = lastUserID
	return m
}

func (m MessRequirement) GetLastUserID() string {
	if m == (MessRequirement{}) {
		return ""
	}
	return m.LastUserID
}
