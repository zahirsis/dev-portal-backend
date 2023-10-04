package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/zahirsis/dev-portal-backend/src/domain/entity"
	"github.com/zahirsis/dev-portal-backend/src/domain/service"
	"github.com/zahirsis/dev-portal-backend/src/pkg/logger"
	"strings"
)

type registryApiService struct {
	logger logger.Logger
	client *ecr.Client
}

func NewRegistryApiService(logger logger.Logger, client *ecr.Client) service.RegistryApiService {
	return &registryApiService{
		logger: logger,
		client: client,
	}
}

func (r *registryApiService) Create(e entity.RegistryEntity) (string, error) {
	var tags []types.Tag
	var msg string
	for _, tag := range e.Tags() {
		tags = append(tags, types.Tag{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}
	ctx := context.Background()
	created, err := r.client.CreateRepository(ctx, &ecr.CreateRepositoryInput{
		RepositoryName: e.Name(),
		ImageScanningConfiguration: &types.ImageScanningConfiguration{
			ScanOnPush: e.Config().ImageScanningConfiguration.ScanOnPush,
		},
		RegistryId: e.Config().RegistryId,
		Tags:       tags,
	}, func(opt *ecr.Options) { opt.Region = e.Config().Region })
	if err != nil && !strings.Contains(err.Error(), "already exists in the registry with id") {
		return msg, err
	}
	if err != nil {
		name := e.Name()
		r.logger.Info("Repository already exists: "+*name, err)
	}
	if created != nil {
		r.logger.Info("Repository created", created.Repository)
	}
	policySet, err := r.client.SetRepositoryPolicy(
		ctx,
		&ecr.SetRepositoryPolicyInput{
			PolicyText:     e.Policy(),
			RepositoryName: e.Name(),
			Force:          false,
			RegistryId:     e.Config().RegistryId,
		},
		func(opt *ecr.Options) { opt.Region = e.Config().Region })
	if err != nil {
		return msg, err
	}
	r.logger.Info("Policy set", policySet)
	return fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s", *e.Config().RegistryId, e.Config().Region, *e.Name()), nil
}
