package db

import (
	"github.com/r-anime/ZeroTsu/entities"
	"log"
)

// GetGuildFeedChecks retrieves all feed checks from MongoDB
func GetGuildFeedChecks(guildID string, limit int) []entities.FeedCheck {
	err := entities.InitGuildIfNotExists(guildID)
	if err != nil {
		log.Println(err)
		return nil
	}

	feedChecks, err := entities.LoadFeedChecks(guildID, limit)
	if err != nil {
		return nil
	}

	return feedChecks
}

// SetGuildFeedCheck saves or deletes a guild feed check in MongoDB
func SetGuildFeedCheck(guildID string, feedCheck entities.FeedCheck, delete ...bool) {
	err := entities.InitGuildIfNotExists(guildID)
	if err != nil {
		log.Println(err)
		return
	}

	if len(delete) == 0 {
		err := entities.SaveFeedCheck(guildID, feedCheck)
		if err != nil {
			log.Printf("Error saving feed check for guild %s: %v\n", guildID, err)
		}
	} else {
		err := entities.DeleteFeedCheck(guildID, feedCheck)
		if err != nil {
			log.Printf("Error deleting feed check for guild %s: %v\n", guildID, err)
		}
	}
}

// SetGuildFeedChecks saves or deletes multiple feed checks in MongoDB
func SetGuildFeedChecks(guildID string, feedChecks []entities.FeedCheck, delete ...bool) {
	err := entities.InitGuildIfNotExists(guildID)
	if err != nil {
		log.Println(err)
		return
	}

	if len(delete) == 0 {
		err := entities.SaveMultipleFeedChecks(guildID, feedChecks)
		if err != nil {
			log.Printf("Error saving feed checks for guild %s: %v\n", guildID, err)
		}
	} else {
		err := entities.DeleteMultipleFeedChecks(guildID, feedChecks)
		if err != nil {
			log.Printf("Error deleting multiple feed checks for guild %s: %v\n", guildID, err)
		}
	}
}
