package transaction

import (
	"encoding/json"
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
)

type ExcludeInstallmentsUntilController struct {
	UpdateTransactionRepository usecase.UpdateTransactionRepository
}

func NewExcludeInstallmentsUntilController(
	updateTransaction usecase.UpdateTransactionRepository,
) *ExcludeInstallmentsUntilController {
	return &ExcludeInstallmentsUntilController{
		UpdateTransactionRepository: updateTransaction,
	}
}

type ExcludeInstallmentsUntilBody struct {
	TransactionId string `json:"transactionId"`
	Until         string `json:"until"`
	Count         int    `json:"count"`
}

func (c *ExcludeInstallmentsUntilController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	// workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	// if err != nil {
	// 	return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
	// 		Error: "Formato do ID da área de trabalho inválido",
	// 	}, http.StatusBadRequest)
	// }

	var body ExcludeInstallmentsUntilBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Formato da solicitação inválido",
		}, http.StatusBadRequest)
	}

	// transaction, err := c.UpdateTransactionRepository.Update(body.TransactionId, body.Until, body.Count, workspaceId)
	// if err != nil {
	// 	return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
	// 		Error: "Erro ao atualizar a transação",
	// 	}, http.StatusInternalServerError)
	// }

	return helpers.CreateResponse(nil, http.StatusOK)
}
