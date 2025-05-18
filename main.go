package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	router := gin.Default()

	router.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		defer conn.Close()

		for {
			conn.WriteMessage(websocket.TextMessage, []byte("Hello websocket"))
			time.Sleep(time.Second)
		}
	})

	log.Println("starting server at :7007")

	router.Run(":7007")
}
