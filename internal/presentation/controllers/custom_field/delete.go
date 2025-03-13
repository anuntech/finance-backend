package custom_field

import (
	"encoding/json"
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeleteCustomFieldController struct {
	DeleteCustomFieldRepository usecase.DeleteCustomFieldRepository
}

func NewDeleteCustomFieldController(deleteCustomFieldRepository usecase.DeleteCustomFieldRepository) *DeleteCustomFieldController {
	return &DeleteCustomFieldController{
		DeleteCustomFieldRepository: deleteCustomFieldRepository,
	}
}

type DeleteCustomFieldsBody struct {
	CustomFieldIds []string `json:"customFieldIds" validate:"required,min=1,dive,required,mongodb"`
}

func (c *DeleteCustomFieldController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body DeleteCustomFieldsBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid request body",
		}, http.StatusBadRequest)
	}

	// Validate customFieldIds
	if len(body.CustomFieldIds) == 0 {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "At least one customFieldId is required",
		}, http.StatusBadRequest)
	}

	// Convert string IDs to ObjectIDs
	customFieldObjectIds := make([]primitive.ObjectID, len(body.CustomFieldIds))
	for i, id := range body.CustomFieldIds {
		objectId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "Invalid custom field ID format: " + id,
			}, http.StatusBadRequest)
		}
		customFieldObjectIds[i] = objectId
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid workspace ID format",
		}, http.StatusBadRequest)
	}

	err = c.DeleteCustomFieldRepository.Delete(customFieldObjectIds, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Error deleting custom fields: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(map[string]bool{"success": true}, http.StatusOK)
}
