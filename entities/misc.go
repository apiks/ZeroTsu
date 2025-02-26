package entities

import (
	"github.com/sasha-s/go-deadlock"
)

var (
	Mutex         deadlock.RWMutex
	AnimeSchedule = &AnimeScheduleMap{AnimeSchedule: make(map[int][]*ShowAirTime)}
)

type AnimeScheduleMap struct {
	deadlock.RWMutex
	AnimeSchedule map[int][]*ShowAirTime
}
