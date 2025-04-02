package workspace_user_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FindWorkspaceUserByIdRepository struct {
	DbWorkspace *mongo.Database
}

func NewFindWorkspaceUserByIdRepository(db *mongo.Database) *FindWorkspaceUserByIdRepository {
	return &FindWorkspaceUserByIdRepository{
		DbWorkspace: db,
	}
}

func (r *FindWorkspaceUserByIdRepository) Find(UserId primitive.ObjectID) (*models.WorkspaceUser, error) {
	collection := r.DbWorkspace.Collection("users")

	var user models.WorkspaceUser
	err := collection.FindOne(context.Background(), bson.M{"_id": UserId}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
