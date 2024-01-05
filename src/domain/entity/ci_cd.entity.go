package entity

import (
	"regexp"
	"strings"
)

type EnvironmentCreatedData struct {
	Label           string
	Code            string
	Url             string
	ApplicationName string
}

type CreatedData struct {
	RegistryUrl   string
	Environments  []*EnvironmentCreatedData
	GitOpsPath    string
	ConfigMapPath string
}

type SetupEnvData interface {
	Env() EnvironmentEntity
	ReplicasMin() int
	ReplicasMax() int
}

type setupEnvData struct {
	env      EnvironmentEntity
	replicas LimitsIntData
}

func (s *setupEnvData) Env() EnvironmentEntity {
	return s.env
}

func (s *setupEnvData) ReplicasMin() int {
	return s.replicas.Min
}

func (s *setupEnvData) ReplicasMax() int {
	return s.replicas.Max
}

func NewSetupEnvData(env EnvironmentEntity, replicasMin int, replicasMax int) SetupEnvData {
	return &setupEnvData{
		env: env,
		replicas: LimitsIntData{
			Min: replicasMin,
			Max: replicasMax,
		},
	}
}

type SetupCiCdEntity interface {
	ProgressId() string
	Template() TemplateEntity
	Envs() []SetupEnvData
	Manifests() []*Manifest
	Squad() SquadEntity
	ApplicationName() string
	ApplicationSlug() string
	ApplicationRootPath() string
	ApplicationHealthCheckPath() string
	ApplicationMinCpu() float32
	ApplicationMaxCpu() float32
	ApplicationMemoryMin() float32
	ApplicationMemoryMax() float32
	ApplicationPort() int
	IngressCustomHost() string
	IngressCustomPath() string
	IngressAuthentication() bool
	IngressStripPath() bool
	IngressHost(env string) string
	IngressPath(env string) string
	IngressFull(env string) string
	CreatedData() *CreatedData
}

type SetupCiCdData struct {
	ID          string
	Template    TemplateEntity
	Envs        []SetupEnvData
	Manifests   []*Manifest
	Squad       SquadEntity
	Application ApplicationData
	Ingress     IngressData
}

type setupCiCdEntity struct {
	ID              string
	template        TemplateEntity
	envs            []SetupEnvData
	manifests       []*Manifest
	applicationSlug string
	squad           SquadEntity
	application     ApplicationData
	ingress         IngressData
	createdData     *CreatedData
}

func (s *setupCiCdEntity) ProgressId() string {
	return s.ID
}

func (s *setupCiCdEntity) Template() TemplateEntity {
	return s.template
}

func (s *setupCiCdEntity) Envs() []SetupEnvData {
	return s.envs
}

func (s *setupCiCdEntity) Manifests() []*Manifest {
	return s.manifests
}

func (s *setupCiCdEntity) Squad() SquadEntity {
	return s.squad
}

func (s *setupCiCdEntity) ApplicationSlug() string {
	return s.applicationSlug
}

func (s *setupCiCdEntity) ApplicationName() string {
	return s.application.Name
}

func (s *setupCiCdEntity) ApplicationRootPath() string {
	return "/" + strings.Trim(s.application.RootPath, "/")
}

func (s *setupCiCdEntity) ApplicationHealthCheckPath() string {
	return "/" + strings.Trim(s.application.HealthCheckPath, "/")
}

func (s *setupCiCdEntity) ApplicationMinCpu() float32 {
	return s.application.Resources.Cpu.Min
}

func (s *setupCiCdEntity) ApplicationMaxCpu() float32 {
	return s.application.Resources.Cpu.Max
}

func (s *setupCiCdEntity) ApplicationMemoryMin() float32 {
	return s.application.Resources.Memory.Min
}

func (s *setupCiCdEntity) ApplicationMemoryMax() float32 {
	return s.application.Resources.Memory.Max
}

func (s *setupCiCdEntity) ApplicationPort() int {
	return s.application.Port
}

func (s *setupCiCdEntity) IngressCustomHost() string {
	return strings.Trim(s.ingress.CustomHost, "/")
}

func (s *setupCiCdEntity) IngressCustomPath() string {
	return strings.Trim(s.ingress.CustomPath, "/")
}

func (s *setupCiCdEntity) IngressAuthentication() bool {
	return s.ingress.Authentication
}

func (s *setupCiCdEntity) IngressStripPath() bool {
	if s.template.IngressDefault().Frontend {
		return false
	}
	parts := strings.Split(strings.Trim(s.application.RootPath, "/"), "/")
	if len(parts) > 0 && parts[0] == strings.Trim(s.ingress.CustomPath, "/") {
		return false
	}
	return true
}

func (s *setupCiCdEntity) IngressHost(env string) string {
	var host string
	if s.IngressCustomHost() != "" {
		host = s.IngressCustomHost()
	}
	host += strings.Trim(s.Template().IngressDefault().Host.Fixed, "/")
	return s.replaceCommonValues(host, env)
}

func (s *setupCiCdEntity) IngressPath(env string) string {
	path := "/" + strings.Trim(s.Template().IngressDefault().Path.Fixed, "/")
	if s.IngressCustomPath() != "" {
		path += "/" + s.IngressCustomPath()
	}
	return s.replaceCommonValues(path, env)
}

func (s *setupCiCdEntity) IngressFull(env string) string {
	return s.IngressHost(env) + s.IngressPath(env)
}

func (s *setupCiCdEntity) CreatedData() *CreatedData {
	return s.createdData
}

func NewSetupCiCdEntity(options SetupCiCdData) SetupCiCdEntity {
	slug := strings.ReplaceAll(options.Application.Name, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	slug = strings.ToLower(slug)
	reg := regexp.MustCompile("[^a-zA-Z0-9-]")
	slug = reg.ReplaceAllString(slug, "")
	return &setupCiCdEntity{
		ID:              options.ID,
		template:        options.Template,
		envs:            options.Envs,
		manifests:       options.Manifests,
		squad:           options.Squad,
		applicationSlug: slug,
		application: ApplicationData{
			Name:            options.Application.Name,
			RootPath:        options.Application.RootPath,
			HealthCheckPath: options.Application.HealthCheckPath,
			Resources: ResourcesDataObject{
				Cpu: LimitsFloatData{
					Min: options.Application.Resources.Cpu.Min,
					Max: options.Application.Resources.Cpu.Max,
				},
				Memory: LimitsFloatData{
					Min: options.Application.Resources.Memory.Min,
					Max: options.Application.Resources.Memory.Max,
				},
			},
			Port: options.Application.Port,
		},
		ingress: IngressData{
			CustomHost:     options.Ingress.CustomHost,
			CustomPath:     options.Ingress.CustomPath,
			Authentication: options.Ingress.Authentication,
		},
		createdData: &CreatedData{
			RegistryUrl:   "",
			Environments:  []*EnvironmentCreatedData{},
			GitOpsPath:    "",
			ConfigMapPath: "",
		},
	}
}

func (s *setupCiCdEntity) replaceCommonValues(value, envCode string) string {
	value = strings.ReplaceAll(value, "<environment>", envCode)
	value = strings.ReplaceAll(value, "{environment}", envCode)
	value = strings.ReplaceAll(value, "{{environment}}", envCode)
	value = strings.ReplaceAll(value, "{{.Environment}}", envCode)
	value = strings.ReplaceAll(value, "<squadName>", s.Squad().Code())
	value = strings.ReplaceAll(value, "{squadName}", s.Squad().Code())
	value = strings.ReplaceAll(value, "{{squadName}}", s.Squad().Code())
	value = strings.ReplaceAll(value, "{{.SquadName}}", s.Squad().Code())
	value = strings.ReplaceAll(value, "<namespace>", s.Squad().Code())
	value = strings.ReplaceAll(value, "{namespace}", s.Squad().Code())
	value = strings.ReplaceAll(value, "{{namespace}}", s.Squad().Code())
	value = strings.ReplaceAll(value, "{{.Namespace}}", s.Squad().Code())
	value = strings.ReplaceAll(value, "<applicationName>", s.ApplicationSlug())
	value = strings.ReplaceAll(value, "{applicationName}", s.ApplicationSlug())
	value = strings.ReplaceAll(value, "{{applicationName}}", s.ApplicationSlug())
	value = strings.ReplaceAll(value, "{{.ApplicationName}}", s.ApplicationSlug())
	return value
}
