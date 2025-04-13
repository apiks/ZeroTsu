package entities

import "github.com/sasha-s/go-deadlock"

type ShowAirTime struct {
	deadlock.RWMutex

	Name     string
	AirTime  string
	AirType  string
	Episode  string
	Delayed  string
	Key      string
	ImageUrl string
	Subbed   bool
	Donghua  bool
}

func NewShowAirTime(name, airTime, airType, episode, delayed, key, imageUrl string, subbed, donghua bool) *ShowAirTime {
	return &ShowAirTime{Name: name, AirTime: airTime, AirType: airType, Episode: episode, Delayed: delayed, Key: key, ImageUrl: imageUrl, Subbed: subbed, Donghua: donghua}
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

func (s *ShowAirTime) SetAirType(airType string) {
	s.Lock()
	s.AirType = airType
	s.Unlock()
}

func (s *ShowAirTime) GetAirType() string {
	s.RLock()
	defer s.RUnlock()
	if s == nil {
		return ""
	}
	return s.AirType
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

func (s *ShowAirTime) SetImageUrl(imageUrl string) {
	s.Lock()
	s.ImageUrl = imageUrl
	s.Unlock()
}

func (s *ShowAirTime) GetImageUrl() string {
	s.RLock()
	defer s.RUnlock()
	if s == nil {
		return ""
	}
	return s.ImageUrl
}

func (s *ShowAirTime) SetSubbed(subbed bool) {
	s.Lock()
	s.Subbed = subbed
	s.Unlock()
}

func (s *ShowAirTime) GetSubbed() bool {
	s.RLock()
	defer s.RUnlock()
	if s == nil {
		return false
	}
	return s.Subbed
}

func (s *ShowAirTime) SetDonghua(donghua bool) {
	s.Lock()
	s.Donghua = donghua
	s.Unlock()
}

func (s *ShowAirTime) GetDonghua() bool {
	s.RLock()
	defer s.RUnlock()
	if s == nil {
		return false
	}
	return s.Donghua
}
