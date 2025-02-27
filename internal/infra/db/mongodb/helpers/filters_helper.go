package helpers

import (
	"time"

	presentationHelpers "github.com/anuntech/finance-backend/internal/presentation/helpers"
	"go.mongodb.org/mongo-driver/bson"
)

func GetFilterByType(filterType string) bson.M {
	if filterType == "RECIPE" {
		return bson.M{"type": "RECIPE"}
	}
	return bson.M{"type": "EXPENSE"}
}

func BuildTransactionFilter(filters *presentationHelpers.GlobalFilterParams) bson.M {
	startOfMonth := time.Date(filters.Year, time.Month(filters.Month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	filter := bson.M{
		"workspace_id": filters.WorkspaceId,
	}
	if filters.Type != "" {
		filter["type"] = filters.Type
	}

	if filters.Month != 0 {
		filter["$or"] = []bson.M{
			{
				"due_date": bson.M{
					"$gte": startOfMonth,
					"$lt":  endOfMonth,
				},
				"is_confirmed": false,
			},
			{
				"confirmation_date": bson.M{
					"$gte": startOfMonth,
					"$lt":  endOfMonth,
				},
				"is_confirmed": true,
			},
		}
	}

	return filter
}
