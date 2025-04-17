package factory

import (
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/account_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/repositories/category_repository"
	controllers "github.com/anuntech/finance-backend/internal/presentation/controllers/category"
	controllers_category "github.com/anuntech/finance-backend/internal/presentation/controllers/category/sub_category"
	"go.mongodb.org/mongo-driver/mongo"
)

func MakeCreateCategoryController(db *mongo.Database) *controllers.CreateCategoryController {
	createCategory := category_repository.NewCreateCategoryRepository(db)
	findAccountById := account_repository.NewFindByIdMongoRepository(db)
	findCategorysByWorkspaceId := category_repository.NewFindCategoriesRepository(db)
	findCategoryByNameAndType := category_repository.NewFindByNameAndTypeMongoRepository(db)
	return controllers.NewCreateCategoryController(createCategory, findAccountById, findCategorysByWorkspaceId, findCategoryByNameAndType)
}

func MakeGetCategorysController(db *mongo.Database) *controllers.GetCategoriesController {
	findCategorysByWorkspaceId := category_repository.NewFindCategoriesRepository(db)
	return controllers.NewGetCategoriesController(findCategorysByWorkspaceId)
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
	findCategoryByNameAndType := category_repository.NewFindByNameAndTypeMongoRepository(db)
	return controllers.NewUpdateCategoryController(updateCategory, findCategoryById, findCategoryByNameAndType)
}

func MakeUpdateSubCategoryController(db *mongo.Database) *controllers_category.UpdateSubCategoryController {
	updateCategory := category_repository.NewUpdateCategoryRepository(db)
	findCategoryById := category_repository.NewFindCategoryByIdRepository(db)
	return controllers_category.NewUpdateSubCategoryController(updateCategory, findCategoryById)
}

func MakeImportCategoryController(db *mongo.Database) *controllers.ImportCategoryController {
	importCategory := category_repository.NewImportCategoriesRepository(db)
	findCategorysByWorkspaceId := category_repository.NewFindCategoriesRepository(db)
	findCategoryByNameAndType := category_repository.NewFindByNameAndTypeMongoRepository(db)
	return controllers.NewImportCategoryController(importCategory, findCategorysByWorkspaceId, findCategoryByNameAndType)
}

func MakeImportSubCategoryController(db *mongo.Database) *controllers_category.ImportSubCategoryController {
	updateCategory := category_repository.NewUpdateCategoryRepository(db)
	findCategoryById := category_repository.NewFindCategoryByIdRepository(db)
	return controllers_category.NewImportSubCategoryController(updateCategory, findCategoryById)
}
