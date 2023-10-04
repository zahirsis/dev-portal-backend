package websocket

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/zahirsis/dev-portal-backend/src/app/interfaces"
	"github.com/zahirsis/dev-portal-backend/src/app/usecase"
	"github.com/zahirsis/dev-portal-backend/src/infrastructure/container"
)

type CiCdHandler struct {
	*container.Container
	uc usecase.ProgressUseCase
	u  *websocket.Upgrader
}

func NewCiCdHandler(
	c *container.Container,
	r interfaces.Router,
	u *websocket.Upgrader,
	uc usecase.ProgressUseCase,
) *CiCdHandler {
	h := &CiCdHandler{
		c,
		uc,
		u,
	}
	r.GET("progress/ws", h.HandleWebSocket)
	return h
}

func (h *CiCdHandler) HandleWebSocket(c interfaces.HttpServerContext) {
	processID := c.DefaultQuery("id", "")
	if processID == "" {
		// TODO: treat error
		h.Logger.Error("Process ID not found")
		return
	}

	conn, err := h.u.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// TODO: treat error
		h.Logger.Error("Error upgrading connection: %s", err.Error())
		return
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			h.Logger.Error("Error closing connection: %s", err.Error())
		}
	}(conn)

	progressChannel := h.MessageManager.Subscribe(processID)

	oldMessages, err := h.uc.Exec(processID)
	if err != nil {
		// TODO: treat error
		h.Logger.Error("Error reading messages from database: %s", err.Error())
	}
	for _, oldMessage := range oldMessages {
		jsonOldMessage, _ := json.Marshal(oldMessage)
		err := conn.WriteMessage(websocket.TextMessage, jsonOldMessage)
		if err != nil {
			// TODO: treat error
			h.Logger.Error("Error writing message to websocket: %s", err.Error())
			return
		}
	}

	for {
		if h.uc.IsFinished(processID) {
			return
		}
		select {
		case message, ok := <-progressChannel:
			if !ok {
				return // Closed channel
			}
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				// TODO: treat error
				h.Logger.Error("Error writing message to websocket: %s", err.Error())
				return
			}
		}
	}
}
