package transaction

import (
	"errors"
	"net/http"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
	infraHelpers "github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
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

type GetTransactionParams struct {
	DateType    string `json:"dateType" validate:"omitempty,oneof=CONFIRMATION DUE REGISTRATION"`
	Sort        string `json:"sort" validate:"omitempty,oneof=ASC DESC"`
	Supplier    string `json:"supplier" validate:"omitempty,max=255"`
	Description string `json:"description" validate:"omitempty,max=255"`
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

	params := &GetTransactionParams{
		DateType:    r.UrlParams.Get("dateType"),
		Sort:        r.UrlParams.Get("sort"),
		Description: r.UrlParams.Get("description"),
		Supplier:    r.UrlParams.Get("supplier"),
	}

	if params.Description != "" || params.Supplier != "" {
		transactions = c.filterTransactionsByDescriptionAndSupplier(transactions, params)
	} else {
		if params.DateType != "" {
			transactions, err = c.filterTransactionsByDateType(transactions, globalFilters, params)
		} else {
			transactions, err = c.filterTransactions(transactions, globalFilters)
		}
	}

	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when filtering transactions",
		}, http.StatusInternalServerError)
	}

	if params.Sort != "" {
		c.sortTransactions(transactions, params)
	}

	transactions, err = c.putTransactionCustomFieldTypes(transactions)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when putting transaction custom field types",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(transactions, http.StatusOK)
}

func (c *GetTransactionController) filterTransactionsByDateType(transactions []models.Transaction, globalFilters *helpers.GlobalFilterParams, params *GetTransactionParams) ([]models.Transaction, error) {
	var filtered []models.Transaction

	startOfMonth := time.Date(globalFilters.Year, time.Month(globalFilters.Month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	for _, tx := range transactions {
		switch params.DateType {
		case "DUE":
			if (tx.DueDate.Equal(startOfMonth) || tx.DueDate.After(startOfMonth)) &&
				tx.DueDate.Before(endOfMonth) {
				filtered = append(filtered, tx)
			}
		case "CONFIRMATION":
			if tx.IsConfirmed && tx.ConfirmationDate != nil &&
				(tx.ConfirmationDate.Equal(startOfMonth) || tx.ConfirmationDate.After(startOfMonth)) &&
				tx.ConfirmationDate.Before(endOfMonth) {
				filtered = append(filtered, tx)
			}
		case "REGISTRATION":
			if (tx.RegistrationDate.Equal(startOfMonth) || tx.RegistrationDate.After(startOfMonth)) &&
				tx.RegistrationDate.Before(endOfMonth) {
				filtered = append(filtered, tx)
			}
		default:
			filtered = append(filtered, tx)
		}
	}

	return filtered, nil
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

	// Ordenar as transações por DueDate por padrão
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].DueDate.After(filtered[j].DueDate)
	})

	return filtered, nil
}

func (c *GetTransactionController) sortTransactions(transactions []models.Transaction, params *GetTransactionParams) {
	switch params.DateType {
	case "DUE":
		sort.Slice(transactions, func(i, j int) bool {
			if params.Sort == "ASC" {
				return transactions[i].DueDate.After(transactions[j].DueDate)
			}

			return transactions[i].DueDate.Before(transactions[j].DueDate)
		})
	case "CONFIRMATION":
		sort.Slice(transactions, func(i, j int) bool {
			if params.Sort == "ASC" {
				return transactions[i].ConfirmationDate.After(*transactions[j].ConfirmationDate)
			}

			return transactions[i].ConfirmationDate.Before(*transactions[j].ConfirmationDate)
		})
	case "REGISTRATION":
		sort.Slice(transactions, func(i, j int) bool {
			if params.Sort == "ASC" {
				return transactions[i].RegistrationDate.After(transactions[j].RegistrationDate)
			}

			return transactions[i].RegistrationDate.Before(transactions[j].RegistrationDate)
		})
	}
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

		transactionCopy := transactions[i]
		transactionCopy.Type = "RECIPE"
		calc := infraHelpers.CalculateOneTransactionBalance(&transactionCopy)
		transactions[i].Balance.NetBalance = calc
	}

	return transactions, nil
}

func (c *GetTransactionController) ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(
		strings.ToLower(s),
		strings.ToLower(substr),
	)
}

func (c *GetTransactionController) filterTransactionsByDescriptionAndSupplier(transactions []models.Transaction, params *GetTransactionParams) []models.Transaction {
	var filtered []models.Transaction

	for _, tx := range transactions {
		matchesDescription := params.Description == "" ||
			(tx.Name != "" && c.ContainsIgnoreCase(tx.Name, params.Description))

		matchesSupplier := params.Supplier == "" ||
			(tx.Supplier != "" && c.ContainsIgnoreCase(tx.Supplier, params.Supplier))

		if matchesDescription && matchesSupplier {
			filtered = append(filtered, tx)
		}
	}

	return filtered
}
