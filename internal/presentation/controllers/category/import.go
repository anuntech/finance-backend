package category

import (
	"encoding/json"
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ImportCategoryController struct {
	ImportCategoriesRepository   usecase.ImportCategoriesRepository
	Validate                     *validator.Validate
	FindCategoriesRepository     usecase.FindCategoriesRepository
	FindCategoryByNameRepository usecase.FindCategoryByNameAndTypeRepository
}

func NewImportCategoryController(
	importUseCase usecase.ImportCategoriesRepository,
	findCategorys usecase.FindCategoriesRepository,
	findCategoryByName usecase.FindCategoryByNameAndTypeRepository,
) *ImportCategoryController {
	validate := validator.New()

	return &ImportCategoryController{
		ImportCategoriesRepository:   importUseCase,
		Validate:                     validate,
		FindCategoriesRepository:     findCategorys,
		FindCategoryByNameRepository: findCategoryByName,
	}
}

type ImportCategoryBody struct {
	Categories []ImportCategoryInput `json:"categories" validate:"required,dive"`
}

type ImportCategoryInput struct {
	Name          string                `json:"name" validate:"required,min=3,max=255"`
	SubCategories []subCategoryCategory `json:"subCategories" validate:"dive"`
	Type          string                `json:"type" validate:"required,oneof=RECIPE EXPENSE TAG"`
	Icon          string                `json:"icon" validate:"required,min=1,max=50"`
}

func (c *ImportCategoryController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body ImportCategoryBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "corpo da requisição inválido",
		}, http.StatusBadRequest)
	}

	if err := c.Validate.Struct(body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: helpers.GetErrorMessages(c.Validate, err),
		}, http.StatusUnprocessableEntity)
	}

	workspaceIdStr := r.Header.Get("workspaceId")
	workspaceId, err := primitive.ObjectIDFromHex(workspaceIdStr)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "formato de workspaceId inválido",
		}, http.StatusBadRequest)
	}

	currentCategories, err := c.FindCategoriesRepository.Find(&helpers.GlobalFilterParams{
		WorkspaceId: workspaceId,
	})
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "erro ao buscar categorias existentes",
		}, http.StatusInternalServerError)
	}

	if len(currentCategories)+len(body.Categories) > 50 {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "importação excede o número máximo de categorias permitidas (50)",
		}, http.StatusBadRequest)
	}

	// Verificar duplicidade de nomes entre as categorias a serem importadas
	categoryNameSet := make(map[string]bool)
	for _, cat := range body.Categories {
		// Verificar se a categoria já existe na importação
		if categoryNameSet[cat.Name] {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "a importação contém categorias com nomes duplicados: " + cat.Name,
			}, http.StatusBadRequest)
		}
		categoryNameSet[cat.Name] = true

		// Verificar se o nome já existe no workspace
		existingCategory, err := c.FindCategoryByNameRepository.Find(cat.Name, cat.Type, workspaceId)
		if err != nil {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "erro ao verificar duplicidade de nomes de categorias: " + err.Error(),
			}, http.StatusInternalServerError)
		}

		if existingCategory != nil {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "já existe uma categoria com o nome '" + cat.Name + "' neste workspace",
			}, http.StatusConflict)
		}

		// Verificar duplicidade de nomes entre as subcategorias
		subCategoryNameSet := make(map[string]bool)
		for _, subCat := range cat.SubCategories {
			if subCategoryNameSet[subCat.Name] {
				return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "a categoria '" + cat.Name + "' contém subcategorias com nomes duplicados: " + subCat.Name,
				}, http.StatusBadRequest)
			}
			subCategoryNameSet[subCat.Name] = true
		}
	}

	var categoryInputs []models.Category
	for _, cat := range body.Categories {
		var mappedSubCategories []models.SubCategoryCategory
		for _, sc := range cat.SubCategories {
			mappedSubCategories = append(mappedSubCategories, models.SubCategoryCategory{
				Name: sc.Name,
				Icon: sc.Icon,
				Id:   primitive.NewObjectID(),
			})
		}

		categoryInputs = append(categoryInputs, models.Category{
			Name:          cat.Name,
			Type:          cat.Type,
			Icon:          cat.Icon,
			WorkspaceId:   workspaceId,
			SubCategories: mappedSubCategories,
		})
	}

	importedCategories, err := c.ImportCategoriesRepository.Import(categoryInputs, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "erro ao importar categorias: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(importedCategories, http.StatusOK)
}
