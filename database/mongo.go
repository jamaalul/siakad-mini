package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"siakad-mini/config"
)

// NewMongoClient membuat dan memverifikasi koneksi ke instance MongoDB.
func NewMongoClient(ctx context.Context) (*mongo.Client, error) {
	uri := config.GetEnv("MONGO_URI", "mongodb://localhost:27017")

	clientOpts := options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(5 * time.Second).
		SetConnectTimeout(5 * time.Second)

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat client MongoDB: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(pingCtx)
		return nil, fmt.Errorf("gagal terhubung ke MongoDB: %w", err)
	}

	return client, nil
}

// GetDatabase mengembalikan instance *mongo.Database sesuai konfigurasi MONGO_DATABASE.
func GetDatabase(client *mongo.Client) *mongo.Database {
	dbName := config.GetEnv("MONGO_DATABASE", "siakad_mini")
	return client.Database(dbName)
}
