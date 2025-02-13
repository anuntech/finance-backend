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
	CreateCategoryRepository             usecase.CreateCategoryRepository
	Validate                             *validator.Validate
	FindAccountById                      usecase.FindAccountByIdRepository
	FindCategorysByWorkspaceIdRepository usecase.FindCategorysByWorkspaceIdRepository
}

func NewCreateCategoryController(createCategory usecase.CreateCategoryRepository, findAccountById usecase.FindAccountByIdRepository, findCategorysByWorkspaceId usecase.FindCategorysByWorkspaceIdRepository) *CreateCategoryController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &CreateCategoryController{
		CreateCategoryRepository:             createCategory,
		Validate:                             validate,
		FindAccountById:                      findAccountById,
		FindCategorysByWorkspaceIdRepository: findCategorysByWorkspaceId,
	}
}

type subCategoryCategory struct {
	Name   string  `json:"name" validate:"required,min=3,max=255"`
	Icon   string  `json:"icon" validate:"required,min=1,max=255"`
	Amount float64 `json:"amount" validate:"required,min=0"`
}

type CreateCategoryBody struct {
	Name          string                `json:"name" validate:"required,min=3,max=255"`
	SubCategories []subCategoryCategory `json:"subCategories" validate:"dive"`
	Icon          string                `json:"icon" validate:"required,min=1,max=50"`
}

func (c *CreateCategoryController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	var body CreateCategoryBody
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

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspace id",
		}, http.StatusBadRequest)
	}

	categorys, err := c.FindCategorysByWorkspaceIdRepository.Find(workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error finding categorys",
		}, http.StatusInternalServerError)
	}

	if len(categorys) >= 50 {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "user has reached the maximum number of categorys",
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
					Id:     primitive.NewObjectID(),
					Name:   subCat.Name,
					Amount: subCat.Amount,
					Icon:   subCat.Icon,
				}
			}
			return result
		}(body.SubCategories),
	})
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error creating category",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(category, http.StatusOK)
}
