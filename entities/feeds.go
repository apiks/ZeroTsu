package entities

import "github.com/sasha-s/go-deadlock"

type Feed struct {
	deadlock.RWMutex

	Subreddit string `json:"Subreddit"`
	Title     string `json:"Title"`
	Author    string `json:"Author"`
	Pin       bool   `json:"Pin"`
	PostType  string `json:"PostType"`
	ChannelID string `json:"ChannelID"`
}

func NewFeed(subreddit string, title string, author string, pin bool, postType string, channelID string) Feed {
	return Feed{Subreddit: subreddit, Title: title, Author: author, Pin: pin, PostType: postType, ChannelID: channelID}
}

func (f Feed) SetSubreddit(subreddit string) Feed {
	f.Subreddit = subreddit
	return f
}

func (f Feed) GetSubreddit() string {
	if f == (Feed{}) {
		return ""
	}
	return f.Subreddit
}

func (f Feed) SetTitle(title string) Feed {
	f.Title = title
	return f
}

func (f Feed) GetTitle() string {
	if f == (Feed{}) {
		return ""
	}
	return f.Title
}

func (f Feed) SetAuthor(author string) Feed {
	f.Author = author
	return f
}

func (f Feed) GetAuthor() string {
	if f == (Feed{}) {
		return ""
	}
	return f.Author
}

func (f Feed) SetPin(pin bool) Feed {
	f.Pin = pin
	return f
}

func (f Feed) GetPin() bool {
	if f == (Feed{}) {
		return false
	}
	return f.Pin
}

func (f Feed) SetPostType(postType string) Feed {
	f.PostType = postType
	return f
}

func (f Feed) GetPostType() string {
	if f == (Feed{}) {
		return ""
	}
	return f.PostType
}

func (f Feed) SetChannelID(channelID string) Feed {
	f.ChannelID = channelID
	return f
}

func (f Feed) GetChannelID() string {
	if f == (Feed{}) {
		return ""
	}
	return f.ChannelID
}
