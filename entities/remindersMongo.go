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

// RemindMeMongo represents how a RemindMe is stored in MongoDB
type RemindMeMongo struct {
	Message        string    `bson:"message"`
	Date           time.Time `bson:"date"`
	CommandChannel string    `bson:"command_channel"`
	RemindID       int       `bson:"remind_id"`
}

// RemindMeSliceMongo represents how a RemindMeSlice is stored in MongoDB
type RemindMeSliceMongo struct {
	ID        string          `bson:"id"`
	IsGuild   bool            `bson:"is_guild"`
	Reminders []RemindMeMongo `bson:"reminders"`
	Premium   bool            `bson:"premium"`
}

// GetReminders retrieves a specific user's or guild's reminders from MongoDB
func GetReminders(id string) (*RemindMeSlice, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var remindMeSlice RemindMeSliceMongo
	err := RemindersCollection.FindOne(ctx, bson.M{"id": id}).Decode(&remindMeSlice)
	if err == mongo.ErrNoDocuments {
		return &RemindMeSlice{RemindMeSlice: []*RemindMe{}, Guild: false, Premium: false}, nil // Return empty slice if no reminders exist
	}
	if err != nil {
		log.Printf("Error fetching reminders for %s: %v\n", id, err)
		return nil, err
	}

	return ConvertMongoToRemindMeSlice(remindMeSlice), nil
}

// SaveReminders updates or inserts a specific user's or guild's reminders in MongoDB
func SaveReminders(id string, remindMeSlice *RemindMeSlice) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// If no reminders left, delete the entry
	if len(remindMeSlice.RemindMeSlice) == 0 {
		_, err := RemindersCollection.DeleteOne(ctx, bson.M{"id": id})
		if err != nil {
			log.Printf("Error deleting reminders for %s: %v\n", id, err)
			return err
		}
		return nil
	}

	// Convert to MongoDB structure
	data := ConvertRemindMeSlice(id, remindMeSlice)

	filter := bson.M{"id": id}
	update := bson.M{"$set": data}

	opts := options.Update().SetUpsert(true)

	_, err := RemindersCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Printf("Error saving updated reminders for %s: %v\n", id, err)
		return err
	}

	return nil
}

// GetDueReminders retrieves reminders that are due for sending
func GetDueReminders() (map[string]*RemindMeSlice, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()

	// Query for reminders with at least one past-due reminder
	filter := bson.M{"reminders": bson.M{"$elemMatch": bson.M{"date": bson.M{"$lte": now}}}}
	cursor, err := RemindersCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch due reminders: %v", err)
	}
	defer cursor.Close(ctx)

	remindersMap := make(map[string]*RemindMeSlice)
	for cursor.Next(ctx) {
		var reminderData RemindMeSliceMongo
		if err := cursor.Decode(&reminderData); err != nil {
			log.Println("Error decoding due reminders from MongoDB:", err)
			continue
		}

		remindersMap[reminderData.ID] = ConvertMongoToRemindMeSlice(reminderData)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error while loading due reminders: %v", err)
	}

	return remindersMap, nil
}

// ConvertRemindMe converts *RemindMe to RemindMeMongo
func ConvertRemindMe(r *RemindMe) RemindMeMongo {
	if r == nil {
		return RemindMeMongo{}
	}

	return RemindMeMongo{
		Message:        r.Message,
		Date:           r.Date,
		CommandChannel: r.CommandChannel,
		RemindID:       r.RemindID,
	}
}

// ConvertMongoToRemindMe converts RemindMeMongo back to entities.RemindMe
func ConvertMongoToRemindMe(r RemindMeMongo) *RemindMe {
	return &RemindMe{
		Message:        r.Message,
		Date:           r.Date,
		CommandChannel: r.CommandChannel,
		RemindID:       r.RemindID,
	}
}

// ConvertRemindMeSlice converts RemindMeSlice to RemindMeSliceMongo
func ConvertRemindMeSlice(id string, remindMeSlice *RemindMeSlice) RemindMeSliceMongo {
	if remindMeSlice == nil {
		return RemindMeSliceMongo{ID: id}
	}

	var reminders []RemindMeMongo
	for _, remind := range remindMeSlice.RemindMeSlice {
		if remind == nil {
			continue
		}
		reminders = append(reminders, ConvertRemindMe(remind))
	}

	return RemindMeSliceMongo{
		ID:        id,
		IsGuild:   remindMeSlice.Guild,
		Reminders: reminders,
		Premium:   remindMeSlice.Premium,
	}
}

// ConvertMongoToRemindMeSlice converts RemindMeSliceMongo back to RemindMeSlice
func ConvertMongoToRemindMeSlice(r RemindMeSliceMongo) *RemindMeSlice {
	if r.Reminders == nil {
		return &RemindMeSlice{}
	}

	reminders := make([]*RemindMe, 0, len(r.Reminders))
	for _, remind := range r.Reminders {
		reminders = append(reminders, ConvertMongoToRemindMe(remind))
	}

	return &RemindMeSlice{
		RemindMeSlice: reminders,
		Premium:       r.Premium,
		Guild:         r.IsGuild,
	}
}

func EnsureRemindersIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.M{"id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.M{"reminders.date": 1},
			Options: options.Index(),
		},
	}

	_, err := RemindersCollection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.Fatal("Failed to create indexes for reminders:", err)
	}
}
