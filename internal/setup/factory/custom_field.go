package factory

import (
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/custom_field_repository"
	controllers "github.com/anuntech/finance-backend/internal/presentation/controllers/custom_field"
	"go.mongodb.org/mongo-driver/mongo"
)

func MakeCreateCustomFieldController(db *mongo.Database) *controllers.CreateCustomFieldController {
	createCustomFieldRepository := custom_field_repository.NewCreateCustomFieldRepository(db)
	return controllers.NewCreateCustomFieldController(createCustomFieldRepository)
}

func MakeGetCustomFieldsController(db *mongo.Database) *controllers.GetCustomFieldsController {
	findCustomFieldsRepository := custom_field_repository.NewFindCustomFieldsRepository(db)
	return controllers.NewGetCustomFieldsController(findCustomFieldsRepository)
}

func MakeGetCustomFieldByIdController(db *mongo.Database) *controllers.GetCustomFieldByIdController {
	findCustomFieldByIdRepository := custom_field_repository.NewFindCustomFieldByIdRepository(db)
	return controllers.NewGetCustomFieldByIdController(findCustomFieldByIdRepository)
}

func MakeUpdateCustomFieldController(db *mongo.Database) *controllers.UpdateCustomFieldController {
	updateCustomFieldRepository := custom_field_repository.NewUpdateCustomFieldRepository(db)
	findCustomFieldByIdRepository := custom_field_repository.NewFindCustomFieldByIdRepository(db)
	return controllers.NewUpdateCustomFieldController(updateCustomFieldRepository, findCustomFieldByIdRepository)
}

func MakeDeleteCustomFieldController(db *mongo.Database) *controllers.DeleteCustomFieldController {
	deleteCustomFieldRepository := custom_field_repository.NewDeleteCustomFieldRepository(db)
	return controllers.NewDeleteCustomFieldController(deleteCustomFieldRepository)
}
