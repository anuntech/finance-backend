package credit_card

import (
	"encoding/json"
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateCreditCardController handles creating new credit cards
type CreateCreditCardController struct {
	Validate                       *validator.Validate
	CreateCreditCardRepository     usecase.CreateCreditCardRepository
	FindCreditCardByNameRepository usecase.FindCreditCardByNameAndWorkspaceIdRepository
}

// NewCreateCreditCardController initializes a CreateCreditCardController
func NewCreateCreditCardController(
	createCreditCardRepository usecase.CreateCreditCardRepository,
	findCreditCardByNameRepository usecase.FindCreditCardByNameAndWorkspaceIdRepository,
) *CreateCreditCardController {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return &CreateCreditCardController{
		Validate:                       validate,
		CreateCreditCardRepository:     createCreditCardRepository,
		FindCreditCardByNameRepository: findCreditCardByNameRepository,
	}
}

// CreateCreditCardControllerBody defines the expected body for creating a credit card
type CreateCreditCardControllerBody struct {
	Name      string  `json:"name" validate:"required,min=3,max=255"`
	DueDate   int     `json:"dueDate" validate:"required,min=1,max=31"`
	CloseDate int     `json:"closeDate" validate:"required,min=1,max=31"`
	Limit     float64 `json:"limit" validate:"min=0"`
	Balance   float64 `json:"balance" validate:"min=0"`
}

// Handle processes the HTTP request for creating a credit card
func (c *CreateCreditCardController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body CreateCreditCardControllerBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid body request",
		}, http.StatusBadRequest)
	}

	if err := c.Validate.Struct(body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: helpers.GetErrorMessages(c.Validate, err),
		}, http.StatusUnprocessableEntity)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspace ID format",
		}, http.StatusBadRequest)
	}

	existing, err := c.FindCreditCardByNameRepository.FindByNameAndWorkspaceId(body.Name, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when checking for credit card name",
		}, http.StatusInternalServerError)
	}
	if existing != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "a credit card with this name already exists in this workspace",
		}, http.StatusConflict)
	}

	creditCard, err := c.CreateCreditCardRepository.Create(&models.CreditCard{
		Name:        body.Name,
		WorkspaceId: workspaceId.Hex(),
		DueDate:     body.DueDate,
		CloseDate:   body.CloseDate,
		Limit:       body.Limit,
		Balance:     body.Balance,
	})
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when creating credit card",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(creditCard, http.StatusCreated)
}
