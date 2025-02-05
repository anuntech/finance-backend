package models

import "time"

type Account struct {
	Id          string    `bson:"_id"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
	Name        string    `bson:"name"`
	Image       string    `bson:"image"`
	Color       string    `bson:"color"`
	WorkspaceId string    `bson:"workspace_id"`
	UserId      string    `bson:"user_id"`
}

type AccountInput struct {
	Name        string `bson:"name"`
	Image       string `bson:"image"`
	Color       string `bson:"color"`
	WorkspaceId string `bson:"workspace_id"`
	UserId      string `bson:"user_id"`
}
