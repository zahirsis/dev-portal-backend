package entity

import "strings"

type SecretConfig struct {
	SecretPath string `json:"secretPath" yaml:"secretPath"`
	RootPath   string `json:"rootPath" yaml:"rootPath"`
}

type SecretEntity interface {
	Data() SetupCiCdEntity
	Config() *SecretConfig
	Tags() []*Tag
}

type secretEntity struct {
	data   SetupCiCdEntity
	config *SecretConfig
	tags   []*Tag
}

func NewSecretEntity(s SetupCiCdEntity, config *SecretConfig, tags []*Tag) SecretEntity {
	config.replace("<namespace>", s.Squad().Code())
	config.replace("<applicationName>", s.ApplicationSlug())
	return &secretEntity{
		config: config,
		tags:   tags,
	}
}

func (t *secretEntity) Data() SetupCiCdEntity {
	return t.data
}

func (t *secretEntity) Config() *SecretConfig {
	return t.config
}

func (t *secretEntity) Tags() []*Tag {
	return t.tags
}

func (t *SecretConfig) GetRootPath(env SetupEnvData) string {
	return strings.Replace(t.RootPath, "<environmentMountPath>", env.Env().SecretsPath(), -1)
}

func (t *SecretConfig) GetSecretPath(env SetupEnvData) string {
	return strings.Replace(t.SecretPath, "<environmentMountPath>", env.Env().SecretsPath(), -1)
}

func (t *SecretConfig) replace(old, new string) {
	t.RootPath = strings.Replace(t.RootPath, old, new, -1)
	t.SecretPath = strings.Replace(t.SecretPath, old, new, -1)
}
