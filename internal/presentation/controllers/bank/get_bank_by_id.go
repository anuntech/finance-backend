package bank

import (
	"net/http"

	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
)

type GetBankByIdController struct {
	FindBankByIdRepository usecase.FindBankByIdRepository
}

func NewGetBankByIdController(findByIdUsecase usecase.FindBankByIdRepository) *GetBankByIdController {
	return &GetBankByIdController{
		FindBankByIdRepository: findByIdUsecase,
	}
}

func (c *GetBankByIdController) Handle(w http.ResponseWriter, r *http.Request) *presentationProtocols.HttpResponse {
	id := r.URL.Query().Get("id")

	bank, err := c.FindBankByIdRepository.Find(id)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when retrieving bank",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(bank, http.StatusOK)
}
