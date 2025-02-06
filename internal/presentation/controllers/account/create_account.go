package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
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
	Id          string `json:"id"`
	Name        string `json:"name" validate:"required"`
	WorkspaceId string `json:"workspaceId" validate:"required"`
	Bank        string `json:"bank" validate:"required"`
}

type CreateAccountControllerBody struct {
	Name string `validate:"required"`
	Bank string `validate:"required"`
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

	bank, err := c.FindBankById.Find(body.Bank)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error ocurred when finding bank",
		}, http.StatusInternalServerError)
	}

	if bank == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "bank not found",
		}, http.StatusNotFound)
	}

	accounts, err := c.FindAccountByWorkspaceIdRepository.Find(r.Header.Get("userId"), r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error ocurred when finding accounts",
		}, http.StatusInternalServerError)
	}

	if len(accounts) >= 50 {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "user has reached the maximum number of accounts",
		}, http.StatusBadRequest)
	}

	account, err := c.CreateAccountRepository.Create(&models.AccountInput{
		Name:        body.Name,
		Bank:        body.Bank,
		WorkspaceId: r.Header.Get("workspaceId"),
	})

	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error ocurred when creating account",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(&CreateAccountControllerResponse{
		Id:          account.Id,
		Name:        account.Name,
		WorkspaceId: account.WorkspaceId,
		Bank:        account.Bank,
	}, http.StatusOK)
}
