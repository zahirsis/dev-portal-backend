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

type SetupGitOpsData struct {
	Entity                   entity.GitOpsEntity
	Env                      entity.SetupEnvData
	TemplatesPathBase        string
	TemplatesPathNamespace   string
	TemplatesPathApplication string
	GitOpsPathBae            string
	GitOpsPathNamespace      string
	GitOpsPathApplication    string
}

type GitOpsService interface {
	LoadData(data entity.SetupCiCdEntity, manifest *entity.Manifest, templatesPath string) (entity.GitOpsEntity, error)
	SetupBaseUtilities(e entity.GitOpsEntity, templatesPath, gitOpsPath string) error
	SetupNamespacedUtilities(e entity.GitOpsEntity, templatesPath, gitOpsPath string) error
	SetupK8sManifests(e entity.GitOpsEntity, templatesPath, gitOpsPath, cmPath string) ([]string, error)
	SetupGitOpsManifests(e entity.GitOpsEntity, templatesPath, gitOpsPath string, env entity.SetupEnvData) error
}

type gitOpsService struct {
	config           *config.Config
	logger           logger.Logger
	directoryService DirectoryService
}

func NewGitOpsService(config *config.Config, logger logger.Logger, directoryService DirectoryService) GitOpsService {
	return &gitOpsService{
		config:           config,
		logger:           logger,
		directoryService: directoryService,
	}
}

func (g *gitOpsService) LoadData(data entity.SetupCiCdEntity, manifest *entity.Manifest, templatesPath string) (entity.GitOpsEntity, error) {
	cfg, err := os.ReadFile(fmt.Sprintf("%s/%s/config.yaml", templatesPath, manifest.Dir))
	if err != nil {
		return nil, err
	}
	configData := &entity.GitOpsConfig{}
	err = yaml.Unmarshal(cfg, configData)
	if err != nil {
		g.logger.Error("Error unmarshalling config", err.Error(), string(cfg))
		return nil, err
	}
	return entity.NewGitOpsEntity(data, configData, entity.DefaultTags(data)), nil
}

func (g *gitOpsService) SetupBaseUtilities(e entity.GitOpsEntity, templatesPath, gitOpsPath string) error {
	templatesPath = templatesPath + "/" + e.Config().K8sBaseTemplatesPath
	gitOpsPath = gitOpsPath + "/" + e.Config().K8sBaseDestinationPath
	parentDir := g.getParentDir(gitOpsPath)
	if exists, err := g.directoryService.DirectoryExists(parentDir); err != nil {
		return err
	} else if !exists {
		if err := g.directoryService.CreateDirectory(parentDir); err != nil {
			return err
		}
	}
	if exists, err := g.directoryService.DirectoryExists(gitOpsPath); err != nil {
		return err
	} else if exists {
		return nil
	}
	return g.directoryService.CopyDirectory(templatesPath, gitOpsPath)
}

func (g *gitOpsService) SetupNamespacedUtilities(e entity.GitOpsEntity, templatesPath, gitOpsPath string) error {
	templatesPath = templatesPath + "/" + e.Config().K8sNamespaceUtilitiesTemplatesPath
	gitOpsPath = gitOpsPath + "/" + e.Config().K8sNamespaceUtilitiesDestinationPath
	parentDir := g.getParentDir(gitOpsPath)
	if exists, err := g.directoryService.DirectoryExists(parentDir); err != nil {
		return err
	} else if !exists {
		if err := g.directoryService.CreateDirectory(parentDir); err != nil {
			return err
		}
	}
	if exists, err := g.directoryService.DirectoryExists(gitOpsPath); err != nil {
		return err
	} else if exists {
		return nil
	}
	if err := g.directoryService.CopyDirectory(templatesPath, gitOpsPath); err != nil {
		return err
	}
	type ReplaceValue struct {
		Namespace string
	}
	replaceValues := ReplaceValue{Namespace: e.Data().Squad().Code()}
	return g.directoryService.ApplyTemplateRecursively(gitOpsPath, replaceValues)
}

type ApplicationData struct {
	Namespace                           string
	ApplicationName                     string
	ApplicationPort                     int
	ApplicationCpuLimit                 string
	ApplicationMemoryLimit              string
	ApplicationCpuRequest               string
	ApplicationMemoryRequest            string
	ApplicationHealthCheckPath          string
	ApplicationInitialDelaySeconds      int
	ApplicationSecondDelaySeconds       int
	ApplicationHealthCheckPeriodSeconds int
	IngressStripPath                    bool
	IngressAuthentication               bool
	IngressFrontend                     bool
	IngressCustomPath                   string
	IngressHost                         string
	IngressPath                         string
	DefaultImageName                    string
	DefaultImageTag                     string
	ApplicationMinReplicas              int
	ApplicationMaxReplicas              int
	EnvironmentMountPath                string
}

func (g *gitOpsService) SetupK8sManifests(e entity.GitOpsEntity, templatesPath, gitOpsPath, cmPath string) ([]string, error) {
	cmTemplatesPath := templatesPath + "/" + e.Config().K8sConfigMapTemplatesPath
	templatesPath = templatesPath + "/" + e.Config().K8sApplicationTemplatesPath
	gitOpsPath = gitOpsPath + "/" + e.Config().K8sApplicationDestinationPath
	if exists, err := g.directoryService.DirectoryExists(gitOpsPath); err != nil {
		return []string{}, err
	} else if exists {
		return []string{}, errors.New("k8s manifests for already exists for application")
	}
	if err := g.directoryService.CreateDirectory(gitOpsPath); err != nil {
		return []string{}, err
	}
	if err := g.directoryService.CopyDirectory(templatesPath+"/base", gitOpsPath+"/base"); err != nil {
		return []string{}, err
	}
	data := g.createApplicationData(e)
	g.logger.Debug("Applying template recursively", gitOpsPath+"/base", data)
	if err := g.directoryService.ApplyTemplateRecursively(gitOpsPath+"/base", data); err != nil {
		return []string{}, err
	}
	if err := g.directoryService.CreateDirectory(gitOpsPath + "/overlays"); err != nil {
		return []string{}, err
	}
	extraData := []string{"Application ingresses:"}
	for _, env := range e.Data().Envs() {
		if err := g.directoryService.CopyDirectory(templatesPath+"/overlays/overlay", gitOpsPath+"/overlays/"+env.Env().Code()); err != nil {
			return []string{}, err
		}
		data.IngressHost = e.Data().IngressHost(env.Env().Code())
		data.IngressPath = e.Data().IngressPath(env.Env().Code())
		data.ApplicationMinReplicas = env.ReplicasMin()
		data.ApplicationMaxReplicas = env.ReplicasMax()
		data.EnvironmentMountPath = env.Env().SecretsPath()
		if err := g.directoryService.ApplyTemplateRecursively(gitOpsPath+"/overlays/"+env.Env().Code(), data); err != nil {
			return []string{}, err
		}
		extraData = append(
			extraData,
			fmt.Sprintf(" -- %s: %s", env.Env().Label(), e.Data().IngressFull(env.Env().Code())),
		)
		e.Data().CreatedData().Environments = append(e.Data().CreatedData().Environments, &entity.EnvironmentCreatedData{
			Label:           env.Env().Label(),
			Code:            env.Env().Code(),
			Url:             e.Data().IngressFull(env.Env().Code()),
			ApplicationName: e.Data().ApplicationName(),
		})
		if !g.config.SetupCiCd.ExternalConfigMap {
			continue
		}
		if err := g.directoryService.CreateDirectory(fmt.Sprintf("%s/%s", cmPath, e.Config().K8sConfigMapDestinationPath)); err != nil {
			return []string{}, err
		}
		if err := g.directoryService.CopyDirectory(cmTemplatesPath+"/overlay", fmt.Sprintf("%s/%s/%s", cmPath, e.Config().K8sConfigMapDestinationPath, env.Env().Code())); err != nil {
			return []string{}, err
		}
		if err := g.directoryService.ApplyTemplateRecursively(fmt.Sprintf("%s/%s/%s", cmPath, e.Config().K8sConfigMapDestinationPath, env.Env().Code()), data); err != nil {
			return []string{}, err
		}
	}
	return extraData, nil
}

type GitOpsManifestsData struct {
	templateKustomizationPath             string
	templateAppPath                       string
	templateNamespaceUtilitiesPath        string
	baseKustomizationDestinationPath      string
	namespaceKustomizationDestinationPath string
	namespaceUtilitiesDestinationPath     string
	namespaceUtilitiesFileName            string
	appDestinationPath                    string
	Namespace                             string
	ApplicationName                       string
	Environment                           string
	DestinationCluster                    string
	Project                               string
	K8sApplicationPath                    string
	K8sNamespaceUtilitiesPath             string
	GitOpsToolsRepository                 string
	GitOpsRepository                      string
	ConfigMapPath                         string
	ConfigMapRepository                   string
}

func (g *gitOpsService) SetupGitOpsManifests(e entity.GitOpsEntity, templatesPath, gitOpsPath string, env entity.SetupEnvData) error {
	gitOpsBaseDestinationPath := gitOpsPath + "/" + e.Config().GitOpsAppsDestination(env.Env().Code())
	gitOpsNamespaceDestinationPath := gitOpsBaseDestinationPath + "/" + e.Data().Squad().Code()
	if exists, err := g.directoryService.DirectoryExists(gitOpsNamespaceDestinationPath); err != nil {
		return err
	} else if !exists {
		if err := g.directoryService.CreateDirectory(gitOpsNamespaceDestinationPath); err != nil {
			return err
		}
	}
	data := &GitOpsManifestsData{
		templateKustomizationPath:             templatesPath + "/" + e.Config().GitOpsKustomizationTemplatePath,
		templateAppPath:                       templatesPath + "/" + e.Config().GitOpsAppTemplatesPath,
		templateNamespaceUtilitiesPath:        templatesPath + "/" + e.Config().GitOpsAppNamespaceUtilitiesTemplatesPath,
		baseKustomizationDestinationPath:      gitOpsBaseDestinationPath + "/kustomization.yaml",
		namespaceKustomizationDestinationPath: gitOpsNamespaceDestinationPath + "/kustomization.yaml",
		namespaceUtilitiesDestinationPath:     gitOpsNamespaceDestinationPath + "/_base.yaml",
		appDestinationPath:                    gitOpsNamespaceDestinationPath + "/" + e.Data().ApplicationSlug() + ".yaml",
		Namespace:                             e.Data().Squad().Code(),
		ApplicationName:                       e.Data().ApplicationSlug(),
		Environment:                           env.Env().Code(),
		DestinationCluster:                    env.Env().DestinationCluster(),
		Project:                               env.Env().Project(),
		K8sApplicationPath:                    e.Config().K8sApplicationDestinationPath + "/overlays/" + env.Env().Code(),
		K8sNamespaceUtilitiesPath:             e.Config().K8sNamespaceUtilitiesDestinationPath + "/overlays/" + env.Env().Code(),
		GitOpsRepository:                      g.config.SetupCiCd.GitOpsRepository,
		GitOpsToolsRepository:                 g.config.SetupCiCd.GitOpsToolsRepository,
		ConfigMapPath:                         e.Config().K8sConfigMapDestinationPath + "/" + env.Env().Code(),
		ConfigMapRepository:                   g.config.SetupCiCd.ConfigMapRepository,
	}
	if err := g.setupGitOpsBaseManifests(data); err != nil {
		return err
	}
	if err := g.setupGitOpsNamespaceManifests(data); err != nil {
		return err
	}
	return g.setupGitOpsApplicationManifests(data)
}

func (g *gitOpsService) setupGitOpsBaseManifests(data *GitOpsManifestsData) error {
	if exists, err := g.directoryService.DirectoryExists(data.baseKustomizationDestinationPath); err != nil {
		return err
	} else if !exists {
		if err := g.directoryService.CopyDirectory(data.templateKustomizationPath, data.baseKustomizationDestinationPath); err != nil {
			return err
		}
	}
	return g.directoryService.VerifyOrInsertLineInFile(
		data.baseKustomizationDestinationPath,
		fmt.Sprintf("- %s/", data.Namespace),
	)
}

func (g *gitOpsService) setupGitOpsNamespaceManifests(data *GitOpsManifestsData) error {
	if exists, err := g.directoryService.DirectoryExists(data.namespaceKustomizationDestinationPath); err != nil {
		return err
	} else if !exists {
		if err := g.directoryService.CopyDirectory(data.templateKustomizationPath, data.namespaceKustomizationDestinationPath); err != nil {
			return err
		}
	}
	if err := g.directoryService.VerifyOrInsertLineInFile(data.namespaceKustomizationDestinationPath, "- _base.yaml"); err != nil {
		return err
	}
	err := g.directoryService.VerifyOrInsertLineInFile(
		data.namespaceKustomizationDestinationPath,
		fmt.Sprintf("- %s.yaml", data.ApplicationName),
	)
	if err != nil {
		return err
	}
	if exists, err := g.directoryService.DirectoryExists(data.namespaceUtilitiesDestinationPath); err != nil {
		return err
	} else if !exists {
		if err := g.directoryService.CopyDirectory(data.templateNamespaceUtilitiesPath, data.namespaceUtilitiesDestinationPath); err != nil {
			return err
		}
	}
	return g.directoryService.ApplyTemplate(data.namespaceUtilitiesDestinationPath, data)
}

func (g *gitOpsService) setupGitOpsApplicationManifests(data *GitOpsManifestsData) error {
	if exists, err := g.directoryService.DirectoryExists(data.appDestinationPath); err != nil {
		return err
	} else if !exists {
		if err := g.directoryService.CopyDirectory(data.templateAppPath, data.appDestinationPath); err != nil {
			return err
		}
	}
	return g.directoryService.ApplyTemplate(data.appDestinationPath, data)
}

func (g *gitOpsService) getParentDir(path string) string {
	return path[0 : len(path)-len(g.getDirName(path))-1]
}

func (g *gitOpsService) getDirName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return ""
}

func (g *gitOpsService) createApplicationData(e entity.GitOpsEntity) *ApplicationData {
	return &ApplicationData{
		Namespace:                           e.Data().Squad().Code(),
		ApplicationName:                     e.Data().ApplicationName(),
		ApplicationPort:                     e.Data().ApplicationPort(),
		ApplicationCpuLimit:                 formatCpu(e.Data().ApplicationMaxCpu()),
		ApplicationMemoryLimit:              formatMemory(e.Data().ApplicationMemoryMax()),
		ApplicationCpuRequest:               formatCpu(e.Data().ApplicationMinCpu()),
		ApplicationMemoryRequest:            formatMemory(e.Data().ApplicationMemoryMin()),
		ApplicationHealthCheckPath:          e.Data().ApplicationHealthCheckPath(),
		ApplicationInitialDelaySeconds:      e.Data().Template().ApplicationDefault().HealthCheckInitialDelaySeconds,
		ApplicationSecondDelaySeconds:       e.Data().Template().ApplicationDefault().HealthCheckSecondDelaySeconds,
		ApplicationHealthCheckPeriodSeconds: e.Data().Template().ApplicationDefault().HealthCheckPeriodSeconds,
		IngressStripPath:                    e.Data().IngressStripPath(),
		IngressAuthentication:               e.Data().IngressAuthentication(),
		IngressCustomPath:                   e.Data().IngressCustomPath(),
		IngressFrontend:                     e.Data().Template().IngressDefault().Frontend,
		DefaultImageName:                    g.config.SetupCiCd.DefaultImageName,
		DefaultImageTag:                     g.config.SetupCiCd.DefaultImageTag,
	}
}
