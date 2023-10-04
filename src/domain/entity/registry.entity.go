package entity

type RegistryConfig struct {
	Region                     string  `json:"region" yaml:"region"`
	RegistryId                 *string `json:"registryId" yaml:"registryId"`
	ImageScanningConfiguration struct {
		ScanOnPush bool `json:"scanOnPush" yaml:"scanOnPush"`
	} `json:"imageScanningConfiguration" yaml:"imageScanningConfiguration"`
}

type RegistryEntity interface {
	Name() *string
	Policy() *string
	Config() *RegistryConfig
	Tags() []*Tag
}

type registryEntity struct {
	name   *string
	policy *string
	config *RegistryConfig
	tags   []*Tag
}

func NewRegistryEntity(name string, policy string, config *RegistryConfig, tags []*Tag) RegistryEntity {
	return &registryEntity{
		name:   &name,
		policy: &policy,
		config: config,
		tags:   tags,
	}
}

func (t *registryEntity) Name() *string {
	return t.name
}

func (t *registryEntity) Policy() *string {
	return t.policy
}

func (t *registryEntity) Config() *RegistryConfig {
	return t.config
}

func (t *registryEntity) Tags() []*Tag {
	return t.tags
}
