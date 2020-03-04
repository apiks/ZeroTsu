package entities

import "sync"

type ShowAirTime struct {
	sync.RWMutex

	Name    string
	AirTime string
	Episode string
	Delayed string
	Key     string
}

func NewShowAirTime(name string, airTime string, episode string, delayed string, key string) *ShowAirTime {
	return &ShowAirTime{Name: name, AirTime: airTime, Episode: episode, Delayed: delayed, Key: key}
}

func (s *ShowAirTime) SetName(name string) {
	s.Lock()
	s.Name = name
	s.Unlock()
}

func (s *ShowAirTime) GetName() string {
	s.RLock()
	defer s.RUnlock()
	if s == nil {
		return ""
	}
	return s.Name
}

func (s *ShowAirTime) SetAirTime(airTime string) {
	s.Lock()
	s.AirTime = airTime
	s.Unlock()
}

func (s *ShowAirTime) GetAirTime() string {
	s.RLock()
	defer s.RUnlock()
	if s == nil {
		return ""
	}
	return s.AirTime
}

func (s *ShowAirTime) SetEpisode(episode string) {
	s.Lock()
	s.Episode = episode
	s.Unlock()
}

func (s *ShowAirTime) GetEpisode() string {
	s.RLock()
	defer s.RUnlock()
	if s == nil {
		return ""
	}
	return s.Episode
}

func (s *ShowAirTime) SetDelayed(delayed string) {
	s.Lock()
	s.Delayed = delayed
	s.Unlock()
}

func (s *ShowAirTime) GetDelayed() string {
	s.RLock()
	defer s.RUnlock()
	if s == nil {
		return ""
	}
	return s.Delayed
}

func (s *ShowAirTime) SetKey(key string) {
	s.Lock()
	s.Key = key
	s.Unlock()
}

func (s *ShowAirTime) GetKey() string {
	s.RLock()
	defer s.RUnlock()
	if s == nil {
		return ""
	}
	return s.Key
}
