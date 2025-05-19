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
	app.router.GET("/connect/:id", app.Connect)
	app.router.POST("/create-chat", app.CreateChat)

	log.Println("starting server at :7007")

	app.router.Run(":7007")
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
	log.Println(app.ConnPool)

	stop := make(chan bool)

	go app.readMessage(conn, stop)
	// go app.sendMessage(conn)

	// go app.ProcessMessage(conn, message)
	if <-stop { // Blocks until a bool is received
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

// func (app *App) ProcessMessage(conn *websocket.Conn, data []byte) {
// 	var msg repository.Message

// 	err := json.Unmarshal(data, &msg)
// 	if err != nil {
// 		return
// 	}

// 	_, err = app.MessageRepo.CreateMessage(context.Background(), &msg)
// 	if err != nil {
// 		conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
// 		return
// 	}

// 	log.Println(msg)

// 	// conn.WriteJSON(msg)

// 	// conn.WriteMessage(websocket.TextMessage, []byte(msgs.Content))

// 	// live update
// 	if connection, ok := app.ConnPool[msg.To]; ok {
// 		log.Println("there is a connection")
// 		connection.WriteMessage(websocket.TextMessage, []byte("hi there 102"))
// 		connection.WriteJSON(msg)
// 	}
// }
