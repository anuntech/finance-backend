package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type SubRecipeCategory struct {
	Id     string  `json:"id" bson:"id"`
	Name   string  `json:"name" bson:"name"`
	Amount float64 `json:"amount" bson:"amount"`
}

type Recipe struct {
	Id            primitive.ObjectID  `json:"id" bson:"_id"`
	Name          string              `json:"name" bson:"name"`
	AccountId     primitive.ObjectID  `json:"accountId" bson:"accountId"`
	SubCategories []SubRecipeCategory `json:"subCategories" bson:"subCategories"`
}
