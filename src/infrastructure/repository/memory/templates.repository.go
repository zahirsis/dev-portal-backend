package memory

import (
	"errors"
	"github.com/zahirsis/dev-portal-backend/src/domain/entity"
	"github.com/zahirsis/dev-portal-backend/src/domain/repository"
	"github.com/zahirsis/dev-portal-backend/src/pkg/logger"
)

const (
	ErrTemplateNotFound = "template not found"
)

type templateRepository struct {
	logger.Logger
}

func NewTemplateRepository(l logger.Logger) repository.TemplateRepository {
	return &templateRepository{l}
}

func (r *templateRepository) List() ([]entity.TemplateEntity, error) {
	return r.memory(), nil
}

func (r *templateRepository) Get(code string) (entity.TemplateEntity, error) {
	for _, v := range r.memory() {
		if v.Code() == code {
			return v, nil
		}
	}
	return nil, errors.New(ErrTemplateNotFound)
}

func (r *templateRepository) memory() []entity.TemplateEntity {
	m := []*entity.Manifest{
		{
			Code:  "aws-ecr",
			Label: "Aws ECR",
			Type:  entity.RegistryManifests,
			Dir:   "manifests/registry/aws-ecr",
		},
		{
			Code:  "argo-cd",
			Label: "Argo manifests",
			Type:  entity.GitOpsManifests,
			Dir:   "manifests/git-ops/argo-cd",
		},
	}
	msb := append(m, &entity.Manifest{
		Code:  "bitbucket-pipelines",
		Label: "Bitbucket pipelines",
		Type:  entity.PipelineManifests,
		Dir:   "manifests/pipeline/bitbucket-pipelines/spring-boot",
	})
	mrj := append(m, &entity.Manifest{
		Code:  "bitbucket-pipelines",
		Label: "Bitbucket pipelines",
		Type:  entity.PipelineManifests,
		Dir:   "manifests/pipeline/bitbucket-pipelines/react-js",
	})
	return []entity.TemplateEntity{
		entity.NewTemplateEntity("spring-boot", "SpringBoot", entity.ApplicationObject{
			RootPath: entity.PathObject{
				Default:      "/{applicationName}",
				Customizable: true,
			},
			HealthCheckPath: entity.PathObject{
				Default:      "/{applicationName}/actuator/health",
				Customizable: true,
			},
			HealthCheckInitialDelaySeconds: 120,
			HealthCheckSecondDelaySeconds:  180,
			HealthCheckPeriodSeconds:       30,
			Port:                           8080,
			Memory: entity.ResourceObject{
				Min: entity.NumberValueObject{
					Value: 256,
					Step:  128,
					Min:   128,
					Max:   2048,
				},
				Max: entity.NumberValueObject{
					Value: 512,
					Step:  128,
					Min:   128,
					Max:   4096,
				},
			},
			Cpu: entity.ResourceObject{
				Min: entity.NumberValueObject{
					Value: 0.05,
					Step:  0.01,
					Min:   0.01,
					Max:   2,
				},
				Max: entity.NumberValueObject{
					Value: 0.3,
					Step:  0.1,
					Min:   0.1,
					Max:   4,
				},
			},
		}, entity.IngressObject{
			Host: entity.PathObject{
				Fixed:        "gw.<environment>.tempoassist.cloud",
				Customizable: false,
			},
			Path: entity.PathObject{
				Fixed:        "/{squadName}/",
				Default:      "{applicationName}",
				Customizable: true,
			},
			Authentication: true,
			Frontend:       false,
			Enabled:        true,
		}, msb),
		entity.NewTemplateEntity("react-js", "ReactJs", entity.ApplicationObject{
			RootPath: entity.PathObject{
				Default:      "/",
				Customizable: true,
			},
			HealthCheckPath: entity.PathObject{
				Default:      "/health",
				Customizable: true,
			},
			Port: 3000,
			Memory: entity.ResourceObject{
				Min: entity.NumberValueObject{
					Value: 64,
					Step:  64,
					Min:   64,
					Max:   512,
				},
				Max: entity.NumberValueObject{
					Value: 128,
					Step:  64,
					Min:   64,
					Max:   1024,
				},
			},
			Cpu: entity.ResourceObject{
				Min: entity.NumberValueObject{
					Value: 0.01,
					Step:  0.01,
					Min:   0.01,
					Max:   0.5,
				},
				Max: entity.NumberValueObject{
					Value: 0.1,
					Step:  0.01,
					Min:   0.1,
					Max:   1,
				},
			},
		}, entity.IngressObject{
			Host: entity.PathObject{
				Fixed:        ".<environment>.tempoassist.cloud",
				Default:      "{applicationName}",
				Customizable: true,
			},
			Path: entity.PathObject{
				Fixed:        "/",
				Default:      "",
				Customizable: false,
			},
			Authentication: false,
			Frontend:       true,
			Enabled:        true,
		}, mrj),
		//entity.NewTemplateEntity("node-js", "Node.Js", entity.ApplicationObject{
		//	RootPath:        entity.PathObject{},
		//	HealthCheckPath: entity.PathObject{},
		//	Port:            8080,
		//	Memory:          entity.ResourceObject{},
		//	Cpu:             entity.ResourceObject{},
		//}, entity.IngressObject{
		//	Host:           entity.PathObject{},
		//	Path:           entity.PathObject{},
		//	Authentication: false,
		//}),
		//entity.NewTemplateEntity("python", "Python", entity.ApplicationObject{
		//	RootPath:        entity.PathObject{},
		//	HealthCheckPath: entity.PathObject{},
		//	Port:            8080,
		//	Memory:          entity.ResourceObject{},
		//	Cpu:             entity.ResourceObject{},
		//}, entity.IngressObject{
		//	Host:           entity.PathObject{},
		//	Path:           entity.PathObject{},
		//	Authentication: false,
		//})
	}
}
