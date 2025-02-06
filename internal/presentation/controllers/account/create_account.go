package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateAccountController struct {
	CreateAccountRepository            usecase.CreateAccountRepository
	Validate                           *validator.Validate
	FindAccountByWorkspaceIdRepository usecase.FindAccountByWorkspaceIdRepository
	FindBankById                       usecase.FindBankByIdRepository
}

func NewCreateAccountController(createAccount usecase.CreateAccountRepository, findManyByUserIdAndWorkspaceId usecase.FindAccountByWorkspaceIdRepository, findBankById usecase.FindBankByIdRepository) *CreateAccountController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &CreateAccountController{
		CreateAccountRepository:            createAccount,
		FindAccountByWorkspaceIdRepository: findManyByUserIdAndWorkspaceId,
		Validate:                           validate,
		FindBankById:                       findBankById,
	}
}

type CreateAccountControllerResponse struct {
	Id          string  `json:"id"`
	Name        string  `json:"name" validate:"required"`
	WorkspaceId string  `json:"workspaceId" validate:"required"`
	BankId      string  `json:"bankId" validate:"required"`
	Balance     float64 `json:"balance" validate:"required"`
}

type CreateAccountControllerBody struct {
	Name    string  `validate:"required,min=3,max=255"`
	Balance float64 `validate:"required"`
	BankId  string  `validate:"required"`
}

func (c *CreateAccountController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body CreateAccountControllerBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid body request",
		}, http.StatusBadRequest)
	}

	if err := c.Validate.Struct(body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: err.Error(),
		}, http.StatusUnprocessableEntity)
	}

	bankId, err := primitive.ObjectIDFromHex(body.BankId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid bank ID format",
		}, http.StatusBadRequest)
	}

	bank, err := c.FindBankById.Find(bankId.Hex())
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "bank not found",
		}, http.StatusNotFound)
	}

	if bank == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "bank not found",
		}, http.StatusNotFound)
	}

	accounts, err := c.FindAccountByWorkspaceIdRepository.Find(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error ocurred when finding accounts: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	if len(accounts) >= 50 {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "user has reached the maximum number of accounts",
		}, http.StatusBadRequest)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid workspace ID format",
		}, http.StatusBadRequest)
	}

	account, err := c.CreateAccountRepository.Create(&models.AccountInput{
		Name:        body.Name,
		BankId:      bankId,
		Balance:     body.Balance,
		WorkspaceId: workspaceId,
	})

	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when creating account",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(&CreateAccountControllerResponse{
		Id:          account.Id.Hex(),
		Name:        account.Name,
		WorkspaceId: account.WorkspaceId.Hex(),
		BankId:      account.BankId.Hex(),
		Balance:     account.Balance,
	}, http.StatusOK)
}
