package transaction

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/workspace_repository/member_repository"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UpdateTransactionController struct {
	UpdateTransactionRepository usecase.UpdateTransactionRepository
	Validate                    *validator.Validate
	FindTransactionById         usecase.FindTransactionByIdRepository
	FindMemberByIdRepository    *member_repository.FindMemberByIdRepository
	FindAccountByIdRepository   usecase.FindAccountByIdRepository
	FindCategoryByIdRepository  usecase.FindCategoryByIdRepository
}

func NewUpdateTransactionController(updateTransaction usecase.UpdateTransactionRepository, findTransactionById usecase.FindTransactionByIdRepository, findMemberByIdRepository *member_repository.FindMemberByIdRepository, findAccountByIdRepository usecase.FindAccountByIdRepository, findCategoryByIdRepository usecase.FindCategoryByIdRepository) *UpdateTransactionController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &UpdateTransactionController{
		UpdateTransactionRepository: updateTransaction,
		Validate:                    validate,
		FindTransactionById:         findTransactionById,
		FindMemberByIdRepository:    findMemberByIdRepository,
		FindAccountByIdRepository:   findAccountByIdRepository,
		FindCategoryByIdRepository:  findCategoryByIdRepository,
	}
}

func (c *UpdateTransactionController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body TransactionBody
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

	transactionId, err := primitive.ObjectIDFromHex(r.Req.PathValue("id"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid transactionId format",
		}, http.StatusBadRequest)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspaceId format",
		}, http.StatusBadRequest)
	}

	transaction, err := c.FindTransactionById.Find(transactionId, workspaceId)
	if transaction == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "transaction not found",
		}, http.StatusNotFound)
	}
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error finding transaction",
		}, http.StatusInternalServerError)
	}

	transaction.Name = body.Name
	transaction.Description = body.Description
	transaction.Type = body.Type
	transaction.Supplier = body.Supplier
	transaction.Balance = models.TransactionBalance{
		Value:    body.Balance.Value,
		Parts:    transaction.Balance.Parts,
		Labor:    transaction.Balance.Labor,
		Discount: transaction.Balance.Discount,
		Interest: transaction.Balance.Interest,
	}
	transaction.Frequency = body.Frequency
	transaction.RepeatSettings = &models.TransactionRepeatSettings{
		InitialInstallment: body.RepeatSettings.InitialInstallment,
		Count:              body.RepeatSettings.Count,
		Interval:           body.RepeatSettings.Interval,
	}

	transactionIdsParsed, err := c.createTransaction(&body)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error creating transaction",
		}, http.StatusInternalServerError)
	}

	transaction.TagId = transactionIdsParsed.TagId
	transaction.SubTagId = transactionIdsParsed.SubTagId
	transaction.AccountId = transactionIdsParsed.AccountId
	transaction.RegistrationDate = transactionIdsParsed.RegistrationDate
	transaction.ConfirmationDate = transactionIdsParsed.ConfirmationDate

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
		if err := c.validateAccount(workspaceId, transaction.AccountId); err != nil {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := c.validateCategory(workspaceId, transaction.CategoryId, transaction.Type, transaction.SubCategoryId); err != nil {
			errChan <- err
		}
	}()

	if transaction.TagId != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.validateTag(workspaceId, *transaction.TagId, *transaction.SubTagId); err != nil {
				errChan <- err
			}
		}()
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return <-errChan
	}

	transactionUpdated, err := c.UpdateTransactionRepository.Update(transactionId, transaction)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error updating transaction",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(transactionUpdated, http.StatusOK)
}

func (c *UpdateTransactionController) createTransaction(body *TransactionBody) (*models.Transaction, error) {
	convertID := func(id string) (primitive.ObjectID, error) {
		return primitive.ObjectIDFromHex(id)
	}

	parseDate := func(date string) (time.Time, error) {
		location := time.UTC
		return time.ParseInLocation("2006-01-02T15:04:05Z", date, location)
	}

	categoryId, err := convertID(body.CategoryId)
	if err != nil {
		return nil, err
	}

	subCategoryId, err := convertID(body.SubCategoryId)
	if err != nil {
		return nil, err
	}

	var tagId *primitive.ObjectID
	if body.TagId != "" {
		tagIdParsed, err := convertID(body.TagId)
		if err != nil {
			return nil, err
		}

		tagId = &tagIdParsed
	}

	var subTagId *primitive.ObjectID
	if body.SubTagId != "" {
		subTagIdParsed, err := convertID(body.SubTagId)
		if err != nil {
			return nil, err
		}

		subTagId = &subTagIdParsed
	}

	accountId, err := convertID(body.AccountId)
	if err != nil {
		return nil, err
	}

	assignedTo, err := convertID(body.AssignedTo)
	if err != nil {
		return nil, err
	}

	registrationDate, err := parseDate(body.RegistrationDate)
	if err != nil {
		return nil, err
	}

	var confirmationDate time.Time
	if body.IsConfirmed {
		confirmationDate, err = parseDate(*body.ConfirmationDate)
		if err != nil {
			return nil, err
		}
	}

	dueDate, err := parseDate(body.DueDate)
	if err != nil {
		return nil, err
	}

	return &models.Transaction{
		Name:        body.Name,
		Description: body.Description,
		Type:        body.Type,
		Supplier:    body.Supplier,
		AssignedTo:  assignedTo,
		Balance: models.TransactionBalance{
			Value:    body.Balance.Value,
			Parts:    body.Balance.Parts,
			Labor:    body.Balance.Labor,
			Discount: body.Balance.Discount,
			Interest: body.Balance.Interest,
		},
		Frequency: body.Frequency,
		RepeatSettings: &models.TransactionRepeatSettings{
			InitialInstallment: body.RepeatSettings.InitialInstallment,
			Count:              body.RepeatSettings.Count,
			Interval:           body.RepeatSettings.Interval,
		},
		IsConfirmed:      body.IsConfirmed,
		CategoryId:       categoryId,
		SubCategoryId:    subCategoryId,
		TagId:            tagId,
		SubTagId:         subTagId,
		AccountId:        accountId,
		RegistrationDate: registrationDate,
		ConfirmationDate: &confirmationDate,
		DueDate:          dueDate,
	}, nil
}

func (c *UpdateTransactionController) validateAssignedMember(workspaceId primitive.ObjectID, assignedTo primitive.ObjectID) *presentationProtocols.HttpResponse {
	member, err := c.FindMemberByIdRepository.Find(workspaceId, assignedTo)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error finding member",
		}, http.StatusInternalServerError)
	}

	if member == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "AssignedTo is not a member of the workspace",
		}, http.StatusNotFound)
	}

	return nil
}

func (c *UpdateTransactionController) validateAccount(workspaceId primitive.ObjectID, accountId primitive.ObjectID) *presentationProtocols.HttpResponse {
	account, err := c.FindAccountByIdRepository.Find(accountId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error finding account",
		}, http.StatusInternalServerError)
	}

	if account == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "account not found",
		}, http.StatusNotFound)
	}

	return nil
}

func (c *UpdateTransactionController) validateCategory(workspaceId primitive.ObjectID, categoryId primitive.ObjectID, transactionType string, subCategoryId primitive.ObjectID) *presentationProtocols.HttpResponse {
	category, err := c.FindCategoryByIdRepository.Find(categoryId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error finding category",
		}, http.StatusInternalServerError)
	}

	if category == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "category not found",
		}, http.StatusNotFound)
	}

	if !strings.EqualFold(category.Type, transactionType) {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "category type does not match transaction type",
		}, http.StatusBadRequest)
	}

	for _, subCategory := range category.SubCategories {
		if subCategory.Id == subCategoryId {
			return nil
		}
	}

	return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
		Error: "sub category not found",
	}, http.StatusNotFound)
}

func (c *UpdateTransactionController) validateTag(workspaceId primitive.ObjectID, categoryId primitive.ObjectID, subCategoryId primitive.ObjectID) *presentationProtocols.HttpResponse {
	category, err := c.FindCategoryByIdRepository.Find(categoryId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error finding tag",
		}, http.StatusInternalServerError)
	}

	if category == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "tag not found",
		}, http.StatusNotFound)
	}

	if !strings.EqualFold(category.Type, "TAG") {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "tag type does not match transaction type",
		}, http.StatusBadRequest)
	}

	for _, subCategory := range category.SubCategories {
		if subCategory.Id == subCategoryId {
			return nil
		}
	}

	return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
		Error: "sub tag not found",
	}, http.StatusNotFound)
}
