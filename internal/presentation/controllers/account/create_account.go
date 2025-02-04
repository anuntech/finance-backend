package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
)

type CreateAccountController struct {
	CreateAccount usecase.CreateAccount
	Validate      *validator.Validate
}

func NewCreateAccountController(createAccount usecase.CreateAccount) *CreateAccountController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &CreateAccountController{
		CreateAccount: createAccount,
		Validate:      validate,
	}
}

type CreateAccountControllerResponse struct {
	Id    string `json:"id"`
	Name  string `json:"name" validate:"required,email"`
	Image string `json:"image" validate:"required,mongodb"`
	Color string `json:"color" validate:"required,hexcolor"`
}

type CreateAccountControllerBody struct {
	Name  string `validate:"required,email"`
	Image string `validate:"required,mongodb"`
	Color string `validate:"required,hexcolor"`
}

func (c *CreateAccountController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body CreateAccountControllerBody
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

	account, err := c.CreateAccount.Create(&models.AccountInput{
		Name:  body.Name,
		Image: body.Image,
		Color: body.Color,
	})

	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error ocurred when creating account",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(&CreateAccountControllerResponse{
		Id:    account.Id,
		Name:  account.Name,
		Image: account.Image,
		Color: account.Color,
	}, http.StatusOK)
}
