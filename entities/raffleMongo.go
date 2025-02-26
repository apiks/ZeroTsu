package entities

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
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

	opts := options.FindOne().SetProjection(bson.M{"raffles": 1})
	err := GuildCollection.FindOne(ctx, bson.M{"_id": guildID}, opts).Decode(&result)
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

	// Check if raffle already exists
	filter := bson.M{"_id": guildID, "raffles.name": raffle.Name}
	update := bson.M{"$set": bson.M{"raffles.$": raffleMongo}}

	result, err := GuildCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update raffle for guild %s: %v", guildID, err)
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
	for i, raffle := range raffles {
		result[i] = ConvertRaffle(raffle)
	}
	return result
}

func ConvertMongoToRaffles(raffles []RaffleMongo) []*Raffle {
	result := make([]*Raffle, len(raffles))
	for i, r := range raffles {
		result[i] = &Raffle{
			Name:           r.Name,
			ParticipantIDs: r.ParticipantIDs,
			ReactMessageID: r.ReactMessageID,
		}
	}
	return result
}
