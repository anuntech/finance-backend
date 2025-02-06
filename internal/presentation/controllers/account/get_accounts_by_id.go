package controllers

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GetAccountByIdController struct {
	FindAccountByIdRepository usecase.FindAccountByIdRepository
}

func NewGetAccountByIdController(findById usecase.FindAccountByIdRepository) *GetAccountByIdController {
	return &GetAccountByIdController{
		FindAccountByIdRepository: findById,
	}
}

func (c *GetAccountByIdController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	id := r.Req.PathValue("id")
	_, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid account ID format",
		}, http.StatusBadRequest)
	}

	account, err := c.FindAccountByIdRepository.Find(id)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when retrieving account: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	if account == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "account not found",
		}, http.StatusNotFound)
	}

	return helpers.CreateResponse(account, http.StatusOK)
}
