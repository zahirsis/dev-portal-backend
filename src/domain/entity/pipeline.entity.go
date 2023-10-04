package entity

import "strings"

type PipelineTriggers struct {
	Name       string `json:"name" yaml:"name"`
	Branch     string `json:"branch" yaml:"branch"`
	Deployment string `json:"deployment" yaml:"deployment"`
}

type PipelineVariable struct {
	Name   string `json:"name" yaml:"name"`
	Value  string `json:"value" yaml:"value"`
	Secure bool   `json:"secure" yaml:"secure"`
}

type PipelineEnvironment struct {
	Triggers  []*PipelineTriggers `json:"triggers" yaml:"triggers"`
	Variables []*PipelineVariable `json:"variables" yaml:"variables"`
}

type PipelineConfig struct {
	TemplatesPath    string                          `json:"templatesPath" yaml:"templatesPath"`
	DestinationPath  string                          `json:"destinationPath" yaml:"destinationPath"`
	InitialPipeline  string                          `json:"initialPipeline" yaml:"initialPipeline"`
	Environments     map[string]*PipelineEnvironment `json:"environments" yaml:"environments"`
	DefaultVariables []*PipelineVariable             `json:"defaultVariables" yaml:"defaultVariables"`
}

type PipelineEntity interface {
	Data() SetupCiCdEntity
	Config() *PipelineConfig
	Tags() []*Tag
}

type pipelineEntity struct {
	data   SetupCiCdEntity
	config *PipelineConfig
	tags   []*Tag
}

func NewPipelineEntity(s SetupCiCdEntity, config *PipelineConfig, tags []*Tag) PipelineEntity {
	config.replace("<namespace>", s.Squad().Code())
	config.replace("<applicationName>", s.ApplicationSlug())
	return &pipelineEntity{
		data:   s,
		config: config,
		tags:   tags,
	}
}

func (g *pipelineEntity) Data() SetupCiCdEntity {
	return g.data
}

func (g *pipelineEntity) Config() *PipelineConfig {
	return g.config
}

func (g *pipelineEntity) Tags() []*Tag {
	return g.tags
}

func (c *PipelineConfig) replace(old, new string) {
	c.DestinationPath = strings.ReplaceAll(c.DestinationPath, old, new)
	for i, e := range c.DefaultVariables {
		c.DefaultVariables[i].Value = strings.ReplaceAll(e.Value, old, new)
	}
	for i, e := range c.Environments {
		for j, v := range e.Variables {
			c.Environments[i].Variables[j].Value = strings.ReplaceAll(v.Value, old, new)
		}
	}
}
