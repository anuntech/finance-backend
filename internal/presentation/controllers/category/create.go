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

type CreateCategoryController struct {
	CreateCategoryRepository     usecase.CreateCategoryRepository
	Validate                     *validator.Validate
	FindAccountById              usecase.FindAccountByIdRepository
	FindCategoriesRepository     usecase.FindCategoriesRepository
	FindCategoryByNameRepository usecase.FindCategoryByNameAndTypeRepository
}

func NewCreateCategoryController(
	createCategory usecase.CreateCategoryRepository,
	findAccountById usecase.FindAccountByIdRepository,
	findCategorysByWorkspaceId usecase.FindCategoriesRepository,
	findCategoryByName usecase.FindCategoryByNameAndTypeRepository,
) *CreateCategoryController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &CreateCategoryController{
		CreateCategoryRepository:     createCategory,
		Validate:                     validate,
		FindAccountById:              findAccountById,
		FindCategoriesRepository:     findCategorysByWorkspaceId,
		FindCategoryByNameRepository: findCategoryByName,
	}
}

type subCategoryCategory struct {
	Name string `json:"name" validate:"required,min=3,max=255"`
	Icon string `json:"icon" validate:"required,min=1,max=255"`
}

type CreateCategoryBody struct {
	Name          string                `json:"name" validate:"required,min=3,max=255"`
	SubCategories []subCategoryCategory `json:"subCategories" validate:"dive"`
	Type          string                `json:"type" validate:"required,oneof=RECIPE EXPENSE TAG"`
	Icon          string                `json:"icon" validate:"required,min=1,max=50"`
}

func (c *CreateCategoryController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body CreateCategoryBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Requisição inválida. Por favor, verifique os dados enviados.",
		}, http.StatusBadRequest)

	}

	if err := c.Validate.Struct(body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: helpers.GetErrorMessages(c.Validate, err),
		}, http.StatusUnprocessableEntity)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "ID do espaço de trabalho inválido.",
		}, http.StatusBadRequest)
	}

	// Verificar se já existe uma categoria com o mesmo nome neste workspace
	existingCategory, err := c.FindCategoryByNameRepository.Find(body.Name, body.Type, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Erro ao verificar o nome da categoria.",
		}, http.StatusInternalServerError)
	}

	if existingCategory != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Já existe uma categoria com este nome neste espaço de trabalho.",
		}, http.StatusConflict)
	}

	// Verificar se há subcategorias com nomes duplicados
	if len(body.SubCategories) > 0 {
		nameSet := make(map[string]bool)
		for _, subCat := range body.SubCategories {
			if nameSet[subCat.Name] {
				return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "Os nomes das subcategorias devem ser únicos dentro da mesma categoria.",
				}, http.StatusBadRequest)
			}
			nameSet[subCat.Name] = true
		}
	}

	categorys, err := c.FindCategoriesRepository.Find(&helpers.GlobalFilterParams{
		WorkspaceId: workspaceId,
		Type:        body.Type,
	})
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Erro ao buscar categorias.",
		}, http.StatusInternalServerError)
	}

	if len(categorys) >= 500 {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Você atingiu o limite máximo de 500 categorias.",
		}, http.StatusBadRequest)
	}

	category, err := c.CreateCategoryRepository.Create(&models.Category{
		Name:        body.Name,
		WorkspaceId: workspaceId,
		Icon:        body.Icon,
		SubCategories: func(subCats []subCategoryCategory) []models.SubCategoryCategory {
			result := make([]models.SubCategoryCategory, len(subCats))
			for i, subCat := range subCats {
				result[i] = models.SubCategoryCategory{
					Id:   primitive.NewObjectID(),
					Name: subCat.Name,
					Icon: subCat.Icon,
				}
			}
			return result
		}(body.SubCategories),
		Type: body.Type,
	})
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "Erro ao criar categoria.",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(category, http.StatusOK)
}
