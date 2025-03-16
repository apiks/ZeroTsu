package db

import (
	"github.com/r-anime/ZeroTsu/entities"
	"log"
	"strings"
)

// GetAnimeSubs retrieves a user's or guild's anime subscriptions from MongoDB
func GetAnimeSubs(id string) []*entities.ShowSub {
	subs, err := entities.GetAnimeSubs(id)
	if err != nil {
		log.Println("Error fetching anime subscriptions for", id, ":", err)
		return nil
	}
	return subs
}

// GetAllAnimeSubs retrieves all anime subscriptions from MongoDB
func GetAllAnimeSubs() map[string][]*entities.ShowSub {
	animeSubs, err := entities.LoadAnimeSubs()
	if err != nil {
		log.Println("Error loading all anime subscriptions:", err)
		return nil
	}

	return animeSubs
}

// AddAnimeSub saves a new anime subscription for a user or guild in MongoDB
func AddAnimeSub(id string, showSub *entities.ShowSub, isGuild bool) {
	subs := GetAnimeSubs(id)

	// Avoid duplicate subscriptions
	for _, existing := range subs {
		if strings.ToLower(existing.GetShow()) == strings.ToLower(showSub.GetShow()) {
			log.Printf("User/Guild %s is already subscribed to %s", id, showSub.GetShow())
			return
		}
	}

	// Append new subscription
	subs = append(subs, showSub)

	// Save updated subscriptions to MongoDB
	err := entities.SetAnimeSubs(id, subs, isGuild)
	if err != nil {
		log.Printf("Error saving anime subscription for %s: %v\n", id, err)
	}
}

// SetAnimeSubs updates a user or guild's anime subscriptions
func SetAnimeSubs(id string, subscriptions []*entities.ShowSub, isGuild bool) {
	err := entities.SetAnimeSubs(id, subscriptions, isGuild)
	if err != nil {
		log.Printf("Error updating anime subscriptions for %s: %v\n", id, err)
	}
}

// RemoveAnimeSub deletes a specific anime subscription for a user or guild in MongoDB
func RemoveAnimeSub(id string, showName string) {
	subs := GetAnimeSubs(id)

	// Filter out the target subscription
	var newSubs []*entities.ShowSub
	for _, showSub := range subs {
		if strings.ToLower(showSub.GetShow()) != strings.ToLower(showName) {
			newSubs = append(newSubs, showSub)
		}
	}

	// Save updated subscriptions or delete the entry if empty
	err := entities.SetAnimeSubs(id, newSubs, len(newSubs) > 0 && newSubs[0].GetGuild())
	if err != nil {
		log.Printf("Error updating anime subscriptions after deletion for %s: %v\n", id, err)
	}
}
