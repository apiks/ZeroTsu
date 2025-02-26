package db

import (
	"github.com/r-anime/ZeroTsu/entities"
	"log"
)

// GetGuildSettings retrieves guild settings from MongoDB
func GetGuildSettings(guildID string) entities.GuildSettings {
	err := entities.InitGuildIfNotExists(guildID)
	if err != nil {
		log.Println(err)
		return entities.GuildSettings{}
	}

	settings, err := entities.LoadGuildSettings(guildID)
	if err != nil {
		log.Printf("Error loading guild settings for guild %s: %v\n", guildID, err)
		return entities.GuildSettings{}
	}

	return settings
}

// SetGuildSettings saves guild settings in MongoDB
func SetGuildSettings(guildID string, guildSettings entities.GuildSettings) {
	err := entities.InitGuildIfNotExists(guildID)
	if err != nil {
		log.Println(err)
		return
	}

	err = entities.SaveGuildSettings(guildID, guildSettings)
	if err != nil {
		log.Printf("Error saving guild settings for guild %s: %v\n", guildID, err)
	}
}
