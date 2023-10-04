package http

import (
	"errors"
	"github.com/gin-gonic/gin"
	customErrors "github.com/zahirsis/dev-portal-backend/pkg/errors"
	"github.com/zahirsis/dev-portal-backend/src/app/interfaces"
	"github.com/zahirsis/dev-portal-backend/src/app/usecase"
	"github.com/zahirsis/dev-portal-backend/src/infrastructure/container"
)

type CiCdHandler struct {
	*container.Container
	setupUseCase   usecase.SetupCiCdUseCase
	getDataUseCase usecase.GetCiCdDataUseCase
}

func NewCiCdHandler(
	c *container.Container,
	r interfaces.Router,
	suc usecase.SetupCiCdUseCase,
	guc usecase.GetCiCdDataUseCase,
) *CiCdHandler {
	h := &CiCdHandler{
		c,
		suc,
		guc,
	}
	r.POST("setup", h.Setup)
	r.GET("data", h.GetData)
	return h
}

func (th *CiCdHandler) GetData(c interfaces.HttpServerContext) {
	out, err := th.getDataUseCase.Exec()
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "success", "data": out})
}

func (th *CiCdHandler) Setup(c interfaces.HttpServerContext) {
	var requestBody usecase.CiCdInputDto
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	out := th.setupUseCase.Exec(requestBody)

	if len(out.Errors) > 0 {
		c.JSON(400, gin.H{"errors": th.formatErrors(out.Errors), "message": "Setup data is invalid, please check the errors"})
		return
	}
	c.JSON(200, gin.H{"status": "success", "data": out, "message": "Process started"})
}

func (th *CiCdHandler) formatErrors(e []error) map[string][]string {
	errs := make(map[string][]string)
	for _, err := range e {
		var ie *customErrors.InputError
		if err == nil {
			continue
		}
		if errors.As(err, &ie) {
			input := ie.Input
			if _, ok := errs[input]; !ok {
				errs[input] = []string{}
			}
			for _, msg := range ie.Messages {
				errs[input] = append(errs[input], msg)
			}
			continue
		}
		th.Logger.Error("INTERNAL ERROR", err)
		if _, ok := errs["internal"]; !ok {
			errs["internal"] = []string{}
		}
		errs["internal"] = append(errs["internal"], err.Error())
	}
	th.Logger.Debug("FORMATTED ERRORS", errs)
	return errs
}
