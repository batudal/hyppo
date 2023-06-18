package main

import (
	"context"
	"errors"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoClient() (*mongo.Client, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		return nil, errors.New("MONGODB_URI is empty")
	}
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	return client, nil
}
