package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Member struct {
	ID       primitive.ObjectID `bson:"_id"`
	Role     string             `bson:"role"`
	MemberId primitive.ObjectID `bson:"memberId"`
}

type Members struct {
	MemberId primitive.ObjectID `bson:"memberId"`
	Id       primitive.ObjectID `bson:"_id"`
}

type AllowedMemberApps struct {
	Members []Members          `bson:"members"`
	AppId   primitive.ObjectID `bson:"appId"`
}

type Rules struct {
	AllowedMemberApps []AllowedMemberApps `bson:"allowedMemberApps"`
}

type Workspace struct {
	ID      primitive.ObjectID `bson:"_id"`
	Owner   primitive.ObjectID `bson:"owner"`
	Members []Member           `bson:"members"`
	Rules   Rules              `bson:"rules"`
}
