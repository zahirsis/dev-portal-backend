package entity

import "strings"

type GitOpsConfig struct {
	// k8s Base Templates
	K8sBaseTemplatesPath   string `json:"k8sBaseTemplatesPath" yaml:"k8sBaseTemplatesPath"`
	K8sBaseDestinationPath string `json:"k8sBaseDestinationPath" yaml:"k8sBaseDestinationPath"`
	// k8s Namespace Utilities Templates
	K8sNamespaceUtilitiesTemplatesPath   string `json:"k8sNamespaceUtilitiesTemplatesPath" yaml:"k8sNamespaceUtilitiesTemplatesPath"`
	K8sNamespaceUtilitiesDestinationPath string `json:"k8sNamespaceUtilitiesDestinationPath" yaml:"k8sNamespaceUtilitiesDestinationPath"`
	// k8s Application Templates
	K8sApplicationTemplatesPath   string `json:"k8sApplicationTemplatesPath" yaml:"k8sApplicationTemplatesPath"`
	K8sApplicationDestinationPath string `json:"k8sApplicationDestinationPath" yaml:"k8sApplicationDestinationPath"`
	// k8s ConfigMap Templates
	K8sConfigMapTemplatesPath   string `json:"k8sConfigMapTemplatesPath" yaml:"k8sConfigMapTemplatesPath"`
	K8sConfigMapDestinationPath string `json:"k8sConfigMapDestinationPath" yaml:"k8sConfigMapDestinationPath"`
	// Apps Base Templates
	GitOpsKustomizationTemplatePath          string `json:"gitOpsKustomizationTemplatePath" yaml:"gitOpsKustomizationTemplatePath"`
	GitOpsAppTemplatesPath                   string `json:"gitOpsAppTemplatesPath" yaml:"gitOpsAppTemplatesPath"`
	GitOpsAppNamespaceUtilitiesTemplatesPath string `json:"gitOpsAppNamespaceUtilitiesTemplatesPath" yaml:"gitOpsAppNamespaceUtilitiesTemplatesPath"`
	GitOpsBaseDestinationPath                string `json:"gitOpsBaseDestinationPath" yaml:"gitOpsBaseDestinationPath"`
}

type GitOpsEntity interface {
	Data() SetupCiCdEntity
	Config() *GitOpsConfig
	Tags() []*Tag
}

type gitOpsEntity struct {
	data   SetupCiCdEntity
	config *GitOpsConfig
	tags   []*Tag
}

func NewGitOpsEntity(s SetupCiCdEntity, config *GitOpsConfig, tags []*Tag) GitOpsEntity {
	config.replace("<namespace>", s.Squad().Code())
	config.replace("<applicationName>", s.ApplicationSlug())
	return &gitOpsEntity{
		data:   s,
		config: config,
		tags:   tags,
	}
}

func (g *gitOpsEntity) Data() SetupCiCdEntity {
	return g.data
}

func (g *gitOpsEntity) Config() *GitOpsConfig {
	return g.config
}

func (g *gitOpsEntity) Tags() []*Tag {
	return g.tags
}

func (c *GitOpsConfig) GitOpsAppsDestination(env string) string {
	return strings.Replace(c.GitOpsBaseDestinationPath, "<environment>", env, -1)
}

func (c *GitOpsConfig) replace(old, new string) {
	c.K8sApplicationDestinationPath = strings.Replace(c.K8sApplicationDestinationPath, old, new, -1)
	c.K8sBaseDestinationPath = strings.Replace(c.K8sBaseDestinationPath, old, new, -1)
	c.K8sNamespaceUtilitiesDestinationPath = strings.Replace(c.K8sNamespaceUtilitiesDestinationPath, old, new, -1)
	c.K8sConfigMapDestinationPath = strings.Replace(c.K8sConfigMapDestinationPath, old, new, -1)
	c.GitOpsBaseDestinationPath = strings.Replace(c.GitOpsBaseDestinationPath, old, new, -1)
}
