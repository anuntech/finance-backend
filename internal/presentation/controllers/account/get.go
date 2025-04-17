package controllers

import (
	"net/http"
	"slices"

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
			Error: "Formato do ID da área de trabalho inválido",
		}, http.StatusBadRequest)
	}

	globalFilters, errHttp := helpers.GetGlobalFilterByQueries(&r.UrlParams, workspaceId, c.Validator)
	if errHttp != nil {
		return errHttp
	}

	accounts, err := c.FindAccountByWorkspaceIdRepository.Find(globalFilters)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "ocorreu um erro ao buscar as contas",
		}, http.StatusInternalServerError)
	}

	slices.Reverse(accounts)

	return helpers.CreateResponse(accounts, http.StatusOK)
}
