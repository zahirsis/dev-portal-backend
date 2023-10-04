package entity

type EnvironmentEntity interface {
	Code() string
	Label() string
	AccentColor() string
	DefaultActive() bool
	Concurrences() []string
	DefaultReplicas() ResourceObject
	RequireApproval() bool
	DestinationCluster() string
	Project() string
	SecretsPath() string
}

type environmentEntity struct {
	label              DataLabelObject
	accentColor        string
	defaultActive      bool
	defaultReplicas    ResourceObject
	concurrences       []string
	requireApproval    bool
	destinationCluster string
	project            string
	secretsPath        string
}

type EnvironmentConfig struct {
	Code               string
	Label              string
	AccentColor        string
	DefaultActive      bool
	DefaultReplicas    ResourceObject
	Concurrences       []string
	RequireApproval    bool
	DestinationCluster string
	Project            string
	SecretsPath        string
}

func NewEnvironmentEntity(e *EnvironmentConfig) EnvironmentEntity {
	return &environmentEntity{
		DataLabelObject{
			Code:  e.Code,
			Label: e.Label,
		},
		e.AccentColor,
		e.DefaultActive,
		e.DefaultReplicas,
		e.Concurrences,
		e.RequireApproval,
		e.DestinationCluster,
		e.Project,
		e.SecretsPath,
	}
}

func (t *environmentEntity) Code() string {
	return t.label.Code
}

func (t *environmentEntity) Label() string {
	return t.label.Label
}

func (t *environmentEntity) AccentColor() string {
	return t.accentColor
}

func (t *environmentEntity) DefaultActive() bool {
	return t.defaultActive
}

func (t *environmentEntity) Concurrences() []string {
	return t.concurrences
}

func (t *environmentEntity) DefaultReplicas() ResourceObject {
	return t.defaultReplicas
}

func (t *environmentEntity) RequireApproval() bool {
	return t.requireApproval
}

func (t *environmentEntity) DestinationCluster() string {
	return t.destinationCluster
}

func (t *environmentEntity) Project() string {
	return t.project
}

func (t *environmentEntity) SecretsPath() string {
	return t.secretsPath
}
