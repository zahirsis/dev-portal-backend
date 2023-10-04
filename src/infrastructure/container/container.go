package container

import (
	"github.com/zahirsis/dev-portal-backend/src/domain/repository"
	"github.com/zahirsis/dev-portal-backend/src/domain/service"
	"github.com/zahirsis/dev-portal-backend/src/pkg/logger"
	"github.com/zahirsis/dev-portal-backend/src/pkg/messenger"
)

type Container struct {
	Logger         logger.Logger
	MessageManager messenger.MessageManager
	Repositories   *repository.Container
	Services       *service.Container
}
