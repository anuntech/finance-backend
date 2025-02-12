package recipe

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeleteRecipeController struct {
	DeleteRecipeRepository usecase.DeleteRecipeRepository
}

func NewDeleteRecipeController(deleteRecipe usecase.DeleteRecipeRepository) *DeleteRecipeController {
	return &DeleteRecipeController{
		DeleteRecipeRepository: deleteRecipe,
	}
}

func (c *DeleteRecipeController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid workspace ID format",
		}, http.StatusBadRequest)
	}

	recipeId, err := primitive.ObjectIDFromHex(r.Req.PathValue("recipeId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid recipe ID format",
		}, http.StatusBadRequest)
	}

	err = c.DeleteRecipeRepository.Delete(recipeId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when deleting recipe: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(nil, http.StatusNoContent)
}
