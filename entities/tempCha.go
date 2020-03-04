package entities

import (
	"sync"
	"time"
)

type TempChaInfo struct {
	sync.RWMutex

	CreationDate time.Time `json:"CreationDate"`
	RoleName     string    `json:"RoleName"`
	Elevated     bool      `json:"Permission"`
}

func NewTempChaInfo(creationDate time.Time, roleName string, elevated bool) *TempChaInfo {
	return &TempChaInfo{CreationDate: creationDate, RoleName: roleName, Elevated: elevated}
}

func (t *TempChaInfo) SetCreationDate(creationDate time.Time) {
	t.Lock()
	t.CreationDate = creationDate
	t.Unlock()
}

func (t *TempChaInfo) GetCreationDate() time.Time {
	t.RLock()
	defer t.RUnlock()
	if t == nil {
		return time.Time{}
	}
	return t.CreationDate
}

func (t *TempChaInfo) SetRoleName(roleName string) {
	t.Lock()
	t.RoleName = roleName
	t.Unlock()
}

func (t *TempChaInfo) GetRoleName() string {
	t.RLock()
	defer t.RUnlock()
	if t == nil {
		return ""
	}
	return t.RoleName
}

func (t *TempChaInfo) SetElevated(elevated bool) {
	t.Lock()
	t.Elevated = elevated
	t.Unlock()
}

func (t *TempChaInfo) GetElevated() bool {
	t.RLock()
	defer t.RUnlock()
	if t == nil {
		return false
	}
	return t.Elevated
}
