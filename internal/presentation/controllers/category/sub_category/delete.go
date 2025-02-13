package sub_category

import (
	"net/http"

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

	subCategoryId, err := primitive.ObjectIDFromHex(r.Req.PathValue("subCategoryId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid subCategoryId format",
		}, http.StatusBadRequest)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspaceId format",
		}, http.StatusBadRequest)
	}

	err = c.UpdateCategoryRepository.DeleteSubCategory(categoryId, subCategoryId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when deleting sub category: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(nil, http.StatusNoContent)
}
