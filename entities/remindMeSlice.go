package entities

import "sync"

type RemindMeSlice struct {
	sync.RWMutex

	RemindMeSlice []*RemindMe
	Premium       bool
}

func NewRemindMeSlice(remindMeSlice []*RemindMe, premium bool) *RemindMeSlice {
	return &RemindMeSlice{RemindMeSlice: remindMeSlice, Premium: premium}
}

func (r *RemindMeSlice) AppendToRemindMeSlice(remindMe *RemindMe) {
	r.Lock()
	r.RemindMeSlice = append(r.RemindMeSlice, remindMe)
	r.Unlock()
}

func (r *RemindMeSlice) RemoveFromRemindMeSlice(index int) {
	r.Lock()
	if index < len(r.RemindMeSlice)-1 {
		copy(r.RemindMeSlice[index:], r.RemindMeSlice[index+1:])
	}
	r.RemindMeSlice[len(r.RemindMeSlice)-1] = nil
	r.RemindMeSlice = r.RemindMeSlice[:len(r.RemindMeSlice)-1]
	r.Unlock()
}

func (r *RemindMeSlice) SetRemindMeSlice(remindMeSlice []*RemindMe) {
	r.Lock()
	r.RemindMeSlice = remindMeSlice
	r.Unlock()
}

func (r *RemindMeSlice) GetRemindMeSlice() []*RemindMe {
	r.RLock()
	defer r.RUnlock()
	if r == nil {
		return nil
	}
	return r.RemindMeSlice
}

func (r *RemindMeSlice) SetPremium(premium bool) {
	r.Lock()
	r.Premium = premium
	r.Unlock()
}

func (r *RemindMeSlice) GetPremium() bool {
	r.RLock()
	defer r.RUnlock()
	if r == nil {
		return false
	}
	return r.Premium
}
