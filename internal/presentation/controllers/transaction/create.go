package transaction

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
)

type CreateTransactionController struct {
	Validate *validator.Validate
}

func NewCreateTransactionController() *CreateTransactionController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &CreateTransactionController{
		Validate: validate,
	}
}

type CreateTransactionBody struct {
	Name        string `json:"name" validate:"required,min=3,max=30"`
	Description string `json:"description" validate:"required,min=3,max=255"`
	Type        string `json:"type" validate:"required,oneof=EXPENSE RECIPE"`
	Supplier    string `json:"supplier" validate:"required,min=3,max=30"`
	AssignedTo  string `json:"assignedTo" validate:"required,min=3,max=30"`
	Balance     struct {
		Value    int `json:"value" validate:"required,min=0"`
		Parts    int `json:"parts" validate:"min=0"`
		Labor    int `json:"labor" validate:"min=0"`
		Discount int `json:"discount" validate:"min=0"`
		Interest int `json:"interest" validate:"min=0"`
	} `json:"balance" validate:"required"`
	Frequency      string `json:"frequency" validate:"oneof=DO_NOT_REPEAT RECURRING REPEAT"`
	RepeatSettings struct {
		InitialInstallment time.Month `json:"initialInstallment" validate:"min=0"`
		Count              int        `json:"count" validate:"min=0"`
		Interval           string     `json:"interval" validate:"oneof=DAILY WEEKLY MONTHLY QUARTERLY YEARLY"`
	} `json:"repeatSettings" validate:"excluded_if=Frequency DO_NOT_REPEAT,excluded_if=Frequency RECURRING,required_if=Frequency REPEAT"`
	DueDate          string `json:"dueDate" validate:"required,datetime=2006-01-02"`
	IsConfirmed      bool   `json:"isConfirmed"`
	CategoryId       string `json:"categoryId" validate:"required"`
	SubCategoryId    string `json:"subCategoryId" validate:"required"`
	TagId            string `json:"tagId" validate:"required"`
	SubTagId         string `json:"subTagId" validate:"required"`
	AccountId        string `json:"accountId" validate:"required"`
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

	// transaction := models.Transaction{
	// 	Name:        body.Name,
	// 	Description: body.Description,
	// }

	// transaction, err := c.CreateTransactionRepository.Create(transaction)
	// if err != nil {
	// 	return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
	// 		Error: "error creating transaction",
	// 	}, http.StatusInternalServerError)
	// }

	return helpers.CreateResponse(body, http.StatusCreated)
}
