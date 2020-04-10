package db

import (
	"github.com/r-anime/ZeroTsu/entities"
	"log"
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

	if member.GetID() == "692715248410558595" {
		log.Println("1.1")
	}

	if len(deleteSlice) == 0 {
		if member.GetID() == "692715248410558595" {
			log.Println("1.2")
		}
		entities.Guilds.DB[guildID].AssignToMemberInfoMap(member.GetID(), member)
		if member.GetID() == "692715248410558595" {
			log.Println("1.3")
		}
	} else {
		entities.Guilds.DB[guildID].RemoveFromMemberInfoMap(member.GetID())
	}

	if member.GetID() == "692715248410558595" {
		log.Println("1.4")
	}

	entities.Guilds.DB[guildID].WriteData("memberInfo", entities.Guilds.DB[guildID].GetMemberInfoMap())

	if member.GetID() == "692715248410558595" {
		log.Println("1.5")
	}
}
