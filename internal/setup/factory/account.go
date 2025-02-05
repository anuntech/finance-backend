package factory

import (
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/account_repository"
	controllers "github.com/anuntech/finance-backend/internal/presentation/controllers/account"
	"go.mongodb.org/mongo-driver/mongo"
)

func MakeCreateAccountController(db *mongo.Database) *controllers.CreateAccountController {
	accountRepository := account_repository.NewCreateAccountMongoRepository(db)
	findManyByUserIdAndWorkspaceId := account_repository.NewFindManyByUserIdAndWorkspaceIdMongoRepository(db)
	return controllers.NewCreateAccountController(accountRepository, findManyByUserIdAndWorkspaceId)
}
