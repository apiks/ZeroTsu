package entities

import "sync"

type Filter struct {
	sync.RWMutex

	Filter string `json:"Filter"`
}

func NewFilter(filter string) *Filter {
	return &Filter{Filter: filter}
}

func (f *Filter) SetFilter(filter string) {
	f.Lock()
	f.Filter = filter
	f.Unlock()
}

func (f *Filter) GetFilter() string {
	f.RLock()
	defer f.RUnlock()
	if f == nil {
		return ""
	}
	return f.Filter
}