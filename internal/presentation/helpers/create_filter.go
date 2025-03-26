package helpers

import (
	"net/http"
	"net/url"
	"strconv"

	presentationProtocols "github.com/anuntech/finance-backend/internal/presentation/protocols"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GlobalFilterParams struct {
	Month       int    `json:"month" validate:"required_with=Year,omitempty,min=1,max=12"`
	Year        int    `json:"year" validate:"required_with=Month,omitempty,min=1,max=9999"`
	Type        string `json:"type" validate:"omitempty,oneof=RECIPE EXPENSE TAG ALL"` // all is only for custom field
	InitialDate string `json:"initialDate" validate:"omitempty,datetime=2006-01-02"`
	FinalDate   string `json:"finalDate" validate:"omitempty,datetime=2006-01-02"`
	WorkspaceId primitive.ObjectID
}

func GetGlobalFilterByQueries(urlQueries *url.Values, workspaceId primitive.ObjectID, validator *validator.Validate) (*GlobalFilterParams, *presentationProtocols.HttpResponse) {
	monthInt, _ := strconv.Atoi(urlQueries.Get("month"))
	yearInt, _ := strconv.Atoi(urlQueries.Get("year"))

	params := &GlobalFilterParams{
		Month:       monthInt,
		Year:        yearInt,
		Type:        urlQueries.Get("type"),
		WorkspaceId: workspaceId,
		InitialDate: urlQueries.Get("initialDate"),
		FinalDate:   urlQueries.Get("finalDate"),
	}

	err := validator.Struct(params)
	if err != nil {
		return nil, CreateResponse(&presentationProtocols.ErrorResponse{
			Error: GetErrorMessages(validator, err),
		}, http.StatusBadRequest)
	}

	return params, nil
}
