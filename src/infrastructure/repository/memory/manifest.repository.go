package memory

import (
	"github.com/zahirsis/dev-portal-backend/src/domain/entity"
	"github.com/zahirsis/dev-portal-backend/src/domain/repository"
	"github.com/zahirsis/dev-portal-backend/src/pkg/logger"
)

type manifestRepository struct {
	logger.Logger
}

func NewManifestRepository(l logger.Logger) repository.ManifestRepository {
	return &manifestRepository{l}
}

func (r *manifestRepository) ListDefault() ([]*entity.Manifest, error) {
	return r.memory(), nil
}

func (r *manifestRepository) memory() []*entity.Manifest {
	return []*entity.Manifest{
		{
			Code:  "confluence",
			Label: "Confluence Wiki",
			Type:  entity.WikiManifests,
			Dir:   "manifests/wiki/confluence",
		},
		{
			Code:  "vault-kv-v2",
			Label: "Vault kv v2",
			Type:  entity.SecretManifests,
			Dir:   "manifests/secret/vault-kv-v2",
		},
	}
}
