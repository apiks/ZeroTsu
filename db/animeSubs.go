package db

import (
	"log"
	"strings"

	"github.com/r-anime/ZeroTsu/cache"
	"github.com/r-anime/ZeroTsu/entities"
)

// GetAllAnimeSubs retrieves all anime subscriptions from MongoDB
func GetAllAnimeSubs() map[string][]*entities.ShowSub {
	animeSubs, err := entities.LoadAnimeSubs()
	if err != nil {
		log.Println("Error loading all anime subscriptions:", err)
		return nil
	}

	return animeSubs
}

// GetGuildAnimeSubs retrieves only guild anime subscriptions from MongoDB
func GetGuildAnimeSubs() map[string][]*entities.ShowSub {
	guildSubs, err := entities.LoadGuildAnimeSubs()
	if err != nil {
		log.Println("Error loading guild anime subscriptions:", err)
		return nil
	}

	return guildSubs
}

// GetUserAnimeSubs retrieves only user anime subscriptions from MongoDB
func GetUserAnimeSubs() map[string][]*entities.ShowSub {
	userSubs, err := entities.LoadUserAnimeSubs()
	if err != nil {
		log.Println("Error loading user anime subscriptions:", err)
		return nil
	}

	return userSubs
}

// GetAnimeSubs retrieves a user's or guild's anime subscriptions
func GetAnimeSubs(id string) []*entities.ShowSub {
	subs, err := entities.GetAnimeSubs(id)
	if err != nil {
		log.Println("Error fetching anime subscriptions for", id, ":", err)
		return nil
	}
	return subs
}

// BulkGetAnimeSubs retrieves multiple users' or guilds' anime subscriptions in one query
func BulkGetAnimeSubs(ids []string) map[string][]*entities.ShowSub {
	subs, err := entities.BulkGetAnimeSubs(ids)
	if err != nil {
		log.Println("Error fetching bulk anime subscriptions:", err)
		return nil
	}
	return subs
}

// CountAnimeSubs returns the total count of anime subscriptions
func CountAnimeSubs() int64 {
	count, err := entities.CountAnimeSubs()
	if err != nil {
		log.Println("Error counting anime subscriptions:", err)
		return 0
	}
	return count
}

// CountGuildAnimeSubs returns the count of guild anime subscriptions
func CountGuildAnimeSubs() int64 {
	count, err := entities.CountGuildAnimeSubs()
	if err != nil {
		log.Println("Error counting guild anime subscriptions:", err)
		return 0
	}
	return count
}

// CountUserAnimeSubs returns the count of user anime subscriptions
func CountUserAnimeSubs() int64 {
	count, err := entities.CountUserAnimeSubs()
	if err != nil {
		log.Println("Error counting user anime subscriptions:", err)
		return 0
	}
	return count
}

// SetAnimeSubs updates a user or guild's anime subscriptions
func SetAnimeSubs(id string, subscriptions []*entities.ShowSub, isGuild bool) {
	err := entities.SetAnimeSubs(id, subscriptions, isGuild)
	if err != nil {
		log.Printf("Error updating anime subscriptions for %s: %v\n", id, err)
		return
	}

	// Update cache with the new data instead of invalidating
	cache.AnimeSubs.Update(id, subscriptions)
}

// AddAnimeSub adds a new anime subscription for a user or guild in MongoDB
func AddAnimeSub(id string, showSub *entities.ShowSub, isGuild bool) {
	subs := GetAnimeSubs(id)
	// Avoid duplicate subscriptions
	for _, existing := range subs {
		if strings.EqualFold(existing.GetShow(), showSub.GetShow()) {
			return
		}
	}
	subs = append(subs, showSub)
	SetAnimeSubs(id, subs, isGuild)
}

// RemoveAnimeSub deletes a specific anime subscription for a user or guild in MongoDB
func RemoveAnimeSub(id string, showName string) {
	subs := GetAnimeSubs(id)
	var newSubs []*entities.ShowSub
	for _, sub := range subs {
		if !strings.EqualFold(sub.GetShow(), showName) {
			newSubs = append(newSubs, sub)
		}
	}
	isGuild := len(newSubs) > 0 && newSubs[0].GetGuild()
	SetAnimeSubs(id, newSubs, isGuild)
}
