package factory

import (
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/account_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/bank_repository"
	controllers "github.com/anuntech/finance-backend/internal/presentation/controllers/account"
	"go.mongodb.org/mongo-driver/mongo"
)

func MakeCreateAccountController(db *mongo.Database) *controllers.CreateAccountController {
	accountRepository := account_repository.NewCreateAccountMongoRepository(db)
	findManyByUserIdAndWorkspaceId := account_repository.NewFindManyByUserIdAndWorkspaceIdMongoRepository(db)
	findBankById := bank_repository.NewFindByIdMongoRepository(db)
	return controllers.NewCreateAccountController(accountRepository, findManyByUserIdAndWorkspaceId, findBankById)
}

func MakeGetAccountsController(db *mongo.Database) *controllers.GetAccountsController {
	findManyByUserIdAndWorkspaceId := account_repository.NewFindManyByUserIdAndWorkspaceIdMongoRepository(db)
	return controllers.NewGetAccountsController(findManyByUserIdAndWorkspaceId)
}

func MakeGetAccountByIdController(db *mongo.Database) *controllers.GetAccountByIdController {
	findById := account_repository.NewFindByIdMongoRepository(db)
	return controllers.NewGetAccountByIdController(findById)
}

func MakeDeleteAccountController(db *mongo.Database) *controllers.DeleteAccountController {
	deleteAccount := account_repository.NewDeleteAccountMongoRepository(db)
	return controllers.NewDeleteAccountController(deleteAccount)
}

func MakeUpdateAccountController(db *mongo.Database) *controllers.UpdateAccountController {
	updateAccount := account_repository.NewUpdateAccountMongoRepository(db)
	findBankById := bank_repository.NewFindByIdMongoRepository(db)
	findAccountById := account_repository.NewFindByIdMongoRepository(db)
	return controllers.NewUpdateAccountController(updateAccount, findBankById, findAccountById)
}

func MakeImportAccountController(db *mongo.Database) *controllers.ImportAccountController {
	importAccounts := account_repository.NewImportAccountsMongoRepository(db)
	findAccountByWorkspaceId := account_repository.NewFindManyByUserIdAndWorkspaceIdMongoRepository(db)
	return controllers.NewImportAccountController(importAccounts, findAccountByWorkspaceId)
}
