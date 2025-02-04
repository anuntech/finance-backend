package models

type Message struct {
	Id         string `json:"id" bson:"_id"`
	ChatId     string `json:"chatId" bson:"chat_id"`
	Message    string `json:"message" bson:"message"`
	AuthorName string `json:"authorName" bson:"author_name"`
	AuthorId   string `json:"authorId" bson:"author_id"`
}
