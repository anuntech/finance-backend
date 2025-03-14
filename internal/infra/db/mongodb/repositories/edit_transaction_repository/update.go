package edit_transaction_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UpdateEditTransactionRepository struct {
	Db *mongo.Database
}

func NewUpdateEditTransactionRepository(db *mongo.Database) *UpdateEditTransactionRepository {
	return &UpdateEditTransactionRepository{
		Db: db,
	}
}

func (r *UpdateEditTransactionRepository) Update(transaction *models.Transaction) (*models.Transaction, error) {
	collection := r.Db.Collection("edit_transaction")

	transaction.UpdatedAt = time.Now().UTC()

	filter := bson.M{"main_id": transaction.MainId, "workspace_id": transaction.WorkspaceId}
	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	// Use $set to update fields without changing _id
	update := bson.M{
		"$set": bson.M{
			"name":              transaction.Name,
			"description":       transaction.Description,
			"invoice":           transaction.Invoice,
			"type":              transaction.Type,
			"supplier":          transaction.Supplier,
			"assigned_to":       transaction.AssignedTo,
			"balance":           transaction.Balance,
			"is_confirmed":      transaction.IsConfirmed,
			"category_id":       transaction.CategoryId,
			"sub_category_id":   transaction.SubCategoryId,
			"tags":              transaction.Tags,
			"account_id":        transaction.AccountId,
			"registration_date": transaction.RegistrationDate,
			"confirmation_date": transaction.ConfirmationDate,
			"due_date":          transaction.DueDate,
			"updated_at":        transaction.UpdatedAt,
			"main_count":        transaction.MainCount,
			// Include other fields as needed, but exclude _id
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	// Get the updated document
	var updatedTransaction models.Transaction
	err = collection.FindOne(ctx, filter).Decode(&updatedTransaction)
	if err != nil {
		return nil, err
	}

	return &updatedTransaction, nil
}
