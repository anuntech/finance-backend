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
	FindAccountByNameRepository        usecase.FindAccountByNameAndWorkspaceIdRepository
}

func NewImportAccountController(
	importUseCase usecase.ImportAccountsRepository,
	findAccounts usecase.FindAccountByWorkspaceIdRepository,
	findByNameRepository usecase.FindAccountByNameAndWorkspaceIdRepository,
) *ImportAccountController {
	validate := validator.New()

	return &ImportAccountController{
		ImportAccountsRepository:           importUseCase,
		Validate:                           validate,
		FindAccountByWorkspaceIdRepository: findAccounts,
		FindAccountByNameRepository:        findByNameRepository,
	}
}

type ImportAccount struct {
	Name    string  `json:"name" validate:"required,min=3,max=255"`
	Balance float64 `json:"balance" validate:"required"`
	BankId  string  `json:"bankId" validate:"required"`
}

type ImportAccountBody struct {
	Accounts []ImportAccount `json:"accounts" validate:"required,dive"`
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

	currentAccounts, err := c.FindAccountByWorkspaceIdRepository.Find(&helpers.GlobalFilterParams{
		WorkspaceId: workspaceId,
	})
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

	// Verificar duplicidade de nomes entre as contas a serem importadas
	nameSet := make(map[string]bool)
	for _, acc := range body.Accounts {
		if nameSet[acc.Name] {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "a importação contém contas com nomes duplicados: " + acc.Name,
			}, http.StatusBadRequest)
		}
		nameSet[acc.Name] = true

		// Verificar se o nome já existe no workspace
		existingAccount, err := c.FindAccountByNameRepository.FindByNameAndWorkspaceId(acc.Name, workspaceId)
		if err != nil {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "erro ao verificar duplicidade de nomes: " + err.Error(),
			}, http.StatusInternalServerError)
		}

		if existingAccount != nil {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "já existe uma conta com o nome '" + acc.Name + "' neste workspace",
			}, http.StatusConflict)
		}
	}

	var Accounts []models.Account
	for _, acc := range body.Accounts {
		bankId, err := primitive.ObjectIDFromHex(acc.BankId)
		if err != nil {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "formato de bankId inválido",
			}, http.StatusBadRequest)
		}

		Accounts = append(Accounts, models.Account{
			Name:        acc.Name,
			Balance:     acc.Balance,
			BankId:      bankId,
			WorkspaceId: workspaceId,
		})
	}

	importedAccounts, err := c.ImportAccountsRepository.Import(Accounts, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "erro ao importar contas: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(importedAccounts, http.StatusOK)
}
