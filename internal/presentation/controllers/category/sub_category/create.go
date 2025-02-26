package sub_category

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateSubCategoryController struct {
	UpdateCategoryRepository usecase.UpdateCategoryRepository
	Validate                 *validator.Validate
	FindCategoryById         usecase.FindCategoryByIdRepository
}

func NewCreateSubCategoryController(updateCategory usecase.UpdateCategoryRepository, findCategoryById usecase.FindCategoryByIdRepository) *CreateSubCategoryController {
	validate := validator.New(validator.WithRequiredStructEnabled())

	return &CreateSubCategoryController{
		UpdateCategoryRepository: updateCategory,
		Validate:                 validate,
		FindCategoryById:         findCategoryById,
	}
}

type subCategory struct {
	Name string `json:"name" validate:"required,min=3,max=255"`
	Icon string `json:"icon" validate:"required,min=1,max=255"`
}

type subCategoryBody struct {
	CategoryId  string      `json:"categoryId" validate:"required,mongodb"`
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
			Error: helpers.GetErrorMessages(c.Validate, err),
		}, http.StatusUnprocessableEntity)
	}

	categoryId, err := primitive.ObjectIDFromHex(body.CategoryId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid account id",
		}, http.StatusBadRequest)
	}

	workspaceId, err := primitive.ObjectIDFromHex(r.Header.Get("workspaceId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "invalid workspace id",
		}, http.StatusBadRequest)
	}

	category, err := c.FindCategoryById.Find(categoryId, workspaceId)
	if err != nil {
		log.Println(err.Error())
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error while finding a category",
		}, http.StatusInternalServerError)
	}
	if category == nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "category not found",
		}, http.StatusNotFound)
	}

	subCategory, err := c.UpdateCategoryRepository.CreateSubCategory(&models.SubCategoryCategory{
		Name: body.SubCategory.Name,
		Icon: body.SubCategory.Icon,
	}, categoryId, workspaceId)
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "error creating SubCategory",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(subCategory, http.StatusOK)
}
