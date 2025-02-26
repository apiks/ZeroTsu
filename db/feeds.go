package db

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/entities"
	"log"
	"strings"
)

// GetGuildFeeds retrieves all feeds from MongoDB
func GetGuildFeeds(guildID string) []entities.Feed {
	err := entities.InitGuildIfNotExists(guildID)
	if err != nil {
		log.Println(err)
		return nil
	}

	feeds, err := entities.LoadAllFeeds(guildID)
	if err != nil {
		return nil
	}

	return feeds
}

// SetGuildFeed saves or deletes a guild feed in MongoDB
func SetGuildFeed(guildID string, feed entities.Feed, delete ...bool) error {
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
	existingFeeds := GetGuildFeeds(guildID)
	limit := 50
	if guildSettings.Premium {
		limit = 400
	}
	if len(delete) == 0 && len(existingFeeds) >= limit {
		if guildSettings.Premium {
			return fmt.Errorf("Error: You have reached the reddit feed autopost limit (400) for this server.")
		}
		return fmt.Errorf("Error: You have reached the reddit feed autopost limit (50) for this server. Please remove some or upgrade to premium at <https://patreon.com/animeschedule>")
	}

	feed = feed.SetSubreddit(strings.ToLower(feed.GetSubreddit()))

	if len(delete) == 0 {
		// Check if feed already exists
		for _, existingFeed := range existingFeeds {
			if existingFeed.GetSubreddit() == feed.GetSubreddit() &&
				existingFeed.GetChannelID() == feed.GetChannelID() &&
				existingFeed.GetPostType() == feed.GetPostType() {
				return fmt.Errorf("Error: That feed already exists.")
			}
		}

		// Save the new feed to MongoDB
		err := entities.SaveFeed(guildID, feed)
		if err != nil {
			log.Printf("Error saving feed for guild %s: %v\n", guildID, err)
			return err
		}
	} else {
		// Delete the feed from MongoDB
		err := entities.DeleteFeed(guildID, feed)
		if err != nil {
			log.Printf("Error deleting feed for guild %s: %v\n", guildID, err)
			return err
		}
	}

	return nil
}
