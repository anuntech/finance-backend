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
	transactions = r.filterRepeatTransactions(transactions, startOfMonth, endOfMonth)

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

func (r *TransactionRepository) computeInstallmentDueDate(initial time.Time, interval string, offset int) time.Time {
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
func (r *TransactionRepository) filterRepeatTransactions(transactions []models.Transaction, startOfMonth, endOfMonth time.Time) []models.Transaction {
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
			valid := false

			for i := int(tx.RepeatSettings.InitialInstallment); i <= tx.RepeatSettings.Count; i++ {
				// Como a primeira parcela usa o DueDate original,
				// usamos (i-1) como offset para calcular a data da parcela.

				installmentDueDate := r.computeInstallmentDueDate(dateRef, tx.RepeatSettings.Interval, i-1)

				// Verifica se o vencimento da parcela está dentro do período desejado.
				if !installmentDueDate.Before(startOfMonth) && installmentDueDate.Before(endOfMonth) {
					// Armazena a data de registro original para manter o dia
					originalRegHour, originalRegMin, originalRegSec := tx.RegistrationDate.Clock()
					originalRegDay := tx.RegistrationDate.Day()

					// Atualiza a transação para exibir apenas a parcela atual...
					tx.DueDate = installmentDueDate

					// Sincroniza o RegistrationDate com base no mês atual
					newRegDate := time.Date(
						startOfMonth.Year(), startOfMonth.Month(), originalRegDay,
						originalRegHour, originalRegMin, originalRegSec, 0, startOfMonth.Location(),
					)

					// Certifica-se de que o dia existe no mês atual (por exemplo, 31 de fevereiro não existe)
					if originalRegDay > daysInMonth(startOfMonth) {
						newRegDate = time.Date(
							startOfMonth.Year(), startOfMonth.Month(), daysInMonth(startOfMonth),
							originalRegHour, originalRegMin, originalRegSec, 0, startOfMonth.Location(),
						)
					}

					tx.RegistrationDate = newRegDate

					// Sincroniza o ConfirmationDate se a transação estiver confirmada
					if tx.IsConfirmed && tx.ConfirmationDate != nil {
						// Mantém a mesma hora do dia do ConfirmationDate original
						origHour, origMin, origSec := tx.ConfirmationDate.Clock()
						newConfDate := installmentDueDate
						newConfDate = time.Date(
							newConfDate.Year(), newConfDate.Month(), newConfDate.Day(),
							origHour, origMin, origSec, 0, newConfDate.Location(),
						)
						tx.ConfirmationDate = &newConfDate
					}

					// E ajusta a quantidade de parcelas restantes (por exemplo, se eram 3 e estamos na 2ª, então resta 2 parcelas)
					tx.RepeatSettings.CurrentCount = i
					balance := tx.Balance.Value
					tx.Balance.Value = balance / float64(tx.RepeatSettings.Count)
					tx.TotalBalance = balance
					valid = true
					break
				}
			}

			// Se nenhuma parcela se encaixar no período, a transação não deverá ser exibida.
			if !valid {
				continue
			}
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

				// Mantém o mesmo dia do mês da transação original
				originalDay := dateRef.Day()
				if originalDay > daysInMonth(currentDate) {
					originalDay = daysInMonth(currentDate)
				}

				// Atualiza o DueDate para este mês
				newDueDate := time.Date(
					currentDate.Year(), currentDate.Month(), originalDay,
					dateRef.Hour(), dateRef.Minute(), dateRef.Second(), 0, dateRef.Location(),
				)

				// Verifica se está dentro do intervalo
				if !newDueDate.Before(startOfMonth) && newDueDate.Before(endOfMonth) {
					txCopy.DueDate = newDueDate

					// Atualiza a contagem atual
					if txCopy.RepeatSettings == nil {
						txCopy.RepeatSettings = &models.TransactionRepeatSettings{
							Interval: "MONTHLY",
						}
					}

					// Atualiza o CurrentCount para refletir a ordem das parcelas no período
					txCopy.RepeatSettings.CurrentCount = installmentCounter
					installmentCounter++ // Incrementa para a próxima parcela
					// Atualiza o RegistrationDate
					originalRegHour, originalRegMin, originalRegSec := tx.RegistrationDate.Clock()
					originalRegDay := tx.RegistrationDate.Day()
					if originalRegDay > daysInMonth(currentDate) {
						originalRegDay = daysInMonth(currentDate)
					}

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
