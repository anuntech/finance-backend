package factory

import (
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/workspace_repository/member_repository"
	"github.com/anuntech/finance-backend/internal/presentation/controllers/transaction"
	"go.mongodb.org/mongo-driver/mongo"
)

func MakeCreateTransactionController(db *mongo.Database) *transaction.CreateTransactionController {
	findMemberByIdRepository := member_repository.NewFindMemberByIdRepository(db)
	return transaction.NewCreateTransactionController(findMemberByIdRepository)
}
