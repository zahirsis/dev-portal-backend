package repository

import (
	"github.com/zahirsis/dev-portal-backend/src/domain/entity"
)

type EnvironmentRepository interface {
	List() ([]entity.EnvironmentEntity, error)
	Get(code string) (entity.EnvironmentEntity, error)
}
