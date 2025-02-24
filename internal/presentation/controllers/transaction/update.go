package transaction

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UpdateTransactionController struct {
	UpdateTransactionRepository usecase.UpdateTransactionRepository
	Validate                    *validator.Validate
	FindTransactionById         usecase.FindTransactionByIdRepository
}

func NewUpdateTransactionController(updateTransaction usecase.UpdateTransactionRepository, findTransactionById usecase.FindTransactionByIdRepository) *UpdateTransactionController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &UpdateTransactionController{
		UpdateTransactionRepository: updateTransaction,
		Validate:                    validate,
		FindTransactionById:         findTransactionById,
	}
}

type UpdateTransactionBody struct {
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
	} `json:"repeatSettings" validate:"excluded_if=Frequency DO_NOT_REPEAT,excluded_if=Frequency RECURRING,required_if=Frequency REPEAT,omitempty"`
	DueDate          string `json:"dueDate" validate:"required,datetime=2006-01-02T15:04:05Z"`
	IsConfirmed      bool   `json:"isConfirmed"`
	CategoryId       string `json:"categoryId" validate:"required,mongodb"`
	SubCategoryId    string `json:"subCategoryId" validate:"required,mongodb"`
	TagId            string `json:"tagId" validate:"required,mongodb"`
	SubTagId         string `json:"subTagId" validate:"required,mongodb"`
	AccountId        string `json:"accountId" validate:"required,mongodb"`
	RegistrationDate string `json:"registrationDate" validate:"required,datetime=2006-01-02T15:04:05Z"`
	ConfirmationDate string `json:"confirmationDate" validate:"datetime=2006-01-02T15:04:05Z,excluded_if=IsConfirmed false,required_if=IsConfirmed true"`
}

func (c *UpdateTransactionController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body UpdateTransactionBody
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
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "transaction not found",
		}, http.StatusNotFound)
	}

	transaction.Name = body.Name
	transaction.Description = body.Description
	transaction.Type = body.Type
	transaction.Supplier = body.Supplier
	transaction.Balance = models.TransactionBalance{
		Value:    body.Balance.Value,
		Parts:    body.Balance.Parts,
		Labor:    body.Balance.Labor,
		Discount: body.Balance.Discount,
		Interest: body.Balance.Interest,
	}
	transaction.Frequency = body.Frequency
	transaction.RepeatSettings = models.TransactionRepeatSettings{
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

	transaction.AssignedTo = transactionIdsParsed.AssignedTo
	transaction.TagId = transactionIdsParsed.TagId
	transaction.SubTagId = transactionIdsParsed.SubTagId
	transaction.AccountId = transactionIdsParsed.AccountId
	transaction.RegistrationDate = transactionIdsParsed.RegistrationDate
	transaction.ConfirmationDate = transactionIdsParsed.ConfirmationDate

	transactionUpdated, err := c.UpdateTransactionRepository.Update(transactionId, transaction)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error updating transaction",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(transactionUpdated, http.StatusOK)
}

func (c *UpdateTransactionController) createTransaction(body *UpdateTransactionBody) (*models.Transaction, error) {
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
