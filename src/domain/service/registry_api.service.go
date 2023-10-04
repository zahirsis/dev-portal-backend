package service

import "github.com/zahirsis/dev-portal-backend/src/domain/entity"

type RegistryApiService interface {
	Create(entity entity.RegistryEntity) (string, error)
}
