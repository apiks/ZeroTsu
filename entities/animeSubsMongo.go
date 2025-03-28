package entities

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type AnimeSubs struct {
	ID      string     `bson:"id"`
	IsGuild bool       `bson:"is_guild"`
	Shows   []*ShowSub `bson:"shows"`
}

// AnimeSubsMongo represents how AnimeSubs are stored in MongoDB
type AnimeSubsMongo struct {
	ID      string          `bson:"id"`
	IsGuild bool            `bson:"is_guild"`
	Shows   []*ShowSubMongo `bson:"shows"`
}

// LoadAnimeSubs retrieves all anime subscriptions from MongoDB
func LoadAnimeSubs() (map[string][]*ShowSub, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cursor, err := AnimeSubsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch anime subscriptions: %v", err)
	}
	defer cursor.Close(ctx)

	animeSubsMap := make(map[string][]*ShowSub)
	for cursor.Next(ctx) {
		var animeSubData AnimeSubs
		if err := cursor.Decode(&animeSubData); err != nil {
			log.Println("Error decoding anime subscriptions from MongoDB:", err)
			continue
		}
		animeSubsMap[animeSubData.ID] = animeSubData.Shows
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error while loading anime subscriptions: %v", err)
	}

	if animeSubsMap == nil {
		return make(map[string][]*ShowSub), nil
	}

	return animeSubsMap, nil
}

// GetAnimeSubs retrieves a specific user's or guild's anime subscriptions from MongoDB
func GetAnimeSubs(id string) ([]*ShowSub, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var animeSubs AnimeSubs
	err := AnimeSubsCollection.FindOne(ctx, bson.M{"id": id}).Decode(&animeSubs)
	if err == mongo.ErrNoDocuments {
		return []*ShowSub{}, nil
	}
	if err != nil {
		log.Printf("Error fetching anime subscriptions for %s: %v\n", id, err)
		return nil, err
	}

	return animeSubs.Shows, nil
}

// SaveAnimeSubs stores all anime subscriptions in MongoDB
func SaveAnimeSubs(animeSubs map[string][]*ShowSub) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	operations := make([]mongo.WriteModel, 0, len(animeSubs))
	for id, shows := range animeSubs {
		sanitizedShows := make([]*ShowSubMongo, len(shows))
		for i, s := range shows {
			sanitizedShows[i] = ConvertShowSub(s)
		}

		filter := bson.M{"id": id}
		update := bson.M{
			"$set": bson.M{
				"is_guild": len(sanitizedShows) > 0 && sanitizedShows[0].Guild,
				"shows":    sanitizedShows,
			},
		}

		model := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)

		operations = append(operations, model)
	}

	if len(operations) > 0 {
		opts := options.BulkWrite().SetOrdered(false)
		_, err := AnimeSubsCollection.BulkWrite(ctx, operations, opts)
		if err != nil {
			return fmt.Errorf("failed to save anime subscriptions: %v", err)
		}
	}

	return nil
}

// SetAnimeSubs updates a specific user's or guild's anime subscriptions in MongoDB
func SetAnimeSubs(id string, subscriptions []*ShowSub, isGuild bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// If no subscriptions left, delete from the database
	if len(subscriptions) == 0 {
		_, err := AnimeSubsCollection.DeleteOne(ctx, bson.M{"id": id})
		if err != nil {
			log.Printf("Error deleting anime subscriptions for %s: %v\n", id, err)
			return err
		}
		return nil
	}

	// Sanitize subscriptions
	sanitizedShows := make([]*ShowSubMongo, len(subscriptions))
	for i, s := range subscriptions {
		sanitizedShows[i] = ConvertShowSub(s)
	}

	// Save to MongoDB
	filter := bson.M{"id": id}
	update := bson.M{
		"$set": bson.M{
			"is_guild": isGuild,
			"shows":    sanitizedShows,
		},
	}

	_, err := AnimeSubsCollection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		log.Printf("Error saving updated anime subscriptions for %s: %v\n", id, err)
		return err
	}

	return nil
}

// ConvertShowSub converts a ShowSub to a MongoDB compatible ShowSubMongo
func ConvertShowSub(s *ShowSub) *ShowSubMongo {
	if s == nil {
		return nil
	}

	return &ShowSubMongo{
		Show:     s.Show,
		Notified: s.Notified,
		Guild:    s.Guild,
	}
}

func EnsureAnimeSubsIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.M{"id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{"is_guild": 1},
		},
	}

	_, err := AnimeSubsCollection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.Fatal("Failed to create indexes for anime_subs:", err)
	}
}
