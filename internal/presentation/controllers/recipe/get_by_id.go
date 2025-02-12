package recipe

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GetRecipeByIdController struct {
	FindRecipeByIdRepository usecase.FindRecipeByIdRepository
}

func NewGetRecipeByIdController(findRecipeByIdRepository usecase.FindRecipeByIdRepository) *GetRecipeByIdController {
	return &GetRecipeByIdController{
		FindRecipeByIdRepository: findRecipeByIdRepository,
	}
}

func (c *GetRecipeByIdController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	recipeId, err := primitive.ObjectIDFromHex(r.Req.PathValue("recipeId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid recipeId format",
		}, http.StatusBadRequest)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspaceId format",
		}, http.StatusBadRequest)
	}

	recipe, err := c.FindRecipeByIdRepository.Find(recipeId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: err.Error(),
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(recipe, http.StatusOK)
}
