package memory

import (
	"errors"
	"github.com/zahirsis/dev-portal-backend/src/domain/entity"
	"github.com/zahirsis/dev-portal-backend/src/domain/repository"
	"github.com/zahirsis/dev-portal-backend/src/pkg/logger"
)

const (
	ErrEnvironmentNotFound = "environment not found"
)

type environmentRepository struct {
	logger.Logger
}

func NewEnvironmentRepository(l logger.Logger) repository.EnvironmentRepository {
	return &environmentRepository{l}
}

func (r *environmentRepository) List() ([]entity.EnvironmentEntity, error) {
	return r.memory(), nil
}

func (r *environmentRepository) Get(code string) (entity.EnvironmentEntity, error) {
	for _, e := range r.memory() {
		if e.Code() == code {
			return e, nil
		}
	}
	return nil, errors.New(ErrEnvironmentNotFound)
}

func (r *environmentRepository) memory() []entity.EnvironmentEntity {
	return []entity.EnvironmentEntity{
		entity.NewEnvironmentEntity(&entity.EnvironmentConfig{
			Code:          "qa",
			Label:         "Quality Assurance",
			AccentColor:   "orange",
			DefaultActive: true,
			DefaultReplicas: entity.ResourceObject{
				Min: entity.NumberValueObject{Value: 1, Step: 1, Min: 1, Max: 2},
				Max: entity.NumberValueObject{Value: 1, Step: 1, Min: 1, Max: 2},
			},
			Concurrences:       []string{"dev"},
			RequireApproval:    false,
			DestinationCluster: "qa",
			Project:            "qa",
			SecretsPath:        "qa",
		}),
		entity.NewEnvironmentEntity(&entity.EnvironmentConfig{
			Code:        "dev",
			Label:       "Development",
			AccentColor: "blue",
			DefaultReplicas: entity.ResourceObject{
				Min: entity.NumberValueObject{Value: 1, Step: 1, Min: 1, Max: 2},
				Max: entity.NumberValueObject{Value: 1, Step: 1, Min: 1, Max: 2},
			},
			Concurrences:       []string{"qa"},
			RequireApproval:    false,
			DestinationCluster: "dev",
			Project:            "dev",
			SecretsPath:        "qa",
		}),
		entity.NewEnvironmentEntity(&entity.EnvironmentConfig{
			Code:          "hml",
			Label:         "Homologation",
			AccentColor:   "green",
			DefaultActive: true,
			DefaultReplicas: entity.ResourceObject{
				Min: entity.NumberValueObject{Value: 1, Step: 1, Min: 1, Max: 5},
				Max: entity.NumberValueObject{Value: 1, Step: 1, Min: 1, Max: 5},
			},
			RequireApproval:    true,
			DestinationCluster: "hml",
			Project:            "hml",
			SecretsPath:        "hml",
		}),
		entity.NewEnvironmentEntity(&entity.EnvironmentConfig{
			Code:          "prd",
			Label:         "Production",
			AccentColor:   "red",
			DefaultActive: true,
			DefaultReplicas: entity.ResourceObject{
				Min: entity.NumberValueObject{Value: 2, Step: 1, Min: 1, Max: 20},
				Max: entity.NumberValueObject{Value: 4, Step: 1, Min: 1, Max: 20},
			},
			RequireApproval:    true,
			DestinationCluster: "PRD",
			Project:            "prd",
			SecretsPath:        "prd",
		}),
	}
}
