package controllers

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeleteAccountController struct {
	DeleteAccountRepository usecase.DeleteAccountRepository
}

func NewDeleteAccountController(deleteAccount usecase.DeleteAccountRepository) *DeleteAccountController {
	return &DeleteAccountController{
		DeleteAccountRepository: deleteAccount,
	}
}

func (c *DeleteAccountController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	id := r.Req.PathValue("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid account ID format",
		}, http.StatusBadRequest)
	}

	err = c.DeleteAccountRepository.Delete(objectId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when deleting account: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(nil, http.StatusNoContent)
}
