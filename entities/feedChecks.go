package entities

import (
	"sync"
	"time"
)

type FeedCheck struct {
	sync.RWMutex

	Feed Feed      `json:"Thread"`
	Date time.Time `json:"Date"`
	GUID string    `json:"GUID"`
}

func NewFeedCheck(feed Feed, date time.Time, GUID string) FeedCheck {
	return FeedCheck{Feed: feed, Date: date, GUID: GUID}
}

func (f FeedCheck) SetFeed(feed Feed) FeedCheck {
	f.Feed = feed
	return f
}

func (f FeedCheck) GetFeed() Feed {
	if f == (FeedCheck{}) {
		return Feed{}
	}
	return f.Feed
}

func (f FeedCheck) SetDate(date time.Time) FeedCheck {
	f.Date = date
	return f
}

func (f FeedCheck) GetDate() time.Time {
	if f == (FeedCheck{}) {
		return time.Time{}
	}
	return f.Date
}

func (f FeedCheck) SetGUID(guid string) FeedCheck {
	f.GUID = guid
	return f
}

func (f FeedCheck) GetGUID() string {
	if f == (FeedCheck{}) {
		return ""
	}
	return f.GUID
}
