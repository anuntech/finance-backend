package factory

import (
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/account_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/category_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/custom_field_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/edit_transaction_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/transaction_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/workspace_repository/member_repository"
	"github.com/anuntech/finance-backend/internal/presentation/controllers/edit_transaction"
	"github.com/anuntech/finance-backend/internal/presentation/controllers/transaction"
	"go.mongodb.org/mongo-driver/mongo"
)

func MakeCreateTransactionController(workspaceDb *mongo.Database, db *mongo.Database) *transaction.CreateTransactionController {
	findMemberByIdRepository := member_repository.NewFindMemberByIdRepository(workspaceDb)
	createTransactionRepository := transaction_repository.NewCreateTransactionRepository(db)
	findAccountByIdRepository := account_repository.NewFindByIdMongoRepository(db)
	findCategoryByIdRepository := category_repository.NewFindCategoryByIdRepository(db)
	findCustomFieldByIdRepository := custom_field_repository.NewFindCustomFieldByIdRepository(db)

	return transaction.NewCreateTransactionController(
		findMemberByIdRepository,
		createTransactionRepository,
		findAccountByIdRepository,
		findCategoryByIdRepository,
		findCustomFieldByIdRepository,
	)
}

func MakeGetTransactionController(db *mongo.Database) *transaction.GetTransactionController {
	findTransactionsByWorkspaceIdAndMonthRepository := transaction_repository.NewTransactionRepository(db)
	findByIdEditTransactionRepository := edit_transaction_repository.NewFindByIdEditTransactionRepository(db)
	findCustomFieldByIdRepository := custom_field_repository.NewFindCustomFieldByIdRepository(db)

	return transaction.NewGetTransactionController(
		findTransactionsByWorkspaceIdAndMonthRepository,
		findByIdEditTransactionRepository,
		findCustomFieldByIdRepository,
	)
}

func MakeUpdateTransactionController(workspaceDb *mongo.Database, db *mongo.Database) *transaction.UpdateTransactionController {
	findTransactionByIdRepository := transaction_repository.NewGetTransactionByIdRepository(db)
	updateTransactionRepository := transaction_repository.NewUpdateTransactionRepository(db)
	findMemberByIdRepository := member_repository.NewFindMemberByIdRepository(workspaceDb)
	findAccountByIdRepository := account_repository.NewFindByIdMongoRepository(db)
	findCategoryByIdRepository := category_repository.NewFindCategoryByIdRepository(db)
	findCustomFieldByIdRepository := custom_field_repository.NewFindCustomFieldByIdRepository(db)

	return transaction.NewUpdateTransactionController(
		updateTransactionRepository,
		findTransactionByIdRepository,
		findMemberByIdRepository,
		findAccountByIdRepository,
		findCategoryByIdRepository,
		findCustomFieldByIdRepository,
	)
}

func MakeGetTransactionByIdController(workspaceDb *mongo.Database) *transaction.GetTransactionByIdController {
	findTransactionByIdRepository := transaction_repository.NewGetTransactionByIdRepository(workspaceDb)

	return transaction.NewGetTransactionByIdController(findTransactionByIdRepository)
}

func MakeDeleteTransactionController(db *mongo.Database) *transaction.DeleteTransactionController {
	deleteTransactionRepository := transaction_repository.NewDeleteTransactionRepository(db)

	return transaction.NewDeleteTransactionController(deleteTransactionRepository)
}

func MakeCreateEditTransactionController(workspaceDb *mongo.Database, db *mongo.Database) *edit_transaction.CreateEditTransactionController {
	findTransactionByIdRepository := transaction_repository.NewGetTransactionByIdRepository(db)
	createEditTransactionRepository := edit_transaction_repository.NewCreateEditTransactionRepository(db)
	findMemberByIdRepository := member_repository.NewFindMemberByIdRepository(workspaceDb)
	findAccountByIdRepository := account_repository.NewFindByIdMongoRepository(db)
	findCategoryByIdRepository := category_repository.NewFindCategoryByIdRepository(db)
	findByIdEditTransactionRepository := edit_transaction_repository.NewFindByIdEditTransactionRepository(db)
	updateEditTransactionRepository := edit_transaction_repository.NewUpdateEditTransactionRepository(db)
	findCustomFieldByIdRepository := custom_field_repository.NewFindCustomFieldByIdRepository(db)

	return edit_transaction.NewCreateEditTransactionController(
		findMemberByIdRepository,
		createEditTransactionRepository,
		findAccountByIdRepository,
		findCategoryByIdRepository,
		findTransactionByIdRepository,
		findByIdEditTransactionRepository,
		updateEditTransactionRepository,
		findCustomFieldByIdRepository,
	)
}
