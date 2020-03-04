package entities

import "sync"

type Filter struct {
	sync.RWMutex

	Filter string `json:"Filter"`
}

func NewFilter(filter string) Filter {
	return Filter{Filter: filter}
}

func (f Filter) SetFilter(filter string) Filter {
	f.Filter = filter
	return f
}

func (f Filter) GetFilter() string {
	if f == (Filter{}) {
		return ""
	}
	return f.Filter
}
