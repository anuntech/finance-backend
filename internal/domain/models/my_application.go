package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type MyApplication struct {
	ID                    primitive.ObjectID `bson:"_id"`
	WorkspaceId           primitive.ObjectID `bson:"workspaceId"`
	AllowedApplicationsId []string           `bson:"allowedApplicationsId"`
}
