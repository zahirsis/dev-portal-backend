package usecase

import (
	"github.com/zahirsis/dev-portal-backend/src/domain/entity"
	"github.com/zahirsis/dev-portal-backend/src/infrastructure/container"
)

type TemplateDto struct {
	Code                string                   `json:"code"`
	Label               string                   `json:"label"`
	ApplicationDefaults entity.ApplicationObject `json:"applicationDefaults"`
	IngressDefaults     entity.IngressObject     `json:"ingressDefaults"`
	Manifests           []*entity.Manifest       `json:"manifests"`
}

type ListTemplatesUseCase interface {
	Exec() ([]TemplateDto, error)
}

type listTemplatesUseCase struct {
	*container.Container
}

func NewListTemplatesUseCase(c *container.Container) ListTemplatesUseCase {
	return &listTemplatesUseCase{c}
}

func (uc *listTemplatesUseCase) Exec() ([]TemplateDto, error) {
	var r []TemplateDto
	l, err := uc.Repositories.TemplateRepository.List()
	if err != nil {
		return nil, err
	}
	for _, v := range l {
		r = append(r, TemplateDto{v.Code(), v.Label(), v.ApplicationDefault(), v.IngressDefault(), v.Manifests()})
	}
	return r, nil
}
