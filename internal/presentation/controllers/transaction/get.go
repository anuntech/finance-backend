package transaction

import (
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GetTransactionController struct {
	FindTransactionsByWorkspaceIdAndMonthRepository usecase.FindTransactionsByWorkspaceIdRepository
	Validator                                       *validator.Validate
}

func NewGetTransactionController(findManyByUserIdAndWorkspaceId usecase.FindTransactionsByWorkspaceIdRepository) *GetTransactionController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &GetTransactionController{
		FindTransactionsByWorkspaceIdAndMonthRepository: findManyByUserIdAndWorkspaceId,
		Validator: validate,
	}
}

func (c *GetTransactionController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspaceId format",
		}, http.StatusBadRequest)
	}

	filters, errHttp := c.getFilters(&r.UrlParams)
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

	for i, j := 0, len(transactions)-1; i < j; i, j = i+1, j-1 {
		transactions[i], transactions[j] = transactions[j], transactions[i]
	}

	return helpers.CreateResponse(transactions, http.StatusOK)
}

type FilterParams struct {
	Month int    `json:"month" validate:"omitempty,min=1,max=12,required_with=Year"`
	Year  int    `json:"year" validate:"omitempty,min=1,max=9999,required_with=Month"`
	Type  string `json:"type" validate:"omitempty,oneof=RECIPE EXPENSE"`
}

func (c *GetTransactionController) getFilters(urlQueries *url.Values) (*usecase.FindTransactionsByWorkspaceIdInputRepository, *presentationProtocols.HttpResponse) {
	monthInt, _ := strconv.Atoi(urlQueries.Get("month"))
	yearInt, _ := strconv.Atoi(urlQueries.Get("year"))

	params := &FilterParams{
		Month: monthInt,
		Year:  yearInt,
		Type:  urlQueries.Get("type"),
	}

	err := c.Validator.Struct(params)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return nil, helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: err.Error(),
		}, http.StatusBadRequest)
	}

	return &usecase.FindTransactionsByWorkspaceIdInputRepository{
		Month: monthInt,
		Year:  yearInt,
		Type:  params.Type,
	}, nil
}
