package usecase

import "github.com/willchat-ofc/api-willchat-golang/internal/domain/models"

type CreateMessageInput struct {
	ChatId     string
	Message    string
	AuthorName string
	AuthorId   string
}

type CreateMessage interface {
	Create(data *CreateMessageInput) (*models.Message, error)
}
