package repository

import (
	"github.com/zahirsis/dev-portal-backend/src/domain/entity"
)

type TemplateRepository interface {
	List() ([]entity.TemplateEntity, error)
	Get(code string) (entity.TemplateEntity, error)
}
