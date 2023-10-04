package service

import (
	"fmt"
	"github.com/zahirsis/dev-portal-backend/pkg/errors"
	"github.com/zahirsis/dev-portal-backend/src/domain/entity"
	"github.com/zahirsis/dev-portal-backend/src/domain/repository"
	"github.com/zahirsis/dev-portal-backend/src/pkg/logger"
)

type CiCdService interface {
	ValidateSetup(setup entity.SetupCiCdEntity) []error
}

type ciCdService struct {
	logger       logger.Logger
	repositories *repository.Container
}

func NewCiCdService(logger logger.Logger, repositories *repository.Container) CiCdService {
	return &ciCdService{
		logger:       logger,
		repositories: repositories,
	}
}

func (c *ciCdService) ValidateSetup(setup entity.SetupCiCdEntity) []error {
	var errs []error
	if len(setup.Envs()) == 0 {
		errs = append(errs, errors.NewInputError("envs", []string{"envs cannot be empty"}))
	}
	for _, env := range setup.Envs() {
		err := c.checkEnvConcurrency(env.Env().Code(), setup.Envs())
		if err != nil {
			errs = append(errs, err)
		}
		errs = append(errs, c.checkEnvReplicas(env)...)
	}
	for _, manifest := range setup.Manifests() {
		err := c.checkManifest(manifest, setup.Template().Manifests())
		if err != nil {
			errs = append(errs, err)
		}
	}
	errs = append(errs, c.checkResources(setup)...)
	errs = append(errs, c.CheckApplication(setup)...)
	errs = append(errs, c.checkIngress(setup)...)
	return errs
}

func (c *ciCdService) checkEnvConcurrency(env string, envs []entity.SetupEnvData) error {
	for _, e := range envs {
		for _, c := range e.Env().Concurrences() {
			if c == env {
				return errors.NewInputError("env."+env, []string{"this env cannot be used in concurrency with " + c + " env"})
			}
		}
	}
	return nil
}

func (c *ciCdService) checkEnvReplicas(env entity.SetupEnvData) []error {
	var errs []error
	if env.ReplicasMin() > env.ReplicasMax() {
		errs = append(errs, errors.NewInputError(
			"env."+env.Env().Code()+".replicas.min",
			[]string{"min cannot be greater than max"}),
		)
	}
	minLimit := int(env.Env().DefaultReplicas().Min.Min)
	maxLimit := int(env.Env().DefaultReplicas().Max.Max)
	if env.ReplicasMin() < minLimit {
		errs = append(errs, errors.NewInputError(
			"env."+env.Env().Code()+".replicas.min",
			[]string{fmt.Sprintf("min cannot be less than %d", minLimit)}),
		)
	}
	if env.ReplicasMax() > maxLimit {
		errs = append(errs, errors.NewInputError(
			"env."+env.Env().Code()+".replicas.max",
			[]string{fmt.Sprintf("max cannot be greater than %d", maxLimit)}),
		)
	}
	return errs
}

func (c *ciCdService) checkManifest(manifest *entity.Manifest, manifests []*entity.Manifest) error {
	for _, m := range manifests {
		if m.Code == manifest.Code {
			return nil
		}
	}
	return errors.NewInputError("manifests."+manifest.Code, []string{"template does not have this manifest"})
}

func (c *ciCdService) checkResources(setup entity.SetupCiCdEntity) []error {
	var errs []error
	// cpu
	if setup.ApplicationMinCpu() > setup.ApplicationMaxCpu() {
		errs = append(errs, errors.NewInputError(
			"application.resources.cpu.min",
			[]string{"min cannot be greater than max"}),
		)
	}
	if setup.ApplicationMinCpu() < setup.Template().ApplicationDefault().Cpu.Min.Min {
		errs = append(errs, errors.NewInputError(
			"application.resources.cpu.min",
			[]string{fmt.Sprintf("min cannot be less than %s", formatCpu(setup.Template().ApplicationDefault().Cpu.Min.Min))}),
		)
	}
	if setup.ApplicationMaxCpu() > setup.Template().ApplicationDefault().Cpu.Max.Max {
		errs = append(errs, errors.NewInputError(
			"application.resources.cpu.max",
			[]string{fmt.Sprintf("max cannot be greater than %s", formatCpu(setup.Template().ApplicationDefault().Cpu.Max.Max))}),
		)
	}
	// memory
	if setup.ApplicationMemoryMin() > setup.ApplicationMemoryMax() {
		errs = append(errs, errors.NewInputError(
			"application.resources.memory.min",
			[]string{"min cannot be greater than max"}),
		)
	}
	if setup.ApplicationMemoryMin() < setup.Template().ApplicationDefault().Memory.Min.Min {
		errs = append(errs, errors.NewInputError(
			"application.resources.memory.min",
			[]string{fmt.Sprintf("min cannot be less than %s", formatMemory(setup.Template().ApplicationDefault().Memory.Min.Min))}),
		)
	}
	if setup.ApplicationMemoryMax() > setup.Template().ApplicationDefault().Memory.Max.Max {
		errs = append(errs, errors.NewInputError(
			"application.resources.memory.max",
			[]string{fmt.Sprintf("max cannot be greater than %s", formatMemory(setup.Template().ApplicationDefault().Memory.Max.Max))}),
		)
	}
	return errs
}

func (c *ciCdService) CheckApplication(setup entity.SetupCiCdEntity) []error {
	var errs []error
	// name
	if setup.ApplicationName() == "" {
		errs = append(errs, errors.NewInputError(
			"application.name",
			[]string{"name cannot be empty"}),
		)
	}
	// root path
	if setup.ApplicationRootPath() == "" {
		errs = append(errs, errors.NewInputError(
			"application.rootPath",
			[]string{"root path cannot be empty"}),
		)
	}
	// health check path
	if setup.ApplicationHealthCheckPath() == "" {
		errs = append(errs, errors.NewInputError(
			"application.healthCheckPath",
			[]string{"health checkPath cannot be empty"}),
		)
	}
	// port
	if setup.ApplicationPort() < 0 || setup.ApplicationPort() > 65535 {
		errs = append(errs, errors.NewInputError(
			"application.port",
			[]string{"port must be between 0 and 65535"}),
		)
	}
	return errs
}

func (c *ciCdService) checkIngress(setup entity.SetupCiCdEntity) []error {
	var errs []error
	// custom host
	if setup.IngressCustomHost() == "" && setup.Template().IngressDefault().Host.Customizable {
		errs = append(errs, errors.NewInputError(
			"ingress.customHost",
			[]string{"ingress host cannot be empty"}),
		)
	}
	if setup.Template().IngressDefault().Host.Customizable {
		return errs
	}
	// custom path
	if setup.IngressCustomPath() == "" && setup.Template().IngressDefault().Path.Customizable {
		errs = append(errs, errors.NewInputError(
			"ingress.customPath",
			[]string{"ingress path cannot be empty"}),
		)
	}
	return errs
}

func formatCpu(cpu float32) string {
	if cpu < 1 {
		return fmt.Sprintf("%dm", int(cpu*1000))
	}
	return fmt.Sprintf("%.2f", cpu)
}

func formatMemory(memory float32) string {
	if memory < 1024 {
		return fmt.Sprintf("%dMi", int(memory))
	}
	return fmt.Sprintf("%.2fGi", memory/1024)
}
