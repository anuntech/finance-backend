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
	workspaceId := r.Header.Get("workspaceId")
	recipeId := r.Req.PathValue("recipeId")

	if recipeId == "" {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "recipe id is required",
		}, http.StatusUnprocessableEntity)
	}

	objectId, err := primitive.ObjectIDFromHex(recipeId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid recipe ID format",
		}, http.StatusBadRequest)
	}

	err = c.DeleteRecipeRepository.Delete(objectId.Hex(), workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when deleting recipe: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(nil, http.StatusNoContent)
}
