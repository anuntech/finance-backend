package usecase

import "github.com/willchat-ofc/api-willchat-golang/internal/domain/models"

type FindChatById interface {
	Find(id string) (*models.Chat, error)
}
