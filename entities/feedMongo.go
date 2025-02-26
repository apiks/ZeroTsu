package entities

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
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

	opts := options.FindOne().SetProjection(bson.M{"feeds": 1})
	err := GuildCollection.FindOne(ctx, bson.M{"_id": guildID}, opts).Decode(&result)
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

	// Check if feed already exists
	filter := bson.M{"_id": guildID, "feeds.subreddit": feed.Subreddit, "feeds.channel_id": feed.ChannelID, "feeds.post_type": feed.PostType}
	update := bson.M{"$set": bson.M{"feeds.$": feedMongo}}

	result, err := GuildCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update feed for guild %s: %v", guildID, err)
	}

	// If no existing feed was updated, insert a new one
	if result.ModifiedCount == 0 {
		filter = bson.M{"_id": guildID}
		update = bson.M{"$push": bson.M{"feeds": feedMongo}}
		_, err = GuildCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			return fmt.Errorf("failed to insert new feed for guild %s: %v", guildID, err)
		}
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

	_, err := GuildCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete feed for guild %s: %v", guildID, err)
	}

	return nil
}
