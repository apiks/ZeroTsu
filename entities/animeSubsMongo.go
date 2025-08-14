package entities

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Use projection to only fetch necessary fields and process in batches
	opts := GetOptimizedFindOptions().SetProjection(bson.M{
		"id":       1,
		"is_guild": 1,
		"shows":    1,
		"_id":      0, // Exclude _id field
	})

	cursor, err := AnimeSubsCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch anime subscriptions: %v", err)
	}
	defer cursor.Close(ctx)

	animeSubsMap := make(map[string][]*ShowSub)

	// Process in batches to reduce memory usage
	batchSize := 50
	batch := make([]AnimeSubs, 0, batchSize)

	for cursor.Next(ctx) {
		var animeSubData AnimeSubs
		if err := cursor.Decode(&animeSubData); err != nil {
			log.Println("Error decoding anime subscriptions from MongoDB:", err)
			continue
		}

		batch = append(batch, animeSubData)

		// Process batch when it reaches the size limit
		if len(batch) >= batchSize {
			for _, data := range batch {
				animeSubsMap[data.ID] = data.Shows
			}
			batch = batch[:0] // Reset slice but keep capacity
		}
	}

	// Process remaining items
	for _, data := range batch {
		animeSubsMap[data.ID] = data.Shows
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error while loading anime subscriptions: %v", err)
	}

	return animeSubsMap, nil
}

// LoadGuildAnimeSubs retrieves only guild anime subscriptions from MongoDB
func LoadGuildAnimeSubs() (map[string][]*ShowSub, error) {
	// Use a short timeout for the initial find, and a longer one for iteration
	findCtx, findCancel := GetContextWithTimeout(30 * time.Second)
	defer findCancel()

	iterCtx, iterCancel := GetContextWithTimeout(2 * time.Minute)
	defer iterCancel()

	// Only fetch guild subscriptions with projection
	filter := bson.M{"is_guild": true}
	opts := GetOptimizedFindOptions().
		SetProjection(bson.M{
			"id":       1,
			"is_guild": 1,
			"shows":    1,
			"_id":      0,
		})

	cursor, err := AnimeSubsCollection.Find(findCtx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch guild anime subscriptions: %v", err)
	}
	defer cursor.Close(iterCtx)

	guildSubsMap := make(map[string][]*ShowSub)

	// Process in batches
	batchSize := 50
	batch := make([]AnimeSubs, 0, batchSize)

	for cursor.Next(iterCtx) {
		var animeSubData AnimeSubs
		if err := cursor.Decode(&animeSubData); err != nil {
			log.Println("Error decoding guild anime subscriptions from MongoDB:", err)
			continue
		}

		batch = append(batch, animeSubData)

		if len(batch) >= batchSize {
			for _, data := range batch {
				guildSubsMap[data.ID] = data.Shows
			}
			batch = batch[:0]
		}
	}

	// Process remaining items
	for _, data := range batch {
		guildSubsMap[data.ID] = data.Shows
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error while loading guild anime subscriptions: %v", err)
	}

	return guildSubsMap, nil
}

// LoadUserAnimeSubs retrieves only user anime subscriptions from MongoDB
func LoadUserAnimeSubs() (map[string][]*ShowSub, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Only fetch user subscriptions with projection
	filter := bson.M{"is_guild": false}
	opts := GetOptimizedFindOptions().SetProjection(bson.M{
		"id":       1,
		"is_guild": 1,
		"shows":    1,
		"_id":      0,
	})

	cursor, err := AnimeSubsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user anime subscriptions: %v", err)
	}
	defer cursor.Close(ctx)

	userSubsMap := make(map[string][]*ShowSub)

	// Process in batches
	batchSize := 50
	batch := make([]AnimeSubs, 0, batchSize)

	for cursor.Next(ctx) {
		var animeSubData AnimeSubs
		if err := cursor.Decode(&animeSubData); err != nil {
			log.Println("Error decoding user anime subscriptions from MongoDB:", err)
			continue
		}

		batch = append(batch, animeSubData)

		if len(batch) >= batchSize {
			for _, data := range batch {
				userSubsMap[data.ID] = data.Shows
			}
			batch = batch[:0]
		}
	}

	// Process remaining items
	for _, data := range batch {
		userSubsMap[data.ID] = data.Shows
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error while loading user anime subscriptions: %v", err)
	}

	return userSubsMap, nil
}

// GetAnimeSubs retrieves a specific user's or guild's anime subscriptions from MongoDB
func GetAnimeSubs(id string) ([]*ShowSub, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var animeSubs AnimeSubs
	err := AnimeSubsCollection.FindOne(ctx, bson.M{"id": id}).Decode(&animeSubs)
	if err == mongo.ErrNoDocuments {
		return []*ShowSub{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error fetching anime subscriptions for %s: %v", id, err)
	}

	return animeSubs.Shows, nil
}

// BulkGetAnimeSubs retrieves multiple users' or guilds' anime subscriptions in one query
func BulkGetAnimeSubs(ids []string) (map[string][]*ShowSub, error) {
	if len(ids) == 0 {
		return make(map[string][]*ShowSub), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	filter := bson.M{"id": bson.M{"$in": ids}}
	opts := GetOptimizedFindOptions().SetProjection(bson.M{
		"id":       1,
		"is_guild": 1,
		"shows":    1,
		"_id":      0,
	})

	cursor, err := AnimeSubsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bulk anime subscriptions: %v", err)
	}
	defer cursor.Close(ctx)

	subsMap := make(map[string][]*ShowSub)

	// Process in batches to reduce memory usage
	batchSize := 25
	batch := make([]AnimeSubs, 0, batchSize)

	for cursor.Next(ctx) {
		var animeSubData AnimeSubs
		if err := cursor.Decode(&animeSubData); err != nil {
			log.Println("Error decoding bulk anime subscriptions from MongoDB:", err)
			continue
		}

		batch = append(batch, animeSubData)

		// Process batch when it reaches the size limit
		if len(batch) >= batchSize {
			for _, data := range batch {
				subsMap[data.ID] = data.Shows
			}
			batch = batch[:0] // Reset slice but keep capacity
		}
	}

	// Process remaining items
	for _, data := range batch {
		subsMap[data.ID] = data.Shows
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error while loading bulk anime subscriptions: %v", err)
	}

	return subsMap, nil
}

// CountAnimeSubs returns the total count of anime subscriptions
func CountAnimeSubs() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := AnimeSubsCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("failed to count anime subscriptions: %v", err)
	}

	return count, nil
}

// CountGuildAnimeSubs returns the count of guild anime subscriptions
func CountGuildAnimeSubs() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := AnimeSubsCollection.CountDocuments(ctx, bson.M{"is_guild": true})
	if err != nil {
		return 0, fmt.Errorf("failed to count guild anime subscriptions: %v", err)
	}

	return count, nil
}

// CountUserAnimeSubs returns the count of user anime subscriptions
func CountUserAnimeSubs() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := AnimeSubsCollection.CountDocuments(ctx, bson.M{"is_guild": false})
	if err != nil {
		return 0, fmt.Errorf("failed to count user anime subscriptions: %v", err)
	}

	return count, nil
}

// SetAnimeSubs updates a specific user's or guild's anime subscriptions in MongoDB
func SetAnimeSubs(id string, subscriptions []*ShowSub, isGuild bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
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
