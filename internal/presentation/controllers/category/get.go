package category

import (
	"net/http"
	"slices"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GetCategoriesController struct {
	FindCategoriesByWorkspaceIdRepository usecase.FindCategoriesByWorkspaceIdRepository
}

func NewGetCategoriesController(findManyByUserIdAndWorkspaceId usecase.FindCategoriesByWorkspaceIdRepository) *GetCategoriesController {
	return &GetCategoriesController{
		FindCategoriesByWorkspaceIdRepository: findManyByUserIdAndWorkspaceId,
	}
}

func (c *GetCategoriesController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspaceId format",
		}, http.StatusBadRequest)
	}

	categoryType := r.UrlParams.Get("type")

	allowedTypes := []string{"RECIPE", "EXPENSE", "TAG", ""}
	if !slices.Contains(allowedTypes, categoryType) {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid category type",
		}, http.StatusBadRequest)
	}

	categories, err := c.FindCategoriesByWorkspaceIdRepository.Find(workspaceId, categoryType)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when retrieving categories",
		}, http.StatusInternalServerError)
	}

	for i := range categories {
		categories[i].CalculateTotalAmount()
		categories[i].InvertSubCategoriesOrder()
	}

	for i, j := 0, len(categories)-1; i < j; i, j = i+1, j-1 {
		categories[i], categories[j] = categories[j], categories[i]
	}

	return helpers.CreateResponse(categories, http.StatusOK)
}
