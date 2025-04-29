package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type CreditCard struct {
	Id          primitive.ObjectID `json:"id" bson:"_id"`
	WorkspaceId primitive.ObjectID `json:"workspaceId" bson:"workspace_id"`
	Name        string             `json:"name" bson:"name"`
	DueDate     int                `json:"dueDate" bson:"due_date"`
	CloseDate   int                `json:"closeDate" bson:"close_date"`
	Limit       float64            `json:"limit" bson:"limit"`
	Balance     float64            `json:"balance" bson:"-"`
	Flag        string             `json:"flag" bson:"flag"`
}
