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

// UpdateCreditCardController handles updating credit cards
type UpdateCreditCardController struct {
	Validate                           *validator.Validate
	UpdateCreditCardRepository         usecase.UpdateCreditCardRepository
	FindCreditCardByIdRepository       usecase.FindCreditCardByIdRepository
	FindByNameAndWorkspaceIdRepository usecase.FindCreditCardByNameAndWorkspaceIdRepository
}

// NewUpdateCreditCardController initializes a new UpdateCreditCardController
func NewUpdateCreditCardController(
	updateRepo usecase.UpdateCreditCardRepository,
	findByIdRepo usecase.FindCreditCardByIdRepository,
	findByNameRepo usecase.FindCreditCardByNameAndWorkspaceIdRepository,
) *UpdateCreditCardController {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return &UpdateCreditCardController{
		Validate:                           validate,
		UpdateCreditCardRepository:         updateRepo,
		FindCreditCardByIdRepository:       findByIdRepo,
		FindByNameAndWorkspaceIdRepository: findByNameRepo,
	}
}

type UpdateCreditCardControllerBody struct {
	Name      string  `json:"name" validate:"required,min=3,max=255"`
	DueDate   int     `json:"dueDate" validate:"required,min=1,max=31"`
	CloseDate int     `json:"closeDate" validate:"required,min=1,max=31"`
	Limit     float64 `json:"limit" validate:"min=0"`
	Balance   float64 `json:"balance" validate:"min=0"`
	Flag      string  `json:"flag" validate:"required,oneof=MASTERCARD VISA ELO HIPERCARD UNIONPAY AURA"`
}

// Handle processes the HTTP request to update a credit card
func (c *UpdateCreditCardController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	id, err := primitive.ObjectIDFromHex(r.Req.PathValue("creditCardId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{Error: "invalid credit card ID format"}, http.StatusBadRequest)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{Error: "invalid workspace ID format"}, http.StatusBadRequest)
	}

	existing, err := c.FindCreditCardByIdRepository.Find(id, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{Error: "an error occurred when finding credit card"}, http.StatusInternalServerError)
	}
	if existing == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{Error: "credit card not found"}, http.StatusNotFound)
	}

	var body UpdateCreditCardControllerBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{Error: "invalid body request"}, http.StatusBadRequest)
	}

	if err := c.Validate.Struct(body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{Error: helpers.GetErrorMessages(c.Validate, err)}, http.StatusUnprocessableEntity)
	}

	// Check for name uniqueness if changed
	if existing.Name != body.Name {
		other, err := c.FindByNameAndWorkspaceIdRepository.FindByNameAndWorkspaceId(body.Name, workspaceId)
		if err != nil {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{Error: "an error occurred when checking credit card name"}, http.StatusInternalServerError)
		}
		if other != nil && other.Id != id {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{Error: "a credit card with this name already exists in this workspace"}, http.StatusConflict)
		}
	}

	updated, err := c.UpdateCreditCardRepository.Update(id, &models.CreditCard{
		Id:          existing.Id,
		WorkspaceId: workspaceId,
		Name:        body.Name,
		DueDate:     body.DueDate,
		CloseDate:   body.CloseDate,
		Limit:       body.Limit,
		Balance:     body.Balance,
		Flag:        body.Flag,
	})
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{Error: "an error occurred when updating credit card"}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(updated, http.StatusOK)
}
