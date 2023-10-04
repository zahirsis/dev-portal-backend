package usecase

import (
	"github.com/zahirsis/dev-portal-backend/src/infrastructure/container"
)

type SquadDto struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

type ListSquadsUseCase interface {
	Exec() ([]SquadDto, error)
}

type listSquadsUseCase struct {
	*container.Container
}

func NewListSquadsUseCase(c *container.Container) ListSquadsUseCase {
	return &listSquadsUseCase{c}
}

func (uc *listSquadsUseCase) Exec() ([]SquadDto, error) {
	var r []SquadDto
	l, err := uc.Repositories.SquadRepository.List()
	if err != nil {
		return nil, err
	}
	for _, v := range l {
		r = append(r, SquadDto{v.Code(), v.Label()})
	}
	return r, nil
}
