package factory

import (
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/transaction_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/workspace_repository/member_repository"
	"github.com/anuntech/finance-backend/internal/presentation/controllers/transaction"
	"go.mongodb.org/mongo-driver/mongo"
)

func MakeCreateTransactionController(workspaceDb *mongo.Database, db *mongo.Database) *transaction.CreateTransactionController {
	findMemberByIdRepository := member_repository.NewFindMemberByIdRepository(workspaceDb)
	createTransactionRepository := transaction_repository.NewCreateTransactionRepository(db)
	return transaction.NewCreateTransactionController(findMemberByIdRepository, createTransactionRepository)
}
