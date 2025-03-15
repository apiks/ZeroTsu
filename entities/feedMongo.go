package entities

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type FeedMongo struct {
	GuildId   string `bson:"guild_id"`
	Subreddit string `bson:"subreddit"`
	Title     string `bson:"title"`
	Author    string `bson:"author"`
	Pin       bool   `bson:"pin"`
	PostType  string `bson:"post_type"`
	ChannelID string `bson:"channel_id"`
}

// ConvertFeed converts `entities.Feed` → `FeedMongo`
func ConvertFeed(feed Feed) FeedMongo {
	return FeedMongo{
		Subreddit: feed.Subreddit,
		Title:     feed.Title,
		Author:    feed.Author,
		Pin:       feed.Pin,
		PostType:  feed.PostType,
		ChannelID: feed.ChannelID,
	}
}

// ConvertFeedSlice converts `[]entities.Feed` → `[]FeedMongo`
func ConvertFeedSlice(feeds []Feed) []FeedMongo {
	result := make([]FeedMongo, len(feeds))
	for i, feed := range feeds {
		result[i] = ConvertFeed(feed)
	}
	return result
}

// ConvertMongoToFeed converts `FeedMongo` → `entities.Feed`
func ConvertMongoToFeed(feed FeedMongo) Feed {
	return Feed{
		Subreddit: feed.Subreddit,
		Title:     feed.Title,
		Author:    feed.Author,
		Pin:       feed.Pin,
		PostType:  feed.PostType,
		ChannelID: feed.ChannelID,
	}
}

// ConvertMongoToFeeds converts `[]FeedMongo` → `[]entities.Feed`
func ConvertMongoToFeeds(feeds []FeedMongo) []Feed {
	result := make([]Feed, len(feeds))
	for i, feed := range feeds {
		result[i] = ConvertMongoToFeed(feed)
	}
	return result
}

// LoadAllFeeds retrieves all feeds for a guild
func LoadAllFeeds(guildID string) ([]Feed, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result struct {
		Feeds []FeedMongo `bson:"feeds"`
	}

	err := GuildCollection.FindOne(ctx, bson.M{"_id": guildID}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return []Feed{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load feeds for guild %s: %v", guildID, err)
	}

	return ConvertMongoToFeeds(result.Feeds), nil
}

// SaveFeed inserts or updates a feed in MongoDB
func SaveFeed(guildID string, feed Feed) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	feedMongo := ConvertFeed(feed)

	filter := bson.M{"_id": guildID}
	update := bson.M{"$addToSet": bson.M{"feeds": feedMongo}} // Prevents duplicate feeds

	_, err := GuildCollection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("failed to save feed for guild %s: %v", guildID, err)
	}

	return nil
}

// DeleteFeed removes a specific feed from MongoDB
func DeleteFeed(guildID string, feed Feed) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": guildID}
	update := bson.M{"$pull": bson.M{"feeds": bson.M{
		"subreddit":  feed.Subreddit,
		"channel_id": feed.ChannelID,
		"post_type":  feed.PostType,
	}}}

	result, err := GuildCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete feed for guild %s: %v", guildID, err)
	}

	if result.ModifiedCount == 0 {
		return fmt.Errorf("feed not found in guild %s", guildID)
	}

	return nil
}

func EnsureFeedsIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.M{"feeds.subreddit": 1},
			Options: options.Index(),
		},
	}

	_, err := GuildCollection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		fmt.Println("Failed to create index for feeds:", err)
	}
}
