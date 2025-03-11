package helpers

import (
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
)

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
		// Para transações com repetição fixa, procuramos qual parcela cai no mês desejado
		for i := int(transaction.RepeatSettings.InitialInstallment); i <= transaction.RepeatSettings.Count; i++ {
			installmentDueDate := computeInstallmentDueDate(dateRef, transaction.RepeatSettings.Interval, i-1)

			// Verificamos se esta parcela cai no mês/período filtrado
			if !installmentDueDate.Before(targetDate) && installmentDueDate.Before(endOfMonth) {
				return i
			}
		}
		return 0 // Nenhuma parcela encontrada para o mês

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
	default:
		// Caso o intervalo não seja reconhecido, utiliza-se mensal como padrão
		return initial.AddDate(0, offset, 0)
	}
}
