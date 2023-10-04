package service

import (
	"errors"
	"fmt"
	"github.com/zahirsis/dev-portal-backend/config"
	"github.com/zahirsis/dev-portal-backend/src/domain/entity"
	"github.com/zahirsis/dev-portal-backend/src/pkg/logger"
	"gopkg.in/yaml.v3"
	"os"
)

type SetupPipelineData struct {
	Entity                   entity.PipelineEntity
	Env                      entity.SetupEnvData
	TemplatesPathBase        string
	TemplatesPathNamespace   string
	TemplatesPathApplication string
	PipelinePathBae          string
	PipelinePathNamespace    string
	PipelinePathApplication  string
}

type PipelineService interface {
	LoadData(data entity.SetupCiCdEntity, manifest *entity.Manifest, templatesPath string) (entity.PipelineEntity, error)
	SetupPipeline(e entity.PipelineEntity, templatesPath, pipelinePath string) error
}

type pipelineService struct {
	config           *config.Config
	logger           logger.Logger
	directoryService DirectoryService
}

func NewPipelineService(config *config.Config, logger logger.Logger, directoryService DirectoryService) PipelineService {
	return &pipelineService{
		config:           config,
		logger:           logger,
		directoryService: directoryService,
	}
}

func (g *pipelineService) LoadData(data entity.SetupCiCdEntity, manifest *entity.Manifest, templatesPath string) (entity.PipelineEntity, error) {
	cfg, err := os.ReadFile(fmt.Sprintf("%s/%s/config.yaml", templatesPath, manifest.Dir))
	if err != nil {
		return nil, err
	}
	configData := &entity.PipelineConfig{}
	err = yaml.Unmarshal(cfg, configData)
	if err != nil {
		g.logger.Error("Error unmarshalling config", err.Error(), string(cfg))
		return nil, err
	}
	return entity.NewPipelineEntity(data, configData, entity.DefaultTags(data)), nil
}

type PipelineData struct {
	Environments     map[string]*entity.PipelineEnvironment `json:"environments" yaml:"environments"`
	DefaultVariables []*entity.PipelineVariable             `json:"defaultVariables" yaml:"defaultVariables"`
}

func (g *pipelineService) SetupPipeline(e entity.PipelineEntity, templatesPath, applicationPath string) error {
	templatesPath = templatesPath + "/" + e.Config().TemplatesPath
	pipelinePath := applicationPath + "/" + e.Config().DestinationPath
	if exists, err := g.directoryService.DirectoryExists(pipelinePath); err != nil {
		return err
	} else if exists {
		return errors.New(fmt.Sprintf("pipeline already exists: %s", pipelinePath))
	}
	if err := g.directoryService.CopyDirectory(templatesPath, pipelinePath); err != nil {
		return err
	}
	environments := make(map[string]*entity.PipelineEnvironment)
	for k, v := range e.Config().Environments {
		for _, env := range e.Data().Envs() {
			if env.Env().Code() == k {
				environments[k] = v
			}
		}
	}
	data := PipelineData{
		Environments:     environments,
		DefaultVariables: e.Config().DefaultVariables,
	}
	return g.directoryService.ApplyTemplateRecursively(pipelinePath, data)
}
