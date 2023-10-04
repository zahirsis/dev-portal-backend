package bitbucket

import (
	"encoding/json"
	"fmt"
	"github.com/ktrysmt/go-bitbucket"
	"github.com/zahirsis/dev-portal-backend/config"
	"github.com/zahirsis/dev-portal-backend/src/domain/service"
	"github.com/zahirsis/dev-portal-backend/src/pkg/logger"
)

type gitApiService struct {
	cfg    *config.GitConfig
	logger logger.Logger
	client *bitbucket.Client
}

type bitbucketEnvironments struct {
	Uuid      string `json:"uuid"`
	Variables []*service.PipelineVariable
}

type variablesResponse struct {
	Variables []struct {
		Uuid  string `json:"uuid"`
		Key   string `json:"key"`
		Value string `json:"value"`
	}
}

type environmentResponse struct {
	Uuid string `json:"uuid"`
	Name string `json:"name"`
}

type environmentsResponse struct {
	Environments []*environmentResponse `json:"environments"`
}

func NewGitApiService(cfg *config.GitConfig, l logger.Logger, client *bitbucket.Client) service.GitApiService {
	return &gitApiService{cfg, l, client}
}

func (a *gitApiService) CreatePullRequest(repository, sourceBranch, destinationBranch, title, message string) (*service.CreatedPullRequest, error) {
	r, err := a.client.Repositories.PullRequests.Create(&bitbucket.PullRequestsOptions{
		RepoSlug:          a.cfg.GetRepositoryPath(repository),
		Title:             title,
		Description:       message,
		CloseSourceBranch: true,
		SourceBranch:      sourceBranch,
		DestinationBranch: destinationBranch,
	})
	pr := &service.CreatedPullRequest{}
	if err != nil {
		a.logger.Error("Error creating pull request", err.Error())
		return pr, err
	}
	if err := a.unmarshalResponse(r, pr, "create pull request"); err != nil {
		return pr, err
	}
	a.logger.Debug("Pull request created", r)
	return pr, nil
}

func (a *gitApiService) MergePullRequest(repository string, pullRequestId int) error {
	r, err := a.client.Repositories.PullRequests.Merge(&bitbucket.PullRequestsOptions{
		ID:                fmt.Sprintf("%d", pullRequestId),
		RepoSlug:          a.cfg.GetRepositoryPath(repository),
		CloseSourceBranch: true,
	})
	if err != nil {
		a.logger.Error("Error merging pull request", err)
		return err
	}
	a.logger.Debug("Pull request merged", r)
	return nil
}

func (a *gitApiService) EnablePipelines(repository string) error {
	a.logger.Debug("Enabling pipelines", repository)
	_, err := a.client.Repositories.Repository.UpdatePipelineConfig(&bitbucket.RepositoryPipelineOptions{
		RepoSlug: a.cfg.GetRepositoryPath(repository),
		Enabled:  true,
	})
	if err != nil {
		a.logger.Error("Error enabling pipelines", err)
		return err
	}
	return nil
}

func (a *gitApiService) SetRepositoryVariables(repository string, variables []*service.PipelineVariable) error {
	repoSlug := a.cfg.GetRepositoryPath(repository)
	a.logger.Debug("Setting repository variables", repository, variables)
	lvr, err := a.client.Repositories.Repository.ListPipelineVariables(&bitbucket.RepositoryPipelineVariablesOptions{
		RepoSlug: repoSlug,
	})
	if err != nil {
		a.logger.Error("Error listing repository variables", repository, err.Error())
		return err
	}
	lv := &variablesResponse{}
	if err := a.unmarshalResponse(lvr, lv, "list repository variables"); err != nil {
		return err
	}
	for _, v := range variables {
		skip := false
		for _, rv := range lv.Variables {
			if rv.Key == v.Key && rv.Value != v.Value {
				a.logger.Debug("Repository variable already exists with different value, removing", repository, v.Key)
				_, err = a.client.Repositories.Repository.DeletePipelineVariable(&bitbucket.RepositoryPipelineVariableDeleteOptions{
					RepoSlug: repoSlug,
					Uuid:     rv.Uuid,
				})
				if err != nil {
					a.logger.Error("Error deleting repository variable", repository, v.Key, err.Error())
					return err
				}
			}
			if rv.Key == v.Key && rv.Value == v.Value {
				a.logger.Debug("Repository variable already exists, skipping", repository, v.Key)
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		_, err = a.client.Repositories.Repository.AddPipelineVariable(&bitbucket.RepositoryPipelineVariableOptions{
			RepoSlug: a.cfg.GetRepositoryPath(repository),
			Key:      v.Key,
			Value:    v.Value,
			Secured:  v.Secure,
		})
		if err != nil {
			a.logger.Error("Error creating repository variable", repository, v.Key, err.Error())
			return err
		}
	}
	return nil
}

func (a *gitApiService) SetRepositoryEnvironmentsVariables(repository string, environments []*service.PipelineEnvironment) error {
	repoSlug := a.cfg.GetRepositoryPath(repository)
	a.logger.Debug("Setting repository environment variables", repository, environments)
	le, err := a.client.Repositories.Repository.ListEnvironments(&bitbucket.RepositoryEnvironmentsOptions{
		RepoSlug: repoSlug,
	})
	if err != nil {
		a.logger.Error("Error listing repository environments", repository, err.Error())
		return err
	}
	e := &environmentsResponse{}
	if err := a.unmarshalResponse(le, e, "list repository environments"); err != nil {
		return err
	}
	for _, v := range e.Environments {
		remove := true
		for _, lv := range environments {
			if v.Name == lv.Name {
				remove = false
			}
		}
		if remove {
			_ = a.removeEnvironment(repoSlug, v.Uuid)
		}
	}
	envs := make(map[string]*bitbucketEnvironments)
	for _, v := range environments {
		for _, lv := range e.Environments {
			if lv.Name == v.Name {
				envs[v.Name] = &bitbucketEnvironments{
					Uuid:      lv.Uuid,
					Variables: v.Variables,
				}
			}
		}
		if _, ok := envs[v.Name]; ok {
			continue
		}

		er, err := a.client.Repositories.Repository.AddEnvironment(&bitbucket.RepositoryEnvironmentOptions{
			RepoSlug: repoSlug,
			Name:     v.Name,
		})
		if err != nil {
			a.logger.Error("Error creating repository environment", repository, v.Name, err.Error())
			return err
		}
		ec := &environmentResponse{}
		if err := a.unmarshalResponse(er, ec, "create repository environment"); err != nil {
			return err
		}
		envs[v.Name] = &bitbucketEnvironments{
			Uuid:      ec.Uuid,
			Variables: v.Variables,
		}
	}
	for _, e := range envs {
		lvr, err := a.client.Repositories.Repository.ListDeploymentVariables(&bitbucket.RepositoryDeploymentVariablesOptions{
			RepoSlug: repoSlug,
			Environment: &bitbucket.Environment{
				Uuid: e.Uuid,
			},
		})
		if err != nil {
			a.logger.Error("Error listing repository variables", repository, err.Error())
			return err
		}
		lv := &variablesResponse{}
		if err := a.unmarshalResponse(lvr, lv, "list repository variables"); err != nil {
			return err
		}
		for _, v := range e.Variables {
			skip := false
			for _, vr := range lv.Variables {
				if vr.Key == v.Key && vr.Value != v.Value {
					a.logger.Debug("Environment Variable already exists with different value, removing", repository, v.Key)
					_, err = a.client.Repositories.Repository.DeleteDeploymentVariable(&bitbucket.RepositoryDeploymentVariableDeleteOptions{
						RepoSlug: repoSlug,
						Uuid:     vr.Uuid,
						Environment: &bitbucket.Environment{
							Uuid: e.Uuid,
						},
					})
					if err != nil {
						a.logger.Error("Error deleting environment variable", repository, v.Key, err.Error())
						return err
					}
				}
				if vr.Key == v.Key && vr.Value == v.Value {
					a.logger.Debug("Environment variable already exists, skipping", repository, v.Key)
					skip = true
				}
			}
			if skip {
				continue
			}
			_, err = a.client.Repositories.Repository.AddDeploymentVariable(&bitbucket.RepositoryDeploymentVariableOptions{
				RepoSlug: a.cfg.GetRepositoryPath(repository),
				Key:      v.Key,
				Value:    v.Value,
				Environment: &bitbucket.Environment{
					Uuid: e.Uuid,
				},
			})
			if err != nil {
				a.logger.Error("Error creating environment variable", repository, v.Key, err.Error())
				return err
			}
		}
	}
	return nil
}

func (a *gitApiService) ActiveRepositoryPipelines(repository string) error {
	a.logger.Debug("Activating repository pipelines", repository)
	return nil
}

func (a *gitApiService) unmarshalResponse(r interface{}, pr any, responseType string) error {
	jr, err := json.Marshal(r)
	if err != nil {
		a.logger.Error("Error marshalling response: "+responseType, r, err.Error())
		return err
	}
	err = json.Unmarshal(jr, pr)
	if err != nil {
		a.logger.Error("Error unmarshalling response: "+responseType, string(jr), err.Error())
		return err
	}
	return nil
}

func (a *gitApiService) removeEnvironment(slug string, uuid string) error {
	a.logger.Debug("Removing environment", slug, uuid)
	_, err := a.client.Repositories.Repository.DeleteEnvironment(&bitbucket.RepositoryEnvironmentDeleteOptions{
		RepoSlug: slug,
		Uuid:     uuid,
	})
	if err != nil {
		a.logger.Error("Error deleting environment", slug, uuid, err.Error())
		return err
	}
	return nil
}
