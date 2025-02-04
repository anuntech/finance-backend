package chat_repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/willchat-ofc/api-willchat-golang/internal/data/protocols"
	"github.com/willchat-ofc/api-willchat-golang/internal/domain/models"
	"github.com/willchat-ofc/api-willchat-golang/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreateChatMongoRepository struct {
	Db *mongo.Database
}

func NewCreateChatMongoRepository(db *mongo.Database) *CreateChatMongoRepository {
	return &CreateChatMongoRepository{
		Db: db,
	}
}

func (c *CreateChatMongoRepository) Create(ownerId string) (*protocols.CreateChatRepositoryOutput, error) {
	collection := c.Db.Collection("chat")

	id := uuid.New().String()
	createdAt := time.Now()
	_, err := collection.InsertOne(helpers.Ctx, &models.Chat{
		Id:        id,
		CreatedAt: createdAt,
		OwnerId:   ownerId,
	})

	if err != nil {
		return nil, err
	}

	return &protocols.CreateChatRepositoryOutput{
		Id:        id,
		OwnerId:   ownerId,
		CreatedAt: createdAt.String(),
	}, nil
}
