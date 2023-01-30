package entities

import "github.com/sasha-s/go-deadlock"

type sharedInfo struct {
	deadlock.RWMutex

	RemindMes map[string]*RemindMeSlice
	AnimeSubs map[string][]*ShowSub
}

func newSharedInfo(remindMes map[string]*RemindMeSlice, animeSubs map[string][]*ShowSub) *sharedInfo {
	return &sharedInfo{RemindMes: remindMes, AnimeSubs: animeSubs}
}

func (s *sharedInfo) SetRemindMesMap(remindMes map[string]*RemindMeSlice) {
	if s == nil {
		return
	}
	s.Lock()
	s.RemindMes = remindMes
	s.Unlock()
}

func (s *sharedInfo) GetRemindMesMap() map[string]*RemindMeSlice {
	if s == nil {
		return nil
	}
	s.RLock()
	defer s.RUnlock()
	return s.RemindMes
}

func (s *sharedInfo) SetAnimeSubsMap(animeSubs map[string][]*ShowSub) {
	if s == nil {
		return
	}
	s.Lock()
	s.AnimeSubs = animeSubs
	s.Unlock()
}

func (s *sharedInfo) GetAnimeSubsMap() map[string][]*ShowSub {
	if s == nil {
		return nil
	}
	s.RLock()
	defer s.RUnlock()
	return s.AnimeSubs
}
