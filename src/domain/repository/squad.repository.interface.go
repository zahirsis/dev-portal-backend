package repository

import (
	"github.com/zahirsis/dev-portal-backend/src/domain/entity"
)

type SquadRepository interface {
	List() ([]entity.SquadEntity, error)
	Get(code string) (entity.SquadEntity, error)
}
