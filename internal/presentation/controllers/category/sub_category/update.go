package sub_category

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

type UpdateSubCategoryController struct {
	UpdateCategoryRepository usecase.UpdateCategoryRepository
	Validate                 *validator.Validate
	FindCategoryById         usecase.FindCategoryByIdRepository
}

func NewUpdateSubCategoryController(updateCategory usecase.UpdateCategoryRepository, findCategoryById usecase.FindCategoryByIdRepository) *UpdateSubCategoryController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &UpdateSubCategoryController{
		UpdateCategoryRepository: updateCategory,
		Validate:                 validate,
		FindCategoryById:         findCategoryById,
	}
}

type subCategoryUpdateBody struct {
	Name string `json:"name" validate:"required,min=3,max=255"`
	Icon string `json:"icon" validate:"required,min=1,max=255"`
}

func (c *UpdateSubCategoryController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	categoryIdStr := r.Req.PathValue("categoryId")
	subCategoryIdStr := r.Req.PathValue("subCategoryId")

	categoryId, err := primitive.ObjectIDFromHex(categoryIdStr)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid categoryId format",
		}, http.StatusBadRequest)
	}

	subCategoryId, err := primitive.ObjectIDFromHex(subCategoryIdStr)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid subCategoryId format",
		}, http.StatusBadRequest)
	}

	workspaceIdStr := r.Header.Get("workspaceId")
	workspaceId, err := primitive.ObjectIDFromHex(workspaceIdStr)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspaceId format",
		}, http.StatusBadRequest)
	}

	var body subCategoryUpdateBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid body request",
		}, http.StatusBadRequest)
	}

	if err := c.Validate.Struct(body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: helpers.GetErrorMessages(c.Validate, err),
		}, http.StatusUnprocessableEntity)
	}

	category, err := c.FindCategoryById.Find(categoryId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error occurred while finding the category",
		}, http.StatusInternalServerError)
	}
	if category == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "category not found",
		}, http.StatusNotFound)
	}

	// Encontrar a subcategoria atual para verificar se o nome está sendo alterado
	var currentSubCategory *models.SubCategoryCategory
	for _, sc := range category.SubCategories {
		if sc.Id == subCategoryId {
			currentSubCategory = &sc
			break
		}
	}

	if currentSubCategory == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "subcategory not found in this category",
		}, http.StatusNotFound)
	}

	// Se o nome for alterado, verificar duplicidade
	if currentSubCategory.Name != body.Name {
		// Verificar se já existe uma subcategoria com o mesmo nome nesta categoria
		for _, existingSubCat := range category.SubCategories {
			if existingSubCat.Name == body.Name && existingSubCat.Id != subCategoryId {
				return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
					Error: "a subcategory with this name already exists in this category",
				}, http.StatusConflict)
			}
		}
	}

	err = c.UpdateCategoryRepository.UpdateSubCategory(&models.SubCategoryCategory{
		Id:   subCategoryId,
		Name: body.Name,
		Icon: body.Icon,
	}, categoryId, subCategoryId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error updating SubCategory",
		}, http.StatusInternalServerError)
	}

	category.CalculateTotalAmount()
	err = c.UpdateCategoryRepository.UpdateCategory(category)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error updating category total amount",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(nil, http.StatusOK)
}
