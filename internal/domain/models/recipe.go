package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubRecipeCategory struct {
	Id     primitive.ObjectID `json:"id" bson:"id"`
	Name   string             `json:"name" bson:"name"`
	Icon   string             `json:"icon" bson:"icon"`
	Amount float64            `json:"amount" bson:"amount"`
}

type Recipe struct {
	Id            primitive.ObjectID  `json:"id" bson:"_id"`
	Name          string              `json:"name" bson:"name"`
	AccountId     primitive.ObjectID  `json:"accountId" bson:"accountId"`
	SubCategories []SubRecipeCategory `json:"subCategories" bson:"subCategories"`
	CreatedAt     time.Time           `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time           `json:"updatedAt" bson:"updatedAt"`
	WorkspaceId   primitive.ObjectID  `json:"workspaceId" bson:"workspaceId"`
}
