package credit_card

import (
	"net/http"
	"slices"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetCreditCardsController handles retrieving all credit cards
type GetCreditCardsController struct {
	FindCreditCardsRepository usecase.FindCreditCardsRepository
	Validate                  *validator.Validate
}

// NewGetCreditCardsController creates a new instance of GetCreditCardsController
func NewGetCreditCardsController(repo usecase.FindCreditCardsRepository) *GetCreditCardsController {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return &GetCreditCardsController{FindCreditCardsRepository: repo, Validate: validate}
}

// Handle processes the HTTP request to retrieve credit cards
func (c *GetCreditCardsController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspace ID format",
		}, http.StatusBadRequest)
	}

	globalFilters, httpResponse := helpers.GetGlobalFilterByQueries(&r.UrlParams, workspaceId, c.Validate)
	if httpResponse != nil {
		return httpResponse
	}

	cards, err := c.FindCreditCardsRepository.Find(globalFilters)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when retrieving credit cards",
		}, http.StatusInternalServerError)
	}

	slices.Reverse(cards)

	return helpers.CreateResponse(cards, http.StatusOK)
}
