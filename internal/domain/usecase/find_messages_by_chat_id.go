package usecase

import "github.com/anuntech/finance-backend/internal/domain/models"

type FindMessagesByChatIdInput struct {
	ChatId string
	Limit  int
	Offset int
}

type FindMessagesByChatId interface {
	Find(data *FindMessagesByChatIdInput) ([]*models.Message, error)
}
