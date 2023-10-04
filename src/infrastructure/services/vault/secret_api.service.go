package vault

import (
	"context"
	"github.com/hashicorp/vault/api"
	"github.com/zahirsis/dev-portal-backend/config"
	"github.com/zahirsis/dev-portal-backend/src/domain/service"
	"github.com/zahirsis/dev-portal-backend/src/pkg/logger"
	"strings"
)

type secretApiService struct {
	cfg    *config.Config
	logger logger.Logger
	api    *api.Client
	auth   api.AuthMethod
}

func NewSecretApiService(cfg *config.Config, logger logger.Logger, api *api.Client, auth api.AuthMethod) service.SecretApiService {
	return &secretApiService{
		cfg:    cfg,
		logger: logger,
		api:    api,
		auth:   auth,
	}
}

func (s *secretApiService) CreateBlank(location, path string) error {
	ctx := context.Background()
	_, err := s.api.Auth().Login(ctx, s.auth)
	if err != nil {
		s.logger.Error("Error logging in to vault", err.Error())
		return err
	}
	_, err = s.api.KVv2(location).Get(ctx, path)
	if err != nil && !strings.HasPrefix(err.Error(), "secret not found") {
		s.logger.Error("Error verifying if secret exists", err.Error(), location, path)
		return err
	} else if err == nil {
		s.logger.Error("Secret already exists, skipping", location, path)
		return nil
	}
	_, err = s.api.KVv2(location).Put(ctx, path, make(map[string]interface{}))
	if err != nil {
		s.logger.Error("Error creating secret", err.Error(), location, path)
		return err
	}
	return nil
}
