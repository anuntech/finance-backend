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

type ImportSubCategoryController struct {
	UpdateCategoryRepository   usecase.UpdateCategoryRepository
	Validate                   *validator.Validate
	FindCategoryByIdRepository usecase.FindCategoryByIdRepository
}

func NewImportSubCategoryController(updateCategory usecase.UpdateCategoryRepository, findCategoryById usecase.FindCategoryByIdRepository) *ImportSubCategoryController {
	validate := validator.New()

	return &ImportSubCategoryController{
		UpdateCategoryRepository:   updateCategory,
		Validate:                   validate,
		FindCategoryByIdRepository: findCategoryById,
	}
}

type ImportSubCategoryBody struct {
	SubCategories []ImportSubCategoryInput `json:"subCategories" validate:"required,dive"`
}

type ImportSubCategoryInput struct {
	Name string `json:"name" validate:"required,min=3,max=255"`
	Icon string `json:"icon" validate:"required,min=1,max=50"`
}

func (c *ImportSubCategoryController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body ImportSubCategoryBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid request body",
		}, http.StatusBadRequest)
	}

	if err := c.Validate.Struct(body); err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: helpers.GetErrorMessages(c.Validate, err),
		}, http.StatusUnprocessableEntity)
	}

	categoryIdStr := r.Req.PathValue("categoryId")
	categoryId, err := primitive.ObjectIDFromHex(categoryIdStr)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid categoryId format",
		}, http.StatusBadRequest)
	}

	workspaceIdStr := r.Header.Get("workspaceId")
	workspaceId, err := primitive.ObjectIDFromHex(workspaceIdStr)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspaceId format",
		}, http.StatusBadRequest)
	}

	category, err := c.FindCategoryByIdRepository.Find(categoryId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error fetching category",
		}, http.StatusInternalServerError)
	}
	if category == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "category not found",
		}, http.StatusNotFound)
	}

	if len(category.SubCategories)+len(body.SubCategories) > 50 {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "import exceeds maximum allowed subcategories (50)",
		}, http.StatusBadRequest)
	}

	var subCategoryInputs []models.SubCategoryCategory
	for _, sc := range body.SubCategories {
		subCategoryInputs = append(subCategoryInputs, models.SubCategoryCategory{
			Name: sc.Name,
			Icon: sc.Icon,
			Id:   primitive.NewObjectID(),
		})
	}

	_, err = c.UpdateCategoryRepository.CreateSubCategories(subCategoryInputs, categoryId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error importing subcategories: " + err.Error(),
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(subCategoryInputs, http.StatusOK)
}
