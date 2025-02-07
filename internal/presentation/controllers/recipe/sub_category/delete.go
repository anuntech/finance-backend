package sub_category

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
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
	recipeId := r.Req.PathValue("recipeId")
	subCategoryId := r.Req.PathValue("subCategoryId")
	workspaceId := r.Header.Get("workspaceId")

	err := c.UpdateRecipeRepository.DeleteSubCategory(recipeId, subCategoryId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when deleting sub category: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(nil, http.StatusNoContent)
}
