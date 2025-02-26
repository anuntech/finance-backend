package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransactionBalance struct {
	Value    int `bson:"value" json:"value"`
	Parts    int `bson:"parts" json:"parts"`       // increase
	Labor    int `bson:"labor" json:"labor"`       // increase
	Discount int `bson:"discount" json:"discount"` // decrease
	Interest int `bson:"interest" json:"interest"` // increase
}

type TransactionRepeatSettings struct {
	InitialInstallment time.Month `bson:"initial_installment" json:"initialInstallment,omitempty"`
	Count              int        `bson:"count" json:"count,omitempty"`
	Interval           string     `bson:"interval" json:"interval,omitempty"` // MONTHLY | DAILY | WEEKLY | QUARTERLY | YEARLY
}

type Transaction struct {
	Id               primitive.ObjectID         `bson:"_id" json:"id"`
	Name             string                     `bson:"name" json:"name"`
	Description      string                     `bson:"description" json:"description,omitempty"`
	CreatedBy        primitive.ObjectID         `bson:"created_by" json:"createdBy"` // email
	Type             string                     `bson:"type" json:"type"`            // EXPENSE, RECIPE
	Supplier         string                     `bson:"supplier" json:"supplier"`
	AssignedTo       primitive.ObjectID         `bson:"assigned_to" json:"assignedTo"`
	Balance          TransactionBalance         `bson:"balance" json:"balance"`
	Frequency        string                     `bson:"frequency" json:"frequency"` // DO_NOT_REPEAT | RECURRING | REPEAT
	RepeatSettings   *TransactionRepeatSettings `bson:"repeat_settings" json:"repeatSettings,omitempty"`
	DueDate          time.Time                  `bson:"due_date" json:"dueDate"`
	IsConfirmed      bool                       `bson:"is_confirmed" json:"isConfirmed"`
	CategoryId       primitive.ObjectID         `bson:"category_id" json:"categoryId"`
	SubCategoryId    primitive.ObjectID         `bson:"sub_category_id" json:"subCategoryId"`
	TagId            *primitive.ObjectID        `bson:"tag_id" json:"tagId"`
	SubTagId         *primitive.ObjectID        `bson:"sub_tag_id" json:"subTagId"`
	AccountId        primitive.ObjectID         `bson:"account_id" json:"accountId"`
	RegistrationDate time.Time                  `bson:"registration_date" json:"registrationDate"`
	ConfirmationDate *time.Time                 `bson:"confirmation_date" json:"confirmationDate,omitempty"`
	IsOverdue        bool                       `bson:"-" json:"isOverdue"`
	CreatedAt        time.Time                  `bson:"created_at" json:"createdAt"`
	UpdatedAt        time.Time                  `bson:"updated_at" json:"updatedAt"`
	WorkspaceId      primitive.ObjectID         `bson:"workspace_id" json:"workspaceId"`
}
