package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Bank struct {
	Id    primitive.ObjectID `bson:"_id" json:"id"`
	Name  string             `bson:"name" json:"name"`
	Image string             `bson:"image" json:"image"`
}
