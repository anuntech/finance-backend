package factory

import (
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/account_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/bank_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/category_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/custom_field_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/edit_transaction_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/transaction_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/workspace_repository/member_repository"
	workspace_user_repository "github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/workspace_repository/user_repository"
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

func MakeGetTransactionController(workspaceDb *mongo.Database, db *mongo.Database) *transaction.GetTransactionController {
	findTransactionsByWorkspaceIdAndMonthRepository := transaction_repository.NewTransactionRepository(
		db,
		edit_transaction_repository.NewFindByIdEditTransactionRepository(db),
	)
	findByIdEditTransactionRepository := edit_transaction_repository.NewFindByIdEditTransactionRepository(db)
	findCustomFieldByIdRepository := custom_field_repository.NewFindCustomFieldByIdRepository(db)
	findCategoryByIdRepository := category_repository.NewFindCategoryByIdRepository(db)
	findWorkspaceUserByIdRepository := workspace_user_repository.NewFindWorkspaceUserByIdRepository(workspaceDb)

	return transaction.NewGetTransactionController(
		findTransactionsByWorkspaceIdAndMonthRepository,
		findByIdEditTransactionRepository,
		findCustomFieldByIdRepository,
		findCategoryByIdRepository,
		*findWorkspaceUserByIdRepository,
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
	findTransactionByIdRepository := transaction_repository.NewGetTransactionByIdRepository(db)

	return transaction.NewDeleteTransactionController(
		deleteTransactionRepository,
		findTransactionByIdRepository,
	)
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

func MakeImportTransactionController(workspaceDb *mongo.Database, db *mongo.Database) *transaction.ImportTransactionController {
	findMemberByIdRepository := member_repository.NewFindMemberByIdRepository(workspaceDb)
	createTransactionRepository := transaction_repository.NewCreateTransactionRepository(db)
	findAccountByIdRepository := account_repository.NewFindByIdMongoRepository(db)
	findCategoryByIdRepository := category_repository.NewFindCategoryByIdRepository(db)
	findCustomFieldByIdRepository := custom_field_repository.NewFindCustomFieldByIdRepository(db)

	findAccountByNameRepository := account_repository.NewFindByNameMongoRepository(db)
	findCategoryByNameAndTypeRepository := category_repository.NewFindByNameAndTypeMongoRepository(db)
	findMemberByEmailRepository := member_repository.NewFindMemberByEmailRepository(workspaceDb)
	findCustomFieldByNameRepository := custom_field_repository.NewFindCustomFieldByNameRepository(db)

	createAccountRepository := account_repository.NewCreateAccountMongoRepository(db)
	createCategoryRepository := category_repository.NewCreateCategoryRepository(db)
	findBankByNameRepository := bank_repository.NewFindByNameMongoRepository(db)
	return transaction.NewImportTransactionController(
		findMemberByIdRepository,
		createTransactionRepository,
		findAccountByIdRepository,
		findCategoryByIdRepository,
		findCustomFieldByIdRepository,
		findAccountByNameRepository,
		findCategoryByNameAndTypeRepository,
		findMemberByEmailRepository,
		findCustomFieldByNameRepository,
		createAccountRepository,
		createCategoryRepository,
		findBankByNameRepository,
	)
}

func MakeUpdateManyTransactionController(db *mongo.Database) *transaction.UpdateManyTransactionController {
	findTransactionByIdRepository := transaction_repository.NewGetTransactionByIdRepository(db)
	findByIdEditTransactionRepository := edit_transaction_repository.NewFindByIdEditTransactionRepository(db)
	updateTransactionRepository := transaction_repository.NewUpdateTransactionRepository(db)
	createEditTransactionRepository := edit_transaction_repository.NewCreateEditTransactionRepository(db)
	findCustomFieldByIdRepository := custom_field_repository.NewFindCustomFieldByIdRepository(db)

	return transaction.NewUpdateManyTransactionController(
		findTransactionByIdRepository,
		findByIdEditTransactionRepository,
		updateTransactionRepository,
		createEditTransactionRepository,
		findCustomFieldByIdRepository,
	)
}
