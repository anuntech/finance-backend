package transaction_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TransactionRepository struct {
	db                                *mongo.Database
	FindByIdEditTransactionRepository usecase.FindByIdEditTransactionRepository
}

func NewTransactionRepository(db *mongo.Database, findByIdEditTransactionRepository usecase.FindByIdEditTransactionRepository) *TransactionRepository {
	return &TransactionRepository{db: db, FindByIdEditTransactionRepository: findByIdEditTransactionRepository}
}

func (r *TransactionRepository) Find(filters *usecase.FindTransactionsByWorkspaceIdInputRepository) ([]models.Transaction, error) {
	collection := r.db.Collection("transaction")

	var startOfMonth, endOfMonth time.Time

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
	} else {
		// for default use a high end month
		endOfMonth = time.Date(2035, 12, 31, 23, 59, 59, 0, time.Local)
	}

	filter := bson.M{
		"workspace_id": filters.WorkspaceId,
	}

	if filters.Type != "" {
		filter["type"] = filters.Type
	}

	if len(filters.AccountIds) > 0 {
		filter["account_id"] = bson.M{"$in": filters.AccountIds}
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

	transactions = r.applyRepeatAndRecurringLogicTransactions(transactions, startOfMonth, endOfMonth)
	transactions, err = r.replaceTransactionIfEditRepeat(transactions)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

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
		if tx.IsConfirmed && tx.ConfirmationDate != nil {
			dateRef = *tx.ConfirmationDate
		} else {
			dateRef = tx.DueDate
		}

		switch tx.Frequency {
		case "REPEAT":
			if tx.RepeatSettings == nil {
				continue
			}
			var txInstances []models.Transaction

			// Calcular a parcela de base - esta será usada para iniciar o contador
			baseInstallment := int(tx.RepeatSettings.InitialInstallment)

			// Ajusta o contador para começar do número correto da parcela
			// com base no mês de início da busca
			if !startOfMonth.IsZero() {
				for i := int(tx.RepeatSettings.InitialInstallment); i <= tx.RepeatSettings.Count; i++ {
					var installmentDueDate time.Time

					if tx.RepeatSettings.Interval == "CUSTOM" {
						installmentDueDate = r.computeInstallmentDueDate(dateRef, tx.RepeatSettings.Interval, i-1, tx.RepeatSettings.CustomDay)
					} else {
						installmentDueDate = r.computeInstallmentDueDate(dateRef, tx.RepeatSettings.Interval, i-1)
					}

					// Se esta data for anterior ao mês de início da busca, incrementamos a base
					if installmentDueDate.Before(startOfMonth) {
						baseInstallment = i + 1
					}
				}
			}

			// Inicializa o contador a partir da parcela de base
			installmentCounter := baseInstallment

			// Agora itera sobre as parcelas como antes
			for i := int(tx.RepeatSettings.InitialInstallment); i <= tx.RepeatSettings.Count; i++ {
				var installmentDueDate time.Time

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

			// Skip confirmed transactions without a confirmation date
			if tx.IsConfirmed && tx.ConfirmationDate == nil {
				continue
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

func (r *TransactionRepository) replaceTransactionIfEditRepeat(transactions []models.Transaction) ([]models.Transaction, error) {
	// Skip processing if there are no transactions
	if len(transactions) == 0 {
		return transactions, nil
	}

	// Prepare params for batch query
	var queryParams []struct {
		MainId      primitive.ObjectID
		MainCount   int
		WorkspaceId primitive.ObjectID
	}

	// Map to track transaction positions for quick access
	transactionMap := make(map[string]int) // key: "mainId_mainCount_workspaceId"

	// Collect all transactions to look up
	for i, tx := range transactions {
		if tx.RepeatSettings == nil {
			tx.RepeatSettings = &models.TransactionRepeatSettings{}
			continue
		}

		param := struct {
			MainId      primitive.ObjectID
			MainCount   int
			WorkspaceId primitive.ObjectID
		}{
			MainId:      tx.Id,
			MainCount:   tx.RepeatSettings.CurrentCount,
			WorkspaceId: tx.WorkspaceId,
		}

		queryParams = append(queryParams, param)

		// Create a unique key for this transaction in our map
		key := tx.Id.Hex() + "_" + string(rune(tx.RepeatSettings.CurrentCount)) + "_" + tx.WorkspaceId.Hex()
		transactionMap[key] = i
	}

	// Fetch all edited transactions with a single database call
	editedTransactions, err := r.FindByIdEditTransactionRepository.FindMany(queryParams)
	if err != nil {
		return nil, err
	}

	// Apply edited transactions
	for _, editTx := range editedTransactions {
		if editTx.MainId == nil || editTx.MainCount == nil {
			continue
		}

		// Create lookup key
		key := editTx.MainId.Hex() + "_" + string(rune(*editTx.MainCount)) + "_" + editTx.WorkspaceId.Hex()
		idx, exists := transactionMap[key]

		if exists && *editTx.MainCount == transactions[idx].RepeatSettings.CurrentCount {
			// Store original values we need to preserve
			repeatSettings := *transactions[idx].RepeatSettings
			frequency := transactions[idx].Frequency
			totalBalance := transactions[idx].TotalBalance
			balance := transactions[idx].Balance
			id := transactions[idx].Id

			// Preserve the installment number/current count
			installmentNumber := repeatSettings.CurrentCount

			// Replace with edited transaction
			transactions[idx] = *editTx
			transactions[idx].Frequency = frequency
			transactions[idx].RepeatSettings = &repeatSettings

			// Restore the installment number
			transactions[idx].RepeatSettings.CurrentCount = installmentNumber

			transactions[idx].Id = id
			transactions[idx].MainCount = nil
			transactions[idx].MainId = nil

			if transactions[idx].Frequency == "DO_NOT_REPEAT" {
				transactions[idx].Balance = balance
			}

			transactions[idx].TotalBalance = totalBalance
		}
	}

	// Calculate balances for all transactions
	for i := range transactions {
		transactionCopy := transactions[i]
		transactionCopy.Type = "RECIPE"
		calc := helpers.CalculateOneTransactionBalance(&transactionCopy)
		transactions[i].Balance.NetBalance = calc
	}

	// Filter out deleted transactions
	filteredTransactions := transactions[:0]
	for _, tx := range transactions {
		if !tx.IsDeleted {
			filteredTransactions = append(filteredTransactions, tx)
		}
	}

	return filteredTransactions, nil
}
