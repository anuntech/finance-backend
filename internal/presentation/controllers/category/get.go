package category

import (
	"net/http"
	"slices"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GetCategoriesController struct {
	FindCategoriesRepository usecase.FindCategoriesRepository
	Validator                *validator.Validate
}

func NewGetCategoriesController(findManyByUserIdAndWorkspaceId usecase.FindCategoriesRepository) *GetCategoriesController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &GetCategoriesController{
		FindCategoriesRepository: findManyByUserIdAndWorkspaceId,
		Validator:                validate,
	}
}

func (c *GetCategoriesController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Formato do ID do espaço de trabalho inválido.",
		}, http.StatusBadRequest)
	}

	globalFilters, errHttp := helpers.GetGlobalFilterByQueries(&r.UrlParams, workspaceId, c.Validator)
	if errHttp != nil {
		return errHttp
	}

	categories, err := c.FindCategoriesRepository.Find(globalFilters)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Ocorreu um erro ao buscar as categorias.",
		}, http.StatusInternalServerError)
	}

	for i := range categories {
		categories[i].CalculateTotalAmount()
		categories[i].CalculateTotalCurrentAmount()
		categories[i].InvertSubCategoriesOrder()
	}

	slices.Reverse(categories)

	return helpers.CreateResponse(categories, http.StatusOK)
}
