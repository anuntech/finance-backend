package factory

import (
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/account_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/category_repository"
	controllers "github.com/anuntech/finance-backend/internal/presentation/controllers/category"
	controllers_category "github.com/anuntech/finance-backend/internal/presentation/controllers/category/sub_category"
	"go.mongodb.org/mongo-driver/mongo"
)

func MakeCreateCategoryController(db *mongo.Database) *controllers.CreateCategoryController {
	createCategory := category_repository.NewCreateCategoryRepository(db)
	findAccountById := account_repository.NewFindByIdMongoRepository(db)
	findCategorysByWorkspaceId := category_repository.NewFindCategorysByWorkspaceIdRepository(db)
	return controllers.NewCreateCategoryController(createCategory, findAccountById, findCategorysByWorkspaceId)
}

func MakeGetCategorysController(db *mongo.Database) *controllers.GetCategorysController {
	findCategorysByWorkspaceId := category_repository.NewFindCategorysByWorkspaceIdRepository(db)
	return controllers.NewGetCategorysController(findCategorysByWorkspaceId)
}

func MakeCreateSubCategoryController(db *mongo.Database) *controllers_category.CreateSubCategoryController {
	updateCategory := category_repository.NewUpdateCategoryRepository(db)
	findCategoryById := category_repository.NewFindCategoryByIdRepository(db)
	return controllers_category.NewCreateSubCategoryController(updateCategory, findCategoryById)
}

func MakeDeleteSubCategoryController(db *mongo.Database) *controllers_category.DeleteSubCategoryController {
	updateCategory := category_repository.NewUpdateCategoryRepository(db)
	return controllers_category.NewDeleteSubCategoryController(updateCategory)
}

func MakeDeleteCategoryController(db *mongo.Database) *controllers.DeleteCategoryController {
	deleteCategory := category_repository.NewDeleteCategoryRepository(db)
	return controllers.NewDeleteCategoryController(deleteCategory)
}

func MakeGetCategoryByIdController(db *mongo.Database) *controllers.GetCategoryByIdController {
	findCategoryById := category_repository.NewFindCategoryByIdRepository(db)
	return controllers.NewGetCategoryByIdController(findCategoryById)
}

func MakeUpdateCategoryController(db *mongo.Database) *controllers.UpdateCategoryController {
	updateCategory := category_repository.NewUpdateCategoryRepository(db)
	findCategoryById := category_repository.NewFindCategoryByIdRepository(db)
	return controllers.NewUpdateCategoryController(updateCategory, findCategoryById)
}

func MakeUpdateSubCategoryController(db *mongo.Database) *controllers_category.UpdateSubCategoryController {
	updateCategory := category_repository.NewUpdateCategoryRepository(db)
	findCategoryById := category_repository.NewFindCategoryByIdRepository(db)
	return controllers_category.NewUpdateSubCategoryController(updateCategory, findCategoryById)
}
