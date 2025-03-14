package member_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FindMemberByIdRepository struct {
	DbWorkspace *mongo.Database
}

func NewFindMemberByIdRepository(db *mongo.Database) *FindMemberByIdRepository {
	return &FindMemberByIdRepository{
		DbWorkspace: db,
	}
}

func (r *FindMemberByIdRepository) Find(WorkspaceId primitive.ObjectID, MemberId primitive.ObjectID) (*models.Member, error) {
	collection := r.DbWorkspace.Collection("workspaces")

	var workspace models.Workspace
	err := collection.FindOne(context.Background(), bson.M{"_id": WorkspaceId}).Decode(&workspace)
	if err != nil {
		return nil, err
	}

	if workspace.Owner == MemberId {
		return &models.Member{
			MemberId: MemberId,
			Role:     "owner",
		}, nil
	}

	for _, member := range workspace.Members {
		if member.MemberId == MemberId {
			return &member, nil
		}
	}

	return nil, nil
}
