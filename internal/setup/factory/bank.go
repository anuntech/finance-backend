package factory

import (
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/bank_repository"
	controllers "github.com/anuntech/finance-backend/internal/presentation/controllers/bank"
	"go.mongodb.org/mongo-driver/mongo"
)

func MakeGetBanksController(db *mongo.Database) *controllers.GetBankController {
	getBankRepository := bank_repository.NewFindAllMongoRepository(db)

	return controllers.NewGetBankController(getBankRepository)
}

func MakeGetBankByIdController(db *mongo.Database) *controllers.GetBankByIdController {
	getBankByIdRepository := bank_repository.NewFindByIdMongoRepository(db)

	return controllers.NewGetBankByIdController(getBankByIdRepository)
}

func MakeGetBankByNameController(db *mongo.Database) *controllers.GetBankByNameController {
	getBankByNameRepository := bank_repository.NewFindByNameMongoRepository(db)

	return controllers.NewGetBankByNameController(getBankByNameRepository)
}
