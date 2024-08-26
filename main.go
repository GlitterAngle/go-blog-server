package main

import (
	"context"
	"fmt"
	"go-blog-server/api"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

    fmt.Printf("App will run on port: %s\n", port)

    // Connect to MongoDB
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
    db := client.Database(databaseName)
	http.HandleFunc("/posts", func(w http.ResponseWriter,r *http.Request){api.PostsHandler(w,r,db)})
	http.HandleFunc("/post/", func(w http.ResponseWriter,r *http.Request){api.PostHandler(w,r,db)})
	http.HandleFunc("/users", func(w http.ResponseWriter,r *http.Request){api.UsersHandler(w,r,db)})
	http.HandleFunc("/user/", func(w http.ResponseWriter,r *http.Request){api.UserHandler(w,r,db)})

   // Run the server in a separate goroutine
	server := &http.Server{Addr: ":" + port}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Wait for an interrupt signal to gracefully shutdown the server and disconnect from MongoDB
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	if err := server.Shutdown(context.TODO()); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	// Disconnect from MongoDB
	if err := client.Disconnect(context.TODO()); err != nil {
		log.Fatalf("Failed to disconnect from MongoDB: %v", err)
	}

	fmt.Println("Disconnected from MongoDB")

	time.Sleep(10* time.Second)
}
