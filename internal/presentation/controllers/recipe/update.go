package recipe

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

type UpdateRecipeController struct {
	UpdateRecipeRepository usecase.UpdateRecipeRepository
	Validate               *validator.Validate
	FindRecipeById         usecase.FindRecipeByIdRepository
}

func NewUpdateRecipeController(updateRecipe usecase.UpdateRecipeRepository, findRecipeById usecase.FindRecipeByIdRepository) *UpdateRecipeController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &UpdateRecipeController{
		UpdateRecipeRepository: updateRecipe,
		Validate:               validate,
		FindRecipeById:         findRecipeById,
	}
}

type UpdateRecipeBody struct {
	Name          string              `json:"name" validate:"required,min=3,max=255"`
	SubCategories []subRecipeCategory `json:"subCategories" validate:"dive"`
}

func (c *UpdateRecipeController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body UpdateRecipeBody
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

	recipeId, err := primitive.ObjectIDFromHex(r.Req.PathValue("recipeId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid recipeId format",
		}, http.StatusBadRequest)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspaceId format",
		}, http.StatusBadRequest)
	}

	recipe, err := c.FindRecipeById.Find(recipeId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "recipe not found",
		}, http.StatusNotFound)
	}

	recipe.Name = body.Name
	recipe.SubCategories = func(subCats []subRecipeCategory) []models.SubRecipeCategory {
		result := make([]models.SubRecipeCategory, len(subCats))
		for i, subCat := range subCats {
			result[i] = models.SubRecipeCategory{
				Id:     primitive.NewObjectID(),
				Name:   subCat.Name,
				Amount: subCat.Amount,
				Icon:   subCat.Icon,
			}
		}
		return result
	}(body.SubCategories)

	err = c.UpdateRecipeRepository.UpdateRecipe(*recipe)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error updating recipe",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(recipe, http.StatusOK)
}
