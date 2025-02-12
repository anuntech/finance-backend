package controllers

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid workspace ID format",
		}, http.StatusBadRequest)
	}

	accounts, err := c.FindAccountByWorkspaceIdRepository.Find(workspaceId)
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
