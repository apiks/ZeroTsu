package entities

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	// Optimized client options for memory efficiency
	clientOptions := options.Client().ApplyURI(uri).
		SetMaxPoolSize(10).                   // Limit connection pool size
		SetMinPoolSize(2).                    // Minimum connections
		SetMaxConnIdleTime(30 * time.Second). // Close idle connections
		SetRetryWrites(true).                 // Enable retry writes
		SetRetryReads(true).                  // Enable retry reads
		SetServerSelectionTimeout(5 * time.Second).
		SetSocketTimeout(30 * time.Second).
		SetConnectTimeout(10 * time.Second)

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
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
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := MongoClient.Disconnect(ctx)
		if err != nil {
			log.Println("Error disconnecting MongoDB:", err)
		} else {
			fmt.Println("MongoDB connection closed successfully.")
		}
	}
}

// GetOptimizedFindOptions returns optimized find options for memory efficiency
func GetOptimizedFindOptions() *options.FindOptions {
	return options.Find().
		SetBatchSize(100).        // Process in smaller batches
		SetNoCursorTimeout(false) // Allow cursor timeout
}

// GetOptimizedUpdateOptions returns optimized update options
func GetOptimizedUpdateOptions() *options.UpdateOptions {
	return options.Update().
		SetUpsert(true).
		SetBypassDocumentValidation(false)
}
