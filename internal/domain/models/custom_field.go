package models

type CustomField struct {
	Id          string   `json:"id" bson:"_id"`
	WorkspaceId string   `json:"workspaceId" bson:"workspace_id"`
	Name        string   `json:"name" bson:"name"`
	Type        string   `json:"type" bson:"type"`                 // SELECT, TEXT, NUMBER, DATE, BOOLEAN
	Options     []string `json:"options,omitempty" bson:"options"` // SELECT
	Required    bool     `json:"required" bson:"required"`         // true, false
}
