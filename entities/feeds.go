package entities

import "sync"

type Feed struct {
	sync.RWMutex

	Subreddit string `json:"Subreddit"`
	Title     string `json:"Title"`
	Author    string `json:"Author"`
	Pin       bool   `json:"Pin"`
	PostType  string `json:"PostType"`
	ChannelID string `json:"ChannelID"`
}

func NewFeed(subreddit string, title string, author string, pin bool, postType string, channelID string) *Feed {
	return &Feed{Subreddit: subreddit, Title: title, Author: author, Pin: pin, PostType: postType, ChannelID: channelID}
}

func (f *Feed) SetSubreddit(subreddit string) {
	f.Lock()
	f.Subreddit = subreddit
	f.Unlock()
}

func (f *Feed) GetSubreddit() string {
	f.RLock()
	defer f.RUnlock()
	if f == nil {
		return ""
	}
	return f.Subreddit
}

func (f *Feed) SetTitle(title string) {
	f.Lock()
	f.Title = title
	f.Unlock()
}

func (f *Feed) GetTitle() string {
	f.RLock()
	defer f.RUnlock()
	if f == nil {
		return ""
	}
	return f.Title
}

func (f *Feed) SetAuthor(author string) {
	f.Lock()
	f.Author = author
	f.Unlock()
}

func (f *Feed) GetAuthor() string {
	f.RLock()
	defer f.RUnlock()
	if f == nil {
		return ""
	}
	return f.Author
}

func (f *Feed) SetPin(pin bool) {
	f.Lock()
	f.Pin = pin
	f.Unlock()
}

func (f *Feed) GetPin() bool {
	f.RLock()
	defer f.RUnlock()
	if f == nil {
		return false
	}
	return f.Pin
}

func (f *Feed) SetPostType(postType string) {
	f.Lock()
	f.PostType = postType
	f.Unlock()
}

func (f *Feed) GetPostType() string {
	f.RLock()
	defer f.RUnlock()
	if f == nil {
		return ""
	}
	return f.PostType
}

func (f *Feed) SetChannelID(channelID string) {
	f.Lock()
	f.ChannelID = channelID
	f.Unlock()
}

func (f *Feed) GetChannelID() string {
	f.RLock()
	defer f.RUnlock()
	if f == nil {
		return ""
	}
	return f.ChannelID
}