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

// AnimeSubsMongo represents how AnimeSubs are stored in MongoDB
type AnimeSubsMongo struct {
	ID      string     `bson:"id"`
	IsGuild bool       `bson:"is_guild"`
	Shows   []*ShowSub `bson:"shows"`
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
		var animeSubData AnimeSubsMongo
		if err := cursor.Decode(&animeSubData); err != nil {
			log.Println("Error decoding anime subscriptions from MongoDB:", err)
			continue
		}

		animeSubsMap[animeSubData.ID] = animeSubData.Shows
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error while loading anime subscriptions: %v", err)
	}

	return animeSubsMap, nil
}

// GetAnimeSubs retrieves a specific user's or guild's anime subscriptions from MongoDB
func GetAnimeSubs(id string) ([]*ShowSub, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var animeSubs AnimeSubsMongo
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

	var operations []mongo.WriteModel
	for id, shows := range animeSubs {
		// Determine whether the ID belongs to a Guild or User
		isGuild := false
		if len(shows) > 0 {
			isGuild = shows[0].GetGuild()
		}

		data := AnimeSubsMongo{
			ID:      id,
			IsGuild: isGuild,
			Shows:   shows,
		}

		filter := bson.M{"id": id}
		update := bson.M{"$set": data}
		operations = append(operations, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}

	if len(operations) > 0 {
		_, err := AnimeSubsCollection.BulkWrite(ctx, operations)
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

	// Convert the data
	data := AnimeSubsMongo{
		ID:      id,
		IsGuild: isGuild,
		Shows:   subscriptions,
	}

	// Save to MongoDB
	filter := bson.M{"id": id}
	update := bson.M{"$set": data}
	_, err := AnimeSubsCollection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		log.Printf("Error saving updated anime subscriptions for %s: %v\n", id, err)
		return err
	}

	return nil
}

// ConvertShowSub creates a MongoDB-compatible version of ShowSub
func ConvertShowSub(s *ShowSub) *ShowSub {
	if s == nil {
		return nil
	}

	return &ShowSub{
		Show:     s.Show,
		Notified: s.Notified,
		Guild:    s.Guild,
	}
}
