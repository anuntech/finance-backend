package transaction

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GetTransactionController struct {
	FindTransactionsByWorkspaceIdAndMonthRepository usecase.FindTransactionsByWorkspaceIdRepository
}

func NewGetTransactionController(findManyByUserIdAndWorkspaceId usecase.FindTransactionsByWorkspaceIdRepository) *GetTransactionController {
	return &GetTransactionController{
		FindTransactionsByWorkspaceIdAndMonthRepository: findManyByUserIdAndWorkspaceId,
	}
}

func (c *GetTransactionController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspaceId format",
		}, http.StatusBadRequest)
	}

	filters, errHttp := c.getFilters(r.Header)
	if errHttp != nil {
		return errHttp
	}

	filters.WorkspaceId = workspaceId
	transactions, err := c.FindTransactionsByWorkspaceIdAndMonthRepository.Find(filters)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when retrieving transactions",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(transactions, http.StatusOK)
}

func (c *GetTransactionController) getFilters(header http.Header) (*usecase.FindTransactionsByWorkspaceIdInputRepository, *presentationProtocols.HttpResponse) {
	categoryType := header.Get("type")
	allowedTypes := []string{"RECIPE", "EXPENSE", ""}
	if !slices.Contains(allowedTypes, categoryType) {
		return nil, helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid category type",
		}, http.StatusBadRequest)
	}

	month := header.Get("month") // for dueDate and confirmationDate
	monthInt, err := strconv.Atoi(month)
	if monthInt < 1 || monthInt > 12 || err != nil {
		return nil, helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid month",
		}, http.StatusBadRequest)
	}

	year := header.Get("year")
	yearInt, err := strconv.Atoi(year)
	if yearInt < 1800 || yearInt > 9999 || err != nil {
		return nil, helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid year",
		}, http.StatusBadRequest)
	}

	return &usecase.FindTransactionsByWorkspaceIdInputRepository{
		Month: monthInt,
		Year:  yearInt,
		Type:  categoryType,
	}, nil
}
