package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransactionBalance struct {
	Value              float64 `bson:"value" json:"value"`
	Discount           float64 `bson:"discount" json:"discount"`                      // decrease
	Interest           float64 `bson:"interest" json:"interest"`                      // increase
	DiscountPercentage float64 `bson:"discount_percentage" json:"discountPercentage"` // decrease
	InterestPercentage float64 `bson:"interest_percentage" json:"interestPercentage"` // increase
	NetBalance         float64 `bson:"-" json:"netBalance,omitempty"`
}

type TransactionRepeatSettings struct {
	InitialInstallment time.Month `bson:"initial_installment" json:"initialInstallment,omitempty"`
	Count              int        `bson:"count" json:"count,omitempty"`
	CurrentCount       int        `bson:"-" json:"currentCount,omitempty"`
	Interval           string     `bson:"interval" json:"interval,omitempty"` // MONTHLY | DAILY | WEEKLY | QUARTERLY | YEARLY | CUSTOM
	CustomDay          int        `bson:"custom_day" json:"customDay,omitempty"`
}

type TransactionTags struct {
	TagId    primitive.ObjectID `bson:"tag_id" json:"tagId"`
	SubTagId primitive.ObjectID `bson:"sub_tag_id" json:"subTagId"`
}

type TransactionCustomField struct {
	CustomFieldId primitive.ObjectID `bson:"custom_field_id" json:"id"`
	Value         string             `bson:"value" json:"value"`
	Type          string             `bson:"-" json:"type"` // TEXT | NUMBER | DATE | BOOLEAN
}

type Transaction struct {
	Id               primitive.ObjectID         `bson:"_id" json:"id"`
	Name             string                     `bson:"name" json:"name"`
	MainId           *primitive.ObjectID        `bson:"main_id" json:"mainId,omitempty"`
	MainCount        *int                       `bson:"main_count" json:"mainCount,omitempty"`
	Description      string                     `bson:"description" json:"description,omitempty"`
	CreatedBy        primitive.ObjectID         `bson:"created_by" json:"createdBy"` // email
	Invoice          string                     `bson:"invoice" json:"invoice,omitempty"`
	Type             string                     `bson:"type" json:"type"` // EXPENSE, RECIPE
	Supplier         string                     `bson:"supplier" json:"supplier"`
	AssignedTo       primitive.ObjectID         `bson:"assigned_to" json:"assignedTo"`
	Balance          TransactionBalance         `bson:"balance" json:"balance"`
	TotalBalance     float64                    `bson:"-" json:"totalBalance,omitempty"`
	Frequency        string                     `bson:"frequency" json:"frequency"` // DO_NOT_REPEAT | RECURRING | REPEAT
	RepeatSettings   *TransactionRepeatSettings `bson:"repeat_settings" json:"repeatSettings,omitempty"`
	DueDate          time.Time                  `bson:"due_date" json:"dueDate"`
	IsConfirmed      bool                       `bson:"is_confirmed" json:"isConfirmed"`
	CategoryId       *primitive.ObjectID        `bson:"category_id" json:"categoryId"`
	SubCategoryId    *primitive.ObjectID        `bson:"sub_category_id" json:"subCategoryId"`
	Tags             []TransactionTags          `bson:"tags" json:"tags"`
	AccountId        *primitive.ObjectID        `bson:"account_id" json:"accountId"`
	RegistrationDate time.Time                  `bson:"registration_date" json:"registrationDate"`
	ConfirmationDate *time.Time                 `bson:"confirmation_date" json:"confirmationDate,omitempty"`
	IsOverdue        bool                       `bson:"-" json:"isOverdue"`
	CreatedAt        time.Time                  `bson:"created_at" json:"createdAt"`
	UpdatedAt        time.Time                  `bson:"updated_at" json:"updatedAt"`
	WorkspaceId      primitive.ObjectID         `bson:"workspace_id" json:"workspaceId"`
	CustomFields     []TransactionCustomField   `bson:"custom_fields" json:"customFields"`
}
