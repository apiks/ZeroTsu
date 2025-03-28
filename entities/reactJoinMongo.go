package entities

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ReactJoinMongoWrap is a wrapper for ReactJoinMongo to include the channel ID
type ReactJoinMongoWrap struct {
	ChannelID string         `bson:"channel_id"`
	ReactJoin ReactJoinMongo `bson:"react_join"`
}

// ReactJoinMongo represents how ReactJoin is stored in MongoDB
type ReactJoinMongo struct {
	RoleEmojiMap []map[string][]string `bson:"role_emoji"`
}

// LoadReactJoinMap retrieves the react join map from MongoDB
func LoadReactJoinMap(guildID string) (map[string]*ReactJoin, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result struct {
		ReactJoinMap []ReactJoinMongoWrap `bson:"react_join_map"`
	}

	err := GuildCollection.FindOne(ctx, bson.M{"_id": guildID}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return map[string]*ReactJoin{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load react join map for guild %s: %v", guildID, err)
	}

	return ConvertMongoToReactJoinMap(result.ReactJoinMap), nil
}

// SaveReactJoinMap saves the entire react join map in MongoDB
func SaveReactJoinMap(guildID string, reactJoin map[string]*ReactJoin) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reactJoinList := ConvertReactJoinMapToSlice(reactJoin)

	filter := bson.M{"_id": guildID}
	update := bson.M{"$set": bson.M{"react_join_map": reactJoinList}}

	_, err := GuildCollection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("failed to save react join map for guild %s: %v", guildID, err)
	}

	return nil
}

// SaveReactJoinEntry inserts or updates a single react join entry in MongoDB
func SaveReactJoinEntry(guildID, messageID string, reactJoin *ReactJoin) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	entry := ReactJoinMongoWrap{
		ChannelID: messageID,
		ReactJoin: ReactJoinMongo{
			RoleEmojiMap: reactJoin.RoleEmojiMap,
		},
	}

	filter := bson.M{"_id": guildID}
	pull := bson.M{"$pull": bson.M{"react_join_map": bson.M{"channel_id": messageID}}}
	_, _ = GuildCollection.UpdateOne(ctx, filter, pull)

	update := bson.M{"$push": bson.M{"react_join_map": entry}}
	_, err := GuildCollection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("failed to save react join entry for guild %s: %v", guildID, err)
	}

	return nil
}

// DeleteReactJoinEntry removes a single react join entry from MongoDB
func DeleteReactJoinEntry(guildID, messageID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": guildID}
	update := bson.M{"$pull": bson.M{"react_join_map": bson.M{"channel_id": messageID}}}

	_, err := GuildCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete react join entry for guild %s: %v", guildID, err)
	}

	return nil
}

// DeleteReactJoinMap removes the entire react join map from MongoDB
func DeleteReactJoinMap(guildID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": guildID}
	update := bson.M{"$unset": bson.M{"react_join_map": ""}}

	_, err := GuildCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete react join map for guild %s: %v", guildID, err)
	}

	return nil
}

// ConvertReactJoinMapToSlice converts map[string]*ReactJoin → []ReactJoinMongoWrap
func ConvertReactJoinMapToSlice(reactJoinMap map[string]*ReactJoin) []ReactJoinMongoWrap {
	result := make([]ReactJoinMongoWrap, len(reactJoinMap))
	i := 0
	for channelID, reactJoin := range reactJoinMap {
		result[i] = ReactJoinMongoWrap{
			ChannelID: channelID,
			ReactJoin: ReactJoinMongo{
				RoleEmojiMap: reactJoin.RoleEmojiMap,
			},
		}
		i++
	}
	return result
}

// ConvertMongoToReactJoinMap converts []ReactJoinMongoWrap → map[string]*ReactJoin
func ConvertMongoToReactJoinMap(reactJoinList []ReactJoinMongoWrap) map[string]*ReactJoin {
	result := make(map[string]*ReactJoin, len(reactJoinList))
	for _, item := range reactJoinList {
		result[item.ChannelID] = &ReactJoin{
			RoleEmojiMap: item.ReactJoin.RoleEmojiMap,
		}
	}
	return result
}

func EnsureReactJoinIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.M{"react_join_map.channel_id": 1},
			Options: options.Index(),
		},
	}

	_, err := GuildCollection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		fmt.Println("Failed to create index for react join map:", err)
	}
}
