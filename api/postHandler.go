package api

import (
	"context"
	"encoding/json"
	"go-blog-server/models"
	"net/http"
	"sync"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var postsMu sync.Mutex

func PostsHandler(w http.ResponseWriter, r *http.Request, db *mongo.Database){
	switch r.Method{
	case "GET":
		handleGetPosts(w,r, db)
	case "POST":
		handlePostPosts(w,r, db)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func PostHandler(w http.ResponseWriter, r *http.Request, db *mongo.Database){
	id  := r.URL.Path[len("/posts/"):]
	if len(id) == 0{
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	switch r.Method{
	case "GET": 
		handleGetPost(w, r, id, db)
	case "PUT":
		handleUpdatePost(w, r, id, db)
	case "DELETE": 
		handleDeletePost(w, r, id, db)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetPosts(w http.ResponseWriter, r *http.Request, db *mongo.Database){
	postsMu.Lock()
	defer postsMu.Unlock()

	collection := db.Collection("posts")
	cur, err := collection.Find(context.TODO(), bson.M{})
	if err != nil{
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.TODO())

	var posts []models.Post
	for cur.Next(context.TODO()){
		var post models.Post 
		err := cur.Decode(&post)
		if err != nil{
			http.Error(w, "Error decoding post", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	if err := cur.Err(); err != nil{
		http.Error(w, "Error iterating through posts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func handlePostPosts(w http.ResponseWriter, r *http.Request, db *mongo.Database){
	var post models.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil{
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	collection := db.Collection("posts")
	result, err := collection.InsertOne(context.TODO(), post)
	if err != nil{
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}
	
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok{
		post.ID = oid
	} else{
		http.Error(w, "Failed to retrieve ID for the new post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

func handleGetPost(w http.ResponseWriter, r *http.Request, id string, db *mongo.Database){
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil{
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": objectId}

	var post models.Post 
	collection := db.Collection("posts")
	err = collection.FindOne(context.TODO(), filter).Decode(&post)
	if err == mongo.ErrNilDocument{
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	} else if err != nil{
		http.Error(w, "Failed to retrieve post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func handleUpdatePost(w http.ResponseWriter, r *http.Request, id string, db *mongo.Database){
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil{
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": objectId}

	var updatedPost models.Post
	if err := json.NewDecoder(r.Body).Decode(&updatedPost); err != nil{
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	update := bson.M{
		"$set": bson.M{
			"postBody": updatedPost.PostBody,
			"img": updatedPost.Img,
		},
	}

	collection := db.Collection("posts")
	err = collection.FindOneAndUpdate(context.TODO(), filter, update).Decode(&updatedPost)
	if err == mongo.ErrNilDocument{
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}else if err != nil{
		http.Error(w, "Failed to update post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedPost)
}

func handleDeletePost(w http.ResponseWriter, r * http.Request, id string, db *mongo.Database){
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil{
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": objectId}

	var deletePost models.Post

	collection := db.Collection("posts")
	err = collection.FindOneAndDelete(context.TODO(),filter).Decode(&deletePost)
	if err == mongo.ErrNoDocuments{
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}else if err != nil{
		http.Error(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deletePost)
}