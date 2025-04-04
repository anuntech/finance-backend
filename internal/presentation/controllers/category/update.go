package category

import (
	"encoding/json"
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UpdateCategoryController struct {
	UpdateCategoryRepository     usecase.UpdateCategoryRepository
	Validate                     *validator.Validate
	FindCategoryById             usecase.FindCategoryByIdRepository
	FindCategoryByNameRepository usecase.FindCategoryByNameAndTypeRepository
}

func NewUpdateCategoryController(
	updateCategory usecase.UpdateCategoryRepository,
	findCategoryById usecase.FindCategoryByIdRepository,
	findCategoryByName usecase.FindCategoryByNameAndTypeRepository,
) *UpdateCategoryController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &UpdateCategoryController{
		UpdateCategoryRepository:     updateCategory,
		Validate:                     validate,
		FindCategoryById:             findCategoryById,
		FindCategoryByNameRepository: findCategoryByName,
	}
}

type UpdateCategoryBody struct {
	Name string `json:"name" validate:"required,min=3,max=255"`
	Type string `json:"type" validate:"required,oneof=RECIPE EXPENSE TAG"`
	Icon string `json:"icon" validate:"required,min=1,max=50"`
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
			Error: helpers.GetErrorMessages(c.Validate, err),
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

	// Verificar se já existe outra categoria com o mesmo nome neste workspace
	if category.Name != body.Name {
		existingCategory, err := c.FindCategoryByNameRepository.Find(body.Name, body.Type, workspaceId)
		if err != nil {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "error checking category name",
			}, http.StatusInternalServerError)
		}

		if existingCategory != nil && existingCategory.Id != categoryId {
			return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
				Error: "a category with this name already exists in this workspace",
			}, http.StatusConflict)
		}
	}

	category.Name = body.Name
	category.Type = body.Type
	category.Icon = body.Icon

	err = c.UpdateCategoryRepository.UpdateCategory(category)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error updating category",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(category, http.StatusOK)
}
