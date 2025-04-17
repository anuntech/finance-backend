package category

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GetCategoryByIdController struct {
	FindCategoryByIdRepository usecase.FindCategoryByIdRepository
}

func NewGetCategoryByIdController(findCategoryByIdRepository usecase.FindCategoryByIdRepository) *GetCategoryByIdController {
	return &GetCategoryByIdController{
		FindCategoryByIdRepository: findCategoryByIdRepository,
	}
}

func (c *GetCategoryByIdController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	categoryId, err := primitive.ObjectIDFromHex(r.Req.PathValue("categoryId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Formato do ID da categoria inválido.",
		}, http.StatusBadRequest)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Formato do ID do espaço de trabalho inválido.",
		}, http.StatusBadRequest)
	}

	category, err := c.FindCategoryByIdRepository.Find(categoryId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Erro ao buscar a categoria.",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(category, http.StatusOK)
}
