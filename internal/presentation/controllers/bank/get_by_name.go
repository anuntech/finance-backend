package bank

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
)

type GetBankByNameController struct {
	FindBankByNameRepository usecase.FindBankByNameRepository
}

func NewGetBankByNameController(findByNameUsecase usecase.FindBankByNameRepository) *GetBankByNameController {
	return &GetBankByNameController{
		FindBankByNameRepository: findByNameUsecase,
	}
}

func (c *GetBankByNameController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	name := r.Req.URL.Query().Get("name")
	if name == "" {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Bank name is required",
		}, http.StatusBadRequest)
	}

	bank, err := c.FindBankByNameRepository.FindByName(name)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when retrieving bank",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(bank, http.StatusOK)
}
