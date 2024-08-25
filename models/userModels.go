package models

import (
	"context"
	"errors"
	"regexp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username string             `bson:"username" json:"username"`
	Email    string             `bson:"email" json:"email"`
	Password string             `bson:"password" json:"password"`
}

func (u *User) Validate(db *mongo.Database) error {
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return errors.New("invalid email address")
	}

	if err := isUnique(db, "email", u.Email); err != nil{
		return err
	}

	if err := isUnique(db, "username", u.Username); err != nil{
		return err
	}

	return nil
}

func isUnique(db *mongo.Database, fieldName string, value string) error{
	collection := db.Collection("user")
	filter := bson.M{fieldName: value}
	var existingUser User

	err := collection.FindOne(context.TODO(), filter).Decode(&existingUser)
	if err == mongo.ErrNoDocuments{
		return nil
	}
	if err != nil{
		return err
	}

	return errors.New(fieldName + " already exists")
}