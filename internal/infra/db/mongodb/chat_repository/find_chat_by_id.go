package chat_repository

import (
	"github.com/willchat-ofc/api-willchat-golang/internal/domain/models"
	"github.com/willchat-ofc/api-willchat-golang/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type FindChatByIdMongoRepository struct {
	Db *mongo.Database
}

func NewFindChatByIdMongoRepository(db *mongo.Database) *FindChatByIdMongoRepository {
	return &FindChatByIdMongoRepository{
		Db: db,
	}
}

func (c *FindChatByIdMongoRepository) Find(chatId string) (*models.Chat, error) {
	collection := c.Db.Collection("chat")

	filter := bson.D{{Key: "_id", Value: chatId}}
	var chat models.Chat
	if err := collection.FindOne(helpers.Ctx, filter).Decode(&chat); err != nil {
		return nil, err
	}

	return &chat, nil
}
