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

type UpdateAccountController struct {
	UpdateAccountRepository     usecase.UpdateAccountRepository
	Validate                    *validator.Validate
	FindBankById                usecase.FindBankByIdRepository
	FindAccountById             usecase.FindAccountByIdRepository
	FindAccountByNameRepository usecase.FindAccountByNameAndWorkspaceIdRepository
}

func NewUpdateAccountController(
	updateAccount usecase.UpdateAccountRepository,
	findBankById usecase.FindBankByIdRepository,
	findAccountById usecase.FindAccountByIdRepository,
	findByNameRepository usecase.FindAccountByNameAndWorkspaceIdRepository,
) *UpdateAccountController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &UpdateAccountController{
		UpdateAccountRepository:     updateAccount,
		Validate:                    validate,
		FindBankById:                findBankById,
		FindAccountById:             findAccountById,
		FindAccountByNameRepository: findByNameRepository,
	}
}

type UpdateAccountControllerBody struct {
	Name    string  `validate:"required"`
	BankId  string  `validate:"required"`
	Balance float64 `validate:"min=0,max=1000000000000000000"`
}

func (c *UpdateAccountController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	id, err := primitive.ObjectIDFromHex(r.Req.PathValue("id"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid account ID format",
		}, http.StatusBadRequest)
	}

	var body UpdateAccountControllerBody
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

	if _, err := primitive.ObjectIDFromHex(body.BankId); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid bank ID format",
		}, http.StatusBadRequest)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid workspace ID format",
		}, http.StatusBadRequest)
	}

	accountToVerify, err := c.FindAccountById.Find(id, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when finding account",
		}, http.StatusInternalServerError)
	}

	if accountToVerify == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "account not found",
		}, http.StatusNotFound)
	}

	// Verificar se já existe outra conta com o mesmo nome neste workspace
	// que não seja a conta sendo atualizada
	if accountToVerify.Name != body.Name {
		existingAccount, err := c.FindAccountByNameRepository.FindByNameAndWorkspaceId(body.Name, workspaceId)
		if err != nil {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "an error occurred when checking for account name",
			}, http.StatusInternalServerError)
		}

		if existingAccount != nil && existingAccount.Id != id {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "an account with this name already exists in this workspace",
			}, http.StatusConflict)
		}
	}

	bank, err := c.FindBankById.Find(body.BankId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when finding bank",
		}, http.StatusInternalServerError)
	}

	if bank == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "bank not found",
		}, http.StatusNotFound)
	}

	account, err := c.UpdateAccountRepository.Update(id, &models.Account{
		Name:    body.Name,
		Balance: body.Balance,
		BankId:  bank.Id,
	})

	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when updating account",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(account, http.StatusOK)
}
