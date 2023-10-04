package service

import (
	"fmt"
	"github.com/zahirsis/dev-portal-backend/src/domain/entity"
	"github.com/zahirsis/dev-portal-backend/src/pkg/logger"
	"gopkg.in/yaml.v3"
	"os"
)

type SecretService interface {
	LoadData(data entity.SetupCiCdEntity, manifest *entity.Manifest, templatesPath string) (entity.SecretEntity, error)
	SetupNewSecret(secretEntity entity.SecretEntity, env entity.SetupEnvData) error
}

type secretService struct {
	logger logger.Logger
	api    SecretApiService
}

func NewSecretService(logger logger.Logger, api SecretApiService) SecretService {
	return &secretService{
		logger: logger,
		api:    api,
	}
}

func (r *secretService) LoadData(data entity.SetupCiCdEntity, manifest *entity.Manifest, templatesPath string) (entity.SecretEntity, error) {
	config, err := os.ReadFile(fmt.Sprintf("%s/%s/config.yaml", templatesPath, manifest.Dir))
	if err != nil {
		return nil, err
	}
	configData := &entity.SecretConfig{}
	err = yaml.Unmarshal(config, configData)
	if err != nil {
		return nil, err
	}
	return entity.NewSecretEntity(data, configData, entity.DefaultTags(data)), nil
}

func (r *secretService) SetupNewSecret(secretEntity entity.SecretEntity, env entity.SetupEnvData) error {
	return r.api.CreateBlank(secretEntity.Config().GetRootPath(env), secretEntity.Config().GetSecretPath(env))
}
