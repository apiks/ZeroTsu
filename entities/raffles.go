package entities

import "sync"

type Raffle struct {
	sync.RWMutex

	Name           string   `json:"Name"`
	ParticipantIDs []string `json:"ParticipantIDs"`
	ReactMessageID string   `json:"ReactMessageID"`
}

func NewRaffle(name string, participantIDs []string, reactMessageID string) *Raffle {
	return &Raffle{Name: name, ParticipantIDs: participantIDs, ReactMessageID: reactMessageID}
}

func (r *Raffle) SetName(name string) {
	r.Lock()
	r.Name = name
	r.Unlock()
}

func (r *Raffle) GetName() string {
	r.RLock()
	defer r.RUnlock()
	if r == nil {
		return ""
	}
	return r.Name
}

func (r *Raffle) AppendToParticipantIDs(participantID string) {
	r.Lock()
	r.ParticipantIDs = append(r.ParticipantIDs, participantID)
	r.Unlock()
}

func (r *Raffle) RemoveFromParticipantIDs(index int) {
	r.Lock()
	r.ParticipantIDs = append(r.ParticipantIDs[:index], r.ParticipantIDs[index+1:]...)
	r.Unlock()
}

func (r *Raffle) SetParticipantIDs(participantIDs []string) {
	r.Lock()
	r.ParticipantIDs = participantIDs
	r.Unlock()
}

func (r *Raffle) GetParticipantIDs() []string {
	r.RLock()
	defer r.RUnlock()
	if r == nil {
		return nil
	}
	return r.ParticipantIDs
}

func (r *Raffle) SetReactMessageID(reactMessageID string) {
	r.Lock()
	r.ReactMessageID = reactMessageID
	r.Unlock()
}

func (r *Raffle) GetReactMessageID() string {
	r.RLock()
	defer r.RUnlock()
	if r == nil {
		return ""
	}
	return r.ReactMessageID
}