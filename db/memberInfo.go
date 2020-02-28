package db

import (
	"github.com/r-anime/ZeroTsu/entities"
)

// GetGuildMemberInfo returns a guild memberInfo map from in-memory
func GetGuildMemberInfo(guildID string) map[string]entities.UserInfo {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	return entities.Guilds.DB[guildID].GetMemberInfoMap()
}

// SetGuildMemberInfo sets a target guild's memberInfo map in-memory
func SetGuildMemberInfo(guildID string, memberInfo map[string]entities.UserInfo) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].SetMemberInfoMap(memberInfo)
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("memberInfo", entities.Guilds.DB[guildID].GetMemberInfoMap())
}

// GetGuildMember a guild's member object from in-memory
func GetGuildMember(guildID string, userID string) entities.UserInfo {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	if member, ok := entities.Guilds.DB[guildID].GetMemberInfoMap()[userID]; ok {
		return member
	}

	return entities.UserInfo{}
}

// SetGuildMember sets a target guild's member object in-memory
func SetGuildMember(guildID string, member entities.UserInfo, deleteSlice ...bool) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	if len(deleteSlice) == 0 {
		entities.Guilds.DB[guildID].AssignToMemberInfoMap(member.GetID(), member)
	} else {
		entities.Guilds.DB[guildID].RemoveFromMemberInfoMap(member.GetID())
	}
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("memberInfo", entities.Guilds.DB[guildID].GetMemberInfoMap())
}
