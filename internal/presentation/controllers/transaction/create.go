package transaction

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/transaction_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/workspace_repository/member_repository"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateTransactionController struct {
	Validate                    *validator.Validate
	Translator                  ut.Translator
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

type TransactionBody struct {
	Name        string `json:"name" validate:"required,min=2,max=30"`
	Description string `json:"description" validate:"omitempty,max=255"`
	Invoice     string `json:"invoice" validate:"omitempty,min=2,max=50"`
	Type        string `json:"type" validate:"required,oneof=EXPENSE RECIPE"`
	Supplier    string `json:"supplier" validate:"required,min=3,max=30"`
	AssignedTo  string `json:"assignedTo" validate:"required,min=3,max=30,mongodb"`
	Balance     struct {
		Value              float64 `json:"value" validate:"required,min=0.01"`
		Parts              float64 `json:"parts" validate:"omitempty,min=0.01"`
		Labor              float64 `json:"labor" validate:"omitempty,min=0.01"`
		Discount           float64 `json:"discount" validate:"omitempty,min=0.01"`
		Interest           float64 `json:"interest" validate:"omitempty,min=0.01"`
		DiscountPercentage float64 `json:"discountPercentage" validate:"omitempty,min=0.01,max=100"`
		InterestPercentage float64 `json:"interestPercentage" validate:"omitempty,min=0.01,max=100"`
	} `json:"balance" validate:"required"`
	Frequency      string `json:"frequency" validate:"oneof=DO_NOT_REPEAT RECURRING REPEAT"`
	RepeatSettings struct {
		InitialInstallment time.Month `json:"initialInstallment" validate:"min=1"`
		Count              int        `json:"count" validate:"min=2"`
		Interval           string     `json:"interval" validate:"oneof=DAILY WEEKLY MONTHLY QUARTERLY YEARLY"`
	} `json:"repeatSettings" validate:"excluded_if=Frequency DO_NOT_REPEAT,excluded_if=Frequency RECURRING,required_if=Frequency REPEAT,omitempty"`
	DueDate       string  `json:"dueDate" validate:"required,datetime=2006-01-02T15:04:05Z"`
	IsConfirmed   bool    `json:"isConfirmed"`
	CategoryId    *string `json:"categoryId" validate:"omitempty,mongodb"`
	SubCategoryId *string `json:"subCategoryId" validate:"omitempty,mongodb"`
	Tags          []struct {
		TagId    string `json:"tagId" validate:"omitempty,mongodb"`
		SubTagId string `json:"subTagId" validate:"required_with=TagId,excluded_if=TagId '',omitempty,mongodb"`
	} `json:"tags" validate:"omitempty"`
	AccountId        string  `json:"accountId" validate:"required,mongodb"`
	RegistrationDate string  `json:"registrationDate" validate:"required,datetime=2006-01-02T15:04:05Z"`
	ConfirmationDate *string `json:"confirmationDate" validate:"excluded_if=IsConfirmed false,required_if=IsConfirmed true,omitempty,datetime=2006-01-02T15:04:05Z"`
}

func (c *CreateTransactionController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body TransactionBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid body request",
		}, http.StatusBadRequest)
	}

	if err := c.Validate.Struct(body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: helpers.GetErrorMessages(c.Validate, err),
		}, http.StatusBadRequest)
	}

	transaction, err := createTransaction(&body)
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

	errChan := make(chan *presentationProtocols.HttpResponse, 4)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := c.validateAssignedMember(workspaceId, assignedTo); err != nil {
			errChan <- err
			return
		}
		transaction.AssignedTo = assignedTo
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
		if transaction.CategoryId == nil {
			return
		}
		if err := c.validateCategory(workspaceId, *transaction.CategoryId, transaction.Type, *transaction.SubCategoryId); err != nil {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		seenTags := make(map[string]bool)

		for _, tag := range transaction.Tags {
			compositeKey := tag.TagId.Hex() + "|" + tag.SubTagId.Hex()

			if seenTags[compositeKey] {
				errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "duplicate tag detected: " + compositeKey,
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

	transaction, err = c.CreateTransactionRepository.Create(transaction)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error creating transaction",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(transaction, http.StatusCreated)
}

func createTransaction(body *TransactionBody) (*models.Transaction, error) {
	convertID := func(id string) (primitive.ObjectID, error) {
		return primitive.ObjectIDFromHex(id)
	}

	parseDate := func(date string) (time.Time, error) {
		location := time.UTC
		return time.ParseInLocation("2006-01-02T15:04:05Z", date, location)
	}

	var categoryId *primitive.ObjectID
	if body.CategoryId != nil {
		categoryIdParsed, err := convertID(*body.CategoryId)
		if err != nil {
			return nil, err
		}

		categoryId = &categoryIdParsed
	}

	var subCategoryId *primitive.ObjectID
	if body.SubCategoryId != nil {
		subCategoryIdParsed, err := convertID(*body.SubCategoryId)
		if err != nil {
			return nil, err
		}

		subCategoryId = &subCategoryIdParsed
	}

	var tags = []models.TransactionTags{}

	for _, tag := range body.Tags {
		var tagId primitive.ObjectID
		if tag.TagId != "" {
			tagIdParsed, err := convertID(tag.TagId)
			if err != nil {
				return nil, err
			}

			tagId = tagIdParsed
		}

		var subTagId primitive.ObjectID
		if tag.SubTagId != "" {
			subTagIdParsed, err := convertID(tag.SubTagId)
			if err != nil {
				return nil, err
			}

			subTagId = subTagIdParsed
		}

		tags = append(tags, models.TransactionTags{
			TagId:    tagId,
			SubTagId: subTagId,
		})
	}

	assignedTo, err := convertID(body.AssignedTo)
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

	var confirmationDate *time.Time
	if body.IsConfirmed {
		confirmationDateParsed, err := parseDate(*body.ConfirmationDate)
		if err != nil {
			return nil, err
		}

		confirmationDate = &confirmationDateParsed
	} else {
		confirmationDate = nil
	}

	dueDate, err := parseDate(body.DueDate)
	if err != nil {
		return nil, err
	}

	return &models.Transaction{
		Name:        body.Name,
		Description: body.Description,
		Invoice:     body.Invoice,
		Type:        body.Type,
		Supplier:    body.Supplier,
		AssignedTo:  assignedTo,
		Balance: models.TransactionBalance{
			Value:              body.Balance.Value,
			Parts:              body.Balance.Parts,
			Labor:              body.Balance.Labor,
			Discount:           body.Balance.Discount,
			Interest:           body.Balance.Interest,
			DiscountPercentage: body.Balance.DiscountPercentage,
			InterestPercentage: body.Balance.InterestPercentage,
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
		Tags:             tags,
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
