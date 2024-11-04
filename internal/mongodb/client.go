package mongodb

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client wraps the MongoDB client.
type Client struct {
	client *mongo.Client
	ctx    context.Context
	cancel context.CancelFunc
}

// Initialize the MongoDB client instance and context.
var (
	clientInstance *Client
	mongoOnce      sync.Once
)

// GetMongoClient initializes a singleton MongoDB client instance.
func GetMongoClient(username string, password string, host string, port string, dbName string) (*Client, error) {
	mongoOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		// Add credentials if provided
		credential := options.Credential{
			AuthMechanism: "SCRAM-SHA-256",
			AuthSource:    dbName,
			Username:      username,
			Password:      password,
		}
		uri := fmt.Sprintf("%s:%s", host, port)
		clientOptions := options.Client().SetHosts(
			[]string{uri},
		).SetAuth(credential)
		mongoClient, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Fatal("Failed to connect to MongoDB:", err)
		}

		// Ping the MongoDB server to verify connection
		if err = mongoClient.Ping(ctx, nil); err != nil {
			cancel()
			log.Fatal("Failed to ping MongoDB:", err)
		}

		clientInstance = &Client{
			client: mongoClient,
			ctx:    ctx,
			cancel: cancel,
		}
	})

	return clientInstance, nil
}

// InsertOne inserts a single document into the specified collection.
func (c *Client) InsertOne(databaseName, collectionName string, document interface{}) (*mongo.InsertOneResult, error) {
	collection := c.client.Database(databaseName).Collection(collectionName)

	// Increase the timeout for the insert operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel() // Ensure we cancel the context to release resources
	// Use the new context for the insert operation
	result, err := collection.InsertOne(ctx, document)
	if err != nil {
		return nil, fmt.Errorf("failed to insert document: %w", err)
	}
	return result, nil
}

// CloseClient closes the MongoDB client connection.
func (c *Client) CloseClient() {
	if err := c.client.Disconnect(c.ctx); err != nil {
		log.Fatal("Failed to disconnect from MongoDB:", err)
	}
	c.cancel() // Call the cancel function to release resources
}
