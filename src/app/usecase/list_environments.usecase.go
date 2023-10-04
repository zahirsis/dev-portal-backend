package usecase

import (
	"github.com/zahirsis/dev-portal-backend/src/domain/entity"
	"github.com/zahirsis/dev-portal-backend/src/infrastructure/container"
)

type EnvironmentDto struct {
	Code            string                `json:"code"`
	Label           string                `json:"label"`
	AccentColor     string                `json:"accent_color"`
	DefaultActive   bool                  `json:"default_active"`
	DefaultReplicas entity.ResourceObject `json:"default_replicas"`
	Concurrences    []string              `json:"concurrences"`
}

type ListEnvironmentsUseCase interface {
	Exec() ([]EnvironmentDto, error)
}

type listEnvironmentsUseCase struct {
	*container.Container
}

func NewListEnvironmentsUseCase(c *container.Container) ListEnvironmentsUseCase {
	return &listEnvironmentsUseCase{c}
}

func (uc *listEnvironmentsUseCase) Exec() ([]EnvironmentDto, error) {
	var r []EnvironmentDto
	l, err := uc.Repositories.EnvironmentRepository.List()
	if err != nil {
		return nil, err
	}
	for _, v := range l {
		r = append(r, EnvironmentDto{v.Code(), v.Label(), v.AccentColor(), v.DefaultActive(), v.DefaultReplicas(), v.Concurrences()})
	}
	return r, nil
}
