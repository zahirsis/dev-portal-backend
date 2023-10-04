package memory

import (
	"errors"
	"github.com/zahirsis/dev-portal-backend/src/domain/entity"
	"github.com/zahirsis/dev-portal-backend/src/domain/repository"
	"github.com/zahirsis/dev-portal-backend/src/pkg/logger"
)

const (
	ErrSquadNotFound = "squad not found"
)

type squadRepository struct {
	logger.Logger
}

func NewSquadRepository(l logger.Logger) repository.SquadRepository {
	return &squadRepository{l}
}

func (r *squadRepository) List() ([]entity.SquadEntity, error) {
	return r.memory(), nil
}

func (r *squadRepository) Get(code string) (entity.SquadEntity, error) {
	for _, v := range r.memory() {
		if v.Code() == code {
			return v, nil
		}
	}
	return nil, errors.New(ErrSquadNotFound)
}

func (r *squadRepository) memory() []entity.SquadEntity {
	return []entity.SquadEntity{
		entity.NewSquadEntity("atendimento", "Atendimento"),
		entity.NewSquadEntity("cca", "CCA"),
		entity.NewSquadEntity("cco", "CCO"),
		entity.NewSquadEntity("cd", "CD"),
		entity.NewSquadEntity("devops", "Devops"),
		entity.NewSquadEntity("erp-prestadores", "Erp Prestadores"),
		entity.NewSquadEntity("mms", "MMS"),
		entity.NewSquadEntity("processamento", "Processamento"),
		entity.NewSquadEntity("rpa", "RPA"),
	}
}
