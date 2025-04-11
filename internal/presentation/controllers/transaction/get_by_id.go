package transaction

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GetTransactionByIdController struct {
	FindTransactionByIdRepository usecase.FindTransactionByIdRepository
}

func NewGetTransactionByIdController(findTransactionByIdRepository usecase.FindTransactionByIdRepository) *GetTransactionByIdController {
	return &GetTransactionByIdController{
		FindTransactionByIdRepository: findTransactionByIdRepository,
	}
}

func (c *GetTransactionByIdController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	transactionId, err := primitive.ObjectIDFromHex(r.Req.PathValue("transactionId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "formato do ID da transação inválido",
		}, http.StatusBadRequest)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "formato do ID da área de trabalho inválido",
		}, http.StatusBadRequest)
	}

	transaction, err := c.FindTransactionByIdRepository.Find(transactionId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "erro ao buscar a transação: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(transaction, http.StatusOK)
}
