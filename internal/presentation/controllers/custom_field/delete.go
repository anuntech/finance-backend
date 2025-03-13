package custom_field

import (
	"net/http"
	"strings"

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

func (c *DeleteCustomFieldController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	ids := r.UrlParams.Get("ids")
	idsSlice := strings.Split(ids, ",")
	idsObjectID := []primitive.ObjectID{}

	for _, id := range idsSlice {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "Invalid category ID format",
			}, http.StatusBadRequest)
		}
		idsObjectID = append(idsObjectID, objectID)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid workspace ID format",
		}, http.StatusBadRequest)
	}

	err = c.DeleteCustomFieldRepository.Delete(idsObjectID, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Error deleting custom fields: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(nil, http.StatusNoContent)
}
