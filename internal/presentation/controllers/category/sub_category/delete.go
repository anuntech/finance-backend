package sub_category

import (
	"net/http"
	"strings"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeleteSubCategoryController struct {
	UpdateCategoryRepository usecase.UpdateCategoryRepository
}

func NewDeleteSubCategoryController(updateCategory usecase.UpdateCategoryRepository) *DeleteSubCategoryController {
	return &DeleteSubCategoryController{
		UpdateCategoryRepository: updateCategory,
	}
}

func (c *DeleteSubCategoryController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	categoryId, err := primitive.ObjectIDFromHex(r.Req.PathValue("categoryId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid categoryId format",
		}, http.StatusBadRequest)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspaceId format",
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

	err = c.UpdateCategoryRepository.DeleteSubCategory(idsObjectID, categoryId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when deleting sub category: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(nil, http.StatusNoContent)
}
