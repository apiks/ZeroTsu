package entities

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
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

	opts := options.FindOne().SetProjection(bson.M{"autoposts": 1})
	err := GuildCollection.FindOne(ctx, bson.M{"_id": guildID}, opts).Decode(&result)
	if err != nil {
		return Cha{}, fmt.Errorf("failed to load autoposts for guild %s: %v", guildID, err)
	}

	autoposts := ConvertMongoToAutoposts(result.Autoposts)

	if autopost, ok := autoposts[postType]; ok {
		return autopost, nil
	}

	return Cha{}, fmt.Errorf("autopost %s not found", postType)
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

	filter := bson.M{"_id": guildID, "autoposts.post_type": postType}
	update := bson.M{"$set": bson.M{"autoposts.$": channelMongo}}

	result, err := GuildCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update autopost for guild %s: %v", guildID, err)
	}

	if result.ModifiedCount == 0 {
		// If it doesn't exist, push a new one
		filter = bson.M{"_id": guildID}
		update = bson.M{"$push": bson.M{"autoposts": channelMongo}}
		_, err = GuildCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			return fmt.Errorf("failed to insert new autopost for guild %s: %v", guildID, err)
		}
	}

	return nil
}

// DeleteAutopost removes a specific autopost from MongoDB
func DeleteAutopost(guildID string, postType string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": guildID}
	update := bson.M{"$pull": bson.M{"autoposts": bson.M{"post_type": postType}}} // ✅ Remove by `post_type`

	_, err := GuildCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete autopost %s for guild %s: %v", postType, guildID, err)
	}

	return nil
}
