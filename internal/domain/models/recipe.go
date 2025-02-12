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
	SubCategories []SubRecipeCategory `json:"subCategories" bson:"sub_categories"`
	CreatedAt     time.Time           `json:"createdAt" bson:"created_at"`
	UpdatedAt     time.Time           `json:"updatedAt" bson:"updated_at"`
	WorkspaceId   primitive.ObjectID  `json:"workspaceId" bson:"workspace_id"`
	TotalAmount   float64             `json:"totalAmount" bson:"-"`
}

func (r *Recipe) CalculateTotalAmount() float64 {
	total := 0.0
	for _, subCategory := range r.SubCategories {
		total += subCategory.Amount
	}
	return total
}
