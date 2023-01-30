package entities

import "github.com/sasha-s/go-deadlock"

type ReactJoin struct {
	deadlock.RWMutex

	RoleEmojiMap []map[string][]string `json:"roleEmoji"`
}

func (r *ReactJoin) AppendToRoleEmojiMap(roleEmoji map[string][]string) {
	r.Lock()
	r.RoleEmojiMap = append(r.RoleEmojiMap, roleEmoji)
	r.Unlock()
}

func (r *ReactJoin) RemoveFromRoleEmojiMap(index int) {
	r.Lock()
	r.RoleEmojiMap = append(r.RoleEmojiMap[:index], r.RoleEmojiMap[index+1:]...)
	r.Unlock()
}

func (r *ReactJoin) SetRoleEmojiMap(roleEmojiMap []map[string][]string) {
	r.Lock()
	r.RoleEmojiMap = roleEmojiMap
	r.Unlock()
}

func (r *ReactJoin) GetRoleEmojiMap() []map[string][]string {
	r.RLock()
	defer r.RUnlock()
	if r == nil {
		return nil
	}
	return r.RoleEmojiMap
}
