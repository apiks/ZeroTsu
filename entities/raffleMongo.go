package entities

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type RaffleMongo struct {
	Name           string   `bson:"name"`
	ParticipantIDs []string `bson:"participant_ids"`
	ReactMessageID string   `bson:"react_message_id"`
}

// LoadRaffles retrieves all raffles for a guild
func LoadRaffles(guildID string) ([]*Raffle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result struct {
		Raffles []RaffleMongo `bson:"raffles"`
	}

	err := GuildCollection.FindOne(ctx, bson.M{"_id": guildID}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return []*Raffle{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load raffles for guild %s: %v", guildID, err)
	}

	return ConvertMongoToRaffles(result.Raffles), nil
}

// SaveRaffle inserts or updates a raffle in MongoDB
func SaveRaffle(guildID string, raffle *Raffle) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	raffleMongo := ConvertRaffle(raffle)

	filter := bson.M{"_id": guildID, "raffles.name": raffle.Name}
	update := bson.M{
		"$set":         bson.M{"raffles.$": raffleMongo},
		"$setOnInsert": bson.M{"_id": guildID},
	}

	result, err := GuildCollection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("failed to save raffle for guild %s: %v", guildID, err)
	}

	// If no existing raffle was updated, insert a new one
	if result.ModifiedCount == 0 {
		filter = bson.M{"_id": guildID}
		update = bson.M{"$push": bson.M{"raffles": raffleMongo}}
		_, err = GuildCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			return fmt.Errorf("failed to insert new raffle for guild %s: %v", guildID, err)
		}
	}

	return nil
}

// DeleteRaffle removes a specific raffle from MongoDB
func DeleteRaffle(guildID string, raffle *Raffle) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": guildID}
	update := bson.M{"$pull": bson.M{"raffles": bson.M{"name": raffle.Name}}}

	_, err := GuildCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete raffle for guild %s: %v", guildID, err)
	}

	return nil
}

// UpdateRaffleParticipant adds or removes a participant in a raffle
func UpdateRaffleParticipant(guildID, userID string, raffle *Raffle, remove bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": guildID, "raffles.name": raffle.Name}
	update := bson.M{}

	if remove {
		update = bson.M{"$pull": bson.M{"raffles.$.participant_ids": userID}}
	} else {
		update = bson.M{"$addToSet": bson.M{"raffles.$.participant_ids": userID}}
	}

	_, err := GuildCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		action := "add"
		if remove {
			action = "remove"
		}
		return fmt.Errorf("failed to %s participant in raffle for guild %s: %v", action, guildID, err)
	}

	return nil
}

func ConvertRaffle(raffle *Raffle) RaffleMongo {
	return RaffleMongo{
		Name:           raffle.Name,
		ParticipantIDs: raffle.ParticipantIDs,
		ReactMessageID: raffle.ReactMessageID,
	}
}

func ConvertRaffleSlice(raffles []*Raffle) []RaffleMongo {
	result := make([]RaffleMongo, len(raffles))
	for i := range raffles {
		result[i] = ConvertRaffle(raffles[i])
	}
	return result
}

func ConvertMongoToRaffles(raffles []RaffleMongo) []*Raffle {
	result := make([]*Raffle, len(raffles))
	for i := range raffles {
		result[i] = &Raffle{
			Name:           raffles[i].Name,
			ParticipantIDs: raffles[i].ParticipantIDs,
			ReactMessageID: raffles[i].ReactMessageID,
		}
	}
	return result
}

func EnsureRaffleIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.M{"raffles.name": 1},
			Options: options.Index(),
		},
	}

	_, err := GuildCollection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		fmt.Println("Failed to create index for raffles:", err)
	}
}
