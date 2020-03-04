package entities

import "sync"

type Waifu struct {
	sync.RWMutex

	Name string `json:"Name"`
}

func NewWaifu(name string) Waifu {
	return Waifu{Name: name}
}

func (w Waifu) SetName(name string) Waifu {
	w.Name = name
	return w
}

func (w Waifu) GetName() string {
	if w == (Waifu{}) {
		return ""
	}
	return w.Name
}
