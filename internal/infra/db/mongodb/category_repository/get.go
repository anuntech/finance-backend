package category_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FindCategorysByWorkspaceIdRepository struct {
	Db *mongo.Database
}

func NewFindCategorysByWorkspaceIdRepository(db *mongo.Database) *FindCategorysByWorkspaceIdRepository {
	return &FindCategorysByWorkspaceIdRepository{
		Db: db,
	}
}

func (r *FindCategorysByWorkspaceIdRepository) Find(workspaceId primitive.ObjectID) ([]models.Category, error) {
	collection := r.Db.Collection("category")

	cursor, err := collection.Find(context.Background(), bson.M{"workspaceId": workspaceId})
	if err != nil {
		return nil, err
	}

	var categorys []models.Category
	if err = cursor.All(context.Background(), &categorys); err != nil {
		return nil, err
	}

	return categorys, nil
}
