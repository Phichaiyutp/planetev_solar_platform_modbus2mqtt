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
func GetMongoClient(uri string) (*Client, error) {
	mongoOnce.Do(func() {
		if uri == "" {
			uri = "mongodb://localhost:27017" // Default URI if none is provided
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		clientOptions := options.Client().ApplyURI(uri)
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

	// Use the context stored in the client
	result, err := collection.InsertOne(c.ctx, document)
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
