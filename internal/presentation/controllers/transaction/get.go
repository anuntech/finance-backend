package transaction

import (
	"net/http"
	"slices"

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
}

func NewGetTransactionController(findManyByUserIdAndWorkspaceId usecase.FindTransactionsByWorkspaceIdRepository, findByIdEditTransactionRepository usecase.FindByIdEditTransactionRepository) *GetTransactionController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &GetTransactionController{
		FindTransactionsByWorkspaceIdAndMonthRepository: findManyByUserIdAndWorkspaceId,
		Validator:                         validate,
		FindByIdEditTransactionRepository: findByIdEditTransactionRepository,
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

	transactions, err = c.ReplaceTransactionIfEditRepeat(transactions)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when replacing transactions",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(transactions, http.StatusOK)
}

func (c *GetTransactionController) ReplaceTransactionIfEditRepeat(transactions []models.Transaction) ([]models.Transaction, error) {
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
