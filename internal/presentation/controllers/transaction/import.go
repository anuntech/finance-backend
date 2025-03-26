package transaction

// import (
// 	"encoding/json"
// 	"errors"
// 	"net/http"
// 	"strings"
// 	"sync"
// 	"time"

// 	"github.com/anuntech/finance-backend/internal/domain/models"
// 	"github.com/anuntech/finance-backend/internal/domain/usecase"
// 	infraHelpers "github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
// 	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/transaction_repository"
// 	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/workspace_repository/member_repository"
// 	"github.com/anuntech/finance-backend/internal/presentation/helpers"
// 	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
// 	ut "github.com/go-playground/universal-translator"
// 	"github.com/go-playground/validator/v10"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// )

// type CreateTransactionController struct {
// 	Validate                      *validator.Validate
// 	Translator                    ut.Translator
// 	CreateTransactionRepository   usecase.CreateTransactionRepository
// 	FindMemberByIdRepository      *member_repository.FindMemberByIdRepository
// 	FindAccountByIdRepository     usecase.FindAccountByIdRepository
// 	FindCategoryByIdRepository    usecase.FindCategoryByIdRepository
// 	FindCustomFieldByIdRepository usecase.FindCustomFieldByIdRepository
// }

// func NewCreateTransactionController(findMemberByIdRepository *member_repository.FindMemberByIdRepository, createTransactionRepository *transaction_repository.CreateTransactionRepository, findAccountByIdRepository usecase.FindAccountByIdRepository, findCategoryByIdRepository usecase.FindCategoryByIdRepository, findCustomFieldByIdRepository usecase.FindCustomFieldByIdRepository) *CreateTransactionController {
// 	validate := validator.New(validator.WithRequiredStructEnabled())

// 	return &CreateTransactionController{
// 		Validate:                      validate,
// 		FindMemberByIdRepository:      findMemberByIdRepository,
// 		CreateTransactionRepository:   createTransactionRepository,
// 		FindAccountByIdRepository:     findAccountByIdRepository,
// 		FindCategoryByIdRepository:    findCategoryByIdRepository,
// 		FindCustomFieldByIdRepository: findCustomFieldByIdRepository,
// 	}
// }

// type TransactionImportBody struct {
// 	Name        string  `json:"name" validate:"required,min=2,max=30"`
// 	Description string  `json:"description" validate:"omitempty,max=255"`
// 	Invoice     string  `json:"invoice" validate:"omitempty,min=2,max=50"`
// 	Type        string  `json:"type" validate:"required,oneof=EXPENSE RECIPE"`
// 	Supplier    string  `json:"supplier" validate:"omitempty,min=3,max=30"`
// 	AssignedTo  string  `json:"assignedTo" validate:"required,min=3,max=30"`
// 	Balance     struct {
// 		Value              float64 `json:"value" validate:"required,min=0.01"`
// 		Parts              float64 `json:"parts" validate:"omitempty,min=0.01"`
// 		Labor              float64 `json:"labor" validate:"omitempty,min=0.01"`
// 		Discount           float64 `json:"discount" validate:"omitempty,min=0.01"`
// 		Interest           float64 `json:"interest" validate:"omitempty,min=0.01"`
// 		DiscountPercentage float64 `json:"discountPercentage" validate:"omitempty,min=0.01,max=100"`
// 		Intem restPercentage float64 `json:"interestPercentage" validate:"omitempty,min=0.01,max=100"`
// 	} `json:"balance" validate:"required"`
// 	Frequency      string `json:"frequency" validate:"oneof=DO_NOT_REPEAT RECURRING REPEAT"`
// 	RepeatSettings struct {
// 		InitialInstallment time.Month `json:"initialInstallment" validate:"min=1"`
// 		Count              int        `json:"count" validate:"min=2,max=367"`
// 		Interval           string     `json:"interval" validate:"oneof=DAILY WEEKLY MONTHLY QUARTERLY YEARLY"`
// 	} `json:"repeatSettings" validate:"excluded_if=Frequency DO_NOT_REPEAT,excluded_if=Frequency RECURRING,required_if=Frequency REPEAT,omitempty"`
// 	DueDate     string  `json:"dueDate" validate:"required,datetime=2006-01-02T15:04:05Z"`
// 	IsConfirmed bool    `json:"isConfirmed"`
// 	Category    *string `json:"category" validate:"omitempty"`
// 	SubCategory *string `json:"subCategory" validate:"omitempty"`
// 	Tags        []struct {
// 		Tag    string `json:"tag" validate:"omitempty"`
// 		SubTag string `json:"subTag" validate:"excluded_if=Tag '',omitempty"`
// 	} `json:"tags" validate:"omitempty"`
// 	CustomFields []struct {
// 		CustomField string `json:"customField" validate:"required"`
// 		Value       string `json:"value" validate:"required,max=100"`
// 	} `json:"customFields"`
// 	Account          *string `json:"account" validate:"required"`
// 	RegistrationDate string  `json:"registrationDate" validate:"required,datetime=2006-01-02T15:04:05Z"`
// 	ConfirmationDate *string `json:"confirmationDate" validate:"excluded_if=IsConfirmed false,required_if=IsConfirmed true,omitempty,datetime=2006-01-02T15:04:05Z"`
// }

// func (c *CreateTransactionController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
// 	var body TransactionBody
// 	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
// 		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
// 			Error: "invalid body request",
// 		}, http.StatusBadRequest)
// 	}

// 	if err := c.Validate.Struct(body); err != nil {
// 		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
// 			Error: helpers.GetErrorMessages(c.Validate, err),
// 		}, http.StatusBadRequest)
// 	}

// 	transaction, err := createTransaction(&body)
// 	if err != nil {
// 		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
// 			Error: "error creating transaction: " + err.Error(),
// 		}, http.StatusInternalServerError)
// 	}

// 	userObjectID, err := primitive.ObjectIDFromHex(r.Header.Get("userId"))
// 	if err != nil {
// 		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
// 			Error: "invalid user ID format",
// 		}, http.StatusBadRequest)
// 	}
// 	transaction.CreatedBy = userObjectID

// 	assignedTo, err := primitive.ObjectIDFromHex(body.AssignedTo)
// 	if err != nil {
// 		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
// 			Error: "invalid assignedTo ID format",
// 		}, http.StatusBadRequest)
// 	}

// 	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
// 	if err != nil {
// 		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
// 			Error: "invalid workspace ID format",
// 		}, http.StatusBadRequest)
// 	}
// 	transaction.WorkspaceId = workspaceId

// 	errChan := make(chan *presentationProtocols.HttpResponse, 4)
// 	var wg sync.WaitGroup

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		if err := c.validateAssignedMember(workspaceId, assignedTo); err != nil {
// 			errChan <- err
// 			return
// 		}
// 		transaction.AssignedTo = assignedTo
// 	}()

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		if transaction.AccountId == nil {
// 			return
// 		}
// 		if err := c.validateAccount(workspaceId, *transaction.AccountId); err != nil {
// 			errChan <- err
// 		}
// 	}()

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		if transaction.CategoryId == nil {
// 			return
// 		}
// 		if err := c.validateCategory(workspaceId, *transaction.CategoryId, transaction.Type, *transaction.SubCategoryId); err != nil {
// 			errChan <- err
// 		}
// 	}()

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()

// 		seenCustomFields := make(map[string]bool)

// 		for _, customField := range transaction.CustomFields {
// 			compositeKey := customField.CustomFieldId.Hex()

// 			if seenCustomFields[compositeKey] {
// 				errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
// 					Error: "duplicate custom field detected: " + compositeKey,
// 				}, http.StatusBadRequest)
// 			}
// 			seenCustomFields[compositeKey] = true

// 			customFieldParsed, err := c.FindCustomFieldByIdRepository.Find(customField.CustomFieldId, workspaceId)
// 			if err != nil {
// 				errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
// 					Error: "error finding custom field",
// 				}, http.StatusInternalServerError)
// 			}

// 			if customFieldParsed == nil {
// 				errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
// 					Error: "custom field not found",
// 				}, http.StatusNotFound)
// 			}
// 		}
// 	}()

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		seenTags := make(map[string]bool)

// 		for _, tag := range transaction.Tags {
// 			compositeKey := tag.TagId.Hex() + "|" + tag.SubTagId.Hex()

// 			if seenTags[compositeKey] {
// 				errChan <- helpers.CreateResponse(&presentationProtocols.ErrorResponse{
// 					Error: "duplicate tag detected: " + compositeKey,
// 				}, http.StatusBadRequest)
// 				return
// 			}
// 			seenTags[compositeKey] = true

// 			if err := c.validateTag(workspaceId, tag.TagId, tag.SubTagId); err != nil {
// 				errChan <- err
// 				return
// 			}
// 		}
// 	}()

// 	wg.Wait()
// 	close(errChan)

// 	if len(errChan) > 0 {
// 		return <-errChan
// 	}

// 	transaction, err = c.CreateTransactionRepository.Create(transaction)
// 	if err != nil {
// 		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
// 			Error: "error creating transaction",
// 		}, http.StatusInternalServerError)
// 	}

// 	recipeTx := *transaction
// 	recipeTx.Type = "RECIPE"
// 	recipeNetBalance := infraHelpers.CalculateOneTransactionBalance(&recipeTx)
// 	transaction.Balance.NetBalance = recipeNetBalance

// 	return helpers.CreateResponse(transaction, http.StatusCreated)
// }
