package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransferenceAccountController struct {
	FindAccountByIdRepository   usecase.FindAccountByIdRepository
	CreateTransactionRepository usecase.CreateTransactionRepository
	Validate                    *validator.Validate
}

func NewTransferenceAccountController(
	findAccountById usecase.FindAccountByIdRepository,
	createTransaction usecase.CreateTransactionRepository,
) *TransferenceAccountController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &TransferenceAccountController{
		FindAccountByIdRepository:   findAccountById,
		CreateTransactionRepository: createTransaction,
		Validate:                    validate,
	}
}

type TransferenceAccountControllerBody struct {
	SourceAccountId      string  `json:"sourceAccountId" validate:"required"`
	DestinationAccountId string  `json:"destinationAccountId" validate:"required"`
	Amount               float64 `json:"amount" validate:"required,gt=0"`
}

type TransferenceAccountControllerResponse struct {
	SourceAccount      *models.Account     `json:"sourceAccount"`
	DestinationAccount *models.Account     `json:"destinationAccount"`
	ExpenseTransaction *models.Transaction `json:"expenseTransaction"`
	ReceiptTransaction *models.Transaction `json:"receiptTransaction"`
}

func (c *TransferenceAccountController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body TransferenceAccountControllerBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid body request",
		}, http.StatusBadRequest)
	}

	if err := c.Validate.Struct(body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: helpers.GetErrorMessages(c.Validate, err),
		}, http.StatusUnprocessableEntity)
	}

	// Get workspaceId from header
	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid workspace ID format",
		}, http.StatusBadRequest)
	}

	// Validate and convert IDs to ObjectID
	sourceAccountId, err := primitive.ObjectIDFromHex(body.SourceAccountId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid source account ID format",
		}, http.StatusBadRequest)
	}

	destinationAccountId, err := primitive.ObjectIDFromHex(body.DestinationAccountId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid destination account ID format",
		}, http.StatusBadRequest)
	}

	// Ensure accounts are different
	if sourceAccountId == destinationAccountId {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Source and destination accounts must be different",
		}, http.StatusBadRequest)
	}

	// Find source account
	sourceAccount, err := c.FindAccountByIdRepository.Find(sourceAccountId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "An error occurred when finding source account",
		}, http.StatusInternalServerError)
	}

	if sourceAccount == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Source account not found",
		}, http.StatusNotFound)
	}

	// Find destination account
	destinationAccount, err := c.FindAccountByIdRepository.Find(destinationAccountId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "An error occurred when finding destination account",
		}, http.StatusInternalServerError)
	}

	if destinationAccount == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Destination account not found",
		}, http.StatusNotFound)
	}

	userId, err := primitive.ObjectIDFromHex(r.Header.Get("userId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid assignedTo ID format",
		}, http.StatusBadRequest)
	}

	// Get current time for transactions
	now := time.Now().UTC()

	// Create expense transaction for source account
	expenseTransaction := &models.Transaction{
		Id:         primitive.NewObjectID(),
		Name:       "Transferência interna para " + destinationAccount.Name,
		IsDeleted:  false,
		CreatedBy:  workspaceId, // Using workspaceId as createdBy since we don't have user info
		Type:       "EXPENSE",
		Supplier:   "Transferência interna",
		AssignedTo: userId, // Using workspaceId as assignedTo since we don't have user info
		Balance: models.TransactionBalance{
			Value:      body.Amount,
			Discount:   0,
			Interest:   0,
			NetBalance: body.Amount,
		},
		Frequency:        "DO_NOT_REPEAT",
		DueDate:          now,
		IsConfirmed:      true,
		AccountId:        &sourceAccountId,
		RegistrationDate: now,
		ConfirmationDate: &now,
		CreatedAt:        now,
		UpdatedAt:        now,
		WorkspaceId:      workspaceId,
	}

	// Create receipt transaction for destination account
	receiptTransaction := &models.Transaction{
		Id:         primitive.NewObjectID(),
		Name:       "Transferência interna de " + sourceAccount.Name,
		IsDeleted:  false,
		CreatedBy:  workspaceId, // Using workspaceId as createdBy since we don't have user info
		Type:       "RECIPE",
		Supplier:   "Transferência interna",
		AssignedTo: userId, // Using workspaceId as assignedTo since we don't have user info
		Balance: models.TransactionBalance{
			Value:      body.Amount,
			Discount:   0,
			Interest:   0,
			NetBalance: body.Amount,
		},
		Frequency:        "DO_NOT_REPEAT",
		DueDate:          now,
		IsConfirmed:      true,
		AccountId:        &destinationAccountId,
		RegistrationDate: now,
		ConfirmationDate: &now,
		CreatedAt:        now,
		UpdatedAt:        now,
		WorkspaceId:      workspaceId,
	}

	// Create expense transaction
	createdExpenseTransaction, err := c.CreateTransactionRepository.Create(expenseTransaction)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "An error occurred when creating expense transaction",
		}, http.StatusInternalServerError)
	}

	// Create receipt transaction
	createdReceiptTransaction, err := c.CreateTransactionRepository.Create(receiptTransaction)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "An error occurred when creating receipt transaction",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(&TransferenceAccountControllerResponse{
		ExpenseTransaction: createdExpenseTransaction,
		ReceiptTransaction: createdReceiptTransaction,
	}, http.StatusOK)
}
