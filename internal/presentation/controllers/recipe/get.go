package recipe

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
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
	recipes, err := c.FindRecipesByWorkspaceIdRepository.Find(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when retrieving recipes",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(recipes, http.StatusOK)
}
