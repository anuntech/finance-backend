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

type UpdateCategoryController struct {
	UpdateCategoryRepository usecase.UpdateCategoryRepository
	Validate                 *validator.Validate
	FindCategoryById         usecase.FindCategoryByIdRepository
}

func NewUpdateCategoryController(updateCategory usecase.UpdateCategoryRepository, findCategoryById usecase.FindCategoryByIdRepository) *UpdateCategoryController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &UpdateCategoryController{
		UpdateCategoryRepository: updateCategory,
		Validate:                 validate,
		FindCategoryById:         findCategoryById,
	}
}

type UpdateCategoryBody struct {
	Name          string                `json:"name" validate:"required,min=3,max=255"`
	Type          string                `json:"type" validate:"required,oneof=recipe expense"`
	SubCategories []subCategoryCategory `json:"subCategories" validate:"dive"`
}

func (c *UpdateCategoryController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body UpdateCategoryBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid body request",
		}, http.StatusBadRequest)
	}

	if err := c.Validate.Struct(body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: err.Error(),
		}, http.StatusUnprocessableEntity)
	}

	categoryId, err := primitive.ObjectIDFromHex(r.Req.PathValue("categoryId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid categoryId format",
		}, http.StatusBadRequest)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspaceId format",
		}, http.StatusBadRequest)
	}

	category, err := c.FindCategoryById.Find(categoryId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "category not found",
		}, http.StatusNotFound)
	}

	category.Name = body.Name
	category.Type = body.Type
	category.SubCategories = func(subCats []subCategoryCategory) []models.SubCategoryCategory {
		result := make([]models.SubCategoryCategory, len(subCats))
		for i, subCat := range subCats {
			result[i] = models.SubCategoryCategory{
				Id:     primitive.NewObjectID(),
				Name:   subCat.Name,
				Icon:   subCat.Icon,
				Amount: 0,
			}
		}
		return result
	}(body.SubCategories)

	err = c.UpdateCategoryRepository.UpdateCategory(category)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error updating category",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(category, http.StatusOK)
}
