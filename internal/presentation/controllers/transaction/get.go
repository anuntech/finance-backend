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
	workspace_user_repository "github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/workspace_repository/user_repository"
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
	FindCategoryByIdRepository                      usecase.FindCategoryByIdRepository
	FindWorkspaceUserByIdRepository                 workspace_user_repository.FindWorkspaceUserByIdRepository
}

func NewGetTransactionController(
	findManyByUserIdAndWorkspaceId usecase.FindTransactionsByWorkspaceIdRepository,
	findByIdEditTransactionRepository usecase.FindByIdEditTransactionRepository,
	findCustomFieldByIdRepository usecase.FindCustomFieldByIdRepository,
	findCategoryByIdRepository usecase.FindCategoryByIdRepository,
	findWorkspaceUserByIdRepository workspace_user_repository.FindWorkspaceUserByIdRepository,
) *GetTransactionController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &GetTransactionController{
		FindTransactionsByWorkspaceIdAndMonthRepository: findManyByUserIdAndWorkspaceId,
		Validator:                         validate,
		FindByIdEditTransactionRepository: findByIdEditTransactionRepository,
		FindCustomFieldByIdRepository:     findCustomFieldByIdRepository,
		FindCategoryByIdRepository:        findCategoryByIdRepository,
		FindWorkspaceUserByIdRepository:   findWorkspaceUserByIdRepository,
	}
}

type GetTransactionParams struct {
	DateType string `json:"dateType" validate:"omitempty,oneof=CONFIRMATION DUE REGISTRATION"`
	Sort     string `json:"sort" validate:"omitempty,oneof=ASC DESC"`
	Search   string `json:"search" validate:"omitempty,max=255"`
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
		DateType: r.UrlParams.Get("dateType"),
		Sort:     r.UrlParams.Get("sort"),
		Search:   r.UrlParams.Get("search"),
	}

	if params.Search != "" {
		transactions, err = c.filterTransactionsBySearch(transactions, params.Search)
		if err != nil {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "an error occurred when filtering transactions by search",
			}, http.StatusInternalServerError)
		}

		return helpers.CreateResponse(transactions, http.StatusOK)
	}

	if params.DateType != "" {
		transactions, err = c.filterTransactionsByDateType(transactions, globalFilters, params)
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

	// Determine the date range based on available filters
	var startDate, endDate time.Time

	if globalFilters.InitialDate != "" && globalFilters.FinalDate != "" {
		// Use date range if provided
		var err error
		startDate, err = time.Parse("2006-01-02", globalFilters.InitialDate)
		if err != nil {
			return nil, err
		}

		endDate, err = time.Parse("2006-01-02", globalFilters.FinalDate)
		if err != nil {
			return nil, err
		}
		// Set end date to the end of the day
		endDate = endDate.Add(24*time.Hour - time.Second)
	}
	for _, tx := range transactions {
		switch params.DateType {
		case "DUE":
			if (tx.DueDate.Equal(startDate) || tx.DueDate.After(startDate)) &&
				tx.DueDate.Before(endDate) {
				filtered = append(filtered, tx)
			}
		case "CONFIRMATION":
			if tx.IsConfirmed && tx.ConfirmationDate != nil &&
				(tx.ConfirmationDate.Equal(startDate) || tx.ConfirmationDate.After(startDate)) &&
				tx.ConfirmationDate.Before(endDate) {
				filtered = append(filtered, tx)
			}
		case "REGISTRATION":
			if (tx.RegistrationDate.Equal(startDate) || tx.RegistrationDate.After(startDate)) &&
				tx.RegistrationDate.Before(endDate) {
				filtered = append(filtered, tx)
			}
		default:
			filtered = append(filtered, tx)
		}
	}

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
	wg := sync.WaitGroup{}
	editErrors := []error{}

	for i, transaction := range transactions {
		wg.Add(1)

		go func(i int, transaction models.Transaction) {
			defer wg.Done()

			// Make sure RepeatSettings exists
			if transaction.RepeatSettings == nil {
				transaction.RepeatSettings = &models.TransactionRepeatSettings{}
			}

			// Store the current count (installment number) before potential replacement
			currentCount := transaction.RepeatSettings.CurrentCount

			editTransaction, err := c.FindByIdEditTransactionRepository.Find(transaction.Id, transaction.RepeatSettings.CurrentCount, transaction.WorkspaceId)
			if err != nil {
				editErrors = append(editErrors, err)
				return
			}

			if editTransaction != nil && *editTransaction.MainCount == transaction.RepeatSettings.CurrentCount {
				repeatSettings := *transaction.RepeatSettings
				frequency := transaction.Frequency
				totalBalance := transaction.TotalBalance
				balance := transaction.Balance
				id := transaction.Id

				// Preserve the installment number/current count
				installmentNumber := repeatSettings.CurrentCount

				transactions[i] = *editTransaction
				transactions[i].Frequency = frequency
				transactions[i].RepeatSettings = &repeatSettings

				// Restore the installment number after replacing with edited transaction
				transactions[i].RepeatSettings.CurrentCount = installmentNumber

				transactions[i].Id = id
				transactions[i].MainCount = nil
				transactions[i].MainId = nil
				if transactions[i].Frequency == "DO_NOT_REPEAT" {
					transactions[i].Balance = balance
				}
				transactions[i].TotalBalance = totalBalance
			} else if currentCount > 0 {
				// If no edit was found but we had a currentCount, make sure to preserve it
				transactions[i].RepeatSettings.CurrentCount = currentCount
			}

			transactionCopy := transactions[i]
			transactionCopy.Type = "RECIPE"
			calc := infraHelpers.CalculateOneTransactionBalance(&transactionCopy)
			transactions[i].Balance.NetBalance = calc
		}(i, transaction)
	}

	wg.Wait()

	if len(editErrors) > 0 {
		return nil, editErrors[0]
	}

	return transactions, nil
}

func (c *GetTransactionController) ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(
		strings.ToLower(s),
		strings.ToLower(substr),
	)
}

func (c *GetTransactionController) filterTransactionsBySearch(transactions []models.Transaction, search string) ([]models.Transaction, error) {
	filtered := make([]models.Transaction, 0, len(transactions)/2)

	filterCategoryById := func(categoryId primitive.ObjectID, workspaceId primitive.ObjectID) (bool, error) {
		category, err := c.FindCategoryByIdRepository.Find(categoryId, workspaceId)
		if err != nil {
			return false, err
		}

		if c.ContainsIgnoreCase(category.Name, search) {
			return true, nil
		}

		for _, subCategory := range category.SubCategories {
			if c.ContainsIgnoreCase(subCategory.Name, search) {
				return true, nil
			}
		}

		return false, nil
	}

	filterByCategory := func(tx *models.Transaction) (bool, error) {
		if tx.CategoryId == nil {
			return false, nil
		}

		return filterCategoryById(*tx.CategoryId, tx.WorkspaceId)
	}

	filterByTag := func(tx *models.Transaction) (bool, error) {
		if tx.Tags == nil {
			return false, nil
		}

		for _, tag := range tx.Tags {
			categoryMatch, err := filterCategoryById(tag.TagId, tx.WorkspaceId)
			if err != nil {
				return false, err
			}

			if categoryMatch {
				return true, nil
			}
		}

		return false, nil
	}

	filterByAssignedTo := func(id primitive.ObjectID) bool {
		user, err := c.FindWorkspaceUserByIdRepository.Find(id)
		if err != nil {
			return false
		}

		if user == nil {
			return false
		}

		return c.ContainsIgnoreCase(user.Name, search)
	}

	wg := sync.WaitGroup{}
	var mu sync.Mutex

	wg.Add(len(transactions))
	for i := range transactions {
		go func(i int) {
			defer wg.Done()

			tx := &transactions[i]

			addToFiltered := func() {
				mu.Lock()
				defer mu.Unlock()
				filtered = append(filtered, *tx)
			}

			if c.ContainsIgnoreCase(tx.Name, search) ||
				c.ContainsIgnoreCase(tx.Description, search) ||
				c.ContainsIgnoreCase(tx.Supplier, search) ||
				c.ContainsIgnoreCase(tx.Type, search) ||
				c.ContainsIgnoreCase(tx.Frequency, search) {
				addToFiltered()
				return
			}

			if c.isDateMatch(tx.DueDate, search) ||
				c.isDateMatch(tx.RegistrationDate, search) ||
				(tx.ConfirmationDate != nil && c.isDateMatch(*tx.ConfirmationDate, search)) {
				addToFiltered()
				return
			}

			for _, cf := range tx.CustomFields {
				if c.ContainsIgnoreCase(cf.Value, search) {
					addToFiltered()
					return
				}
			}

			categoryMatch, err := filterByCategory(tx)
			if err != nil {
				return
			}

			if categoryMatch {
				addToFiltered()
				return
			}

			tagMatch, err := filterByTag(tx)
			if err != nil {
				return
			}

			if tagMatch {
				addToFiltered()
				return
			}

			if filterByAssignedTo(tx.AssignedTo) {
				addToFiltered()
				return
			}
		}(i)
	}

	wg.Wait()

	return filtered, nil
}

func (c *GetTransactionController) isDateMatch(date time.Time, search string) bool {
	formattedDate := date.Format("02/01/2006")
	if c.ContainsIgnoreCase(formattedDate, search) {
		return true
	}

	dayStr := ""
	if date.Day() < 10 {
		dayStr = date.Format("2")
	} else {
		dayStr = date.Format("02")
	}

	return c.ContainsIgnoreCase(dayStr, search)
}
