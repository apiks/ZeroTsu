package entities

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
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

	var guildIds []string
	results, err := GuildCollection.Distinct(ctx, "_id", bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch guild IDs: %v", err)
	}

	for _, id := range results {
		guildIds = append(guildIds, id.(string))
	}

	return guildIds, nil
}

// LoadGuildBotLogs retrieves only the bot log channel IDs for all guilds
func LoadGuildBotLogs() (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := GuildCollection.Find(ctx, bson.M{}, options.Find())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bot logs: %v", err)
	}
	defer cursor.Close(ctx)

	guildLogs := make(map[string]string)

	for cursor.Next(ctx) {
		var result struct {
			ID            string `bson:"_id"`
			GuildSettings struct {
				BotLogID string `bson:"bot_log_id"`
			} `bson:"guild_settings"`
		}

		if err := cursor.Decode(&result); err != nil {
			log.Println("Error decoding bot log entry from MongoDB:", err)
			continue
		}

		guildLogs[result.ID] = result.GuildSettings.BotLogID
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
	err := GuildCollection.FindOne(ctx, filter).Err()

	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check guild existence: %v", err)
	}

	return true, nil
}

// InitGuildIfNotExists ensures a guild exists in MongoDB and initializes it if missing
func InitGuildIfNotExists(guildID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": guildID}

	// Use `$setOnInsert` to only insert if the document does not exist
	update := bson.M{"$setOnInsert": GuildInfoMongo{
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
	}}

	opts := options.Update().SetUpsert(true)
	_, err := GuildCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to initialize new guild %s: %v", guildID, err)
	}

	return nil
}

func EnsureGuildsIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.M{"guild_settings.bot_log_id": 1},
			Options: options.Index(),
		},
	}

	_, err := GuildCollection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.Fatal("Failed to create indexes for guilds:", err)
	}
}
