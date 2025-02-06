package models

import "time"

type Account struct {
	Id          string    `bson:"_id" json:"id"`
	CreatedAt   time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updatedAt"`
	Name        string    `bson:"name" json:"name"`
	Bank        string    `bson:"bank" json:"bank"`
	WorkspaceId string    `bson:"workspace_id" json:"workspaceId"`
}

type AccountInput struct {
	Name        string `bson:"name"`
	Bank        string `bson:"bank"`
	WorkspaceId string `bson:"workspace_id"`
}
