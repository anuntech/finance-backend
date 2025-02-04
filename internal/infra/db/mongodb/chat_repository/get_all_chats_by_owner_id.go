package chat_repository

import (
	"github.com/willchat-ofc/api-willchat-golang/internal/data/protocols"
	"github.com/willchat-ofc/api-willchat-golang/internal/domain/models"
	"github.com/willchat-ofc/api-willchat-golang/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type GetAllChatsByOwnerIdMongoRepository struct {
	Db *mongo.Database
}

func NewGetAllChatsByOwnerIdMongoRepository(db *mongo.Database) *GetAllChatsByOwnerIdMongoRepository {
	return &GetAllChatsByOwnerIdMongoRepository{
		Db: db,
	}
}

func (c *GetAllChatsByOwnerIdMongoRepository) Get(ownerId string) ([]*protocols.GetAllChatsByOwnerIdRepositoryOutput, error) {
	collection := c.Db.Collection("chat")

	filter := bson.D{{Key: "owner_id", Value: ownerId}}

	cursor, err := collection.Find(helpers.Ctx, filter)

	if err != nil {
		return nil, err
	}

	var chats []models.Chat
	err = cursor.All(helpers.Ctx, &chats)
	if err != nil {
		return nil, err
	}

	var result []*protocols.GetAllChatsByOwnerIdRepositoryOutput
	for _, chat := range chats {
		convertedChat := &protocols.GetAllChatsByOwnerIdRepositoryOutput{
			Id:        chat.Id,
			OwnerId:   chat.OwnerId,
			CreatedAt: chat.CreatedAt.String(),
		}
		result = append(result, convertedChat)
	}

	return result, nil
}
