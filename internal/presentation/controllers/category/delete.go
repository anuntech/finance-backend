package category

import (
	"net/http"
	"strings"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeleteCategoryController struct {
	DeleteCategoryRepository usecase.DeleteCategoryRepository
}

func NewDeleteCategoryController(deleteCategory usecase.DeleteCategoryRepository) *DeleteCategoryController {
	return &DeleteCategoryController{
		DeleteCategoryRepository: deleteCategory,
	}
}

func (c *DeleteCategoryController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid workspace ID format",
		}, http.StatusBadRequest)
	}

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

	err = c.DeleteCategoryRepository.Delete(idsObjectID, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when deleting category: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(nil, http.StatusNoContent)
}
