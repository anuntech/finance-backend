package controllers

import (
	"net/http"

	"github.com/willchat-ofc/api-willchat-golang/internal/domain/usecase"
	"github.com/willchat-ofc/api-willchat-golang/internal/presentation/helpers"
	presentationProtocols "github.com/willchat-ofc/api-willchat-golang/internal/presentation/protocols"
)

type CreateChatController struct {
	CreateChat usecase.CreateChat
}

func NewCreateChatController(createChat usecase.CreateChat) *CreateChatController {
	return &CreateChatController{
		CreateChat: createChat,
	}
}

type CreateChatControllerResponse struct {
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
	OwnerId   string `json:"owner_id"`
}

func (c *CreateChatController) Handle(r presentationProtocols.HttpRequest) *presentationProtocols.HttpResponse {
	chat, err := c.CreateChat.Create(r.Header.Get("UserId"))
	if err != nil {
		return helpers.CreateResponse(&presentationProtocols.ErrorResponse{
			Error: "an error ocurred when creating chat",
		}, http.StatusInternalServerError)
	}

	return helpers.CreateResponse(&CreateChatControllerResponse{
		Id:        chat.Id,
		CreatedAt: chat.CreatedAt,
		OwnerId:   chat.OwnerId,
	}, http.StatusOK)
}
