package entities

import (
	"time"

	"github.com/sasha-s/go-deadlock"
)

type RemindMe struct {
	deadlock.RWMutex

	Message        string
	Date           time.Time
	CommandChannel string
	RemindID       int
}

func NewRemindMe(message string, date time.Time, commandChannel string, remindID int) *RemindMe {
	return &RemindMe{Message: message, Date: date, CommandChannel: commandChannel, RemindID: remindID}
}

func (r *RemindMe) SetMessage(message string) {
	r.Lock()
	r.Message = message
	r.Unlock()
}

func (r *RemindMe) GetMessage() string {
	r.RLock()
	defer r.RUnlock()
	if r == nil {
		return ""
	}
	return r.Message
}

func (r *RemindMe) SetDate(date time.Time) {
	r.Lock()
	r.Date = date
	r.Unlock()
}

func (r *RemindMe) GetDate() time.Time {
	r.RLock()
	defer r.RUnlock()
	if r == nil {
		return time.Time{}
	}
	return r.Date
}

func (r *RemindMe) SetCommandChannel(commandChannel string) {
	r.Lock()
	r.CommandChannel = commandChannel
	r.Unlock()
}

func (r *RemindMe) GetCommandChannel() string {
	r.RLock()
	defer r.RUnlock()
	if r == nil {
		return ""
	}
	return r.CommandChannel
}

func (r *RemindMe) AddToRemindID(remindID int) {
	r.Lock()
	r.RemindID += remindID
	r.Unlock()
}

func (r *RemindMe) SetRemindID(remindID int) {
	r.Lock()
	r.RemindID = remindID
	r.Unlock()
}

func (r *RemindMe) GetRemindID() int {
	r.RLock()
	defer r.RUnlock()
	if r == nil {
		return 0
	}
	return r.RemindID
}
