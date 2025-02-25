package factory

import (
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/account_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/category_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/transaction_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/workspace_repository/member_repository"
	"github.com/anuntech/finance-backend/internal/presentation/controllers/transaction"
	"go.mongodb.org/mongo-driver/mongo"
)

func MakeCreateTransactionController(workspaceDb *mongo.Database, db *mongo.Database) *transaction.CreateTransactionController {
	findMemberByIdRepository := member_repository.NewFindMemberByIdRepository(workspaceDb)
	createTransactionRepository := transaction_repository.NewCreateTransactionRepository(db)
	findAccountByIdRepository := account_repository.NewFindByIdMongoRepository(db)
	findCategoryByIdRepository := category_repository.NewFindCategoryByIdRepository(db)

	return transaction.NewCreateTransactionController(findMemberByIdRepository, createTransactionRepository, findAccountByIdRepository, findCategoryByIdRepository)
}

func MakeGetTransactionController(workspaceDb *mongo.Database) *transaction.GetTransactionController {
	findTransactionsByWorkspaceIdAndMonthRepository := transaction_repository.NewTransactionRepository(workspaceDb)
	return transaction.NewGetTransactionController(findTransactionsByWorkspaceIdAndMonthRepository)
}

func MakeUpdateTransactionController(workspaceDb *mongo.Database, db *mongo.Database) *transaction.UpdateTransactionController {
	findTransactionByIdRepository := transaction_repository.NewGetTransactionByIdRepository(db)
	updateTransactionRepository := transaction_repository.NewUpdateTransactionRepository(db)
	findMemberByIdRepository := member_repository.NewFindMemberByIdRepository(workspaceDb)
	findAccountByIdRepository := account_repository.NewFindByIdMongoRepository(db)
	findCategoryByIdRepository := category_repository.NewFindCategoryByIdRepository(db)

	return transaction.NewUpdateTransactionController(updateTransactionRepository, findTransactionByIdRepository, findMemberByIdRepository, findAccountByIdRepository, findCategoryByIdRepository)
}

func MakeGetTransactionByIdController(workspaceDb *mongo.Database) *transaction.GetTransactionByIdController {
	findTransactionByIdRepository := transaction_repository.NewGetTransactionByIdRepository(workspaceDb)
	return transaction.NewGetTransactionByIdController(findTransactionByIdRepository)
}
