package entities

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FeedCheckMongo represents how a FeedCheck is stored in MongoDB
type FeedCheckMongo struct {
	GuildID string    `bson:"guild_id"`
	GUID    string    `bson:"guid"`
	Feed    FeedMongo `bson:"feed"`
	Date    time.Time `bson:"date"`
}

// LoadFeedChecks retrieves recent FeedChecks for a guild (default limit: 50, no limit if limit <= 0)
func LoadFeedChecks(guildID string, limit int) ([]FeedCheck, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"guild_id": guildID}
	opts := options.Find().SetSort(bson.M{"date": -1})

	// Apply limit only if it's greater than 0
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := FeedCheckCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to load feed checks: %v", err)
	}
	defer cursor.Close(ctx)

	var feedChecksMongo []FeedCheckMongo
	if err := cursor.All(ctx, &feedChecksMongo); err != nil {
		return nil, fmt.Errorf("error decoding feed checks: %v", err)
	}

	return ConvertMongoToFeedChecks(feedChecksMongo), nil
}

// SaveFeedCheck stores a FeedCheck in MongoDB
func SaveFeedCheck(guildID string, feedCheck FeedCheck) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	feedCheckData := FeedCheckMongo{
		GuildID: guildID,
		GUID:    feedCheck.GUID,
		Feed:    ConvertFeed(feedCheck.Feed),
		Date:    feedCheck.Date,
	}

	filter := bson.M{"guild_id": guildID, "guid": feedCheck.GUID}
	update := bson.M{"$set": feedCheckData}

	_, err := FeedCheckCollection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("failed to save feed check: %v", err)
	}

	return nil
}

// SaveMultipleFeedChecks stores multiple FeedChecks in MongoDB
func SaveMultipleFeedChecks(guildID string, feedChecks []FeedCheck) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var operations []mongo.WriteModel
	for _, feedCheck := range feedChecks {
		feedCheckData := FeedCheckMongo{
			GuildID: guildID,
			GUID:    feedCheck.GUID,
			Feed:    ConvertFeed(feedCheck.Feed),
			Date:    feedCheck.Date,
		}

		filter := bson.M{"guild_id": guildID, "guid": feedCheck.GUID}
		update := bson.M{"$set": feedCheckData}
		operations = append(operations, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}

	if len(operations) > 0 {
		const batchSize = 50

		for i := 0; i < len(operations); i += batchSize {
			end := i + batchSize
			if end > len(operations) {
				end = len(operations)
			}

			_, err := FeedCheckCollection.BulkWrite(ctx, operations[i:end])
			if err != nil {
				return fmt.Errorf("failed to save multiple feed checks: %v", err)
			}
		}
	}

	return nil
}

// DeleteFeedCheck removes a single FeedCheck from MongoDB
func DeleteFeedCheck(guildID string, feedCheck FeedCheck) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"guild_id": guildID, "guid": feedCheck.GUID}
	_, err := FeedCheckCollection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete feed check: %v", err)
	}

	return nil
}

// DeleteMultipleFeedChecks removes multiple FeedChecks from MongoDB
func DeleteMultipleFeedChecks(guildID string, feedChecks []FeedCheck) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var guids []string
	for _, feedCheck := range feedChecks {
		guids = append(guids, feedCheck.GUID)
	}

	filter := bson.M{"guild_id": guildID, "guid": bson.M{"$in": guids}}
	_, err := FeedCheckCollection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete multiple feed checks: %v", err)
	}

	return nil
}

// ConvertMongoToFeedChecks converts MongoDB FeedCheck to an Entity FeedCheck
func ConvertMongoToFeedChecks(feedChecks []FeedCheckMongo) []FeedCheck {
	result := make([]FeedCheck, len(feedChecks))
	for i, fc := range feedChecks {
		result[i] = FeedCheck{
			Feed: ConvertMongoToFeed(fc.Feed),
			Date: fc.Date,
			GUID: fc.GUID,
		}
	}
	return result
}

func EnsureFeedCheckIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.D{{"guild_id", 1}, {"guid", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{"date", -1}},
			Options: options.Index(),
		},
		{
			Keys:    bson.D{{"feed.subreddit", 1}},
			Options: options.Index(),
		},
		{
			Keys:    bson.D{{"feed.channel_id", 1}},
			Options: options.Index(),
		},
	}

	_, err := FeedCheckCollection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		fmt.Println("Failed to create index for feed checks:", err)
	}
}
