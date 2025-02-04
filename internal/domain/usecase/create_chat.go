package usecase

type CreateChatOutput struct {
	Id        string
	OwnerId   string
	CreatedAt string
}

type CreateChat interface {
	Create(ownerId string) (*CreateChatOutput, error)
}
