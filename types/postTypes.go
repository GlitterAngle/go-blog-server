package types

import(
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct{
	ID	primitive.ObjectID	`bson:"_id,omitempty" json:"id"`
	User	primitive.ObjectID	`bson:"userId" json:"userId"`
	PostBody	string	`bson:"postBody" json:"postBody"`
	Img	string `bson:"img" json:"img"`
}