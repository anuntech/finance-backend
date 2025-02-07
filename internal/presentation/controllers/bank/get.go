package bank

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
)

type GetBankController struct {
	GetAllBanksRepository usecase.FindAllBankRepository
}

func NewGetBankController(getAllBanksRepository usecase.FindAllBankRepository) *GetBankController {
	return &GetBankController{
		GetAllBanksRepository: getAllBanksRepository,
	}
}

func (c *GetBankController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	bank, err := c.GetAllBanksRepository.Find()
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when retrieving bank",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(bank, http.StatusOK)
}
