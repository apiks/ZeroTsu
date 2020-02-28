package db

import (
	"github.com/r-anime/ZeroTsu/entities"
)

// GetGuildPunishedUsers returns the guild's punished users from in-memory
func GetGuildPunishedUsers(guildID string) []entities.PunishedUsers {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	return entities.Guilds.DB[guildID].GetPunishedUsers()
}

// GetGuildPunishedUser returns a guild's punished user object from in-memory
func GetGuildPunishedUser(guildID string, userID string) entities.PunishedUsers {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	for _, guildPunishedUser := range entities.Guilds.DB[guildID].GetPunishedUsers() {
		if guildPunishedUser.GetID() == userID {
			return guildPunishedUser
		}
	}

	return entities.PunishedUsers{}
}

// SetGuildPunishedUser sets a guild's punished user object in-memory
func SetGuildPunishedUser(guildID string, punishedUser entities.PunishedUsers, delete ...bool) error {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()

	if len(delete) == 0 {
		var exists bool
		for i, guildPunishedUser := range entities.Guilds.DB[guildID].GetPunishedUsers() {
			if guildPunishedUser.ID == punishedUser.ID {
				entities.Guilds.DB[guildID].AssignToPunishedUsers(i, punishedUser)
				exists = true
				break
			}
		}

		if !exists {
			entities.Guilds.DB[guildID].AppendToPunishedUsers(punishedUser)
		}
	} else {
		deleteGuildPunishedUser(guildID, punishedUser)
	}

	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("punishedUsers", entities.Guilds.DB[guildID].GetPunishedUsers())

	return nil
}

// deleteGuildMessageRequirement safely deletes a message requirement from the message requirements slice
func deleteGuildPunishedUser(guildID string, punishedUser entities.PunishedUsers) {
	for i, guildPunishedUser := range entities.Guilds.DB[guildID].GetPunishedUsers() {
		if guildPunishedUser.ID == punishedUser.ID {
			entities.Guilds.DB[guildID].RemoveFromPunishedUsers(i)
			break
		}
	}
}
