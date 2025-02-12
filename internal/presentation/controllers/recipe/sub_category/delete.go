package sub_category

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeleteSubCategoryController struct {
	UpdateRecipeRepository usecase.UpdateRecipeRepository
}

func NewDeleteSubCategoryController(updateRecipe usecase.UpdateRecipeRepository) *DeleteSubCategoryController {
	return &DeleteSubCategoryController{
		UpdateRecipeRepository: updateRecipe,
	}
}

func (c *DeleteSubCategoryController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	recipeId, err := primitive.ObjectIDFromHex(r.Req.PathValue("recipeId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid recipeId format",
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

	err = c.UpdateRecipeRepository.DeleteSubCategory(recipeId, subCategoryId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when deleting sub category: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(nil, http.StatusNoContent)
}
