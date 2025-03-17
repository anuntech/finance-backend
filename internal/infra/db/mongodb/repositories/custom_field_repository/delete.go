package custom_field_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DeleteCustomFieldRepository struct {
	Db *mongo.Database
}

func NewDeleteCustomFieldRepository(db *mongo.Database) *DeleteCustomFieldRepository {
	return &DeleteCustomFieldRepository{
		Db: db,
	}
}

func (r *DeleteCustomFieldRepository) Delete(customFieldIds []primitive.ObjectID, workspaceId primitive.ObjectID) error {
	// Primeiro, excluir os campos personalizados
	collection := r.Db.Collection("custom_field")

	filter := bson.M{
		"_id":          bson.M{"$in": customFieldIds},
		"workspace_id": workspaceId,
	}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	_, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}

	// Em seguida, remover referências a esses campos personalizados de todas as transações
	transactionCollection := r.Db.Collection("transaction")

	// Atualiza todas as transações do mesmo workspace para remover os campos personalizados excluídos
	// Usando o operador $pull para remover elementos de um array que correspondem à condição
	update := bson.M{
		"$pull": bson.M{
			"custom_fields": bson.M{
				"custom_field_id": bson.M{"$in": customFieldIds},
			},
		},
	}

	// Filtro para encontrar transações que contenham qualquer um dos campos personalizados excluídos
	// e que pertençam ao mesmo workspace
	transactionFilter := bson.M{
		"workspace_id": workspaceId,
		"custom_fields": bson.M{
			"$elemMatch": bson.M{
				"custom_field_id": bson.M{"$in": customFieldIds},
			},
		},
	}

	// Criamos um novo contexto para a operação de atualização das transações
	updateCtx, updateCancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer updateCancel()

	// Atualiza todas as transações que atendem aos critérios
	_, err = transactionCollection.UpdateMany(updateCtx, transactionFilter, update)
	return err
}
