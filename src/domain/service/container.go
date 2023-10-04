package service

type Container struct {
	CiCdService        CiCdService
	RegistryService    RegistryService
	RegistryApiService RegistryApiService
	GitService         GitService
	GitApiService      GitApiService
	DirectoryService   DirectoryService
	GitOpsService      GitOpsService
	PipelineService    PipelineService
	WikiService        WikiService
	WikiApiService     WikiApiService
	SecretService      SecretService
	SecretApiService   SecretApiService
}
