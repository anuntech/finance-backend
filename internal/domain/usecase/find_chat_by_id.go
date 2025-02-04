package usecase

import "github.com/anuntech/finance-backend/internal/domain/models"

type FindChatById interface {
	Find(id string) (*models.Chat, error)
}
