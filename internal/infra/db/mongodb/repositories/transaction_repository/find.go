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

	var startOfMonth, endOfMonth time.Time

	if filters.Month != 0 {
		startOfMonth = time.Date(filters.Year, time.Month(filters.Month), 1, 0, 0, 0, 0, time.UTC)
		endOfMonth = startOfMonth.AddDate(0, 1, 0).Add(-time.Second)
	}

	if filters.InitialDate != "" && filters.FinalDate != "" {
		startDate, err := time.Parse("2006-01-02", filters.InitialDate)
		if err != nil {
			return nil, err
		}
		endDate, err := time.Parse("2006-01-02", filters.FinalDate)
		if err != nil {
			return nil, err
		}
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())

		startOfMonth = startDate
		endOfMonth = endDate
	}

	filter := bson.M{
		"workspace_id": filters.WorkspaceId,
	}
	if filters.Type != "" {
		filter["type"] = filters.Type
	}

	// filter["$or"] = r.createNormalFilter(startOfMonth, endOfMonth)

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
	transactions = r.applyRepeatAndRecurringLogicTransactions(transactions, startOfMonth, endOfMonth)

	return transactions, nil
}

// func (r *TransactionRepository) createNormalFilter(startOfMonth, endOfMonth time.Time) []bson.M {
// 	orRepeatAndRecurringLogic := []bson.M{
// 		{
// 			"$and": []bson.M{
// 				{"is_confirmed": false},
// 				{"due_date": bson.M{"$lt": endOfMonth}},
// 			},
// 		},
// 		{
// 			"$and": []bson.M{
// 				{"is_confirmed": true},
// 				{"confirmation_date": bson.M{"$lt": endOfMonth}},
// 			},
// 		},
// 	}

// 	filter := []bson.M{
// 		{
// 			"frequency": "DO_NOT_REPEAT",
// 			"$or": []bson.M{
// 				{
// 					"$and": []bson.M{
// 						{"is_confirmed": false},
// 						{"due_date": bson.M{"$gte": startOfMonth, "$lt": endOfMonth}},
// 					},
// 				},
// 				{
// 					"$and": []bson.M{
// 						{"is_confirmed": true},
// 						{"confirmation_date": bson.M{"$gte": startOfMonth, "$lt": endOfMonth}},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			"frequency": "RECURRING",
// 			"$or":       orRepeatAndRecurringLogic,
// 		},
// 		{
// 			"frequency": "REPEAT",
// 			"$or":       orRepeatAndRecurringLogic,
// 		},
// 	}

// 	return filter
// }

func (r *TransactionRepository) computeInstallmentDueDate(initial time.Time, interval string, offset int, customDay ...int) time.Time {
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
	case "CUSTOM":
		// Se temos o customDay como parâmetro, usamos ele
		if len(customDay) > 0 && customDay[0] > 0 {
			return initial.AddDate(0, 0, customDay[0]*offset)
		}
		// Fallback para intervalo mensal
		return initial.AddDate(0, offset, 0)
	default:
		// Caso o intervalo não seja reconhecido, utiliza-se mensal como padrão
		return initial.AddDate(0, offset, 0)
	}
}

// filterRepeatTransactions percorre todas as transações e, para aquelas com frequência "REPEAT"
// aplica a lógica de parcelas: considera o initialInstallment e ignora as parcelas já passadas.
// Se a parcela para o mês (ou o período escolhido) não existir, a transação é descartada da lista.
func (r *TransactionRepository) applyRepeatAndRecurringLogicTransactions(transactions []models.Transaction, startOfMonth, endOfMonth time.Time) []models.Transaction {
	var filtered []models.Transaction

	for _, tx := range transactions {
		var dateRef time.Time
		if tx.IsConfirmed {
			dateRef = *tx.ConfirmationDate
		} else {
			dateRef = tx.DueDate
		}

		switch tx.Frequency {
		case "REPEAT":
			// Create multiple instances like the RECURRING case
			var txInstances []models.Transaction

			installmentCounter := 1
			for i := int(tx.RepeatSettings.InitialInstallment); i <= tx.RepeatSettings.Count; i++ {
				// Calculate the due date for this installment
				var installmentDueDate time.Time

				// Se for intervalo CUSTOM, passa o CustomDay
				if tx.RepeatSettings.Interval == "CUSTOM" {
					installmentDueDate = r.computeInstallmentDueDate(dateRef, tx.RepeatSettings.Interval, i-1, tx.RepeatSettings.CustomDay)
				} else {
					installmentDueDate = r.computeInstallmentDueDate(dateRef, tx.RepeatSettings.Interval, i-1)
				}

				// Check if this installment is within the date range
				if (!installmentDueDate.Before(startOfMonth) && installmentDueDate.Before(endOfMonth)) || startOfMonth.IsZero() || endOfMonth.IsZero() {
					// Create a copy of the transaction for this installment
					txCopy := tx

					// Make a deep copy of RepeatSettings
					if tx.RepeatSettings != nil {
						repeatSettingsCopy := *tx.RepeatSettings
						txCopy.RepeatSettings = &repeatSettingsCopy
					}

					// Update the due date
					txCopy.DueDate = installmentDueDate

					// Update the current count (installment number)
					txCopy.RepeatSettings.CurrentCount = installmentCounter
					installmentCounter++

					// Update the balance (divide by total installments)
					balance := tx.Balance.Value
					txCopy.Balance.Value = balance / float64(tx.RepeatSettings.Count)
					txCopy.TotalBalance = balance

					// Sync the registration date
					originalRegHour, originalRegMin, originalRegSec := tx.RegistrationDate.Clock()
					originalRegDay := tx.RegistrationDate.Day()

					// Ensure the day exists in the month using min
					originalRegDay = min(originalRegDay, daysInMonth(installmentDueDate))

					newRegDate := time.Date(
						installmentDueDate.Year(), installmentDueDate.Month(), originalRegDay,
						originalRegHour, originalRegMin, originalRegSec, 0, installmentDueDate.Location(),
					)
					txCopy.RegistrationDate = newRegDate

					// Sync the confirmation date if the transaction is confirmed
					if txCopy.IsConfirmed && txCopy.ConfirmationDate != nil {
						origHour, origMin, origSec := txCopy.ConfirmationDate.Clock()
						newConfDate := time.Date(
							installmentDueDate.Year(), installmentDueDate.Month(), installmentDueDate.Day(),
							origHour, origMin, origSec, 0, installmentDueDate.Location(),
						)
						txCopy.ConfirmationDate = &newConfDate
					}

					txInstances = append(txInstances, txCopy)
				}
			}

			if len(txInstances) > 0 {
				filtered = append(filtered, txInstances...)
				continue
			}

			// Skip this transaction if no instances were found
			continue
		case "RECURRING":
			if tx.RepeatSettings == nil {
				tx.RepeatSettings = &models.TransactionRepeatSettings{
					Interval: "MONTHLY", // Assume monthly as default interval
				}
			}

			// Para transações recorrentes, precisamos criar uma instância para cada mês no intervalo
			// Calculamos o intervalo entre os meses de início e fim
			startYear, startMonth, _ := startOfMonth.Date()
			endYear, endMonth, _ := endOfMonth.Date()

			// Número total de meses no intervalo
			totalMonths := (endYear-startYear)*12 + int(endMonth-startMonth)
			if totalMonths < 0 {
				totalMonths = 0
			}

			// Referência para a data original da transação
			refYear, refMonth, _ := dateRef.Date()

			// Para cada mês no intervalo, crie uma cópia da transação
			var txInstances []models.Transaction

			// Contador para manter o controle da exibição sequencial (1, 2, 3, ...)
			installmentCounter := 1

			for monthOffset := 0; monthOffset <= totalMonths; monthOffset++ {
				// Calcula a data para esta instância
				currentDate := time.Date(startYear, startMonth, 1, 0, 0, 0, 0, time.UTC)
				currentDate = currentDate.AddDate(0, monthOffset, 0)

				// Verifica se a data é posterior à data original da transação
				currentYear, currentMonth, _ := currentDate.Date()
				monthsSinceOriginal := (currentYear-refYear)*12 + int(currentMonth-refMonth)

				if monthsSinceOriginal < 0 {
					continue // Pula meses anteriores à data original
				}

				// Cria uma cópia da transação para este mês
				txCopy := tx

				// Cria uma cópia profunda do RepeatSettings
				if tx.RepeatSettings != nil {
					repeatSettingsCopy := *tx.RepeatSettings
					txCopy.RepeatSettings = &repeatSettingsCopy
				}

				// Mantém o mesmo dia do mês da transação original
				originalDay := dateRef.Day()
				originalDay = min(originalDay, daysInMonth(currentDate))

				// Atualiza o DueDate para este mês
				newDueDate := time.Date(
					currentDate.Year(), currentDate.Month(), originalDay,
					dateRef.Hour(), dateRef.Minute(), dateRef.Second(), 0, dateRef.Location(),
				)

				// Verifica se está dentro do intervalo
				if (!newDueDate.Before(startOfMonth) && newDueDate.Before(endOfMonth)) || startOfMonth.IsZero() || endOfMonth.IsZero() {
					txCopy.DueDate = newDueDate

					// Atualiza a contagem atual
					txCopy.RepeatSettings.CurrentCount = installmentCounter
					installmentCounter++ // Incrementa para a próxima parcela
					// Atualiza o RegistrationDate
					originalRegHour, originalRegMin, originalRegSec := tx.RegistrationDate.Clock()
					originalRegDay := tx.RegistrationDate.Day()
					originalRegDay = min(originalRegDay, daysInMonth(currentDate))

					newRegDate := time.Date(
						currentDate.Year(), currentDate.Month(), originalRegDay,
						originalRegHour, originalRegMin, originalRegSec, 0, currentDate.Location(),
					)
					txCopy.RegistrationDate = newRegDate

					// Sincroniza o ConfirmationDate se a transação estiver confirmada
					if txCopy.IsConfirmed && txCopy.ConfirmationDate != nil {
						origHour, origMin, origSec := txCopy.ConfirmationDate.Clock()
						newConfDate := time.Date(
							newDueDate.Year(), newDueDate.Month(), newDueDate.Day(),
							origHour, origMin, origSec, 0, newDueDate.Location(),
						)
						txCopy.ConfirmationDate = &newConfDate
					}

					txInstances = append(txInstances, txCopy)
				}
			}

			// Se encontrou instâncias, adiciona-as aos resultados filtrados
			if len(txInstances) > 0 {
				filtered = append(filtered, txInstances...)
				continue
			}

			// Se não houver instâncias, pulamos esta transação
			continue
		}

		filtered = append(filtered, tx)
	}

	return filtered
}

// Função auxiliar para obter o número de dias em um mês
func daysInMonth(date time.Time) int {
	year, month, _ := date.Date()
	// Pegar o primeiro dia do próximo mês e subtrair 1 dia
	return time.Date(year, month+1, 0, 0, 0, 0, 0, date.Location()).Day()
}
