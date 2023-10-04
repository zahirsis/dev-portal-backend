package http

import (
	"github.com/gin-gonic/gin"
	"github.com/zahirsis/dev-portal-backend/src/app/interfaces"
	"github.com/zahirsis/dev-portal-backend/src/app/usecase"
	"github.com/zahirsis/dev-portal-backend/src/infrastructure/container"
	_ "net/http"
)

type EnvironmentHandler struct {
	*container.Container
	listEnvironmentsUseCase usecase.ListEnvironmentsUseCase
}

func NewEnvironmentHandler(
	c *container.Container,
	r interfaces.Router,
	uc usecase.ListEnvironmentsUseCase,
) *EnvironmentHandler {
	h := &EnvironmentHandler{
		c,
		uc,
	}
	r.GET("", h.ListEnvironments)
	return h
}

func (th *EnvironmentHandler) ListEnvironments(c interfaces.HttpServerContext) {
	l, err := th.listEnvironmentsUseCase.Exec()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "success", "data": l})
}
