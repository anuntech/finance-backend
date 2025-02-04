package usecase

import "github.com/willchat-ofc/api-willchat-golang/internal/domain/models"

type FindMessagesByChatIdInput struct {
	ChatId string
	Limit  int
	Offset int
}

type FindMessagesByChatId interface {
	Find(data *FindMessagesByChatIdInput) ([]*models.Message, error)
}
