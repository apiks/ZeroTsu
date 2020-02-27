package entities

import (
	"sync"
	"time"
)

type FeedCheck struct {
	sync.RWMutex

	Feed *Feed     `json:"Thread"`
	Date time.Time `json:"Date"`
	GUID string    `json:"GUID"`
}

func NewFeedCheck(feed *Feed, date time.Time, GUID string) *FeedCheck {
	return &FeedCheck{Feed: feed, Date: date, GUID: GUID}
}

func (f *FeedCheck) SetFeed(feed *Feed) {
	f.Lock()
	f.Feed = feed
	f.Unlock()
}

func (f *FeedCheck) GetFeed() *Feed {
	f.RLock()
	defer f.RUnlock()
	if f == nil {
		return nil
	}
	return f.Feed
}

func (f *FeedCheck) SetDate(date time.Time) {
	f.Lock()
	f.Date = date
	f.Unlock()
}

func (f *FeedCheck) GetDate() time.Time {
	f.RLock()
	defer f.RUnlock()
	if f == nil {
		return time.Time{}
	}
	return f.Date
}

func (f *FeedCheck) SetGUID(guid string) {
	f.Lock()
	f.GUID = guid
	f.Unlock()
}

func (f *FeedCheck) GetGUID() string {
	f.RLock()
	defer f.RUnlock()
	if f == nil {
		return ""
	}
	return f.GUID
}