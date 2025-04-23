package credit_card

import (
	"net/http"
	"strings"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeleteCreditCardController handles deleting credit cards
type DeleteCreditCardController struct {
	DeleteCreditCardRepository usecase.DeleteCreditCardRepository
}

// NewDeleteCreditCardController initializes a new DeleteCreditCardController
func NewDeleteCreditCardController(deleteRepo usecase.DeleteCreditCardRepository) *DeleteCreditCardController {
	return &DeleteCreditCardController{DeleteCreditCardRepository: deleteRepo}
}

// Handle processes the HTTP request to delete credit cards
func (c *DeleteCreditCardController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspace ID format",
		}, http.StatusBadRequest)
	}

	ids := r.UrlParams.Get("ids")
	idsSlice := strings.Split(ids, ",")
	var idsObjectID []primitive.ObjectID
	for _, id := range idsSlice {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "invalid credit card ID format",
			}, http.StatusBadRequest)
		}
		idsObjectID = append(idsObjectID, objectID)
	}

	if err := c.DeleteCreditCardRepository.Delete(idsObjectID, workspaceId); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when deleting credit cards",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(nil, http.StatusNoContent)
}
