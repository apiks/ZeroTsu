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
	// Optimized client options for better concurrent performance
	clientOptions := options.Client().ApplyURI(uri).
		SetMaxPoolSize(50).
		SetMinPoolSize(5).
		SetMaxConnIdleTime(60 * time.Second).
		SetRetryWrites(true).
		SetRetryReads(true).
		SetServerSelectionTimeout(10 * time.Second).
		SetSocketTimeout(60 * time.Second).
		SetConnectTimeout(15 * time.Second).
		SetHeartbeatInterval(10 * time.Second).
		SetMaxConnecting(10)

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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

// GetContextWithTimeout returns a context with appropriate timeout for MongoDB operations
func GetContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// CheckMongoDBHealth checks if the MongoDB connection is healthy
func CheckMongoDBHealth() error {
	if MongoClient == nil {
		return fmt.Errorf("MongoDB client is nil")
	}

	ctx, cancel := GetContextWithTimeout(5 * time.Second)
	defer cancel()

	return MongoClient.Ping(ctx, nil)
}

// ReconnectMongoDB attempts to reconnect to MongoDB if the connection is unhealthy
func ReconnectMongoDB(uri string) error {
	if err := CheckMongoDBHealth(); err == nil {
		return nil // Connection is healthy
	}

	log.Println("MongoDB connection unhealthy, attempting to reconnect...")

	// Close existing connection
	if MongoClient != nil {
		CloseMongoDB()
	}

	// Reinitialize connection
	InitMongoDB(uri)

	// Test the new connection
	return CheckMongoDBHealth()
}

// LogMongoDBStats logs MongoDB connection pool statistics
func LogMongoDBStats() {
	if MongoClient == nil {
		return
	}

	// Get connection pool statistics
	stats := MongoClient.NumberSessionsInProgress()
	log.Printf("MongoDB connection pool stats - Active sessions: %d", stats)
}
