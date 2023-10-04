package service

import (
	"fmt"
	"github.com/zahirsis/dev-portal-backend/src/domain/entity"
	"github.com/zahirsis/dev-portal-backend/src/domain/repository"
	"github.com/zahirsis/dev-portal-backend/src/pkg/logger"
	"gopkg.in/yaml.v3"
	"os"
)

type RegistryService interface {
	LoadData(data entity.SetupCiCdEntity, manifest *entity.Manifest, templatesPath string) (entity.RegistryEntity, error)
}

type registryService struct {
	logger       logger.Logger
	repositories *repository.Container
}

func NewRegistryService(logger logger.Logger, repositories *repository.Container) RegistryService {
	return &registryService{
		logger:       logger,
		repositories: repositories,
	}
}

func (r *registryService) LoadData(data entity.SetupCiCdEntity, manifest *entity.Manifest, templatesPath string) (entity.RegistryEntity, error) {
	path := fmt.Sprintf("%s/%s/policy.json", templatesPath, manifest.Dir)
	dat, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config, err := os.ReadFile(fmt.Sprintf("%s/%s/config.yaml", templatesPath, manifest.Dir))
	if err != nil {
		return nil, err
	}
	configData := &entity.RegistryConfig{}
	err = yaml.Unmarshal(config, configData)
	if err != nil {
		return nil, err
	}
	return entity.NewRegistryEntity(data.ApplicationSlug(), string(dat), configData, entity.DefaultTags(data)), nil
}
