package service

import (
	"fmt"
	"github.com/zahirsis/dev-portal-backend/config"
	"github.com/zahirsis/dev-portal-backend/src/domain/entity"
	"github.com/zahirsis/dev-portal-backend/src/pkg/logger"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

type WikiService interface {
	LoadData(data entity.SetupCiCdEntity, v *entity.Manifest, dir string) (entity.WikiEntity, error)
	SetupWiki(wiki entity.WikiEntity, templatesPath string) ([]string, error)
}

type wikiService struct {
	config *config.Config
	logger logger.Logger
	api    WikiApiService
	ds     DirectoryService
}

func NewWikiService(config *config.Config, logger logger.Logger, api WikiApiService, ds DirectoryService) WikiService {
	return &wikiService{config, logger, api, ds}
}

func (g *wikiService) LoadData(data entity.SetupCiCdEntity, manifest *entity.Manifest, templatesPath string) (entity.WikiEntity, error) {
	cfg, err := os.ReadFile(fmt.Sprintf("%s/%s/config.yaml", templatesPath, manifest.Dir))
	if err != nil {
		return nil, err
	}
	configData := &entity.WikiConfig{}
	err = yaml.Unmarshal(cfg, configData)
	if err != nil {
		g.logger.Error("Error unmarshalling config", err.Error(), string(cfg))
		return nil, err
	}
	return entity.NewWikiEntity(data, configData, entity.DefaultTags(data)), nil
}

type WikiServiceEnvironment struct {
	Label string
	Code  string
}

type WikiServiceEndpoint struct {
	Name string
	Url  string
}

type WikiServiceData struct {
	ApplicationName string
	GitRepository   string
	Squad           string
	Environments    []*entity.EnvironmentCreatedData
	Exposed         bool
	HealthCheck     string
	Port            int
	RegistryUrl     string
	GitOpsUrl       string
	ConfigMapUrl    string
	Template        string
}

type WikiPagesData struct {
	Pages []*PageList
}

func (g *wikiService) SetupWiki(wiki entity.WikiEntity, templatesPath string) ([]string, error) {
	data := wiki.Data().CreatedData()
	wsd := &WikiServiceData{
		ApplicationName: wiki.Data().ApplicationName(),
		GitRepository:   g.config.GitConfig.GetRepositoryUrl(wiki.Data().ApplicationName()),
		Squad:           wiki.Data().Squad().Label(),
		Environments:    data.Environments,
		Exposed:         wiki.Data().Template().IngressDefault().Enabled,
		HealthCheck:     wiki.Data().ApplicationHealthCheckPath(),
		Port:            wiki.Data().ApplicationPort(),
		GitOpsUrl:       data.GitOpsPath,
		ConfigMapUrl:    data.ConfigMapPath,
		Template:        wiki.Data().Template().Label(),
		RegistryUrl:     data.RegistryUrl,
	}
	c, err := g.ds.LoadTemplate(fmt.Sprintf("%s/%s", templatesPath, wiki.Config().TemplateServicePath), wsd, true)
	if err != nil {
		return []string{}, err
	}
	title := fmt.Sprintf("[%s] %s", strings.ToUpper(wiki.Data().Squad().Label()), strings.ToTitle(wiki.Data().ApplicationSlug()))
	url, err := g.api.CreatePage(title, wiki.Config().SpaceId, wiki.Config().ServicesPageId, c)
	if err != nil {
		return []string{}, err
	}
	list, err := g.api.ListSubPages(wiki.Config().SpaceId, wiki.Config().ServicesPageId)
	if err != nil {
		return []string{}, err
	}
	wpd := &WikiPagesData{Pages: list}
	c, err = g.ds.LoadTemplate(fmt.Sprintf("%s/%s", templatesPath, wiki.Config().TemplatePagePath), wpd, true)
	if err != nil {
		return []string{}, err
	}
	message := fmt.Sprintf("Add [%s] %s", strings.ToUpper(wiki.Data().Squad().Label()), strings.ToTitle(wiki.Data().ApplicationSlug()))
	err = g.api.UpdatePage(wiki.Config().ServicesPageId, c, message)
	if err != nil {
		return []string{}, err
	}
	return []string{url}, nil
}
