package transaction

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ExcludeInstallmentsUntilController struct {
	UpdateTransactionRepository   usecase.UpdateTransactionRepository
	Validate                      *validator.Validate
	FindTransactionByIdRepository usecase.FindTransactionByIdRepository
}

func NewExcludeInstallmentsUntilController(
	updateTransaction usecase.UpdateTransactionRepository,
	findTransactionById usecase.FindTransactionByIdRepository,
) *ExcludeInstallmentsUntilController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &ExcludeInstallmentsUntilController{
		UpdateTransactionRepository:   updateTransaction,
		Validate:                      validate,
		FindTransactionByIdRepository: findTransactionById,
	}
}

type ExcludeInstallmentsUntilBody struct {
	TransactionId string  `json:"transactionId" validate:"required,mongodb"`
	Until         *string `json:"until" validate:"omitempty,datetime=2006-01-02T15:04:05Z"`
	Count         *int    `json:"count" validate:"omitempty,min=1"`
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
			Error: "Transação não encontrada",
		}, http.StatusNotFound)
	}

	switch transactionFound.Frequency {
	case "REPEAT":
		if body.Count != nil {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "Não é possível definir o número de parcelas de uma transação parcelada",
			}, http.StatusBadRequest)
		}

		transactionFound.RepeatSettings.Count = *body.Count
	case "RECURRING":
		until, err := time.Parse("2006-01-02T15:04:05Z", *body.Until)
		if err != nil {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "Formato da data inválido",
			}, http.StatusBadRequest)
		}

		if body.Until != nil {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "Não é possível definir o número de parcelas de uma transação recorrente",
			}, http.StatusBadRequest)
		}

		transactionFound.ExcludeInstallmentsUntil = &until
	default:
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Frequência inválida, não é possível excluir parcelas de uma transação não recorrente",
		}, http.StatusBadRequest)
	}

	transaction, err := c.UpdateTransactionRepository.Update(transactionId, transactionFound)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Erro ao atualizar a transação",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(transaction, http.StatusOK)
}
