package usecase

type GetAllChatsByOwnerIdOutput struct {
	Id        string `json:"id"`
	CreatedAt string `json:"createdAt"`
	OwnerId   string `json:"ownerId"`
}

type GetAllChatsByOwnerId interface {
	Get(ownerId string) ([]*GetAllChatsByOwnerIdOutput, error)
}
