package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type UpdateAccountController struct {
	UpdateAccount usecase.UpdateAccount
	Validate      *validator.Validate
}

func NewUpdateAccountController(updateAccount usecase.UpdateAccount) *UpdateAccountController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &UpdateAccountController{
		UpdateAccount: updateAccount,
		Validate:      validate,
	}
}

type UpdateAccountControllerBody struct {
	Name string `validate:"required"`
	Bank string `validate:"required"`
}

func (c *UpdateAccountController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	id := r.Req.PathValue("id")
	err := uuid.Validate(id)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid account ID format",
		}, http.StatusBadRequest)
	}

	var body UpdateAccountControllerBody
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

	account, err := c.UpdateAccount.Update(id, &models.AccountInput{
		Name: body.Name,
		Bank: body.Bank,
	})

	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when updating account",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(account, http.StatusOK)
}
