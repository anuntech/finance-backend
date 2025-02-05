package models

import "time"

type Account struct {
	Id          string    `bson:"_id" json:"id"`
	CreatedAt   time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updatedAt"`
	Name        string    `bson:"name" json:"name"`
	Image       string    `bson:"image" json:"image"`
	Color       string    `bson:"color" json:"color"`
	WorkspaceId string    `bson:"workspace_id" json:"workspaceId"`
	UserId      string    `bson:"user_id" json:"userId"`
}

type AccountInput struct {
	Name        string `bson:"name"`
	Image       string `bson:"image"`
	Color       string `bson:"color"`
	WorkspaceId string `bson:"workspace_id"`
	UserId      string `bson:"user_id"`
}
