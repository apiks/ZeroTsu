package entities

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo"
	"io/ioutil"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MigrateGuilds moves all guilds and their related data (FeedChecks) from JSON storage to MongoDB.
func MigrateGuilds() {
	// Read all guild folders from the local JSON storage path
	folders, err := ioutil.ReadDir("database/guilds")
	if err != nil {
		log.Panicln("Error reading guild directory:", err)
		return
	}

	for _, folder := range folders {
		if !folder.IsDir() {
			continue
		}

		guildID := folder.Name()
		guild := GuildInfo{ID: guildID}

		// Load Guild Settings
		err := guild.Load("guildSettings.json", guildID)
		if err != nil {
			log.Printf("Warning: Failed to load guildSettings.json for %s: %v\n", guildID, err)
		}

		// Load Feeds
		err = guild.Load("rssThreads.json", guildID)
		if err != nil {
			log.Printf("Warning: Failed to load rssThreads.json (feeds) for %s: %v\n", guildID, err)
		}

		// Load FeedChecks
		cutoffDate := time.Now().AddDate(0, 0, -30)
		validFeedChecks := []FeedCheck{}
		for _, feedCheck := range guild.GetFeedChecks() {
			if feedCheck.GetDate().After(cutoffDate) {
				validFeedChecks = append(validFeedChecks, feedCheck)
			}
		}
		guild.SetFeedChecks(validFeedChecks)

		// Load Raffles
		err = guild.Load("raffles.json", guildID)
		if err != nil {
			log.Printf("Warning: Failed to load raffles.json for %s: %v\n", guildID, err)
		}

		// Load Autoposts
		err = guild.Load("autoposts.json", guildID)
		if err != nil {
			log.Printf("Warning: Failed to load autoposts.json for %s: %v\n", guildID, err)
		}

		// Load ReactJoin
		err = guild.Load("reactJoin.json", guildID)
		if err != nil {
			log.Printf("Warning: Failed to load reactJoin.json for %s: %v\n", guildID, err)
		}

		// Convert to MongoDB format
		guildData := GuildInfoMongo{
			ID:            guild.ID,
			GuildSettings: ConvertGuildSettings(guild.GetGuildSettings()),
			Feeds:         ConvertFeedSlice(guild.GetFeeds()),
			Raffles:       ConvertRaffleSlice(guild.GetRaffles()),
			Autoposts:     ConvertAutopostsMapToSlice(guild.GetAutoposts()),
			ReactJoinMap:  ConvertReactJoinMapToSlice(guild.GetReactJoinMap()),
		}

		// Save Guild Data to MongoDB
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		filter := bson.M{"_id": guild.ID}
		update := bson.M{"$set": guildData}

		_, err = GuildCollection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
		if err != nil {
			log.Printf("Error: Failed to save guild %s to MongoDB: %v\n", guildID, err)
			continue
		}

		// Migrate FeedChecks separately
		for _, feedCheck := range validFeedChecks {
			err := SaveFeedCheck(guildID, feedCheck)
			if err != nil {
				log.Printf("Error: Failed to migrate feed check %s for guild %s: %v\n", feedCheck.GUID, guildID, err)
			}
		}

		log.Printf("Successfully migrated guild %s to MongoDB\n", guildID)
	}
}

// MigrateReminders moves reminders from JSON storage to MongoDB
func MigrateReminders() {
	// Read remindMes.json
	data, err := ioutil.ReadFile("database/shared/remindMes.json")
	if err != nil {
		log.Panicln("Error reading remindMes.json:", err)
		return
	}

	// Unmarshal JSON into a map
	var remindMeMap map[string]*RemindMeSlice
	err = json.Unmarshal(data, &remindMeMap)
	if err != nil {
		log.Panicln("Error unmarshaling remindMes.json:", err)
		return
	}

	// Prepare MongoDB context
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var operations []mongo.WriteModel

	// Iterate through the reminders map
	for id, remindMeSlice := range remindMeMap {
		// Convert to MongoDB format
		data := ConvertRemindMeSlice(id, remindMeSlice)

		// Define update operation
		filter := bson.M{"id": id}
		update := bson.M{"$set": data}

		// Add to bulk operations
		operations = append(operations, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}

	// Execute Bulk Write if there are operations
	if len(operations) > 0 {
		_, err := RemindersCollection.BulkWrite(ctx, operations)
		if err != nil {
			log.Printf("Error migrating remindMes to MongoDB: %v\n", err)
			return
		}
	}

	log.Println("Successfully migrated remindMes to MongoDB!")
}

// MigrateAnimeSubs moves anime subscription data from JSON storage to MongoDB
func MigrateAnimeSubs() {
	// Read animeSubs.json
	data, err := ioutil.ReadFile("database/shared/animeSubs.json")
	if err != nil {
		log.Panicln("Error reading animeSubs.json:", err)
		return
	}

	// Unmarshal JSON into a map
	var animeSubsMap map[string][]*ShowSub
	err = json.Unmarshal(data, &animeSubsMap)
	if err != nil {
		log.Panicln("Error unmarshaling animeSubs.json:", err)
		return
	}

	// Prepare MongoDB context
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var operations []mongo.WriteModel

	// Iterate through the anime subscriptions map
	for id, shows := range animeSubsMap {
		// Remove mutex from ShowSub before inserting
		cleanedShows := make([]*ShowSub, 0, len(shows))
		for _, show := range shows {
			if show != nil {
				cleanedShows = append(cleanedShows, ConvertShowSub(show))
			}
		}

		// Determine if it's a guild subscription
		isGuild := false
		if len(cleanedShows) > 0 {
			isGuild = cleanedShows[0].GetGuild()
		}

		// Convert to MongoDB struct
		data := AnimeSubsMongo{
			ID:      id,
			IsGuild: isGuild,
			Shows:   cleanedShows,
		}

		// Define update operation
		filter := bson.M{"id": id}
		update := bson.M{"$set": data}

		// Add to bulk operations
		operations = append(operations, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}

	// Execute Bulk Write if there are operations
	if len(operations) > 0 {
		_, err := AnimeSubsCollection.BulkWrite(ctx, operations)
		if err != nil {
			log.Printf("Error migrating animeSubs to MongoDB: %v\n", err)
			return
		}
	}

	log.Println("Successfully migrated animeSubs to MongoDB!")
}
