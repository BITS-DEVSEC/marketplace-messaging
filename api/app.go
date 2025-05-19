package api

import (
	"context"
	"encoding/json"
	"log"
	"marketplace-messageing/storage/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type App struct {
	router *gin.Engine

	ChatRepo    *repository.ChatRepository
	MessageRepo *repository.MessageRepository

	ConnPool map[int64]*websocket.Conn
}

func NewApp(chatRepo *repository.ChatRepository, messageRepo *repository.MessageRepository) (*App, error) {
	app := &App{
		router:      gin.New(),
		ChatRepo:    chatRepo,
		MessageRepo: messageRepo,
		ConnPool:    make(map[int64]*websocket.Conn),
	}
	return app, nil
}

func (app *App) Run() {
	app.router.GET("/chats/:userID", app.GetUserChats)
	app.router.GET("/connect/:id", app.Connect)
	app.router.POST("/create-chat", app.CreateChat)
	app.router.GET("/messages/:chatID", app.GetUserMessages)

	log.Println("starting server at :7007")

	app.router.Run(":7007")
}

func (app *App) GetUserMessages(ctx *gin.Context) {
	var req = struct {
		ChatID int64 `uri:"chatID"`
	}{}
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	messages, err := app.MessageRepo.GetAllMessages(ctx, req.ChatID)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (app *App) GetUserChats(ctx *gin.Context) {
	var req = struct {
		UserID int64 `uri:"userID"`
	}{}
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	chats, err := app.ChatRepo.GetAllUserChats(ctx, req.UserID)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	type chatResponse struct {
		Chat        *repository.Chat    `json:"chat"`
		LastMessage *repository.Message `json:"last_message"`
	}

	var response []chatResponse

	for _, chat := range chats {
		msg, err := app.MessageRepo.GetChatLastMessage(ctx, chat.ID)
		if err != nil {
			log.Printf("failed to fetch last message for chat %d: %v", chat.ID, err)
			continue
		}

		resp := chatResponse{
			&chat,
			msg,
		}
		response = append(response, resp)
	}

	ctx.JSON(http.StatusOK, gin.H{"response": response})
}

func (app *App) CreateChat(ctx *gin.Context) {
	var chat *repository.Chat
	if err := ctx.ShouldBindBodyWithJSON(&chat); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	chat, err := app.ChatRepo.CreateChat(ctx, chat)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"chat": chat})
}

func (app *App) Connect(ctx *gin.Context) {
	var req = struct {
		ID int64 `uri:"id"`
	}{}
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	app.ConnPool[req.ID] = conn

	stop := make(chan bool)

	go app.readMessage(conn, stop)
	if <-stop {
		return
	}
}

func (app *App) readMessage(conn *websocket.Conn, stop chan bool) {
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			stop <- true
			break
		}

		var msg repository.Message

		err = json.Unmarshal(data, &msg)
		if err != nil {
			return
		}

		_, err = app.MessageRepo.CreateMessage(context.Background(), &msg)
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			return
		}

		log.Println(msg)

		if connection, ok := app.ConnPool[msg.To]; ok {
			log.Println("there is a connection")
			go app.sendMessage(connection, &msg)
		}
	}
}

func (app *App) sendMessage(conn *websocket.Conn, msg *repository.Message) {
	conn.WriteJSON(msg)
}
