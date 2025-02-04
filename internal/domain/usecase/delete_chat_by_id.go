package usecase

type DeleteChatById interface {
	Delete(chatId string) error
}
