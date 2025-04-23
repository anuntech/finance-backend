package models

type CreditCard struct {
	Id          string  `json:"id" bson:"_id"`
	WorkspaceId string  `json:"workspaceId" bson:"workspace_id"`
	Name        string  `json:"name" bson:"name"`
	DueDate     int     `json:"dueDate" bson:"due_date"`
	CloseDate   int     `json:"closeDate" bson:"close_date"`
	Limit       float64 `json:"limit" bson:"limit"`
	Balance     float64 `json:"balance" bson:"balance"`
}
