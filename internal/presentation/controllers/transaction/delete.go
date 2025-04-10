package transaction

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeleteTransactionController struct {
	DeleteTransactionRepository   usecase.DeleteTransactionRepository
	FindTransactionByIdRepository usecase.FindTransactionByIdRepository
}

func NewDeleteTransactionController(
	deleteTransaction usecase.DeleteTransactionRepository,
	findTransactionById usecase.FindTransactionByIdRepository,
) *DeleteTransactionController {
	return &DeleteTransactionController{
		DeleteTransactionRepository:   deleteTransaction,
		FindTransactionByIdRepository: findTransactionById,
	}
}

func (c *DeleteTransactionController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid workspace ID format",
		}, http.StatusBadRequest)
	}

	ids := r.UrlParams.Get("ids")
	idsSlice := strings.Split(ids, ",")

	// Regular transaction IDs
	idsObjectID := []primitive.ObjectID{}

	// Edit transaction parameters
	var editTransactionParams []struct {
		MainId      primitive.ObjectID
		MainCount   int
		WorkspaceId primitive.ObjectID
	}

	for _, idString := range idsSlice {
		// Check if it's in format "mainId-count" (edit transaction)
		parts := strings.Split(idString, "-")

		if len(parts) == 2 {
			// This is an edit transaction ID with format "mainId-count"
			mainId, err := primitive.ObjectIDFromHex(parts[0])
			if err != nil {
				return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "Invalid transaction main ID format: " + idString,
				}, http.StatusBadRequest)
			}

			count, err := strconv.Atoi(parts[1])
			if err != nil {
				return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "Invalid installment number in: " + idString,
				}, http.StatusBadRequest)
			}

			editTransactionParams = append(editTransactionParams, struct {
				MainId      primitive.ObjectID
				MainCount   int
				WorkspaceId primitive.ObjectID
			}{
				MainId:      mainId,
				MainCount:   count,
				WorkspaceId: workspaceId,
			})
		} else if len(parts) == 1 {
			// This is a regular transaction ID
			objectID, err := primitive.ObjectIDFromHex(idString)
			if err != nil {
				return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "Invalid transaction ID format: " + idString,
				}, http.StatusBadRequest)
			}
			idsObjectID = append(idsObjectID, objectID)
		} else {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "Invalid ID format: " + idString,
			}, http.StatusBadRequest)
		}
	}

	// Process regular transaction deletions
	if len(idsObjectID) > 0 {
		err = c.DeleteTransactionRepository.Delete(idsObjectID, workspaceId)
		if err != nil {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "An error occurred when deleting transactions: " + err.Error(),
			}, http.StatusInternalServerError)
		}
	}

	// Process edit transaction deletions
	if len(editTransactionParams) > 0 {
		err = c.DeleteTransactionRepository.DeleteEditTransactions(editTransactionParams, c.FindTransactionByIdRepository)
		if err != nil {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "An error occurred when marking edit transactions as deleted: " + err.Error(),
			}, http.StatusInternalServerError)
		}
	}

	return helpers.CreateResponse(nil, http.StatusNoContent)
}
