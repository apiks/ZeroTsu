package entities

import "sync"

type Waifu struct {
	sync.RWMutex

	Name string `json:"Name"`
}

func NewWaifu(name string) *Waifu {
	return &Waifu{Name: name}
}

func (w *Waifu) SetName(name string) {
	w.Lock()
	w.Name = name
	w.Unlock()
}

func (w *Waifu) GetName() string {
	w.RLock()
	defer w.RUnlock()
	if w == nil {
		return ""
	}
	return w.Name
}