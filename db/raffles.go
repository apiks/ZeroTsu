package db

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/entities"
	"log"
	"strings"
)

// GetGuildRaffles retrieves all raffles from MongoDB
func GetGuildRaffles(guildID string) []*entities.Raffle {
	err := entities.InitGuildIfNotExists(guildID)
	if err != nil {
		log.Println(err)
		return nil
	}

	raffles, err := entities.LoadRaffles(guildID)
	if err != nil {
		return nil
	}

	return raffles
}

// SetGuildRaffle saves or deletes a guild raffle in MongoDB
func SetGuildRaffle(guildID string, raffle *entities.Raffle, delete ...bool) error {
	err := entities.InitGuildIfNotExists(guildID)
	if err != nil {
		return err
	}

	// Fetch premium status from MongoDB
	guildSettings, err := entities.LoadGuildSettings(guildID)
	if err != nil {
		log.Printf("Error fetching guild settings for guild %s: %v\n", guildID, err)
		return fmt.Errorf("Error: Could not determine premium status.")
	}

	// Enforce premium limit
	existingRaffles := GetGuildRaffles(guildID)
	limit := 50
	if guildSettings.Premium {
		limit = 200
	}
	if len(delete) == 0 && len(existingRaffles) >= limit {
		if guildSettings.Premium {
			return fmt.Errorf("Error: You have reached the raffle limit (200) for this premium server.")
		}
		return fmt.Errorf("Error: You have reached the raffle limit (50) for this server. Please remove some or upgrade to premium at <https://patreon.com/animeschedule>")
	}

	raffle.SetName(strings.ToLower(raffle.GetName()))

	if len(delete) == 0 {
		// Check if raffle already exists
		for _, existingRaffle := range existingRaffles {
			if strings.ToLower(existingRaffle.GetName()) == raffle.GetName() {
				return fmt.Errorf("Error: That raffle already exists.")
			}
		}

		// Save the new raffle to MongoDB
		err := entities.SaveRaffle(guildID, raffle)
		if err != nil {
			log.Printf("Error saving raffle for guild %s: %v\n", guildID, err)
			return err
		}
	} else {
		// Delete the raffle from MongoDB
		err := entities.DeleteRaffle(guildID, raffle)
		if err != nil {
			log.Printf("Error deleting raffle for guild %s: %v\n", guildID, err)
			return err
		}
	}

	return nil
}

// SetGuildRaffleParticipant adds or removes a participant from a raffle in MongoDB
func SetGuildRaffleParticipant(guildID, userID string, raffle *entities.Raffle, delete ...bool) {
	err := entities.InitGuildIfNotExists(guildID)
	if err != nil {
		log.Println(err)
		return
	}

	remove := len(delete) > 0

	err = entities.UpdateRaffleParticipant(guildID, userID, raffle, remove)
	if err != nil {
		log.Printf("Error updating participant for raffle in guild %s: %v\n", guildID, err)
	}
}
