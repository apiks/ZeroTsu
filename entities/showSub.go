package entities

import "github.com/sasha-s/go-deadlock"

type ShowSub struct {
	deadlock.RWMutex

	Show     string `json:"Show"`
	Notified bool   `json:"Notified"`
	Guild    bool   `json:"Guild"`
}

type ShowSubMongo struct {
	Show     string `json:"Show"`
	Notified bool   `json:"Notified"`
	Guild    bool   `json:"Guild"`
}

func NewShowSub(show string, notified bool, guild bool) *ShowSub {
	return &ShowSub{Show: show, Notified: notified, Guild: guild}
}

func (s *ShowSub) SetShow(show string) {
	if s == nil {
		return
	}
	s.Lock()
	s.Show = show
	s.Unlock()
}

func (s *ShowSub) GetShow() string {
	if s == nil {
		return ""
	}
	s.RLock()
	defer s.RUnlock()
	return s.Show
}

func (s *ShowSub) SetNotified(notified bool) {
	if s == nil {
		return
	}
	s.Lock()
	s.Notified = notified
	s.Unlock()
}

func (s *ShowSub) GetNotified() bool {
	if s == nil {
		return false
	}
	s.RLock()
	defer s.RUnlock()
	return s.Notified
}

func (s *ShowSub) SetGuild(guild bool) {
	if s == nil {
		return
	}
	s.Lock()
	s.Guild = guild
	s.Unlock()
}

func (s *ShowSub) GetGuild() bool {
	if s == nil {
		return false
	}
	s.RLock()
	defer s.RUnlock()
	return s.Guild
}
