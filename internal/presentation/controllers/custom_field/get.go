package custom_field

import (
	"net/http"
	"slices"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GetCustomFieldsController struct {
	FindCustomFieldsRepository usecase.FindCustomFieldsRepository
	Validate                   *validator.Validate
}

func NewGetCustomFieldsController(findCustomFieldsRepository usecase.FindCustomFieldsRepository) *GetCustomFieldsController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &GetCustomFieldsController{
		FindCustomFieldsRepository: findCustomFieldsRepository,
		Validate:                   validate,
	}
}

func (c *GetCustomFieldsController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Formato do ID do espaço de trabalho inválido.",
		}, http.StatusBadRequest)
	}

	globalFilters, httpResponse := helpers.GetGlobalFilterByQueries(&r.UrlParams, workspaceId, c.Validate)
	if httpResponse != nil {
		return httpResponse
	}

	customFields, err := c.FindCustomFieldsRepository.Find(globalFilters)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Ocorreu um erro ao buscar os campos personalizados.",
		}, http.StatusInternalServerError)
	}

	slices.Reverse(customFields)

	return helpers.CreateResponse(customFields, http.StatusOK)
}
