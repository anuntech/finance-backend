package factory

import (
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/account_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/bank_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/transaction_repository"
	controllers "github.com/anuntech/finance-backend/internal/presentation/controllers/account"
	"go.mongodb.org/mongo-driver/mongo"
)

func MakeCreateAccountController(db *mongo.Database) *controllers.CreateAccountController {
	accountRepository := account_repository.NewCreateAccountMongoRepository(db)
	findManyByUserIdAndWorkspaceId := account_repository.NewFindAccountsRepository(db)
	findBankById := bank_repository.NewFindByIdMongoRepository(db)
	findByNameRepository := account_repository.NewFindByNameMongoRepository(db)
	return controllers.NewCreateAccountController(accountRepository, findManyByUserIdAndWorkspaceId, findBankById, findByNameRepository)
}

func MakeGetAccountsController(db *mongo.Database) *controllers.GetAccountsController {
	findManyByUserIdAndWorkspaceId := account_repository.NewFindAccountsRepository(db)
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
	findByNameRepository := account_repository.NewFindByNameMongoRepository(db)
	return controllers.NewUpdateAccountController(updateAccount, findBankById, findAccountById, findByNameRepository)
}

func MakeImportAccountController(db *mongo.Database) *controllers.ImportAccountController {
	importAccounts := account_repository.NewImportAccountsMongoRepository(db)
	findAccountByWorkspaceId := account_repository.NewFindAccountsRepository(db)
	findByNameRepository := account_repository.NewFindByNameMongoRepository(db)
	return controllers.NewImportAccountController(importAccounts, findAccountByWorkspaceId, findByNameRepository)
}

func MakeTransferenceAccountController(db *mongo.Database) *controllers.TransferenceAccountController {
	findAccountById := account_repository.NewFindByIdMongoRepository(db)
	createTransaction := transaction_repository.NewCreateTransactionRepository(db)
	return controllers.NewTransferenceAccountController(findAccountById, createTransaction)
}
