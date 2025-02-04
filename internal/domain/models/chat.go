package models

import "time"

type Chat struct {
	Id        string    `bson:"_id"`
	CreatedAt time.Time `bson:"created_at"`
	OwnerId   string    `bson:"owner_id"`
}
