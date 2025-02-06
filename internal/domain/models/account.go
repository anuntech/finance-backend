package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Account struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updatedAt"`
	Name        string             `bson:"name" json:"name"`
	BankId      primitive.ObjectID `bson:"bank_id" json:"bankId"`
	WorkspaceId primitive.ObjectID `bson:"workspace_id" json:"workspaceId"`
}

type AccountInput struct {
	Name        string             `bson:"name"`
	BankId      primitive.ObjectID `bson:"bank_id"`
	WorkspaceId primitive.ObjectID `bson:"workspace_id"`
}
