package repository

type Container struct {
	TemplateRepository    TemplateRepository
	EnvironmentRepository EnvironmentRepository
	SquadRepository       SquadRepository
	ProgressRepository    ProcessRepository
	ManifestRepository    ManifestRepository
}
