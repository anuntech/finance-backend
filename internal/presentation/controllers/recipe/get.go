package recipe

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GetRecipesController struct {
	FindRecipesByWorkspaceIdRepository usecase.FindRecipesByWorkspaceIdRepository
}

func NewGetRecipesController(findManyByUserIdAndWorkspaceId usecase.FindRecipesByWorkspaceIdRepository) *GetRecipesController {
	return &GetRecipesController{
		FindRecipesByWorkspaceIdRepository: findManyByUserIdAndWorkspaceId,
	}
}

func (c *GetRecipesController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspaceId format",
		}, http.StatusBadRequest)
	}

	recipes, err := c.FindRecipesByWorkspaceIdRepository.Find(workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when retrieving recipes",
		}, http.StatusInternalServerError)
	}

	for i := range recipes {
		recipes[i].TotalAmount = recipes[i].CalculateTotalAmount()
	}

	return helpers.CreateResponse(recipes, http.StatusOK)
}
