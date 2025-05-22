package transaction

import (
	"encoding/json"
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/edit_transaction_repository"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ExcludeInstallmentsUntilController struct {
	UpdateTransactionRepository     usecase.UpdateTransactionRepository
	Validate                        *validator.Validate
	FindTransactionByIdRepository   usecase.FindTransactionByIdRepository
	DeleteEditTransactionRepository edit_transaction_repository.DeleteEditTransactionRepository
}

func NewExcludeInstallmentsUntilController(
	updateTransaction usecase.UpdateTransactionRepository,
	findTransactionById usecase.FindTransactionByIdRepository,
	deleteEditTransactionRepository edit_transaction_repository.DeleteEditTransactionRepository,
) *ExcludeInstallmentsUntilController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &ExcludeInstallmentsUntilController{
		UpdateTransactionRepository:     updateTransaction,
		Validate:                        validate,
		FindTransactionByIdRepository:   findTransactionById,
		DeleteEditTransactionRepository: deleteEditTransactionRepository,
	}
}

type ExcludeInstallmentsUntilBody struct {
	TransactionId string `json:"transactionId" validate:"required,mongodb"`
	Count         *int   `json:"count" validate:"omitempty,min=2,max=367"`
}

func (c *ExcludeInstallmentsUntilController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body ExcludeInstallmentsUntilBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Formato da solicitação inválido",
		}, http.StatusBadRequest)
	}

	if err := c.Validate.Struct(body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: helpers.GetErrorMessages(c.Validate, err),
		}, http.StatusBadRequest)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Formato do ID da área de trabalho inválido",
		}, http.StatusBadRequest)
	}

	transactionId, err := primitive.ObjectIDFromHex(body.TransactionId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Formato do ID da transação inválido",
		}, http.StatusBadRequest)
	}

	transactionFound, err := c.FindTransactionByIdRepository.Find(transactionId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Erro ao buscar a transação",
		}, http.StatusInternalServerError)
	}

	if transactionFound == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Transação não encontrada",
		}, http.StatusNotFound)
	}

	if transactionFound.Frequency != "REPEAT" && transactionFound.Frequency != "RECURRING" {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Frequência inválida, não é possível excluir parcelas de uma transação não recorrente",
		}, http.StatusBadRequest)
	}

	if body.Count == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Não é possível definir o número de parcelas de uma transação parcelada",
		}, http.StatusBadRequest)
	}

	err = c.DeleteEditTransactionRepository.DeleteAllAfterCount(transactionId, *body.Count, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Erro ao excluir as transações de edição",
		}, http.StatusInternalServerError)
	}

	transactionFound.RepeatSettings.Count = *body.Count

	transaction, err := c.UpdateTransactionRepository.Update(transactionId, transactionFound)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Erro ao atualizar a transação",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(transaction, http.StatusOK)
}
