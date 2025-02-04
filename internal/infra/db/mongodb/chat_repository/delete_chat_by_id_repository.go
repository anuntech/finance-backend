package chat_repository

import (
	"github.com/willchat-ofc/api-willchat-golang/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type DeleteChatByIdMongoRepository struct {
	Db *mongo.Database
}

func NewDeleteChatByIdMongoRepository(db *mongo.Database) *DeleteChatByIdMongoRepository {
	return &DeleteChatByIdMongoRepository{
		Db: db,
	}
}

func (c *DeleteChatByIdMongoRepository) Delete(chatId string) error {
	collection := c.Db.Collection("chat")

	filter := bson.D{{Key: "_id", Value: chatId}}
	_, err := collection.DeleteOne(helpers.Ctx, filter)

	return err
}
