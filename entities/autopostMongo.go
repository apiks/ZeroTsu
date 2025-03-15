package entities

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// AutopostChannelMongo represents a Discord channel for autoposting in MongoDB
type AutopostChannelMongo struct {
	PostType string `bson:"post_type"`
	Name     string `bson:"name"`
	ID       string `bson:"id"`
	RoleID   string `bson:"role_id"`
}

func ConvertAutopostsMapToSlice(autoposts map[string]Cha) []AutopostChannelMongo {
	result := make([]AutopostChannelMongo, 0, len(autoposts))
	for postType, ch := range autoposts {
		result = append(result, AutopostChannelMongo{
			PostType: postType,
			Name:     ch.Name,
			ID:       ch.ID,
			RoleID:   ch.RoleID,
		})
	}
	return result
}

func ConvertMongoToAutoposts(channels []AutopostChannelMongo) map[string]Cha {
	result := make(map[string]Cha, len(channels))
	for _, ch := range channels {
		result[ch.PostType] = Cha{
			Name:   ch.Name,
			ID:     ch.ID,
			RoleID: ch.RoleID,
		}
	}
	return result
}

// LoadAutopost retrieves a specific autopost entry from MongoDB
func LoadAutopost(guildID string, postType string) (Cha, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result struct {
		Autoposts []AutopostChannelMongo `bson:"autoposts"`
	}

	err := GuildCollection.FindOne(ctx, bson.M{"_id": guildID}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return Cha{}, nil
	}
	if err != nil {
		return Cha{}, fmt.Errorf("failed to load autoposts for guild %s: %v", guildID, err)
	}

	autoposts := ConvertMongoToAutoposts(result.Autoposts)
	if autopost, ok := autoposts[postType]; ok {
		return autopost, nil
	}

	return Cha{}, nil
}

// SaveAutopost updates a single autopost in MongoDB
func SaveAutopost(guildID string, postType string, autopost Cha) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	channelMongo := AutopostChannelMongo{
		PostType: postType,
		Name:     autopost.Name,
		ID:       autopost.ID,
		RoleID:   autopost.RoleID,
	}

	filter := bson.M{"_id": guildID}
	update := bson.M{"$addToSet": bson.M{"autoposts": channelMongo}}

	_, err := GuildCollection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("failed to save autopost for guild %s: %v", guildID, err)
	}

	return nil
}

// DeleteAutopost removes a specific autopost from MongoDB
func DeleteAutopost(guildID string, postType string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": guildID}
	update := bson.M{"$pull": bson.M{"autoposts": bson.M{"post_type": postType}}}

	result, err := GuildCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete autopost %s for guild %s: %v", postType, guildID, err)
	}

	if result.ModifiedCount == 0 {
		return fmt.Errorf("autopost %s not found in guild %s", postType, guildID)
	}

	return nil
}

func EnsureAutopostIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.M{"autoposts.post_type": 1},
			Options: options.Index(),
		},
	}

	_, err := GuildCollection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		fmt.Println("Failed to create index for autoposts:", err)
	}
}
