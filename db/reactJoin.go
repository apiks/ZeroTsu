package db

import (
	"github.com/r-anime/ZeroTsu/entities"
	"log"
)

// GetGuildReactJoin retrieves the ReactJoin map from MongoDB
func GetGuildReactJoin(guildID string) map[string]*entities.ReactJoin {
	err := entities.InitGuildIfNotExists(guildID)
	if err != nil {
		log.Println(err)
		return nil
	}

	reactJoinMap, err := entities.LoadReactJoinMap(guildID)
	if err != nil {
		return nil
	}

	return reactJoinMap
}

// SetGuildReactJoin saves or deletes a guild's ReactJoin map in MongoDB
func SetGuildReactJoin(guildID string, reactJoin map[string]*entities.ReactJoin, delete ...bool) {
	err := entities.InitGuildIfNotExists(guildID)
	if err != nil {
		log.Println(err)
		return
	}

	if len(delete) == 0 {
		err := entities.SaveReactJoinMap(guildID, reactJoin)
		if err != nil {
			log.Printf("Error saving react join map for guild %s: %v\n", guildID, err)
		}
	} else {
		err := entities.DeleteReactJoinMap(guildID)
		if err != nil {
			log.Printf("Error deleting react join map for guild %s: %v\n", guildID, err)
		}
	}
}

// SetGuildReactJoinEmoji saves or removes a ReactJoin emoji entry in MongoDB
func SetGuildReactJoinEmoji(guildID, messageID string, reactJoinEmoji *entities.ReactJoin, delete ...bool) {
	err := entities.InitGuildIfNotExists(guildID)
	if err != nil {
		log.Println(err)
		return
	}

	if len(delete) == 0 {
		err := entities.SaveReactJoinEntry(guildID, messageID, reactJoinEmoji)
		if err != nil {
			log.Printf("Error saving react join emoji for guild %s: %v\n", guildID, err)
		}
	} else {
		err := entities.DeleteReactJoinEntry(guildID, messageID)
		if err != nil {
			log.Printf("Error deleting react join emoji for guild %s: %v\n", guildID, err)
		}
	}
}
