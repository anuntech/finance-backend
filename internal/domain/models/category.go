package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubCategoryCategory struct {
	Id            primitive.ObjectID `json:"id" bson:"_id"`
	Name          string             `json:"name" bson:"name"`
	Icon          string             `json:"icon" bson:"icon"`
	Amount        float64            `json:"amount" bson:"-"`
	CurrentAmount float64            `json:"currentAmount" bson:"-"`
}

type Category struct {
	Id            primitive.ObjectID `json:"id" bson:"_id"`
	Name          string             `json:"name" bson:"name"`
	Amount        float64            `json:"amount" bson:"-"`
	CurrentAmount float64            `json:"currentAmount" bson:"-"`
	Type          string             `json:"type" bson:"type"` // EXPENSE, RECIPE, TAG, PERSONALIZED
	Icon          string             `json:"icon" bson:"icon"`
	// PersonalizedType string              `json:"personalizedType" bson:"personalized_type"` // NUMBER | TEXT | PHONE_NUMBER
	SubCategories []SubCategoryCategory `json:"subCategories" bson:"sub_categories"`
	CreatedAt     time.Time             `json:"createdAt" bson:"created_at"`
	UpdatedAt     time.Time             `json:"updatedAt" bson:"updated_at"`
	WorkspaceId   primitive.ObjectID    `json:"workspaceId" bson:"workspace_id"`
}

func (r *Category) CalculateTotalAmount() {
	total := 0.0
	for _, subCategory := range r.SubCategories {
		total += subCategory.Amount
	}
	r.Amount = total
}

func (r *Category) CalculateTotalCurrentAmount() {
	total := 0.0
	for _, subCategory := range r.SubCategories {
		total += subCategory.CurrentAmount
	}
	r.CurrentAmount = total
}

func (r *Category) InvertSubCategoriesOrder() {
	for i, j := 0, len(r.SubCategories)-1; i < j; i, j = i+1, j-1 {
		r.SubCategories[i], r.SubCategories[j] = r.SubCategories[j], r.SubCategories[i]
	}
}
