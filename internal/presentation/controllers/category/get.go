package category

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GetCategorysController struct {
	FindCategorysByWorkspaceIdRepository usecase.FindCategorysByWorkspaceIdRepository
}

func NewGetCategorysController(findManyByUserIdAndWorkspaceId usecase.FindCategorysByWorkspaceIdRepository) *GetCategorysController {
	return &GetCategorysController{
		FindCategorysByWorkspaceIdRepository: findManyByUserIdAndWorkspaceId,
	}
}

func (c *GetCategorysController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspaceId format",
		}, http.StatusBadRequest)
	}

	categorys, err := c.FindCategorysByWorkspaceIdRepository.Find(workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when retrieving categorys",
		}, http.StatusInternalServerError)
	}

	for i := range categorys {
		categorys[i].Amount = categorys[i].CalculateTotalAmount()
	}

	return helpers.CreateResponse(categorys, http.StatusOK)
}
