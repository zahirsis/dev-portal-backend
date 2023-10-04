package http

import (
	"github.com/gin-gonic/gin"
	"github.com/zahirsis/dev-portal-backend/src/app/interfaces"
	"github.com/zahirsis/dev-portal-backend/src/app/usecase"
	"github.com/zahirsis/dev-portal-backend/src/infrastructure/container"
	_ "net/http"
)

type SquadHandler struct {
	*container.Container
	listSquadsUseCase usecase.ListSquadsUseCase
}

func NewSquadHandler(
	c *container.Container,
	r interfaces.Router,
	uc usecase.ListSquadsUseCase,
) *SquadHandler {
	h := &SquadHandler{
		c,
		uc,
	}
	r.GET("", h.ListSquads)
	return h
}

func (th *SquadHandler) ListSquads(c interfaces.HttpServerContext) {
	l, err := th.listSquadsUseCase.Exec()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "success", "data": l})
}
