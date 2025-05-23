package transaction

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
	infraHelpers "github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/workspace_repository/member_repository"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/anuntech/finance-backend/internal/utils"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UpdateTransactionController struct {
	UpdateTransactionRepository   usecase.UpdateTransactionRepository
	Validate                      *validator.Validate
	FindTransactionById           usecase.FindTransactionByIdRepository
	FindMemberByIdRepository      *member_repository.FindMemberByIdRepository
	FindAccountByIdRepository     usecase.FindAccountByIdRepository
	FindCategoryByIdRepository    usecase.FindCategoryByIdRepository
	FindCustomFieldByIdRepository usecase.FindCustomFieldByIdRepository
}

func NewUpdateTransactionController(updateTransaction usecase.UpdateTransactionRepository, findTransactionById usecase.FindTransactionByIdRepository, findMemberByIdRepository *member_repository.FindMemberByIdRepository, findAccountByIdRepository usecase.FindAccountByIdRepository, findCategoryByIdRepository usecase.FindCategoryByIdRepository, findCustomFieldByIdRepository usecase.FindCustomFieldByIdRepository) *UpdateTransactionController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &UpdateTransactionController{
		UpdateTransactionRepository:   updateTransaction,
		Validate:                      validate,
		FindTransactionById:           findTransactionById,
		FindMemberByIdRepository:      findMemberByIdRepository,
		FindAccountByIdRepository:     findAccountByIdRepository,
		FindCategoryByIdRepository:    findCategoryByIdRepository,
		FindCustomFieldByIdRepository: findCustomFieldByIdRepository,
	}
}

func (c *UpdateTransactionController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body TransactionBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "formato da solicitação inválido",
		}, http.StatusBadRequest)
	}

	if err := c.Validate.Struct(body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: helpers.GetErrorMessages(c.Validate, err),
		}, http.StatusUnprocessableEntity)
	}

	transactionId, err := primitive.ObjectIDFromHex(r.Req.PathValue("id"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "formato do ID da transação inválido",
		}, http.StatusBadRequest)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "formato do ID da área de trabalho inválido",
		}, http.StatusBadRequest)
	}

	transaction, err := c.FindTransactionById.Find(transactionId, workspaceId)
	if transaction == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "transação não encontrada",
		}, http.StatusNotFound)
	}
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "erro ao buscar a transação",
		}, http.StatusInternalServerError)
	}

	transaction.Name = body.Name
	transaction.Description = body.Description
	transaction.Type = body.Type
	transaction.Supplier = body.Supplier
	transaction.Balance = models.TransactionBalance{
		Value:              body.Balance.Value,
		Discount:           body.Balance.Discount,
		Interest:           body.Balance.Interest,
		DiscountPercentage: body.Balance.DiscountPercentage,
		InterestPercentage: body.Balance.InterestPercentage,
	}
	transaction.Frequency = body.Frequency
	transaction.RepeatSettings = &models.TransactionRepeatSettings{
		InitialInstallment: body.RepeatSettings.InitialInstallment,
		Count:              body.RepeatSettings.Count,
		Interval:           body.RepeatSettings.Interval,
	}

	transactionIdsParsed, err := createTransaction(&body)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "erro ao criar a transação",
		}, http.StatusInternalServerError)
	}

	transaction.Tags = transactionIdsParsed.Tags
	transaction.AccountId = transactionIdsParsed.AccountId
	transaction.RegistrationDate = transactionIdsParsed.RegistrationDate
	transaction.DueDate = transactionIdsParsed.DueDate
	transaction.ConfirmationDate = transactionIdsParsed.ConfirmationDate
	transaction.IsConfirmed = transactionIdsParsed.IsConfirmed
	transaction.CustomFields = transactionIdsParsed.CustomFields
	transaction.CategoryId = transactionIdsParsed.CategoryId
	transaction.SubCategoryId = transactionIdsParsed.SubCategoryId
	errChan := make(chan *presentationProtocols.HttpResponse, 4)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := c.validateAssignedMember(workspaceId, transactionIdsParsed.AssignedTo); err != nil {
			errChan <- err
			return
		}
		transaction.AssignedTo = transactionIdsParsed.AssignedTo
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer utils.RecoveryWithCallback(&wg, func(r interface{}) {
			errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "erro na validação da conta: ocorreu um erro inesperado",
			}, http.StatusInternalServerError)
		})

		if transaction.AccountId == nil {
			return
		}
		if err := c.validateAccount(workspaceId, *transaction.AccountId); err != nil {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer utils.RecoveryWithCallback(&wg, func(r interface{}) {
			errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "erro na validação da categoria: ocorreu um erro inesperado",
			}, http.StatusInternalServerError)
		})
		if transaction.CategoryId == nil {
			return
		}

		if transaction.SubCategoryId == nil {
			errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "subcategoria não encontrada",
			}, http.StatusNotFound)
			return
		}

		if err := c.validateCategory(workspaceId, *transaction.CategoryId, transaction.Type, *transaction.SubCategoryId); err != nil {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer utils.RecoveryWithCallback(&wg, func(r any) {
			errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "erro na validação dos campos personalizados: ocorreu um erro inesperado",
			}, http.StatusInternalServerError)
		})

		defer utils.RecoveryWithCallback(&wg, func(r any) {
			errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "erro ao buscar campo personalizado",
			}, http.StatusInternalServerError)
		})

		seenCustomFields := make(map[string]bool)

		for _, customField := range transaction.CustomFields {
			compositeKey := customField.CustomFieldId.Hex()

			if seenCustomFields[compositeKey] {
				errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "campo personalizado duplicado detectado: " + compositeKey,
				}, http.StatusBadRequest)
				return
			}
			seenCustomFields[compositeKey] = true

			customFieldParsed, err := c.FindCustomFieldByIdRepository.Find(customField.CustomFieldId, workspaceId)
			if err != nil {
				errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "erro ao buscar campo personalizado",
				}, http.StatusInternalServerError)
				return
			}

			if customFieldParsed == nil {
				errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "campo personalizado não encontrado",
				}, http.StatusNotFound)
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer utils.RecoveryWithCallback(&wg, func(r any) {
			errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "erro ao buscar campo personalizado",
			}, http.StatusInternalServerError)
		})
		seenTags := make(map[string]bool)

		for _, tag := range transaction.Tags {
			compositeKey := tag.TagId.Hex() + "|" + tag.SubTagId.Hex()

			if seenTags[compositeKey] {
				errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "tag duplicada detectada: " + compositeKey,
				}, http.StatusBadRequest)
				return
			}
			seenTags[compositeKey] = true

			if err := c.validateTag(workspaceId, tag.TagId, tag.SubTagId); err != nil {
				errChan <- err
				return
			}
		}
	}()

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return <-errChan
	}

	transactionUpdated, err := c.UpdateTransactionRepository.Update(transactionId, transaction)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "erro ao atualizar a transação",
		}, http.StatusInternalServerError)
	}

	recipeTx := *transactionUpdated
	recipeTx.Type = "RECIPE"
	recipeNetBalance := infraHelpers.CalculateOneTransactionBalance(&recipeTx)
	transactionUpdated.Balance.NetBalance = recipeNetBalance

	return helpers.CreateResponse(transactionUpdated, http.StatusOK)
}

func (c *UpdateTransactionController) validateAssignedMember(workspaceId primitive.ObjectID, assignedTo primitive.ObjectID) *presentationProtocols.HttpResponse {
	member, err := c.FindMemberByIdRepository.Find(workspaceId, assignedTo)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "erro ao buscar membro",
		}, http.StatusInternalServerError)
	}

	if member == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "o responsável não é um membro da área de trabalho",
		}, http.StatusNotFound)
	}

	return nil
}

func (c *UpdateTransactionController) validateAccount(workspaceId primitive.ObjectID, accountId primitive.ObjectID) *presentationProtocols.HttpResponse {
	account, err := c.FindAccountByIdRepository.Find(accountId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "erro ao buscar conta",
		}, http.StatusInternalServerError)
	}

	if account == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "conta não encontrada",
		}, http.StatusNotFound)
	}

	return nil
}

func (c *UpdateTransactionController) validateCategory(workspaceId primitive.ObjectID, categoryId primitive.ObjectID, transactionType string, subCategoryId primitive.ObjectID) *presentationProtocols.HttpResponse {
	category, err := c.FindCategoryByIdRepository.Find(categoryId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "erro ao buscar categoria",
		}, http.StatusInternalServerError)
	}

	if category == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "categoria não encontrada",
		}, http.StatusNotFound)
	}

	if !strings.EqualFold(category.Type, transactionType) {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "o tipo da categoria não corresponde ao tipo da transação",
		}, http.StatusBadRequest)
	}

	for _, subCategory := range category.SubCategories {
		if subCategory.Id == subCategoryId {
			return nil
		}
	}

	return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
		Error: "subcategoria não encontrada",
	}, http.StatusNotFound)
}

func (c *UpdateTransactionController) validateTag(workspaceId primitive.ObjectID, categoryId primitive.ObjectID, subCategoryId primitive.ObjectID) *presentationProtocols.HttpResponse {
	category, err := c.FindCategoryByIdRepository.Find(categoryId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "erro ao buscar tag",
		}, http.StatusInternalServerError)
	}

	if category == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "tag não encontrada",
		}, http.StatusNotFound)
	}

	if !strings.EqualFold(category.Type, "TAG") {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "o tipo da tag não corresponde ao tipo da transação",
		}, http.StatusBadRequest)
	}

	for _, subCategory := range category.SubCategories {
		if subCategory.Id == subCategoryId {
			return nil
		}
	}

	if subCategoryId == primitive.NilObjectID {
		return nil
	}

	return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
		Error: "subtag não encontrada",
	}, http.StatusNotFound)
}
