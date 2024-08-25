package api

import(
	"context"
	"encoding/json"
	"go-blog-server/models"
	"net/http"
	"sync"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var usersMu sync.Mutex

func UsersHandler(w http.ResponseWriter, r *http.Request, db *mongo.Database){
	switch r.Method{
	case "GET":
		handleGetUsers(w,r,db)
	case "POST":
		handlePostUser(w,r,db)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func UserHandler(w http.ResponseWriter, r *http.Request, db *mongo.Database){
	id := r.URL.Path[len("/user/"):]
	if len(id)==0{
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	switch r.Method{
	case "GET":
		handleGetUser(w, r, id, db)
	case "PUT":
		handleUpdateUser(w, r, id, db)
	case "DELETE":
		handleDeleteUser(w, r, id, db)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetUsers(w http.ResponseWriter, r *http.Request, db *mongo.Database){
	usersMu.Lock()
	defer usersMu.Unlock()

	collection := db.Collection("users")
	cur, err := collection.Find(context.TODO(), bson.M{})
	if err != nil{
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.TODO())

	var users []models.User
	for cur.Next(context.TODO()){
		var user models.User
		err := cur.Decode(&user)
		if err != nil{
			http.Error(w, "Error decoding user", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	if err := cur.Err(); err !=nil{
		http.Error(w, "Error iterating through users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func handlePostUser(w http.ResponseWriter, r *http.Request, db *mongo.Database){
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err!= nil{
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	collection := db.Collection("users")
	result, err := collection.InsertOne(context.TODO(), user)
	if err != nil{
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok{
		user.ID = oid
	} else{
		http.Error(w, "Failed to retrieve ID for the new user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func handleGetUser(w http.ResponseWriter, r *http.Request, id string, db *mongo.Database){
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil{
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": objectId}

	var user models.User
	collection := db.Collection("users")
	err = collection.FindOne(context.TODO(), filter).Decode(&user)
	if err == mongo.ErrNilDocument{
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}else if err != nil{
		http.Error(w, "Failed to retrieve user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func handleUpdateUser(w http.ResponseWriter, r *http.Request, id string, db *mongo.Database){
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil{
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": objectId}

	var updatedUser models.User
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil{
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	update := bson.M{
		"$set": bson.M{
			"username": updatedUser.Username,
		},
	}

	collection := db.Collection("users")
	err = collection.FindOneAndUpdate(context.TODO(), filter,update).Decode(updatedUser)
	if err == mongo.ErrNilDocument{
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}else if err !=nil{
		http.Error(w, "Failed to update post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "apllication/json")
	json.NewEncoder(w).Encode(updatedUser)
}

func handleDeleteUser(w http.ResponseWriter, r *http.Request, id string, db *mongo.Database){
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil{
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	userFilter := bson.M{"_id": objectId}
	postFilter := bson.M{"userId": objectId}

	var deleteUser models.User

	userCollection := db.Collection("users")
	postCollection := db.Collection("posts")

	err = userCollection.FindOneAndDelete(context.TODO(), userFilter).Decode(&deleteUser)
	if err == mongo.ErrNoDocuments{
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil{
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	deleteResult, err := postCollection.DeleteMany(context.TODO(), postFilter)
	if err != nil{
		http.Error(w, "Failed to delete posts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deleteUser": deleteUser,
		"deletedPosts": deleteResult.DeletedCount,
	})
}