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

type CreateSubCategoryController struct {
	UpdateRecipeRepository usecase.UpdateRecipeRepository
	Validate               *validator.Validate
	FindRecipeById         usecase.FindRecipeByIdRepository
}

func NewCreateSubCategoryController(updateRecipe usecase.UpdateRecipeRepository, findRecipeById usecase.FindRecipeByIdRepository) *CreateSubCategoryController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &CreateSubCategoryController{
		UpdateRecipeRepository: updateRecipe,
		Validate:               validate,
		FindRecipeById:         findRecipeById,
	}
}

type subCategory struct {
	Name   string  `json:"name" validate:"required,min=3,max=255"`
	Icon   string  `json:"icon" validate:"required,min=1,max=255"`
	Amount float64 `json:"amount" validate:"required"`
}

type subCategoryBody struct {
	RecipeId    string      `json:"recipeId" validate:"required,mongodb"`
	SubCategory subCategory `json:"subCategory"`
}

func (c *CreateSubCategoryController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body subCategoryBody
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

	recipeId, err := primitive.ObjectIDFromHex(body.RecipeId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid account id",
		}, http.StatusBadRequest)
	}

	recipe, err := c.FindRecipeById.Find(recipeId.Hex(), r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "account not found",
		}, http.StatusNotFound)
	}
	if recipe == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "account not found",
		}, http.StatusNotFound)
	}

	workspaceId := r.Header.Get("workspaceId")

	subCategory, err := c.UpdateRecipeRepository.CreateSubCategory(models.SubRecipeCategory{
		Name:   body.SubCategory.Name,
		Icon:   body.SubCategory.Icon,
		Amount: body.SubCategory.Amount,
	}, recipeId.Hex(), workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error creating SubCategory",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(subCategory, http.StatusOK)
}
