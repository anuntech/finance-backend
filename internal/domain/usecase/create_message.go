package usecase

import "github.com/anuntech/finance-backend/internal/domain/models"

type CreateMessageInput struct {
	ChatId     string
	Message    string
	AuthorName string
	AuthorId   string
}

type CreateMessage interface {
	Create(data *CreateMessageInput) (*models.Message, error)
}
