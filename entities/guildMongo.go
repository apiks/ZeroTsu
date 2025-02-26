package entities

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GuildInfoMongo struct {
	ID            string                 `bson:"_id"`
	GuildSettings GuildSettingsMongo     `bson:"guild_settings"`
	Feeds         []FeedMongo            `bson:"feeds"`
	Raffles       []RaffleMongo          `bson:"raffles"`
	Autoposts     []AutopostChannelMongo `bson:"autoposts"`
	ReactJoinMap  []ReactJoinMongoWrap   `bson:"react_join_map"`
}

type ChannelMongo struct {
	Name   string `bson:"name"`
	ID     string `bson:"id"`
	RoleID string `bson:"role_id"`
}

type VoiceChannelMongo struct {
	Name  string      `bson:"name"`
	ID    string      `bson:"id"`
	Roles []RoleMongo `bson:"roles"`
}

type RoleMongo struct {
	Name     string `bson:"name"`
	ID       string `bson:"id"`
	Position int    `bson:"position"`
}

// LoadAllGuildIDs retrieves only the guild IDs from MongoDB (without loading full guild data)
func LoadAllGuildIDs() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Find().SetProjection(bson.M{"_id": 1})
	cursor, err := GuildCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch guild IDs: %v", err)
	}
	defer cursor.Close(ctx)

	var guildIds []string
	for cursor.Next(ctx) {
		var result struct {
			ID string `bson:"_id"`
		}
		if err := cursor.Decode(&result); err != nil {
			log.Println("Error decoding guild ID from MongoDB:", err)
			continue
		}
		guildIds = append(guildIds, result.ID)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error while loading guild IDs: %v", err)
	}

	return guildIds, nil
}

// LoadGuildBotLogs retrieves only the bot log channel IDs for all guilds
func LoadGuildBotLogs() (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Fetch only `_id` (guildID) and `guild_settings.bot_log_id` (bot log channel)
	opts := options.Find().SetProjection(bson.M{"_id": 1, "guild_settings.bot_log_id": 1})

	cursor, err := GuildCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bot logs: %v", err)
	}
	defer cursor.Close(ctx)

	guildLogs := make(map[string]string)

	for cursor.Next(ctx) {
		var result struct {
			ID            string `bson:"_id"`
			GuildSettings struct {
				BotLog ChannelMongo `bson:"bot_log_id"`
			} `bson:"guild_settings"`
		}

		if err := cursor.Decode(&result); err != nil {
			log.Println("Error decoding bot log entry from MongoDB:", err)
			continue
		}

		// Store guild ID -> bot log channel ID
		guildLogs[result.ID] = result.GuildSettings.BotLog.ID
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error while loading bot logs: %v", err)
	}

	return guildLogs, nil
}

// DoesGuildExist checks if a guild exists in MongoDB
func DoesGuildExist(guildID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": guildID}
	count, err := GuildCollection.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check guild existence: %v", err)
	}

	return count > 0, nil
}

// InitGuildIfNotExists ensures a guild exists in MongoDB and initializes it if missing
func InitGuildIfNotExists(guildID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if the guild already exists
	filter := bson.M{"_id": guildID}
	count, err := GuildCollection.CountDocuments(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to check existing guild: %v", err)
	}

	// If the guild exists, no need to initialize
	if count > 0 {
		return nil
	}

	// Create a new MongoDB-compatible guild entry
	newGuild := GuildInfoMongo{
		ID: guildID,
		GuildSettings: GuildSettingsMongo{
			Prefix:       ".",
			ReactsModule: true,
			PingMessage:  "Hmmm~ So this is what you do all day long?",
		},
		Feeds:        []FeedMongo{},
		Raffles:      []RaffleMongo{},
		Autoposts:    []AutopostChannelMongo{},
		ReactJoinMap: []ReactJoinMongoWrap{},
	}

	// Insert the new guild into MongoDB
	_, err = GuildCollection.InsertOne(ctx, newGuild)
	if err != nil {
		return fmt.Errorf("failed to initialize new guild %s: %v", guildID, err)
	}

	return nil
}
