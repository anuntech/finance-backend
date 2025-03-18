package transaction

import (
	"errors"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GetTransactionController struct {
	FindTransactionsByWorkspaceIdAndMonthRepository usecase.FindTransactionsByWorkspaceIdRepository
	Validator                                       *validator.Validate
	FindByIdEditTransactionRepository               usecase.FindByIdEditTransactionRepository
	FindCustomFieldByIdRepository                   usecase.FindCustomFieldByIdRepository
}

func NewGetTransactionController(findManyByUserIdAndWorkspaceId usecase.FindTransactionsByWorkspaceIdRepository, findByIdEditTransactionRepository usecase.FindByIdEditTransactionRepository, findCustomFieldByIdRepository usecase.FindCustomFieldByIdRepository) *GetTransactionController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &GetTransactionController{
		FindTransactionsByWorkspaceIdAndMonthRepository: findManyByUserIdAndWorkspaceId,
		Validator:                         validate,
		FindByIdEditTransactionRepository: findByIdEditTransactionRepository,
		FindCustomFieldByIdRepository:     findCustomFieldByIdRepository,
	}
}

func (c *GetTransactionController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspaceId format",
		}, http.StatusBadRequest)
	}

	globalFilters, errHttp := helpers.GetGlobalFilterByQueries(&r.UrlParams, workspaceId, c.Validator)
	if errHttp != nil {
		return errHttp
	}

	transactions, err := c.FindTransactionsByWorkspaceIdAndMonthRepository.Find(globalFilters)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when retrieving transactions",
		}, http.StatusInternalServerError)
	}

	slices.Reverse(transactions)

	transactions, err = c.replaceTransactionIfEditRepeat(transactions)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when replacing transactions",
		}, http.StatusInternalServerError)
	}

	transactions, err = c.filterTransactions(transactions, globalFilters)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when filtering transactions",
		}, http.StatusInternalServerError)
	}

	transactions, err = c.putTransactionCustomFieldTypes(transactions)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when putting transaction custom field types",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(transactions, http.StatusOK)
}

func (c *GetTransactionController) filterTransactions(transactions []models.Transaction, globalFilters *helpers.GlobalFilterParams) ([]models.Transaction, error) {
	var filtered []models.Transaction

	startOfMonth := time.Date(globalFilters.Year, time.Month(globalFilters.Month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	for _, tx := range transactions {
		var dateToCheck time.Time
		if tx.IsConfirmed && tx.ConfirmationDate != nil {
			dateToCheck = *tx.ConfirmationDate
		} else {
			dateToCheck = tx.DueDate
		}

		switch tx.Frequency {
		case "DO_NOT_REPEAT":
			// Check if transaction date falls within the target month
			if (dateToCheck.Equal(startOfMonth) || dateToCheck.After(startOfMonth)) &&
				dateToCheck.Before(endOfMonth) {
				filtered = append(filtered, tx)
			}
		case "RECURRING", "REPEAT":
			// Check if transaction date is before end of month
			if dateToCheck.Before(endOfMonth) {
				filtered = append(filtered, tx)
			}
		default:
			filtered = append(filtered, tx)
		}
	}

	return filtered, nil
}

func (c *GetTransactionController) putTransactionCustomFieldTypes(transactions []models.Transaction) ([]models.Transaction, error) {
	wg := sync.WaitGroup{}
	customErrors := []error{}
	for _, transaction := range transactions {
		for l, customField := range transaction.CustomFields {
			wg.Add(1)

			go func(customField models.TransactionCustomField) {
				defer wg.Done()

				customFieldFound, err := c.FindCustomFieldByIdRepository.Find(customField.CustomFieldId, transaction.WorkspaceId)
				if err != nil {
					customErrors = append(customErrors, err)
					return
				}

				if customFieldFound == nil {
					customErrors = append(customErrors, errors.New("custom field not found"))
					return
				}

				transaction.CustomFields[l].Type = customFieldFound.Type
			}(customField)
		}
	}

	wg.Wait()

	if len(customErrors) > 0 {
		return nil, customErrors[0]
	}

	return transactions, nil
}

func (c *GetTransactionController) replaceTransactionIfEditRepeat(transactions []models.Transaction) ([]models.Transaction, error) {
	for i, transaction := range transactions {
		editTransaction, err := c.FindByIdEditTransactionRepository.Find(transaction.Id, transaction.RepeatSettings.CurrentCount, transaction.WorkspaceId)
		if err != nil {
			return nil, err
		}

		if editTransaction != nil && *editTransaction.MainCount == transaction.RepeatSettings.CurrentCount {
			repeatSettings := *transaction.RepeatSettings
			frequency := transaction.Frequency
			totalBalance := transaction.TotalBalance
			balance := transaction.Balance
			id := transaction.Id

			transactions[i] = *editTransaction
			transactions[i].Frequency = frequency
			transactions[i].RepeatSettings = &repeatSettings
			transactions[i].Id = id
			transactions[i].MainCount = nil
			transactions[i].MainId = nil
			if transactions[i].Frequency == "DO_NOT_REPEAT" {
				transactions[i].Balance = balance
			}
			transactions[i].TotalBalance = totalBalance
		}
	}

	return transactions, nil
}
