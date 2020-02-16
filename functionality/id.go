package functionality

import (
	"fmt"
	"math/rand"
)

const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"

// GenerateID generates 4 char random unused ID
func GenerateID(guildID string) (string, error) {
	var (
		isUnusedIDFlag bool
		randID         string
	)

	for !isUnusedIDFlag {
		randID = generateRandID(4)
		isUnusedIDFlag = isUnusedID(randID, guildID)
	}

	if randID == "" {
		return "", fmt.Errorf("error: cannot generate a random ID")
	}

	return randID, nil
}

// generateRandID generates a random ID of n char length
func generateRandID(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = characters[rand.Intn(len(characters))]
	}
	return string(b)
}

// isUnusedID checks if an id is already in use
func isUnusedID(id string, guildID string) bool {
	for _, messReq := range GuildMap[guildID].MessageRequirements {
		if messReq != nil && messReq.ID != "" && messReq.ID == id {
			return false
		}
	}

	for _, feed := range GuildMap[guildID].Feeds {
		if feed != nil && feed.ID != "" && feed.ID == id {
			return false
		}
	}

	for _, feedCheck := range GuildMap[guildID].FeedChecks {
		if feedCheck != nil && feedCheck.ID != "" && feedCheck.ID == id {
			return false
		}
	}

	for _, raffle := range GuildMap[guildID].Raffles {
		if raffle != nil && raffle.ID != "" && raffle.ID == id {
			return false
		}
	}

	for _, waifu := range GuildMap[guildID].Waifus {
		if waifu != nil && waifu.ID != "" && waifu.ID == id {
			return false
		}
	}

	for _, trade := range GuildMap[guildID].WaifuTrades {
		if trade != nil && trade.ID != "" && trade.ID == id {
			return false
		}
	}

	for _, member := range GuildMap[guildID].MemberInfoMap {
		for _, timestamp := range member.Timestamps {
			if timestamp != nil && timestamp.ID != "" && timestamp.ID == id {
				return false
			}
		}
	}

	return true
}
