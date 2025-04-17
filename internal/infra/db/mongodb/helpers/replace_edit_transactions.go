package helpers

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	presentationHelpers "github.com/anuntech/finance-backend/internal/presentation/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func ReplaceEditTransactions(transactions []models.Transaction, db *mongo.Database, globalFilters *presentationHelpers.GlobalFilterParams) ([]models.Transaction, error) {
	if db == nil {
		return transactions, nil
	}

	editCollection := db.Collection("edit_transaction")

	for i, transaction := range transactions {
		if transaction.RepeatSettings != nil {
			transaction.RepeatSettings.CurrentCount = CalculateCurrentCount(&transaction, globalFilters.Year, globalFilters.Month)
			transactions[i].RepeatSettings.CurrentCount = transaction.RepeatSettings.CurrentCount
		}

		result := editCollection.FindOne(context.Background(), bson.M{"main_id": transaction.Id, "workspace_id": transaction.WorkspaceId})

		if result.Err() == mongo.ErrNoDocuments {
			continue
		}

		if result.Err() != nil {
			return nil, result.Err()
		}

		var editTransaction models.Transaction
		if err := result.Decode(&editTransaction); err != nil {
			return nil, err
		}

		if transaction.RepeatSettings != nil {

			if editTransaction.MainCount != nil && transaction.RepeatSettings.CurrentCount == *editTransaction.MainCount {
				transactions[i].Balance = editTransaction.Balance
			}
		}
	}

	return transactions, nil
}

// CalculateCurrentCount calcula em qual parcela (count) a transação está com base no intervalo da repetição e nas datas
func CalculateCurrentCount(transaction *models.Transaction, year int, month int) int {
	if transaction.RepeatSettings == nil {
		return 0
	}

	targetDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := targetDate.AddDate(0, 1, 0).Add(-time.Second)

	var dateRef time.Time
	if transaction.IsConfirmed && transaction.ConfirmationDate != nil {
		dateRef = *transaction.ConfirmationDate
	} else {
		dateRef = transaction.DueDate
	}

	switch transaction.Frequency {
	case "REPEAT":
		// Determinar qual parcela vem antes do mês alvo para encontrar a base
		baseInstallment := int(transaction.RepeatSettings.InitialInstallment)

		for i := int(transaction.RepeatSettings.InitialInstallment); i <= transaction.RepeatSettings.Count; i++ {
			var installmentDueDate time.Time

			// Para o caso CUSTOM, aplicamos o CustomDay direto aqui
			if transaction.RepeatSettings.Interval == "CUSTOM" && transaction.RepeatSettings.CustomDay > 0 {
				installmentDueDate = dateRef.AddDate(0, 0, transaction.RepeatSettings.CustomDay*(i-1))
			} else {
				installmentDueDate = computeInstallmentDueDate(dateRef, transaction.RepeatSettings.Interval, i-1)
			}

			// Se esta instalação vem antes do mês alvo, atualizamos a base
			if installmentDueDate.Before(targetDate) {
				baseInstallment = i
			}

			// Verificamos se esta parcela cai no mês/período filtrado
			if !installmentDueDate.Before(targetDate) && installmentDueDate.Before(endOfMonth) {
				return i
			}
		}

		// Se não encontramos uma parcela exata neste mês, retornamos a base + 1
		// Isso mantém a sequência consistente
		if baseInstallment < transaction.RepeatSettings.Count {
			return baseInstallment + 1
		}

		return baseInstallment

	case "RECURRING":
		// Para transações recorrentes, calculamos o número de meses entre a data de referência e a data alvo
		refYear, refMonth, _ := dateRef.Date()
		months := (year-refYear)*12 + (month - int(refMonth))

		if months < 0 {
			return 0
		}

		// Incrementa em 1 porque a primeira parcela é considerada 1, não 0
		return months + 1

	default:
		return 0
	}
}

// Adicionando a função computeInstallmentDueDate para manter a consistência
func computeInstallmentDueDate(initial time.Time, interval string, offset int) time.Time {
	switch interval {
	case "DAILY":
		return initial.AddDate(0, 0, offset)
	case "WEEKLY":
		return initial.AddDate(0, 0, 7*offset)
	case "MONTHLY":
		return initial.AddDate(0, offset, 0)
	case "QUARTERLY":
		return initial.AddDate(0, 3*offset, 0)
	case "YEARLY":
		return initial.AddDate(offset, 0, 0)
	case "CUSTOM":
		// Para CUSTOM no helper, usamos mensal como padrão
		// (O valor real será calculado no repository que tem acesso ao CustomDay)
		return initial.AddDate(0, offset, 0)
	default:
		// Caso o intervalo não seja reconhecido, utiliza-se mensal como padrão
		return initial.AddDate(0, offset, 0)
	}
}

// Função auxiliar para obter o número de dias em um mês
func daysInMonth(date time.Time) int {
	year, month, _ := date.Date()
	// Pegar o primeiro dia do próximo mês e subtrair 1 dia
	return time.Date(year, month+1, 0, 0, 0, 0, 0, date.Location()).Day()
}

// Função auxiliar para obter o mínimo entre dois inteiros
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
