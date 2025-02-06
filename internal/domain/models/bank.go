package models

type Bank struct {
	Id    string `bson:"_id" json:"id"`
	Name  string `bson:"name" json:"name"`
	Image string `bson:"image" json:"image"`
}
