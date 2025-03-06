package transaction_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	presentationHelpers "github.com/anuntech/finance-backend/internal/presentation/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TransactionRepository struct {
	db *mongo.Database
}

func NewTransactionRepository(db *mongo.Database) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Find(filters *presentationHelpers.GlobalFilterParams) ([]models.Transaction, error) {
	collection := r.db.Collection("transaction")

	startOfMonth := time.Date(filters.Year, time.Month(filters.Month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	filter := bson.M{
		"workspace_id": filters.WorkspaceId,
	}
	if filters.Type != "" {
		filter["type"] = filters.Type
	}

	if filters.Month != 0 {
		orRepeatAndRecurringLogic := []bson.M{
			{
				"$and": []bson.M{
					{"is_confirmed": false},
					{"due_date": bson.M{"$lt": endOfMonth}},
				},
			},
			{
				"$and": []bson.M{
					{"is_confirmed": true},
					{"confirmation_date": bson.M{"$lt": endOfMonth}},
				},
			},
		}

		filter["$or"] = []bson.M{
			{
				"frequency": "DO_NOT_REPEAT",
				"$or": []bson.M{
					{
						"$and": []bson.M{
							{"is_confirmed": false},
							{"due_date": bson.M{"$gte": startOfMonth, "$lt": endOfMonth}},
						},
					},
					{
						"$and": []bson.M{
							{"is_confirmed": true},
							{"confirmation_date": bson.M{"$gte": startOfMonth, "$lt": endOfMonth}},
						},
					},
				},
			},
			{
				"frequency": "RECURRING",
				"$or":       orRepeatAndRecurringLogic,
			},
			{
				"frequency": "REPEAT",
				"$or":       orRepeatAndRecurringLogic,
			},
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err = cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}

	// Lógica para filtrar parcelas já passadas e ajustar o initialInstallment
	if filters.Month != 0 {
		transactions = filterRepeatTransactions(transactions, startOfMonth, endOfMonth)
	}

	return transactions, nil
}

func computeInstallmentDueDate(initial time.Time, interval string, offset int) time.Time {
	switch interval {
	// case "DAILY":
	// 	return initial.AddDate(0, 0, offset)
	// case "WEEKLY":
	// 	return initial.AddDate(0, 0, 7*offset)
	case "MONTHLY":
		return initial.AddDate(0, offset, 0)
	case "QUARTERLY":
		return initial.AddDate(0, 3*offset, 0)
	case "YEARLY":
		return initial.AddDate(offset, 0, 0)
	default:
		// Caso o intervalo não seja reconhecido, utiliza-se mensal como padrão
		return initial.AddDate(0, offset, 0)
	}
}

// filterRepeatTransactions percorre todas as transações e, para aquelas com frequência "REPEAT"
// aplica a lógica de parcelas: considera o initialInstallment e ignora as parcelas já passadas.
// Se a parcela para o mês (ou o período escolhido) não existir, a transação é descartada da lista.
func filterRepeatTransactions(transactions []models.Transaction, startOfMonth, endOfMonth time.Time) []models.Transaction {
	var filtered []models.Transaction

	for _, tx := range transactions {
		if tx.Frequency == "REPEAT" && tx.RepeatSettings != nil {
			valid := false

			// Começa da parcela definida pelo initialInstallment e itera até o número total de parcelas.
			for i := int(tx.RepeatSettings.InitialInstallment); i <= tx.RepeatSettings.Count; i++ {
				// Como a primeira parcela usa o DueDate original,
				// usamos (i-1) como offset para calcular a data da parcela.
				var dateRef time.Time
				if tx.IsConfirmed {
					dateRef = *tx.ConfirmationDate
				} else {
					dateRef = tx.DueDate
				}
				installmentDueDate := computeInstallmentDueDate(dateRef, tx.RepeatSettings.Interval, i-1)

				// Verifica se o vencimento da parcela está dentro do período desejado.
				if !installmentDueDate.Before(startOfMonth) && installmentDueDate.Before(endOfMonth) {
					// Atualiza a transação para exibir apenas a parcela atual...
					tx.DueDate = installmentDueDate
					// E ajusta a quantidade de parcelas restantes (por exemplo, se eram 3 e estamos na 2ª, então resta 2 parcelas)
					tx.RepeatSettings.CurrentCount = i
					valid = true
					break
				}
			}

			// Se nenhuma parcela se encaixar no período, a transação não deverá ser exibida.
			if !valid {
				continue
			}
		}

		filtered = append(filtered, tx)
	}

	return filtered
}
