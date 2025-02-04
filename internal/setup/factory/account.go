package factory

import (
	controllers "github.com/anuntech/finance-backend/internal/presentation/controllers/account"
	"go.mongodb.org/mongo-driver/mongo"
)

func MakeCreateAccountController(db *mongo.Database) *controllers.CreateAccountController {
	// createChatRepository := chat_repository.NewCreateChatMongoRepository(db)
	// dbCreateChat := usecase.NewDbCreateChat(createChatRepository)

	// return controllers.NewCreateChatController(dbCreateChat)
	// createAccountController := controllers.NewCreateAccountController(usecase.CreateAccount{})

	return &controllers.CreateAccountController{
		// CreateAccount: dbCreateAccount,
	}
}
