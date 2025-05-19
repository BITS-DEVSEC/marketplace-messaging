package main

import (
	"context"
	"log"

	"github.com/gorilla/websocket"

	"marketplace-messageing/api"
	"marketplace-messageing/libs"
	"marketplace-messageing/storage"
	"marketplace-messageing/storage/repository"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var connId = make(map[int64]*websocket.Conn)

func main() {
	ctx := context.Background()

	config, err := libs.NewConfig()
	if err != nil {
		panic(err)
	}

	db, err := storage.New(ctx, config.DB)
	if err != nil {
		panic(err)
	}

	if err = db.Migrate(); err != nil {
		panic(err)
	}

	log.Println("finished db setup")

	chatRepo := repository.NewChatRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	app, err := api.NewApp(chatRepo, messageRepo)
	if err != nil {
		panic(err)
	}

	app.Run()
}
