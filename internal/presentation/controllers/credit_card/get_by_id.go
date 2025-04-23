package credit_card

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetCreditCardByIdController handles retrieving a credit card by its ID
type GetCreditCardByIdController struct {
	FindCreditCardByIdRepository usecase.FindCreditCardByIdRepository
}

// NewGetCreditCardByIdController initializes a new GetCreditCardByIdController
func NewGetCreditCardByIdController(repo usecase.FindCreditCardByIdRepository) *GetCreditCardByIdController {
	return &GetCreditCardByIdController{FindCreditCardByIdRepository: repo}
}

// Handle processes the HTTP request to retrieve a single credit card
func (c *GetCreditCardByIdController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	id, err := primitive.ObjectIDFromHex(r.Req.PathValue("creditCardId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid credit card ID format",
		}, http.StatusBadRequest)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspace ID format",
		}, http.StatusBadRequest)
	}

	card, err := c.FindCreditCardByIdRepository.Find(id, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when retrieving credit card",
		}, http.StatusInternalServerError)
	}
	if card == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "credit card not found",
		}, http.StatusNotFound)
	}

	return helpers.CreateResponse(card, http.StatusOK)
}
