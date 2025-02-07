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
	UpdateAccountRepository usecase.UpdateAccountRepository
	Validate                *validator.Validate
	FindBankById            usecase.FindBankByIdRepository
	FindAccountById         usecase.FindAccountByIdRepository
}

func NewUpdateAccountController(updateAccount usecase.UpdateAccountRepository, findBankById usecase.FindBankByIdRepository, findAccountById usecase.FindAccountByIdRepository) *UpdateAccountController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &UpdateAccountController{
		UpdateAccountRepository: updateAccount,
		Validate:                validate,
		FindBankById:            findBankById,
		FindAccountById:         findAccountById,
	}
}

type UpdateAccountControllerBody struct {
	Name    string  `validate:"required"`
	BankId  string  `validate:"required"`
	Balance float64 `validate:"required"`
}

func (c *UpdateAccountController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	id := r.Req.PathValue("id")
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
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
			Error: err.Error(),
		}, http.StatusUnprocessableEntity)
	}

	if _, err := primitive.ObjectIDFromHex(body.BankId); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid bank ID format",
		}, http.StatusBadRequest)
	}

	accountToVerify, err := c.FindAccountById.Find(id, r.Header.Get("workspaceId"))
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

	account, err := c.UpdateAccountRepository.Update(id, &models.AccountInput{
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
