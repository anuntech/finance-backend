package factory

import (
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/credit_card_repository"
	controllers "github.com/anuntech/finance-backend/internal/presentation/controllers/credit_card"
	"go.mongodb.org/mongo-driver/mongo"
)

// MakeCreateCreditCardController creates the controller for creating credit cards
func MakeCreateCreditCardController(db *mongo.Database) *controllers.CreateCreditCardController {
	createRepo := credit_card_repository.NewCreateCreditCardRepository(db)
	findByNameRepo := credit_card_repository.NewFindByNameMongoRepository(db)
	return controllers.NewCreateCreditCardController(createRepo, findByNameRepo)
}

// MakeGetCreditCardsController creates the controller for retrieving credit cards
func MakeGetCreditCardsController(db *mongo.Database) *controllers.GetCreditCardsController {
	findRepo := credit_card_repository.NewFindCreditCardsRepository(db)
	return controllers.NewGetCreditCardsController(findRepo)
}

// MakeGetCreditCardByIdController creates the controller for retrieving a credit card by ID
func MakeGetCreditCardByIdController(db *mongo.Database) *controllers.GetCreditCardByIdController {
	findByIdRepo := credit_card_repository.NewFindCreditCardByIdRepository(db)
	return controllers.NewGetCreditCardByIdController(findByIdRepo)
}

// MakeUpdateCreditCardController creates the controller for updating credit cards
func MakeUpdateCreditCardController(db *mongo.Database) *controllers.UpdateCreditCardController {
	updateRepo := credit_card_repository.NewUpdateCreditCardRepository(db)
	findByIdRepo := credit_card_repository.NewFindCreditCardByIdRepository(db)
	findByNameRepo := credit_card_repository.NewFindByNameMongoRepository(db)
	return controllers.NewUpdateCreditCardController(updateRepo, findByIdRepo, findByNameRepo)
}

// MakeDeleteCreditCardController creates the controller for deleting credit cards
func MakeDeleteCreditCardController(db *mongo.Database) *controllers.DeleteCreditCardController {
	deleteRepo := credit_card_repository.NewDeleteCreditCardRepository(db)
	return controllers.NewDeleteCreditCardController(deleteRepo)
}
