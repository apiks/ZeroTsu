package entities

import "github.com/sasha-s/go-deadlock"

type ShowSub struct {
	deadlock.RWMutex

	Show     string `json:"Show"`
	Notified bool   `json:"Notified"`
	Guild    bool   `json:"Guild"`
}

func NewShowSub(show string, notified bool, guild bool) *ShowSub {
	return &ShowSub{Show: show, Notified: notified, Guild: guild}
}

func (s *ShowSub) SetShow(show string) {
	s.Lock()
	s.Show = show
	s.Unlock()
}

func (s *ShowSub) GetShow() string {
	s.RLock()
	defer s.RUnlock()
	if s == nil {
		return ""
	}
	return s.Show
}

func (s *ShowSub) SetNotified(notified bool) {
	s.Lock()
	s.Notified = notified
	s.Unlock()
}

func (s *ShowSub) GetNotified() bool {
	s.RLock()
	defer s.RUnlock()
	if s == nil {
		return false
	}
	return s.Notified
}

func (s *ShowSub) SetGuild(guild bool) {
	s.Lock()
	s.Guild = guild
	s.Unlock()
}

func (s *ShowSub) GetGuild() bool {
	s.RLock()
	defer s.RUnlock()
	if s == nil {
		return false
	}
	return s.Guild
}
