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

type ImportAccountController struct {
	ImportAccountsRepository           usecase.ImportAccountsRepository
	Validate                           *validator.Validate
	FindAccountByWorkspaceIdRepository usecase.FindAccountByWorkspaceIdRepository
}

func NewImportAccountController(importUseCase usecase.ImportAccountsRepository, findAccounts usecase.FindAccountByWorkspaceIdRepository) *ImportAccountController {
	validate := validator.New()

	return &ImportAccountController{
		ImportAccountsRepository:           importUseCase,
		Validate:                           validate,
		FindAccountByWorkspaceIdRepository: findAccounts,
	}
}

type ImportAccountBody struct {
	Accounts []models.ImportAccountInput `json:"accounts" validate:"required,dive"`
}

func (c *ImportAccountController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body ImportAccountBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "corpo da requisição inválido",
		}, http.StatusBadRequest)
	}

	if err := c.Validate.Struct(body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: helpers.GetErrorMessages(c.Validate, err),
		}, http.StatusUnprocessableEntity)
	}

	workspaceIdStr := r.Header.Get("workspaceId")
	workspaceId, err := primitive.ObjectIDFromHex(workspaceIdStr)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "formato de workspaceId inválido",
		}, http.StatusBadRequest)
	}

	currentAccounts, err := c.FindAccountByWorkspaceIdRepository.Find(workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "erro ao buscar contas existentes",
		}, http.StatusInternalServerError)
	}

	if len(currentAccounts)+len(body.Accounts) > 50 {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "importação excede o número máximo de contas permitidas (50)",
		}, http.StatusBadRequest)
	}

	var accountInputs []models.AccountInput
	for _, acc := range body.Accounts {
		bankId, err := primitive.ObjectIDFromHex(acc.BankId)
		if err != nil {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "formato de bankId inválido",
			}, http.StatusBadRequest)
		}

		accountInputs = append(accountInputs, models.AccountInput{
			Name:        acc.Name,
			Balance:     acc.Balance,
			BankId:      bankId,
			WorkspaceId: workspaceId,
		})
	}

	importedAccounts, err := c.ImportAccountsRepository.Import(accountInputs, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "erro ao importar contas: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(importedAccounts, http.StatusOK)
}
