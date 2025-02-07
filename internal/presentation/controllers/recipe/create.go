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

type CreateRecipeController struct {
	CreateRecipeRepository usecase.CreateRecipeRepository
	Validate               *validator.Validate
	FindAccountById        usecase.FindAccountByIdRepository
}

func NewCreateRecipeController(createRecipe usecase.CreateRecipeRepository, findAccountById usecase.FindAccountByIdRepository) *CreateRecipeController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &CreateRecipeController{
		CreateRecipeRepository: createRecipe,
		Validate:               validate,
		FindAccountById:        findAccountById,
	}
}

type subRecipeCategory struct {
	Name   string  `json:"name" validate:"required,min=3,max=255"`
	Icon   string  `json:"icon" validate:"required,min=1,max=255"`
	Amount float64 `json:"amount" validate:"required"`
}

type CreateRecipeBody struct {
	Name        string              `json:"name" validate:"required,min=3,max=255"`
	AccountId   string              `json:"accountId" validate:"required"`
	SubCategory []subRecipeCategory `json:"subCategory"`
}

func (c *CreateRecipeController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body CreateRecipeBody
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

	accountId, err := primitive.ObjectIDFromHex(body.AccountId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid account id",
		}, http.StatusBadRequest)
	}

	account, err := c.FindAccountById.Find(accountId.Hex())
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "account not found",
		}, http.StatusNotFound)
	}
	if account == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "account not found",
		}, http.StatusNotFound)
	}

	recipe, err := c.CreateRecipeRepository.Create(models.Recipe{
		Name:      body.Name,
		AccountId: accountId,
		SubCategories: func(subCats []subRecipeCategory) []models.SubRecipeCategory {
			result := make([]models.SubRecipeCategory, len(subCats))
			for i, subCat := range subCats {
				result[i] = models.SubRecipeCategory{
					Name:   subCat.Name,
					Amount: subCat.Amount,
					Icon:   subCat.Icon,
				}
			}
			return result
		}(body.SubCategory),
	})
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error creating recipe",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(recipe, http.StatusOK)
}
