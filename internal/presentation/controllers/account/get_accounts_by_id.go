package controllers

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
)

type GetAccountByIdController struct {
	FindById usecase.FindById
}

func NewGetAccountByIdController(findById usecase.FindById) *GetAccountByIdController {
	return &GetAccountByIdController{
		FindById: findById,
	}
}

func (c *GetAccountByIdController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	id := r.Req.PathValue("id")
	account, err := c.FindById.Find(id)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when retrieving account: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(account, http.StatusOK)
}
