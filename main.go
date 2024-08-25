package main

import (
    "fmt"
    "log"
    "os"

    "github.com/joho/godotenv"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "context"
)

func main() {
    // Load the .env file
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file")
    }

    // Get the environment variables
    databaseURI := os.Getenv("DATABASE_URI")
    databaseName := os.Getenv("DATABASE_NAME")
    port := os.Getenv("PORT")

    fmt.Printf("Connecting to MongoDB at URI: %s\n", databaseURI)
    fmt.Printf("Using database: %s\n", databaseName)
    fmt.Printf("App will run on port: %s\n", port)

    // Example: Connect to MongoDB
    clientOptions := options.Client().ApplyURI(databaseURI)
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)
    }

    // Check the connection
    err = client.Ping(context.TODO(), nil)
    if err != nil {
        log.Fatalf("Failed to ping MongoDB: %v", err)
    }

    fmt.Println("Connected to MongoDB!")

    // Do something with the database
    // db := client.Database(databaseName)

    // Your logic here...

    // Remember to disconnect from MongoDB when you're done
    err = client.Disconnect(context.TODO())
    if err != nil {
        log.Fatalf("Failed to disconnect from MongoDB: %v", err)
    }

    fmt.Println("Disconnected from MongoDB")
}
