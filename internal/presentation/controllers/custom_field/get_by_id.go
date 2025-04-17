package custom_field

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GetCustomFieldByIdController struct {
	FindCustomFieldByIdRepository usecase.FindCustomFieldByIdRepository
}

func NewGetCustomFieldByIdController(findCustomFieldByIdRepository usecase.FindCustomFieldByIdRepository) *GetCustomFieldByIdController {
	return &GetCustomFieldByIdController{
		FindCustomFieldByIdRepository: findCustomFieldByIdRepository,
	}
}

func (c *GetCustomFieldByIdController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	customFieldId, err := primitive.ObjectIDFromHex(r.Req.PathValue("customFieldId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Formato do ID do campo personalizado inválido.",
		}, http.StatusBadRequest)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Formato do ID do espaço de trabalho inválido.",
		}, http.StatusBadRequest)
	}

	customField, err := c.FindCustomFieldByIdRepository.Find(customFieldId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Ocorreu um erro ao buscar o campo personalizado.",
		}, http.StatusInternalServerError)
	}

	if customField == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Campo personalizado não encontrado.",
		}, http.StatusNotFound)
	}

	return helpers.CreateResponse(customField, http.StatusOK)
}
