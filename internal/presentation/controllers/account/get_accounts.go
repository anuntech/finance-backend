package controllers

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
)

type GetAccountsController struct {
	FindByUserIdAndWorkspaceId usecase.FindByUserIdAndWorkspaceId
}

func NewGetAccountsController(findManyByUserIdAndWorkspaceId usecase.FindByUserIdAndWorkspaceId) *GetAccountsController {
	return &GetAccountsController{
		FindByUserIdAndWorkspaceId: findManyByUserIdAndWorkspaceId,
	}
}

func (c *GetAccountsController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	accounts, err := c.FindByUserIdAndWorkspaceId.Find(r.Header.Get("UserId"), r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when retrieving accounts",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(accounts, http.StatusOK)
}
