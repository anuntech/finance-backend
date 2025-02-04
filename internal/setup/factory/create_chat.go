package factory

import (
	controllers "github.com/willchat-ofc/api-willchat-golang/internal/presentation/controllers/chat"
	"go.mongodb.org/mongo-driver/mongo"
)

func MakeCreateChatController(db *mongo.Database) *controllers.CreateChatController {
	// createChatRepository := chat_repository.NewCreateChatMongoRepository(db)
	// dbCreateChat := usecase.NewDbCreateChat(createChatRepository)

	// return controllers.NewCreateChatController(dbCreateChat)

	return &controllers.CreateChatController{}
}
