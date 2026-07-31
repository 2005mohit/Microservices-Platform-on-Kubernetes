package api

import (
	"log"
	"net/http"

	"github.com/devplatform/api/internal/db"
	"github.com/devplatform/api/internal/queue"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSHandler struct{}

func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	deployID := c.Param("id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}
	defer conn.Close()

	var logs string
	db.Pool().QueryRow(`SELECT logs FROM deployments WHERE id = $1`, deployID).Scan(&logs)
	if logs != "" {
		conn.WriteMessage(websocket.TextMessage, []byte(logs))
	}

	ch := make(chan string, 100)
	sub := queue.Subscribe("deployment:"+deployID, ch)
	defer queue.Unsubscribe("deployment:"+deployID, sub)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for msg := range ch {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				return
			}
		}
	}()

	conn.SetCloseHandler(func(code int, text string) error {
		return nil
	})

	<-done
}
