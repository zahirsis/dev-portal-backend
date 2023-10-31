package usecase

import (
	"encoding/json"
	"fmt"
	"github.com/zahirsis/dev-portal-backend/config"
	"github.com/zahirsis/dev-portal-backend/pkg/errors"
	"github.com/zahirsis/dev-portal-backend/src/domain/entity"
	"github.com/zahirsis/dev-portal-backend/src/domain/service"
	"github.com/zahirsis/dev-portal-backend/src/infrastructure/container"
	"strings"
	"time"
)

type updateProgressData struct {
	ID      string
	Step    string
	Message string
	Type    string
	IsNode  bool
}

type processData struct {
	id                        string
	data                      entity.SetupCiCdEntity
	rootDestinationDir        string
	templatesRepository       string
	templatesBranch           string
	templatesDestinationDir   string
	gitOpsRepository          string
	gitOpsBranch              string
	gitOpsDestinationDir      string
	gitOpsToolsRepository     string
	gitOpsToolsBranch         string
	gitOpsToolsDestinationDir string
	configMapRepository       string
	configMapBranch           string
	configMapDestinationDir   string
	applicationBranch         string
	applicationDestination    string
	defaultManifests          []*entity.Manifest
}

func (p *processData) customBranch(additionalName string) string {
	if additionalName == "" {
		return fmt.Sprintf("feature/%s", p.data.ApplicationSlug())
	}
	return fmt.Sprintf("feature/%s/%s", p.data.ApplicationSlug(), additionalName)
}

type CiCdInputDto struct {
	Template string `json:"template"`
	Envs     []struct {
		Code     string               `json:"code"`
		Replicas entity.LimitsIntData `json:"replicas"`
	} `json:"envs"`
	Manifests   []string               `json:"manifests"`
	Squad       string                 `json:"squad"`
	Application entity.ApplicationData `json:"application"`
	Ingress     entity.IngressData     `json:"ingress"`
}

type CiCdOutputDto struct {
	Errors    []error `json:"errors"`
	ProcessId string  `json:"code"`
}

type SetupCiCdUseCase interface {
	Exec(i CiCdInputDto) CiCdOutputDto
}

type setupCiCdUseCase struct {
	*container.Container
	config *config.Config
}

func NewSetupCiCdUseCase(c *container.Container, cfg *config.Config) SetupCiCdUseCase {
	return &setupCiCdUseCase{c, cfg}
}

func (uc *setupCiCdUseCase) Exec(i CiCdInputDto) CiCdOutputDto {
	uc.Logger.Debug("RECEIVED REQUEST: ci-cd/setup", i)
	processID := uc.MessageManager.GenerateID()
	var errs []error
	e, errs := uc.makeEntity(i, processID)
	errs = append(errs, uc.Services.CiCdService.ValidateSetup(e)...)
	dm, err := uc.defaultManifests()
	if err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		uc.Logger.Debug("ERRORS VALIDATE SETUP:", errs)
		return CiCdOutputDto{Errors: errs}
	}
	// process
	sc := uc.config.SetupCiCd
	data := &processData{
		id:                        processID,
		data:                      e,
		rootDestinationDir:        strings.Replace(sc.RootDestinationsPath, "{{process-id}}", processID, -1),
		templatesRepository:       sc.TemplatesRepository,
		templatesBranch:           sc.TemplatesRepositoryBranch,
		templatesDestinationDir:   strings.Replace(sc.TemplatesDestinationDir, "{{process-id}}", processID, -1),
		gitOpsRepository:          sc.GitOpsRepository,
		gitOpsBranch:              sc.GitOpsRepositoryBranch,
		gitOpsDestinationDir:      strings.Replace(sc.GitOpsDestinationDir, "{{process-id}}", processID, -1),
		gitOpsToolsRepository:     sc.GitOpsToolsRepository,
		gitOpsToolsBranch:         sc.GitOpsToolsRepositoryBranch,
		gitOpsToolsDestinationDir: strings.Replace(sc.GitOpsToolsDestinationDir, "{{process-id}}", processID, -1),
		configMapRepository:       sc.ConfigMapRepository,
		configMapBranch:           sc.ConfigMapRepositoryBranch,
		configMapDestinationDir:   strings.Replace(sc.ConfigMapDestinationDir, "{{process-id}}", processID, -1),
		applicationBranch:         sc.ApplicationMainBranch,
		applicationDestination:    strings.Replace(sc.ApplicationDestinationDir, "{{process-id}}", processID, -1),
		defaultManifests:          dm,
	}
	go uc.process(data)
	return CiCdOutputDto{Errors: nil, ProcessId: processID}
}

func (uc *setupCiCdUseCase) process(pd *processData) {
	defer func() {
		if r := recover(); r != nil {
			uc.Logger.Error("Recovered in process", r)
			uc.finish(pd, []string{"Process interrupted by internal error"}, true)
		}
	}()
	uc.Logger.Debug("PROCESSING", pd.id)
	var additionalData []string

	// Step: Pre Process Setup Ci/CD Automation
	ud := updateProgressData{
		ID:      pd.id,
		Step:    strings.ToLower("pre-process-setup-ci-cd-automation"),
		Message: "Pre Process Setup Ci/CD Automation",
		Type:    "progress",
		IsNode:  true,
	}
	uc.updateProgress(ud, "")
	if err := uc.stepClone(pd.id, pd.data.ApplicationName(), pd.data.ApplicationName(), pd.applicationBranch, pd.applicationDestination, ud.Step); err != nil {
		uc.finish(pd, additionalData, true)
		return
	}
	// Step: Clone templates repository
	if err := uc.stepClone(pd.id, "Templates", pd.templatesRepository, pd.templatesBranch, pd.templatesDestinationDir, ""); err != nil {
		uc.finish(pd, additionalData, true)
		return
	}
	// Step: Create secrets
	sm := uc.getManifests(pd, entity.SecretManifests)
	if len(sm) > 0 {
		sw, err := uc.setupSecret(pd, sm)
		additionalData = append(additionalData, sw...)
		if err != nil {
			uc.finish(pd, additionalData, true)
			return
		}
	}

	// Step: Create registry
	rm := uc.getManifests(pd, entity.RegistryManifests)
	if len(rm) > 0 {
		cr, err := uc.createRegistry(pd, rm)
		additionalData = append(additionalData, cr...)
		if err != nil {
			uc.finish(pd, additionalData, true)
			return
		}
	}
	// Group Step: GitOps and k8s manifests
	gm := uc.getManifests(pd, entity.GitOpsManifests)
	if len(gm) > 0 {
		ud := updateProgressData{
			ID:      pd.id,
			Step:    strings.ToLower("clone-git-ops-repositories"),
			Message: "Cloning GitOps repositories",
			Type:    "progress",
			IsNode:  true,
		}
		uc.updateProgress(ud, "")
		// Step: Clone gitOps repository
		err := uc.stepClone(pd.id, "GitOps", pd.gitOpsRepository, pd.gitOpsBranch, pd.gitOpsDestinationDir, ud.Step)
		if err != nil {
			uc.finish(pd, additionalData, true)
			return
		}
		// Step: Clone gitOps-tools repository
		err = uc.stepClone(pd.id, "GitOps-Tools", pd.gitOpsToolsRepository, pd.gitOpsToolsBranch, pd.gitOpsToolsDestinationDir, ud.Step)
		if err != nil {
			uc.finish(pd, additionalData, true)
			return
		}
		if uc.config.SetupCiCd.ExternalConfigMap {
			// Step: Clone configMap repository
			err = uc.stepClone(pd.id, "ConfigMap", pd.configMapRepository, pd.configMapBranch, pd.configMapDestinationDir, ud.Step)
			if err != nil {
				uc.finish(pd, additionalData, true)
				return
			}
		}
		// Step: Create k8s manifests
		ck8s, err := uc.createK8sManifests(pd, gm)
		additionalData = append(additionalData, ck8s...)
		if err != nil {
			uc.finish(pd, additionalData, true)
			return
		}
		// Step: Create GitOps manifests
		cgm, err := uc.createGitOpsManifests(pd, gm)
		additionalData = append(additionalData, cgm...)
		if err != nil {
			uc.finish(pd, additionalData, true)
			return
		}
	}
	// Step Group: Pipeline
	pm := uc.getManifests(pd, entity.PipelineManifests)
	if len(pm) > 0 {
		// Step: Create pipeline
		cp, err := uc.createPipeline(pd, pm)
		additionalData = append(additionalData, cp...)
		if err != nil {
			uc.finish(pd, additionalData, true)
			return
		}
	}
	// Step: Create wiki
	wm := uc.getManifests(pd, entity.WikiManifests)
	if len(wm) > 0 {
		sw, err := uc.setupWiki(pd, wm)
		additionalData = append(additionalData, sw...)
		if err != nil {
			uc.finish(pd, additionalData, true)
			return
		}
	}
	// Step: Mark as finish
	uc.finish(pd, additionalData, false)
}

func (uc *setupCiCdUseCase) makeEntity(i CiCdInputDto, ID string) (entity.SetupCiCdEntity, []error) {
	var envs []entity.SetupEnvData
	var errs []error
	for i, v := range i.Envs {
		if v.Code == "" {
			errs = append(errs, errors.NewInputError(
				fmt.Sprintf("env.%d.code", i),
				[]string{"env cannot be empty"},
			))
			continue
		}
		env, err := uc.Repositories.EnvironmentRepository.Get(v.Code)
		if err != nil {
			errs = append(errs, errors.NewInputError("envs."+v.Code, []string{err.Error()}))
		}
		envs = append(envs, entity.NewSetupEnvData(env, v.Replicas.Min, v.Replicas.Max))
	}
	var template entity.TemplateEntity
	var err error
	if i.Template == "" {
		errs = append(errs, errors.NewInputError("template", []string{"template cannot be empty"}))
	} else {
		template, err = uc.Repositories.TemplateRepository.Get(i.Template)
		if err != nil {
			errs = append(errs, errors.NewInputError("template", []string{err.Error()}))
		}
	}
	var squad entity.SquadEntity
	if i.Squad == "" {
		errs = append(errs, errors.NewInputError("squad", []string{"squad cannot be empty"}))
	} else {
		squad, err = uc.Repositories.SquadRepository.Get(i.Squad)
		if err != nil {
			errs = append(errs, errors.NewInputError("squad", []string{err.Error()}))
		}
	}

	var manifests []*entity.Manifest
	for _, m := range template.Manifests() {
		for _, v := range i.Manifests {
			if m.Code == v {
				manifests = append(manifests, m)
			}
		}
	}

	return entity.NewSetupCiCdEntity(entity.SetupCiCdData{
		ID:          ID,
		Template:    template,
		Envs:        envs,
		Manifests:   manifests,
		Squad:       squad,
		Application: i.Application,
		Ingress:     i.Ingress,
	}), errs
}

func (uc *setupCiCdUseCase) setupSecret(pd *processData, manifests []*entity.Manifest) ([]string, error) {
	data := updateProgressData{
		ID:      pd.id,
		Step:    "setup-secrets",
		Message: "Creating Secrets",
		Type:    "progress",
		IsNode:  true,
	}
	uc.updateProgress(data, "")
	data.IsNode = false
	var extraData []string
	for _, v := range manifests {
		data.Type = "progress"
		uc.updateProgress(data, fmt.Sprintf("Creating %s for %s using %s manifests", v.Label, pd.data.ApplicationName(), v.Code))
		secret, err := uc.Services.SecretService.LoadData(pd.data, v, pd.templatesDestinationDir)
		if err != nil {
			uc.updateProgressError(data, err, fmt.Sprintf("Error creating secret with %s manifests", v.Code))
			return []string{}, err
		}
		for _, env := range pd.data.Envs() {
			err = uc.Services.SecretService.SetupNewSecret(secret, env)
			if err != nil {
				uc.updateProgressError(data, err, fmt.Sprintf("Error creating secret with %s manifests", v.Code))
				return []string{}, err
			}
			extraData = append(extraData, fmt.Sprintf(" -- %s: %s - %s", v.Label, secret.Config().GetRootPath(env), secret.Config().GetSecretPath(env)))
			data.Type = "success"
			uc.updateProgress(data, fmt.Sprintf("%s's secrets created for %s's service", v.Label, pd.data.ApplicationSlug()))
		}
	}
	if len(extraData) > 0 {
		extraData = append([]string{"Secrets created:"}, extraData...)
	}
	return extraData, nil
}

func (uc *setupCiCdUseCase) createRegistry(pd *processData, manifests []*entity.Manifest) ([]string, error) {
	data := updateProgressData{
		ID:      pd.id,
		Step:    "create-registry",
		Message: "Creating Registry",
		Type:    "progress",
		IsNode:  true,
	}
	uc.updateProgress(data, "")
	data.IsNode = false
	var extraData []string
	for _, v := range manifests {
		data.Type = "progress"
		uc.updateProgress(data, fmt.Sprintf("Creating %s %s using %s manifests", v.Label, pd.data.ApplicationName(), v.Code))
		registry, err := uc.Services.RegistryService.LoadData(pd.data, v, pd.templatesDestinationDir)
		if err != nil {
			uc.updateProgressError(data, err, fmt.Sprintf("Error creating registry with %s manifests", v.Code))
			return []string{}, err
		}
		url, err := uc.Services.RegistryApiService.Create(registry)
		if err != nil {
			uc.updateProgressError(data, err, fmt.Sprintf("Error creating registry with %s manifests", v.Code))
			return []string{}, err
		}
		extraData = append(extraData, fmt.Sprintf("Registry url: https://%s", url))
		pd.data.CreatedData().RegistryUrl = url
		data.Type = "success"
		uc.updateProgress(data, fmt.Sprintf("%s's registry created for %s", v.Label, pd.data.ApplicationName()))
	}
	return extraData, nil
}

func (uc *setupCiCdUseCase) createK8sManifests(pd *processData, gm []*entity.Manifest) ([]string, error) {
	data := updateProgressData{
		ID:      pd.id,
		Step:    "create-k8s-manifests",
		Message: "Creating K8s manifests",
		Type:    "progress",
		IsNode:  true,
	}
	uc.updateProgress(data, "")
	data.IsNode = false
	var extraData []string

	for _, m := range gm {
		data.Type = "progress"
		uc.updateProgress(data, "Creating repositories branch for changes")
		customBranch := pd.customBranch(m.Code)
		if err := uc.newBranchFromDefault(data, pd.gitOpsDestinationDir, pd.gitOpsBranch, customBranch); err != nil {
			return []string{}, err
		}
		if err := uc.newBranchFromDefault(data, pd.gitOpsToolsDestinationDir, pd.gitOpsToolsBranch, customBranch); err != nil {
			return []string{}, err
		}
		if uc.config.SetupCiCd.ExternalConfigMap {
			if err := uc.newBranchFromDefault(data, pd.configMapDestinationDir, pd.configMapBranch, customBranch); err != nil {
				return []string{}, err
			}
		}

		uc.updateProgress(data, fmt.Sprintf("Creating %s k8s manifests", m.Code))
		ge, err := uc.Services.GitOpsService.LoadData(pd.data, m, pd.templatesDestinationDir)
		if err != nil {
			uc.updateProgressError(data, err, fmt.Sprintf("Error loading data from %s manifest", m.Code))
			return []string{}, err
		}
		uc.updateProgress(data, "Configuring common utilities with k8s manifests")
		err = uc.Services.GitOpsService.SetupBaseUtilities(ge, pd.templatesDestinationDir, pd.gitOpsToolsDestinationDir)
		if err != nil {
			uc.updateProgressError(data, err, fmt.Sprintf("Error creating base utilities manifests from %s k8s templates", m.Code))
			return []string{}, err
		}
		uc.updateProgress(data, "Configuring namespace utilities with k8s manifests")
		err = uc.Services.GitOpsService.SetupNamespacedUtilities(ge, pd.templatesDestinationDir, pd.gitOpsToolsDestinationDir)
		if err != nil {
			uc.updateProgressError(data, err, fmt.Sprintf("Error creating namespace utilities manifests from %s k8s templates", m.Code))
			return []string{}, err
		}

		uc.updateProgress(data, "Configuring k8s manifests for service")
		kd, err := uc.Services.GitOpsService.SetupK8sManifests(ge, pd.templatesDestinationDir, pd.gitOpsDestinationDir, pd.configMapDestinationDir)
		extraData = append(extraData, kd...)
		if err != nil {
			uc.updateProgressError(data, err, fmt.Sprintf("Error creating k8s manifests from %s templates", m.Code))
			return []string{}, err
		}
		commitMessage := fmt.Sprintf("feat: add %s - %s manifests [Setup Ci/CD Automation]", pd.data.ApplicationSlug(), m.Label)
		prd := pullRequestData{
			pd:           data,
			localDir:     pd.gitOpsDestinationDir,
			repository:   uc.config.SetupCiCd.GitOpsRepository,
			targetBranch: pd.gitOpsBranch,
			actualBranch: customBranch,
			message:      commitMessage,
			title:        fmt.Sprintf("Create %s's %s manifests", pd.data.ApplicationSlug(), m.Label),
			merge:        true,
		}
		if _, err := uc.makePr(prd, true); err != nil {
			return []string{}, err
		}
		pd.data.CreatedData().GitOpsPath = uc.config.GitConfig.GetRepositoryUrl(prd.repository) + "/" + ge.Config().K8sApplicationDestinationPath

		prd.localDir = pd.gitOpsToolsDestinationDir
		prd.targetBranch = pd.gitOpsToolsBranch
		prd.repository = uc.config.SetupCiCd.GitOpsToolsRepository
		if _, err := uc.makePr(prd, true); err != nil {
			return []string{}, err
		}

		if uc.config.SetupCiCd.ExternalConfigMap {
			prd.localDir = pd.configMapDestinationDir
			prd.targetBranch = pd.configMapBranch
			prd.repository = uc.config.SetupCiCd.ConfigMapRepository
			if _, err := uc.makePr(prd, true); err != nil {
				return []string{}, err
			}
		}
		pd.data.CreatedData().ConfigMapPath = uc.config.GitConfig.GetRepositoryUrl(prd.repository) + "/" + ge.Config().K8sConfigMapDestinationPath

		data.Type = "success"
		uc.updateProgress(data, fmt.Sprintf("%s's manifests created for %s's service", m.Code, pd.data.ApplicationSlug()))
	}
	return extraData, nil
}

func (uc *setupCiCdUseCase) createGitOpsManifests(pd *processData, gm []*entity.Manifest) ([]string, error) {
	data := updateProgressData{
		ID:      pd.id,
		Step:    "create-git-ops-manifests",
		Message: "Creating GitOps manifests",
		Type:    "progress",
		IsNode:  true,
	}
	uc.updateProgress(data, "")
	data.IsNode = false
	var extraData []string
	prd := pullRequestData{
		pd:           data,
		localDir:     pd.gitOpsToolsDestinationDir,
		repository:   uc.config.SetupCiCd.GitOpsToolsRepository,
		targetBranch: pd.gitOpsToolsBranch,
	}
	for _, m := range gm {
		data.Type = "progress"
		uc.updateProgress(data, fmt.Sprintf("Creating %s gitOps manifests", m.Code))
		ge, err := uc.Services.GitOpsService.LoadData(pd.data, m, pd.templatesDestinationDir)
		if err != nil {
			uc.updateProgressError(data, err, fmt.Sprintf("Error loading data from %s manifest", m.Code))
			return []string{}, err
		}
		for _, e := range pd.data.Envs() {
			data.Type = "progress"
			// Create branch for changes
			uc.updateProgress(data, fmt.Sprintf("Creating repositories branch for changes on %s environment", e.Env().Code()))
			customBranch := pd.customBranch(fmt.Sprintf("%s/%s", e.Env().Code(), m.Code))
			if err := uc.newBranchFromDefault(data, pd.gitOpsToolsDestinationDir, pd.gitOpsToolsBranch, customBranch); err != nil {
				return []string{}, err
			}
			uc.updateProgress(data, fmt.Sprintf("Creating manifests for %s environment", e.Env().Code()))
			err = uc.Services.GitOpsService.SetupGitOpsManifests(ge, pd.templatesDestinationDir, pd.gitOpsToolsDestinationDir, e)
			if err != nil {
				uc.updateProgressError(data, err, fmt.Sprintf("Error creating manifests from %s gitOps templates on environment %s", m.Code))
				return []string{}, err
			}
			commitMessage := fmt.Sprintf("feat: add %s - %s manifests at %s environment [Setup Ci/CD Automation]", pd.data.ApplicationSlug(), m.Label, e.Env().Label())
			prd.actualBranch = customBranch
			prd.message = commitMessage
			prd.title = fmt.Sprintf("Deploy %s at %s environment with %s", pd.data.ApplicationSlug(), e.Env().Label(), m.Label)
			prd.merge = !e.Env().RequireApproval()
			if prUrl, err := uc.makePr(prd, true); err != nil {
				return []string{}, err
			} else if prUrl != "" {
				extraData = append(extraData, " -- "+prUrl)
			}
			data.Type = "success"
			uc.updateProgress(data, fmt.Sprintf("%s's manifests created for %s's environment of %s's service", m.Code, e.Env().Code(), pd.data.ApplicationSlug()))
		}
	}
	if len(extraData) > 0 {
		extraData = append([]string{"Pull requests:"}, extraData...)
	}
	return extraData, nil
}

func (uc *setupCiCdUseCase) createPipeline(pd *processData, pm []*entity.Manifest) ([]string, error) {
	data := updateProgressData{
		ID:      pd.id,
		Step:    "create-pipeline-manifests",
		Message: "Creating Pipeline",
		Type:    "progress",
		IsNode:  true,
	}
	uc.updateProgress(data, "")
	data.IsNode = false
	var extraData []string

	for _, m := range pm {
		data.Type = "progress"
		pe, err := uc.Services.PipelineService.LoadData(pd.data, m, pd.templatesDestinationDir)
		if err != nil {
			uc.updateProgressError(data, err, fmt.Sprintf("Error loading data from %s manifest", m.Code))
			return []string{}, err
		}
		// Enabling pipelines and setting up variables
		uc.updateProgress(data, fmt.Sprintf("Enabling pipelines on %s repository", pd.data.ApplicationName()))
		if err := uc.Services.GitApiService.EnablePipelines(pd.data.ApplicationName()); err != nil {
			uc.updateProgressError(data, err, fmt.Sprintf("Error enabling pipelines on %s repository", pd.data.ApplicationName()))
			return extraData, err
		}
		// Setting up variables
		uc.updateProgress(data, fmt.Sprintf("Setting up variables on %s repository", pd.data.ApplicationName()))
		if err := uc.Services.GitApiService.SetRepositoryVariables(pd.data.ApplicationName(), uc.getRepositoryVariables(pe.Config().DefaultVariables)); err != nil {
			uc.updateProgressError(data, err, fmt.Sprintf("Error setting up variables on %s repository", pd.data.ApplicationName()))
			return extraData, err
		}
		// Setting up environment variables
		var environments []*service.PipelineEnvironment
		for _, e := range pd.data.Envs() {
			uc.updateProgress(data, fmt.Sprintf("Setting up variables on %s repository for %s environment", pd.data.ApplicationName(), e.Env().Code()))
			if _, ok := pe.Config().Environments[e.Env().Code()]; !ok {
				continue
			}
			variables := uc.getRepositoryVariables(pe.Config().Environments[e.Env().Code()].Variables)
			environments = append(environments, &service.PipelineEnvironment{
				Name:      e.Env().Code(),
				Variables: variables,
			})
		}
		if err := uc.Services.GitApiService.SetRepositoryEnvironmentsVariables(pd.data.ApplicationName(), environments); err != nil {
			uc.updateProgressError(data, err, fmt.Sprintf("Error setting up environments' variables on %s repository", pd.data.ApplicationName()))
			return extraData, err
		}

		// Create branch for changes
		uc.updateProgress(data, "Creating new branch for add pipeline files")
		customBranch := pd.customBranch(m.Code)
		if err := uc.newBranchFromDefault(data, pd.applicationDestination, pd.applicationBranch, customBranch); err != nil {
			return []string{}, err
		}
		uc.updateProgress(data, fmt.Sprintf("Creating %s pipeline", m.Code))
		err = uc.Services.PipelineService.SetupPipeline(pe, pd.templatesDestinationDir, pd.applicationDestination)
		if err != nil {
			uc.updateProgressError(data, err, fmt.Sprintf("Error creating pipeline from %s templates", m.Code))
			return []string{}, err
		}
		commitMessage := "feat: add pipeline files [Setup Ci/CD Automation] [skip ci]"
		prd := pullRequestData{
			pd:           data,
			localDir:     pd.applicationDestination,
			repository:   pd.data.ApplicationName(),
			targetBranch: pd.applicationBranch,
			actualBranch: customBranch,
			message:      commitMessage,
			title:        fmt.Sprintf("Create pipeline [Setup Ci/CD Automation] [skip ci]"),
			merge:        true,
		}
		if _, err := uc.makePr(prd, true); err != nil {
			return []string{}, err
		}
		data.Type = "success"
		uc.updateProgress(data, fmt.Sprintf("%s's pipeline created for %s's service", m.Code, pd.data.ApplicationSlug()))
	}

	return extraData, nil
}

func (uc *setupCiCdUseCase) setupWiki(pd *processData, manifests []*entity.Manifest) ([]string, error) {
	data := updateProgressData{
		ID:      pd.id,
		Step:    "setup-wiki",
		Message: "Creating Wiki",
		Type:    "progress",
		IsNode:  true,
	}
	uc.updateProgress(data, "")
	data.IsNode = false
	var extraData []string
	for _, v := range manifests {
		data.Type = "progress"
		uc.updateProgress(data, fmt.Sprintf("Creating %s wiki for %s using %s manifests", v.Label, pd.data.ApplicationName(), v.Code))
		wiki, err := uc.Services.WikiService.LoadData(pd.data, v, pd.templatesDestinationDir)
		if err != nil {
			uc.updateProgressError(data, err, fmt.Sprintf("Error creating wiki with %s manifests", v.Code))
			return []string{}, err
		}
		output, err := uc.Services.WikiService.SetupWiki(wiki, pd.templatesDestinationDir)
		if err != nil {
			uc.updateProgressError(data, err, fmt.Sprintf("Error creating wiki with %s manifests", v.Code))
			return []string{}, err
		}
		extraData = append(extraData, output...)
		data.Type = "success"
		uc.updateProgress(data, fmt.Sprintf("%s's wiki created for %s's service", v.Label, pd.data.ApplicationName()))
	}
	return extraData, nil
}

func (uc *setupCiCdUseCase) finish(pd *processData, additionalData []string, errs bool) {
	data := updateProgressData{
		ID:      pd.id,
		Step:    "finish-setup",
		Message: "",
		Type:    "",
		IsNode:  true,
	}
	if errs {
		data.Type = "error"
		data.Message = "Process finish with errors"
	} else {
		data.Type = "success"
		data.Message = "Process finish with success"
	}
	uc.updateProgress(data, "")
	data.IsNode = false
	for _, v := range additionalData {
		uc.updateProgress(data, v)
	}
	uc.updateProgress(data, "Cleaning setup state")
	defer func() {
		if r := recover(); r != nil {
			uc.Logger.Error("Recovered in finish: cleaning state", r)
			uc.markAsFinished(pd.id)
		}
	}()
	// TODO: uncomment when directory service is implemented
	//err := uc.Services.DirectoryService.RemoveDirectory(pd.rootDestinationDir)
	//if err != nil {
	//	uc.Logger.Error(err.Error())
	//}
	uc.updateProgress(data, "Setup state cleaned")
	uc.markAsFinished(pd.id)
}

func (uc *setupCiCdUseCase) markAsFinished(ID string) {
	uc.Logger.Debug("FINISHING PROCESS", ID)
	err := uc.Repositories.ProgressRepository.MarkAsFinished(ID)
	if err != nil {
		uc.Logger.Error("Error marking process as finish: %s", err.Error())
	}
	uc.MessageManager.Close(ID)
	uc.Logger.Debug("PROCESSING FINISHED", ID)
}

func (uc *setupCiCdUseCase) stepClone(ID, name, repository, branch, destination, step string) error {
	data := updateProgressData{
		ID:      ID,
		Step:    step,
		Message: "",
		Type:    "progress",
		IsNode:  false,
	}
	if step == "" {
		data.IsNode = true
		data.Step = strings.ToLower(fmt.Sprintf("clone-%s-repository", name))
		uc.updateProgress(data, fmt.Sprintf("Cloning %s repository", name))
	}
	data.IsNode = false

	uc.updateProgress(data, fmt.Sprintf("Cloning %s on branch %s into %s", repository, branch, destination))
	err := uc.Services.GitService.CloneRepository(repository, branch, destination)
	if err != nil {
		uc.updateProgressError(data, err, fmt.Sprintf("Error cloning %s into %s", repository, destination))
		return err
	}
	data.Type = "success"
	uc.updateProgress(data, fmt.Sprintf("Repository %s cloned into %s", repository, destination))
	return nil
}

func (uc *setupCiCdUseCase) updateProgressError(data updateProgressData, err error, message string) {
	data.Type = "error"
	uc.updateProgress(data, message)
	uc.updateProgress(data, "Error: "+err.Error())
}

func (uc *setupCiCdUseCase) updateProgress(data updateProgressData, customMessage string) {
	uc.Logger.Debug("UPDATE PROGRESS", data.ID, data.Step, data.Message, data.Type)
	if customMessage != "" {
		data.Message = customMessage
	}
	update := entity.Progress{
		Time:    time.Now(),
		Step:    data.Step,
		Message: data.Message,
		Kind:    data.Type,
		Node:    data.IsNode,
	}
	jsonUpdate, _ := json.Marshal(update)
	uc.MessageManager.Broadcast(data.ID, jsonUpdate)
	_ = uc.Repositories.ProgressRepository.SaveMessage(data.ID, entity.NewProgressEntity(update))
}

func (uc *setupCiCdUseCase) getManifests(pd *processData, manifestType entity.ManifestType) []*entity.Manifest {
	var m []*entity.Manifest
	for _, v := range pd.data.Manifests() {
		if v.Type == manifestType {
			m = append(m, v)
		}
	}
	for _, v := range pd.defaultManifests {
		if v.Type == manifestType {
			m = append(m, v)
		}
	}
	return m
}

func (uc *setupCiCdUseCase) newBranchFromDefault(data updateProgressData, path, defaultBranch, newBranch string) error {
	uc.updateProgress(data, fmt.Sprintf("Checking out default branch on %s", path))
	err := uc.Services.GitService.Checkout(path, defaultBranch)
	if err != nil {
		uc.updateProgressError(data, err, fmt.Sprintf("Error checking out default branch on %s", path))
		return err
	}
	uc.updateProgress(data, fmt.Sprintf("Pulling default branch on %s", path))
	err = uc.Services.GitService.Pull(path, defaultBranch)
	if err != nil {
		uc.updateProgressError(data, err, fmt.Sprintf("Error pulling default branch on %s", path))
	}
	uc.updateProgress(data, fmt.Sprintf("Creating %s branch on %s", newBranch, path))
	err = uc.Services.GitService.Branch(path, newBranch)
	if err != nil {
		uc.updateProgressError(data, err, fmt.Sprintf("Error creating %s branch on %s", newBranch, path))
		return err
	}
	return nil
}

type pullRequestData struct {
	pd           updateProgressData
	localDir     string
	repository   string
	targetBranch string
	actualBranch string
	message      string
	title        string
	merge        bool
}

func (uc *setupCiCdUseCase) makePr(data pullRequestData, commit bool) (string, error) {
	hasChanges, err := uc.Services.GitService.HasChanges(data.localDir)
	if err != nil {
		uc.updateProgressError(data.pd, err, fmt.Sprintf("Error checking changes on %s", data.localDir))
		return "", err
	}
	if !hasChanges {
		uc.updateProgress(data.pd, fmt.Sprintf("No changes on %s", data.localDir))
		return "", nil
	}
	if commit {
		uc.updateProgress(data.pd, fmt.Sprintf("Committing changes on %s", data.localDir))
		if err := uc.Services.GitService.Commit(data.localDir, data.message); err != nil {
			uc.updateProgressError(data.pd, err, fmt.Sprintf("Error committing changes on %s", data.localDir))
			return "", err
		}
	}
	uc.updateProgress(data.pd, fmt.Sprintf("Pushing changes on %s", data.localDir))
	if err := uc.Services.GitService.Push(data.localDir, data.actualBranch); err != nil {
		uc.updateProgressError(data.pd, err, fmt.Sprintf("Error pushing changes on %s", data.localDir))
		return "", err
	}
	pr, err := uc.Services.GitApiService.CreatePullRequest(data.repository, data.actualBranch, data.targetBranch, data.title, data.message)
	if err != nil {
		uc.updateProgressError(data.pd, err, fmt.Sprintf("Error creating PR on %s", data.repository))
		return "", err
	}
	if data.merge {
		err := uc.Services.GitApiService.MergePullRequest(data.repository, pr.Id)
		if err != nil {
			uc.updateProgressError(data.pd, err, fmt.Sprintf("Error merging PR on %s", data.repository))
			return "", err
		}
		return "", nil
	}
	return pr.Links.Html.Href, nil
}

func (uc *setupCiCdUseCase) getRepositoryVariables(variables []*entity.PipelineVariable) []*service.PipelineVariable {
	var v []*service.PipelineVariable
	for _, variable := range variables {
		v = append(v, &service.PipelineVariable{
			Key:    variable.Name,
			Value:  variable.Value,
			Secure: variable.Secure,
		})
	}
	return v
}

func (uc *setupCiCdUseCase) defaultManifests() ([]*entity.Manifest, error) {
	var m []*entity.Manifest
	r, err := uc.Repositories.ManifestRepository.ListDefault()
	if err != nil {
		return nil, err
	}

	for _, v := range r {
		m = append(m, v)
	}
	return m, nil
}
