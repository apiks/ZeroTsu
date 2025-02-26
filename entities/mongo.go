package entities

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

var (
	MongoClient         *mongo.Client
	GuildCollection     *mongo.Collection
	FeedCheckCollection *mongo.Collection
	RemindersCollection *mongo.Collection
	AnimeSubsCollection *mongo.Collection
)

// InitMongoDB initializes the MongoDB connection and collections
func InitMongoDB(uri string) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal("MongoDB ping failed:", err)
	}

	MongoClient = client
	db := client.Database("zerotsu")
	GuildCollection = db.Collection("guilds")
	FeedCheckCollection = db.Collection("feed_checks")
	RemindersCollection = db.Collection("reminders")
	AnimeSubsCollection = db.Collection("anime_subs")

	fmt.Println("Connected to MongoDB!")
}

// CloseMongoDB closes the MongoDB connection gracefully
func CloseMongoDB() {
	if MongoClient != nil {
		err := MongoClient.Disconnect(context.Background())
		if err != nil {
			log.Println("Error disconnecting MongoDB:", err)
		} else {
			fmt.Println("MongoDB connection closed successfully.")
		}
	}
}
