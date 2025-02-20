package transaction

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/transaction_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/workspace_repository/member_repository"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateTransactionController struct {
	Validate                    *validator.Validate
	CreateTransactionRepository usecase.CreateTransactionRepository
	FindMemberByIdRepository    *member_repository.FindMemberByIdRepository
	FindAccountByIdRepository   usecase.FindAccountByIdRepository
	FindCategoryByIdRepository  usecase.FindCategoryByIdRepository
}

func NewCreateTransactionController(findMemberByIdRepository *member_repository.FindMemberByIdRepository, createTransactionRepository *transaction_repository.CreateTransactionRepository, findAccountByIdRepository usecase.FindAccountByIdRepository, findCategoryByIdRepository usecase.FindCategoryByIdRepository) *CreateTransactionController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &CreateTransactionController{
		Validate:                    validate,
		FindMemberByIdRepository:    findMemberByIdRepository,
		CreateTransactionRepository: createTransactionRepository,
		FindAccountByIdRepository:   findAccountByIdRepository,
		FindCategoryByIdRepository:  findCategoryByIdRepository,
	}
}

type CreateTransactionBody struct {
	Name        string `json:"name" validate:"required,min=3,max=30"`
	Description string `json:"description" validate:"min=3,max=255"`
	Type        string `json:"type" validate:"required,oneof=EXPENSE RECIPE"`
	Supplier    string `json:"supplier" validate:"required,min=3,max=30"`
	AssignedTo  string `json:"assignedTo" validate:"required,min=3,max=30,mongodb"`
	Balance     struct {
		Value    int `json:"value" validate:"required,min=0"`
		Parts    int `json:"parts" validate:"min=0"`
		Labor    int `json:"labor" validate:"min=0"`
		Discount int `json:"discount" validate:"min=0"`
		Interest int `json:"interest" validate:"min=0"`
	} `json:"balance" validate:"required"`
	Frequency      string `json:"frequency" validate:"oneof=DO_NOT_REPEAT RECURRING REPEAT"`
	RepeatSettings struct {
		InitialInstallment time.Month `json:"initialInstallment" validate:"min=1"`
		Count              int        `json:"count" validate:"min=2"`
		Interval           string     `json:"interval" validate:"oneof=DAILY WEEKLY MONTHLY QUARTERLY YEARLY"`
	} `json:"repeatSettings" validate:"excluded_if=Frequency DO_NOT_REPEAT,excluded_if=Frequency RECURRING,required_if=Frequency REPEAT"`
	DueDate          string `json:"dueDate" validate:"required,datetime=2006-01-02"`
	IsConfirmed      bool   `json:"isConfirmed"`
	CategoryId       string `json:"categoryId" validate:"required,mongodb"`
	SubCategoryId    string `json:"subCategoryId" validate:"required,mongodb"`
	TagId            string `json:"tagId" validate:"required,mongodb"`
	SubTagId         string `json:"subTagId" validate:"required,mongodb"`
	AccountId        string `json:"accountId" validate:"required,mongodb"`
	RegistrationDate string `json:"registrationDate" validate:"required,datetime=2006-01-02"`
	ConfirmationDate string `json:"confirmationDate" validate:"datetime=2006-01-02,excluded_if=IsConfirmed false,required_if=IsConfirmed true"`
}

func (c *CreateTransactionController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body CreateTransactionBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid body request",
		}, http.StatusBadRequest)
	}

	if err := c.Validate.Struct(body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid body validation: " + err.Error(),
		}, http.StatusBadRequest)
	}

	transaction, err := c.createTransaction(&body)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error creating transaction: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	userObjectID, err := primitive.ObjectIDFromHex(r.Header.Get("userId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid user ID format",
		}, http.StatusBadRequest)
	}
	transaction.CreatedBy = userObjectID

	assignedTo, err := primitive.ObjectIDFromHex(body.AssignedTo)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid assignedTo ID format",
		}, http.StatusBadRequest)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspace ID format",
		}, http.StatusBadRequest)
	}
	transaction.WorkspaceId = workspaceId

	if err := c.validateAssignedMember(workspaceId, assignedTo); err != nil {
		return err
	}
	transaction.AssignedTo = assignedTo

	log.Println(transaction.AccountId)
	log.Println(workspaceId)

	if err := c.validateAccount(workspaceId, transaction.AccountId); err != nil {
		return err
	}

	if err := c.validateCategory(workspaceId, transaction.CategoryId, transaction.Type, transaction.SubCategoryId); err != nil {
		return err
	}

	if err := c.validateTag(workspaceId, transaction.TagId, transaction.SubTagId); err != nil {
		return err
	}

	transaction, err = c.CreateTransactionRepository.Create(transaction)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error creating transaction",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(transaction, http.StatusCreated)
}

func (c *CreateTransactionController) createTransaction(body *CreateTransactionBody) (*models.Transaction, error) {
	convertID := func(id string) (primitive.ObjectID, error) {
		return primitive.ObjectIDFromHex(id)
	}

	parseDate := func(date string) (time.Time, error) {
		return time.Parse("2006-01-02", date)
	}

	categoryId, err := convertID(body.CategoryId)
	if err != nil {
		return nil, err
	}

	subCategoryId, err := convertID(body.SubCategoryId)
	if err != nil {
		return nil, err
	}

	tagId, err := convertID(body.TagId)
	if err != nil {
		return nil, err
	}

	subTagId, err := convertID(body.SubTagId)
	if err != nil {
		return nil, err
	}

	accountId, err := convertID(body.AccountId)
	if err != nil {
		return nil, err
	}

	registrationDate, err := parseDate(body.RegistrationDate)
	if err != nil {
		return nil, err
	}

	confirmationDate, err := parseDate(body.ConfirmationDate)
	if err != nil {
		return nil, err
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
		Balance: models.TransactionBalance{
			Value:    body.Balance.Value,
			Parts:    body.Balance.Parts,
			Labor:    body.Balance.Labor,
			Discount: body.Balance.Discount,
			Interest: body.Balance.Interest,
		},
		Frequency: body.Frequency,
		RepeatSettings: models.TransactionRepeatSettings{
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
		ConfirmationDate: confirmationDate,
		DueDate:          dueDate,
	}, nil
}

func (c *CreateTransactionController) validateAssignedMember(workspaceId primitive.ObjectID, assignedTo primitive.ObjectID) *presentationProtocols.HttpResponse {
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

func (c *CreateTransactionController) validateAccount(workspaceId primitive.ObjectID, accountId primitive.ObjectID) *presentationProtocols.HttpResponse {
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

func (c *CreateTransactionController) validateCategory(workspaceId primitive.ObjectID, categoryId primitive.ObjectID, transactionType string, subCategoryId primitive.ObjectID) *presentationProtocols.HttpResponse {
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

func (c *CreateTransactionController) validateTag(workspaceId primitive.ObjectID, categoryId primitive.ObjectID, subCategoryId primitive.ObjectID) *presentationProtocols.HttpResponse {
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
