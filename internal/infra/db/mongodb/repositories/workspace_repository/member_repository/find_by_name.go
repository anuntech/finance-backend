package member_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FindMemberByNameRepository struct {
	db *mongo.Database
}

func NewFindMemberByNameRepository(db *mongo.Database) *FindMemberByNameRepository {
	return &FindMemberByNameRepository{db}
}

func (r *FindMemberByNameRepository) FindByNameAndWorkspaceId(name string, workspaceId primitive.ObjectID) (*models.Member, error) {
	collection := r.db.Collection("workspaces")

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	// Buscar workspace e membros
	var workspace models.Workspace
	err := collection.FindOne(ctx, bson.M{"_id": workspaceId}).Decode(&workspace)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	// Buscar usuários por nome para obter seus IDs
	usersCollection := r.db.Collection("users")
	var user models.WorkspaceUser
	err = usersCollection.FindOne(ctx, bson.M{"name": name}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	// Verificar se o usuário é membro do workspace
	for _, member := range workspace.Members {
		if member.MemberId == user.Id {
			return &member, nil
		}
	}

	return nil, nil
}
