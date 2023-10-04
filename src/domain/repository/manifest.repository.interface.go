package repository

import (
	"github.com/zahirsis/dev-portal-backend/src/domain/entity"
)

type ManifestRepository interface {
	ListDefault() ([]*entity.Manifest, error)
}
