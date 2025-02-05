package controllers

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/google/uuid"
)

type DeleteAccountController struct {
	DeleteAccount usecase.DeleteAccount
}

func NewDeleteAccountController(deleteAccount usecase.DeleteAccount) *DeleteAccountController {
	return &DeleteAccountController{
		DeleteAccount: deleteAccount,
	}
}

func (c *DeleteAccountController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	id := r.Req.PathValue("id")
	err := uuid.Validate(id)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid account ID format",
		}, http.StatusBadRequest)
	}

	err = c.DeleteAccount.Delete(id)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when deleting account: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(nil, http.StatusNoContent)
}
