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
	app.router.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		defer conn.Close()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			go app.ProcessMessage(conn, message)
			// conn.WriteMessage(websocket.TextMessage, []byte("Hello websocket"))
			// time.Sleep(time.Second)
		}
	})

	app.router.POST("/create-chat", func(ctx *gin.Context) {
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
	})

	log.Println("starting server at :7007")

	app.router.Run(":7007")
}

func (app *App) ProcessMessage(conn *websocket.Conn, data []byte) {
	var msg repository.Message

	log.Println(data)

	err := json.Unmarshal(data, &msg)
	if err != nil {
		log.Println("error parsing ", err)
		return
	}

	log.Println(msg)

	app.ConnPool[msg.From] = conn

	msgs, err := app.MessageRepo.CreateMessage(context.Background(), &msg)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		return
	}

	conn.WriteMessage(websocket.TextMessage, []byte(msgs.Content))
}
