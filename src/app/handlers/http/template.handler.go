package http

import (
	"github.com/gin-gonic/gin"
	"github.com/zahirsis/dev-portal-backend/src/app/interfaces"
	"github.com/zahirsis/dev-portal-backend/src/app/usecase"
	"github.com/zahirsis/dev-portal-backend/src/infrastructure/container"
	_ "net/http"
)

type TemplateHandler struct {
	*container.Container
	listTemplatesUseCase usecase.ListTemplatesUseCase
}

func NewTemplateHandler(
	c *container.Container,
	r interfaces.Router,
	uc usecase.ListTemplatesUseCase,
) *TemplateHandler {
	h := &TemplateHandler{
		c,
		uc,
	}
	r.GET("", h.ListTemplates)
	return h
}

func (th *TemplateHandler) ListTemplates(c interfaces.HttpServerContext) {
	l, err := th.listTemplatesUseCase.Exec()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "success", "data": l})
}
