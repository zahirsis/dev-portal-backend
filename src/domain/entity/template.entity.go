package entity

type ManifestType string

const (
	GitOpsManifests   ManifestType = "gitOps"
	PipelineManifests ManifestType = "pipeline"
	RegistryManifests ManifestType = "registry"
	WikiManifests     ManifestType = "wiki"
	SecretManifests   ManifestType = "secret"
)

type ApplicationObject struct {
	RootPath                       PathObject     `json:"rootPath"`
	HealthCheckPath                PathObject     `json:"healthCheckPath"`
	Port                           int            `json:"port"`
	Memory                         ResourceObject `json:"memory"`
	Cpu                            ResourceObject `json:"cpu"`
	HealthCheckInitialDelaySeconds int            `json:"healthCheckInitialDelaySeconds"` // readiness probe
	HealthCheckSecondDelaySeconds  int            `json:"healthCheckSecondDelaySeconds"`  // liveness probe
	HealthCheckPeriodSeconds       int            `json:"healthCheckPeriodSeconds"`       // both
}

type IngressObject struct {
	Enabled        bool       `json:"enabled"`
	Host           PathObject `json:"host"`
	Path           PathObject `json:"path"`
	Authentication bool       `json:"authentication"`
	Frontend       bool       `json:"frontend"`
}

type TemplateEntity interface {
	Code() string
	Label() string
	ApplicationDefault() ApplicationObject
	IngressDefault() IngressObject
	Manifests() []*Manifest
}

type templateEntity struct {
	code                string
	label               string
	applicationDefaults ApplicationObject
	ingressDefaults     IngressObject
	manifests           []*Manifest
}

func NewTemplateEntity(
	code string,
	label string,
	applicationDefault ApplicationObject,
	ingressDefault IngressObject,
	manifests []*Manifest,
) TemplateEntity {
	return &templateEntity{
		code,
		label,
		applicationDefault,
		ingressDefault,
		manifests,
	}
}

func (t *templateEntity) Code() string {
	return t.code
}

func (t *templateEntity) Label() string {
	return t.label
}

func (t *templateEntity) ApplicationDefault() ApplicationObject {
	return t.applicationDefaults
}

func (t *templateEntity) IngressDefault() IngressObject {
	return t.ingressDefaults
}

func (t *templateEntity) Manifests() []*Manifest {
	return t.manifests
}
