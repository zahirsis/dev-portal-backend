package service

type CreatedPullRequest struct {
	Id    int `json:"id"`
	Links struct {
		Html struct {
			Href string `json:"href"`
		} `json:"html"`
	} `json:"links"`
}

type PipelineEnvironment struct {
	Name      string              `json:"name"`
	Variables []*PipelineVariable `json:"variables"`
}

type PipelineVariable struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Secure bool   `json:"secured"`
}

type GitApiService interface {
	EnablePipelines(repository string) error
	CreatePullRequest(repository, sourceBranch, destinationBranch, title, message string) (*CreatedPullRequest, error)
	MergePullRequest(repository string, pullRequestId int) error
	SetRepositoryVariables(repository string, variables []*PipelineVariable) error
	SetRepositoryEnvironmentsVariables(repository string, environments []*PipelineEnvironment) error
	ActiveRepositoryPipelines(repository string) error
}
