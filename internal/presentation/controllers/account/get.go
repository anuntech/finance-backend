package controllers

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GetAccountsController struct {
	FindAccountByWorkspaceIdRepository usecase.FindAccountByWorkspaceIdRepository
	Validator                          *validator.Validate
}

func NewGetAccountsController(findManyByUserIdAndWorkspaceId usecase.FindAccountByWorkspaceIdRepository) *GetAccountsController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &GetAccountsController{
		FindAccountByWorkspaceIdRepository: findManyByUserIdAndWorkspaceId,
		Validator:                          validate,
	}
}

func (c *GetAccountsController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid workspace ID format",
		}, http.StatusBadRequest)
	}

	globalFilters, errHttp := helpers.GetGlobalFilterByQueries(&r.UrlParams, workspaceId, c.Validator)
	if errHttp != nil {
		return errHttp
	}

	accounts, err := c.FindAccountByWorkspaceIdRepository.Find(globalFilters)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when retrieving accounts",
		}, http.StatusInternalServerError)
	}

	for i, j := 0, len(accounts)-1; i < j; i, j = i+1, j-1 {
		accounts[i], accounts[j] = accounts[j], accounts[i]
	}

	return helpers.CreateResponse(accounts, http.StatusOK)
}
