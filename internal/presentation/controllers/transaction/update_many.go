package transaction

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdateManyRequest is the request body for updating multiple transactions
type UpdateManyRequest struct {
	Name             *string                   `json:"name,omitempty"`
	Description      *string                   `json:"description,omitempty"`
	Invoice          *string                   `json:"invoice,omitempty"`
	Type             *string                   `json:"type,omitempty"`
	Supplier         *string                   `json:"supplier,omitempty"`
	AssignedTo       *string                   `json:"assignedTo,omitempty"`
	Balance          *UpdateManyBalanceRequest `json:"balance,omitempty"`
	DueDate          *string                   `json:"dueDate,omitempty"`
	IsConfirmed      *bool                     `json:"isConfirmed,omitempty"`
	CategoryId       *string                   `json:"categoryId,omitempty"`
	SubCategoryId    *string                   `json:"subCategoryId,omitempty"`
	AccountId        *string                   `json:"accountId,omitempty"`
	RegistrationDate *string                   `json:"registrationDate,omitempty"`
	ConfirmationDate *string                   `json:"confirmationDate,omitempty"`
	CustomFields     []CustomFieldUpdate       `json:"customFields,omitempty"`
}

// UpdateManyBalanceRequest is the request body for updating balance fields in multiple transactions
type UpdateManyBalanceRequest struct {
	Value              *float64 `json:"value,omitempty"`
	Discount           *float64 `json:"discount,omitempty"`
	Interest           *float64 `json:"interest,omitempty"`
	DiscountPercentage *float64 `json:"discountPercentage,omitempty"`
	InterestPercentage *float64 `json:"interestPercentage,omitempty"`
}

// CustomFieldUpdate is the request body for updating custom fields
type CustomFieldUpdate struct {
	CustomFieldId string  `json:"id"`
	Value         *string `json:"value,omitempty"`
}

type UpdateManyTransactionController struct {
	FindTransactionByIdRepository     usecase.FindTransactionByIdRepository
	FindByIdEditTransactionRepository usecase.FindByIdEditTransactionRepository
	UpdateTransactionRepository       usecase.UpdateTransactionRepository
	CreateEditTransactionRepository   usecase.CreateEditTransactionRepository
	FindCustomFieldByIdRepository     usecase.FindCustomFieldByIdRepository
}

func NewUpdateManyTransactionController(
	findTransactionById usecase.FindTransactionByIdRepository,
	findByIdEditTransaction usecase.FindByIdEditTransactionRepository,
	updateTransaction usecase.UpdateTransactionRepository,
	createEditTransaction usecase.CreateEditTransactionRepository,
	findCustomFieldById usecase.FindCustomFieldByIdRepository,
) *UpdateManyTransactionController {
	return &UpdateManyTransactionController{
		FindTransactionByIdRepository:     findTransactionById,
		FindByIdEditTransactionRepository: findByIdEditTransaction,
		UpdateTransactionRepository:       updateTransaction,
		CreateEditTransactionRepository:   createEditTransaction,
		FindCustomFieldByIdRepository:     findCustomFieldById,
	}
}

func (c *UpdateManyTransactionController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body UpdateManyRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid body request",
		}, http.StatusBadRequest)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Invalid workspace ID format",
		}, http.StatusBadRequest)
	}

	ids := r.UrlParams.Get("ids")
	idsSlice := strings.Split(ids, ",")

	type TransactionIdentifier struct {
		ID                primitive.ObjectID
		IsInstallment     bool
		InstallmentNumber int
	}

	transactionIdentifiers := []TransactionIdentifier{}

	for _, idString := range idsSlice {
		// Check if ID contains installment number (format: objectID-installmentNumber)
		parts := strings.Split(idString, "-")

		if len(parts) == 2 {
			// This is an installment transaction
			mainID, err := primitive.ObjectIDFromHex(parts[0])
			if err != nil {
				return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "Invalid transaction ID format: " + idString,
				}, http.StatusBadRequest)
			}

			installmentNumber, err := strconv.Atoi(parts[1])
			if err != nil {
				return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "Invalid installment number in transaction ID: " + idString,
				}, http.StatusBadRequest)
			}

			transactionIdentifiers = append(transactionIdentifiers, TransactionIdentifier{
				ID:                mainID,
				IsInstallment:     true,
				InstallmentNumber: installmentNumber,
			})
		} else if len(parts) == 1 {
			// Regular transaction
			objectID, err := primitive.ObjectIDFromHex(parts[0])
			if err != nil {
				return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "Invalid transaction ID format: " + idString,
				}, http.StatusBadRequest)
			}

			transactionIdentifiers = append(transactionIdentifiers, TransactionIdentifier{
				ID:                objectID,
				IsInstallment:     false,
				InstallmentNumber: 0,
			})
		} else {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "Invalid transaction ID format: " + idString,
			}, http.StatusBadRequest)
		}
	}

	successCount := 0
	failedCount := 0
	updatedTransactions := []any{}

	for _, identifier := range transactionIdentifiers {
		var transaction *models.Transaction
		var err error

		if identifier.IsInstallment {
			// For installment transactions, first check for edited transactions
			transaction, err = c.FindByIdEditTransactionRepository.Find(identifier.ID, identifier.InstallmentNumber, workspaceId)
			if err != nil {
				failedCount++
				continue
			}

			// If no edited transaction, try to find the main transaction
			if transaction == nil {
				mainTransaction, err := c.FindTransactionByIdRepository.Find(identifier.ID, workspaceId)
				if err != nil || mainTransaction == nil {
					failedCount++
					continue
				}

				transaction = mainTransaction
				transaction.MainCount = &identifier.InstallmentNumber
				transaction.MainId = &identifier.ID
			}
		} else {
			transaction, err = c.FindTransactionByIdRepository.Find(identifier.ID, workspaceId)
			if err != nil || transaction == nil {
				failedCount++
				continue
			}
		}

		// Update only non-nil fields
		if body.Name != nil {
			transaction.Name = *body.Name
		}
		if body.Description != nil {
			transaction.Description = *body.Description
		}
		if body.Invoice != nil {
			transaction.Invoice = *body.Invoice
		}
		if body.Type != nil {
			transaction.Type = *body.Type
		}
		if body.Supplier != nil {
			transaction.Supplier = *body.Supplier
		}
		if body.AssignedTo != nil {
			assignedTo, err := primitive.ObjectIDFromHex(*body.AssignedTo)
			if err == nil {
				transaction.AssignedTo = assignedTo
			}
		}
		if body.Balance != nil {
			if body.Balance.Value != nil {
				transaction.Balance.Value = *body.Balance.Value
			}
			if body.Balance.Discount != nil {
				transaction.Balance.Discount = *body.Balance.Discount
			}
			if body.Balance.Interest != nil {
				transaction.Balance.Interest = *body.Balance.Interest
			}
			if body.Balance.DiscountPercentage != nil {
				transaction.Balance.DiscountPercentage = *body.Balance.DiscountPercentage
			}
			if body.Balance.InterestPercentage != nil {
				transaction.Balance.InterestPercentage = *body.Balance.InterestPercentage
			}
		}
		if body.DueDate != nil {
			dueDate, err := time.Parse(time.RFC3339, *body.DueDate)
			if err == nil {
				transaction.DueDate = dueDate
			}
		}
		if body.IsConfirmed != nil {
			transaction.IsConfirmed = *body.IsConfirmed

			if !*body.IsConfirmed {
				transaction.ConfirmationDate = nil
			} else if body.ConfirmationDate != nil {
				confirmationDate, err := time.Parse(time.RFC3339, *body.ConfirmationDate)
				if err == nil {
					transaction.ConfirmationDate = &confirmationDate
				}
			} else {
				return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "Confirmation date is required when transaction is confirmed! ID: " + identifier.ID.Hex(),
				}, http.StatusBadRequest)
			}
		}
		if body.CategoryId != nil {
			if *body.CategoryId != "" {
				categoryId, err := primitive.ObjectIDFromHex(*body.CategoryId)
				if err == nil {
					transaction.CategoryId = &categoryId
				}
			} else {
				transaction.CategoryId = nil
			}
		}
		if body.SubCategoryId != nil {
			if *body.SubCategoryId != "" {
				subCategoryId, err := primitive.ObjectIDFromHex(*body.SubCategoryId)
				if err == nil {
					transaction.SubCategoryId = &subCategoryId
				}
			} else {
				transaction.SubCategoryId = nil
			}
		}
		if body.AccountId != nil {
			if *body.AccountId != "" {
				accountId, err := primitive.ObjectIDFromHex(*body.AccountId)
				if err == nil {
					transaction.AccountId = &accountId
				}
			} else {
				transaction.AccountId = nil
			}
		}
		if body.RegistrationDate != nil {
			registrationDate, err := time.Parse(time.RFC3339, *body.RegistrationDate)
			if err == nil {
				transaction.RegistrationDate = registrationDate
			}
		}

		if len(body.CustomFields) > 0 {
			existingCustomFields := make(map[string]int)
			for i, cf := range transaction.CustomFields {
				existingCustomFields[cf.CustomFieldId.Hex()] = i
			}

			for _, newCF := range body.CustomFields {
				customFieldId, err := primitive.ObjectIDFromHex(newCF.CustomFieldId)
				if err != nil {
					continue
				}

				customField, err := c.FindCustomFieldByIdRepository.Find(customFieldId, workspaceId)
				if err != nil || customField == nil {
					continue
				}

				if idx, exists := existingCustomFields[customFieldId.Hex()]; exists {
					if newCF.Value != nil {
						transaction.CustomFields[idx].Value = *newCF.Value
					}

					continue
				}

				if newCF.Value != nil {
					transaction.CustomFields = append(transaction.CustomFields, models.TransactionCustomField{
						CustomFieldId: customFieldId,
						Value:         *newCF.Value,
					})
				}
			}
		}

		transaction.UpdatedAt = time.Now()

		if identifier.IsInstallment {
			response, err := c.CreateEditTransactionRepository.Create(transaction)
			if err != nil {
				failedCount++
				continue
			}

			successCount++
			updatedTransactions = append(updatedTransactions, response)
			continue
		}

		updatedTransaction, err := c.UpdateTransactionRepository.Update(transaction.Id, transaction)
		if err != nil {
			failedCount++
		} else {
			successCount++
			updatedTransactions = append(updatedTransactions, updatedTransaction)
		}
	}

	return helpers.CreateResponse(map[string]any{
		"success":      successCount,
		"failed":       failedCount,
		"total":        len(transactionIdentifiers),
		"transactions": updatedTransactions,
	}, http.StatusOK)
}
