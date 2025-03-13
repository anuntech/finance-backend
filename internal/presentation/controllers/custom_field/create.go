package custom_field

import (
	"encoding/json"
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateCustomFieldController struct {
	Validate                    *validator.Validate
	Translator                  ut.Translator
	CreateCustomFieldRepository usecase.CreateCustomFieldRepository
}

func NewCreateCustomFieldController(createCustomFieldRepository usecase.CreateCustomFieldRepository) *CreateCustomFieldController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &CreateCustomFieldController{
		Validate:                    validate,
		CreateCustomFieldRepository: createCustomFieldRepository,
	}
}

type CustomFieldBody struct {
	Name     string   `json:"name" validate:"required,min=2,max=50"`
	Type     string   `json:"type" validate:"required,oneof=SELECT TEXT NUMBER DATE BOOLEAN"`
	Options  []string `json:"options,omitempty" validate:"omitempty,required_if=Type SELECT,dive,min=1,max=50"`
	Required bool     `json:"required"`
}

func (c *CreateCustomFieldController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	// Parsing workspace_id from header
	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid workspace ID format",
		}, http.StatusBadRequest)
	}

	// Parsing body
	var body CustomFieldBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid request body",
		}, http.StatusBadRequest)
	}

	// Validating body
	if err := c.Validate.Struct(body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: helpers.GetErrorMessages(c.Validate, err),
		}, http.StatusUnprocessableEntity)
	}

	// For SELECT type, options are required
	if body.Type == "SELECT" && len(body.Options) == 0 {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Options are required for SELECT type",
		}, http.StatusBadRequest)
	}

	// Create custom field
	customField := &models.CustomField{
		WorkspaceId: workspaceId.Hex(),
		Name:        body.Name,
		Type:        body.Type,
		Options:     body.Options,
		Required:    body.Required,
	}

	createdCustomField, err := c.CreateCustomFieldRepository.Create(customField)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: err.Error(),
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(createdCustomField, http.StatusCreated)
}
