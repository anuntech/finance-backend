package custom_field

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

type UpdateCustomFieldController struct {
	Validate                      *validator.Validate
	UpdateCustomFieldRepository   usecase.UpdateCustomFieldRepository
	FindCustomFieldByIdRepository usecase.FindCustomFieldByIdRepository
}

func NewUpdateCustomFieldController(
	updateCustomFieldRepository usecase.UpdateCustomFieldRepository,
	findCustomFieldByIdRepository usecase.FindCustomFieldByIdRepository,
) *UpdateCustomFieldController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &UpdateCustomFieldController{
		Validate:                      validate,
		UpdateCustomFieldRepository:   updateCustomFieldRepository,
		FindCustomFieldByIdRepository: findCustomFieldByIdRepository,
	}
}

type UpdateCustomFieldBody struct {
	Name     string   `json:"name" validate:"required,min=2,max=50"`
	Type     string   `json:"type" validate:"required,oneof=SELECT TEXT NUMBER DATE BOOLEAN"`
	Options  []string `json:"options" validate:"omitempty,required_if=Type SELECT,dive,min=1,max=50"`
	Required bool     `json:"required"`
}

func (c *UpdateCustomFieldController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	customFieldId, err := primitive.ObjectIDFromHex(r.Req.PathValue("customFieldId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Formato do ID do campo personalizado inválido.",
		}, http.StatusBadRequest)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Formato do ID do espaço de trabalho inválido.",
		}, http.StatusBadRequest)
	}

	// Check if custom field exists
	existingCustomField, err := c.FindCustomFieldByIdRepository.Find(customFieldId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Erro ao buscar o campo personalizado.",
		}, http.StatusInternalServerError)
	}

	if existingCustomField == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Campo personalizado não encontrado.",
		}, http.StatusNotFound)
	}

	// Parsing body
	var body UpdateCustomFieldBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Requisição inválida. Por favor, verifique os dados enviados.",
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
			Error: "Opções são obrigatórias para o tipo SELEÇÃO.",
		}, http.StatusBadRequest)
	}

	// Update custom field
	customField := &models.CustomField{
		Id:          existingCustomField.Id,
		WorkspaceId: workspaceId.Hex(),
		Name:        body.Name,
		Type:        body.Type,
		Options:     body.Options,
		Required:    body.Required,
	}

	updatedCustomField, err := c.UpdateCustomFieldRepository.Update(customFieldId, customField)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Erro ao atualizar o campo personalizado.",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(updatedCustomField, http.StatusOK)
}
