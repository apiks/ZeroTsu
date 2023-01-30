package entities

import (
	"time"

	"github.com/sasha-s/go-deadlock"
)

type ASAnime struct {
	deadlock.RWMutex

	Title             string    `json:"title"`
	Route             string    `json:"route"`
	EpisodeDate       time.Time `json:"episodeDate"`
	EpisodeNumber     int       `json:"episodeNumber"`
	Episodes          int       `json:"episodes`
	DelayedFrom       time.Time `json:"delayedFrom"`
	DelayedUntil      time.Time `json:"delayedUntil"`
	AirType           string    `json:"airType"`
	ImageVersionRoute string    `json:"imageVersionRoute"`
	Donghua           bool      `json:"donghua"`
}

func NewASAnime(title, route string, episodeDate time.Time, episodeNumber, episodes int, delayedFrom time.Time, delayedUntil time.Time,
	airType, imageVersionRoute string, donghua bool) *ASAnime {
	return &ASAnime{
		Title:             title,
		Route:             route,
		EpisodeDate:       episodeDate,
		EpisodeNumber:     episodeNumber,
		Episodes:          episodes,
		DelayedFrom:       delayedFrom,
		DelayedUntil:      delayedFrom,
		AirType:           airType,
		ImageVersionRoute: imageVersionRoute,
		Donghua:           donghua,
	}
}

func (a *ASAnime) SetTitle(title string) {
	a.Lock()
	a.Title = title
	a.Unlock()
}

func (a *ASAnime) GetTitle() string {
	a.RLock()
	defer a.RUnlock()
	if a == nil {
		return ""
	}
	return a.Title
}

func (a *ASAnime) SetRoute(route string) {
	a.Lock()
	a.Route = route
	a.Unlock()
}

func (a *ASAnime) GetRoute() string {
	a.RLock()
	defer a.RUnlock()
	if a == nil {
		return ""
	}
	return a.Route
}

func (a *ASAnime) SetEpisodeDate(episodeDate time.Time) {
	a.Lock()
	a.EpisodeDate = episodeDate
	a.Unlock()
}

func (a *ASAnime) GetEpisodeDate() time.Time {
	a.RLock()
	defer a.RUnlock()
	if a == nil {
		return (time.Time{})
	}
	return a.EpisodeDate
}

func (a *ASAnime) SetEpisodeNumber(episodeNumber int) {
	a.Lock()
	a.EpisodeNumber = episodeNumber
	a.Unlock()
}

func (a *ASAnime) GetEpisodeNumber() int {
	a.RLock()
	defer a.RUnlock()
	if a == nil {
		return 0
	}
	return a.EpisodeNumber
}

func (a *ASAnime) SetEpisodes(episodes int) {
	a.Lock()
	a.Episodes = episodes
	a.Unlock()
}

func (a *ASAnime) GetEpisodes() int {
	a.RLock()
	defer a.RUnlock()
	if a == nil {
		return 0
	}
	return a.Episodes
}

func (a *ASAnime) SetDelayedFrom(delayedFrom time.Time) {
	a.Lock()
	a.DelayedFrom = delayedFrom
	a.Unlock()
}

func (a *ASAnime) GetDelayedFrom() time.Time {
	a.RLock()
	defer a.RUnlock()
	if a == nil {
		return (time.Time{})
	}
	return a.DelayedFrom
}

func (a *ASAnime) SetDelayedUntil(delayedUntil time.Time) {
	a.Lock()
	a.DelayedUntil = delayedUntil
	a.Unlock()
}

func (a *ASAnime) GetDelayedUntil() time.Time {
	a.RLock()
	defer a.RUnlock()
	if a == nil {
		return (time.Time{})
	}
	return a.DelayedUntil
}

func (a *ASAnime) SetAirType(airType string) {
	a.Lock()
	a.AirType = airType
	a.Unlock()
}

func (a *ASAnime) GetAirType() string {
	a.RLock()
	defer a.RUnlock()
	if a == nil {
		return ""
	}
	return a.AirType
}

func (a *ASAnime) SetImageVersionRoute(imageVersionRoute string) {
	a.Lock()
	a.ImageVersionRoute = imageVersionRoute
	a.Unlock()
}

func (a *ASAnime) GetImageVersionRoute() string {
	a.RLock()
	defer a.RUnlock()
	if a == nil {
		return ""
	}
	return a.ImageVersionRoute
}

func (a *ASAnime) SetDonghua(donghua bool) {
	a.Lock()
	a.Donghua = donghua
	a.Unlock()
}

func (a *ASAnime) GetDonghua() bool {
	a.RLock()
	defer a.RUnlock()
	if a == nil {
		return false
	}
	return a.Donghua
}
