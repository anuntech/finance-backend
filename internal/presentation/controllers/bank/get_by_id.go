package bank

import (
	"net/http"

	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"

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

func (c *GetBankByIdController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	id := r.Req.PathValue("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid bank ID format",
		}, http.StatusBadRequest)
	}

	bank, err := c.FindBankByIdRepository.Find(objectId.Hex())
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred when retrieving bank",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(bank, http.StatusOK)
}
