package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func connectToMongoDB() (*mongo.Client, error) {
	uri := os.Getenv("DATABASE_URI")
	clientOptions := options.Client().ApplyURI(uri)

	// Use mongo.Connect instead of mongo.NewClient and client.Connect
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the database to verify connection
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	log.Println("Connected to MongoDB!")
	return client, nil
}

func main() {
	_, err := connectToMongoDB()
	if err != nil {
		log.Fatalf("Could not connect to MongoDB: %v", err)
	}

	// Add handlers once made
	listenAddr := flag.String("listenaddr", ":4999", "HTTP listen address")
	flag.Parse()

	log.Printf("Server is running at http://localhost%s", *listenAddr)
	http.ListenAndServe(*listenAddr, nil)
}
