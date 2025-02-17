package controllers

import (
	"log"
	"net/http"
	"strings"

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
	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid workspace ID format",
		}, http.StatusBadRequest)
	}

	ids := r.UrlParams.Get("ids")
	idsSlice := strings.Split(ids, ",")
	idsObjectID := []primitive.ObjectID{}
	log.Println(idsSlice)

	for _, id := range idsSlice {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "Invalid category ID format",
			}, http.StatusBadRequest)
		}
		idsObjectID = append(idsObjectID, objectID)
	}

	err = c.DeleteAccountRepository.Delete(idsObjectID, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when deleting account: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(nil, http.StatusNoContent)
}
