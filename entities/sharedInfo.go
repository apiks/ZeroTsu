package entities

import "sync"

type sharedInfo struct {
	sync.RWMutex

	RemindMes map[string]*RemindMeSlice
	AnimeSubs map[string][]*ShowSub
}

func newSharedInfo(remindMes map[string]*RemindMeSlice, animeSubs map[string][]*ShowSub) *sharedInfo {
	return &sharedInfo{RemindMes: remindMes, AnimeSubs: animeSubs}
}

func (s *sharedInfo) SetRemindMesMap(remindMes map[string]*RemindMeSlice) {
	s.Lock()
	s.RemindMes = remindMes
	s.Unlock()
}

func (s *sharedInfo) GetRemindMesMap() map[string]*RemindMeSlice {
	s.RLock()
	defer s.RUnlock()
	if s == nil {
		return nil
	}
	return s.RemindMes
}

func (s *sharedInfo) SetAnimeSubsMap(animeSubs map[string][]*ShowSub) {
	s.Lock()
	s.AnimeSubs = animeSubs
	s.Unlock()
}

func (s *sharedInfo) GetAnimeSubsMap() map[string][]*ShowSub {
	s.RLock()
	defer s.RUnlock()
	if s == nil {
		return nil
	}
	return s.AnimeSubs
}