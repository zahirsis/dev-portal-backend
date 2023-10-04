package entity

type WikiConfig struct {
	TemplatePagePath    string `json:"templatePagePath" yaml:"templatePagePath"`
	TemplateServicePath string `json:"templateServicePath" yaml:"templateServicePath"`
	SpaceId             string `json:"spaceId" yaml:"spaceId"`
	ServicesPageId      string `json:"servicesPageId" yaml:"servicesPageId"`
	ServicesPageTitle   string `json:"servicesPageTitle" yaml:"servicesPageTitle"`
}

type WikiEntity interface {
	Data() SetupCiCdEntity
	Config() *WikiConfig
	Tags() []*Tag
}

type wikiEntity struct {
	data   SetupCiCdEntity
	config *WikiConfig
	tags   []*Tag
}

func NewWikiEntity(s SetupCiCdEntity, config *WikiConfig, tags []*Tag) WikiEntity {
	return &wikiEntity{
		data:   s,
		config: config,
		tags:   tags,
	}
}

func (g *wikiEntity) Data() SetupCiCdEntity {
	return g.data
}

func (g *wikiEntity) Config() *WikiConfig {
	return g.config
}

func (g *wikiEntity) Tags() []*Tag {
	return g.tags
}
