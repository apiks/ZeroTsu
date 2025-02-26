package db

import (
	"github.com/r-anime/ZeroTsu/entities"
	"log"
)

// GetGuildAutopost retrieves an autopost from MongoDB
func GetGuildAutopost(guildID string, postType string) entities.Cha {
	err := entities.InitGuildIfNotExists(guildID)
	if err != nil {
		log.Println(err)
		return entities.Cha{}
	}

	autopost, err := entities.LoadAutopost(guildID, postType)
	if err != nil {
		return entities.Cha{}
	}

	return autopost
}

// SetGuildAutopost saves an autopost entry in MongoDB
func SetGuildAutopost(guildID string, postType string, autopost entities.Cha) {
	err := entities.InitGuildIfNotExists(guildID)
	if err != nil {
		log.Println(err)
		return
	}

	err = entities.SaveAutopost(guildID, postType, autopost)
	if err != nil {
		log.Printf("Error saving autopost for guild %s: %v\n", guildID, err)
	}
}

// RemoveGuildAutopost deletes an autopost entry from MongoDB
func RemoveGuildAutopost(guildID string, postType string) {
	err := entities.InitGuildIfNotExists(guildID)
	if err != nil {
		log.Println(err)
		return
	}

	err = entities.DeleteAutopost(guildID, postType)
	if err != nil {
		log.Printf("Error deleting autopost %s for guild %s: %v\n", postType, guildID, err)
	}
}
