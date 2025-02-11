package controllers

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
)

type GetAccountsController struct {
	FindAccountByWorkspaceIdRepository usecase.FindAccountByWorkspaceIdRepository
}

func NewGetAccountsController(findManyByUserIdAndWorkspaceId usecase.FindAccountByWorkspaceIdRepository) *GetAccountsController {
	return &GetAccountsController{
		FindAccountByWorkspaceIdRepository: findManyByUserIdAndWorkspaceId,
	}
}

func (c *GetAccountsController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	accounts, err := c.FindAccountByWorkspaceIdRepository.Find(r.Header.Get("workspaceId"))
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
