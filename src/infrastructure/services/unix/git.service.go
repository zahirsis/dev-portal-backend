package unix

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/zahirsis/dev-portal-backend/config"
	"github.com/zahirsis/dev-portal-backend/src/domain/service"
	"github.com/zahirsis/dev-portal-backend/src/pkg/logger"
	"os/exec"
	"strings"
)

type gitService struct {
	cfg    *config.GitConfig
	logger logger.Logger
}

func NewGitService(cfg *config.GitConfig, logger logger.Logger) service.GitService {
	configGit(logger)
	return &gitService{
		cfg:    cfg,
		logger: logger,
	}
}

func (g *gitService) CloneRepository(repository string, branch string, path string) error {
	url := g.cfg.GetRemoteUrl(repository)
	cmd := exec.Command("git", "clone", "-b", branch, url, path)
	return g.execCommand(cmd, fmt.Sprintf("cloning %s repository", url))
}

func (g *gitService) Checkout(path string, branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = path
	return g.execCommand(cmd, fmt.Sprintf("checking out to %s branch", branch))
}

func (g *gitService) Branch(path string, branch string) error {
	cmd := exec.Command("git", "checkout", "-b", branch)
	cmd.Dir = path
	return g.execCommand(cmd, fmt.Sprintf("creating %s branch", branch))
}

func (g *gitService) Commit(path string, message string) error {
	hasChanges, err := g.HasChanges(path)
	if err != nil {
		return err
	}
	if !hasChanges {
		g.logger.Debug("No changes to commit, skipping...")
		return nil
	}
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = path
	err = g.execCommand(cmd, fmt.Sprintf("adding files to stage"))
	if err != nil {
		return err
	}
	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = path
	return g.execCommand(cmd, fmt.Sprintf("commiting files to git"))
}

func (g *gitService) Push(path string, branch string) error {
	cmd := exec.Command("git", "push", "-u", "origin", branch)
	cmd.Dir = path
	return g.execCommand(cmd, fmt.Sprintf("pushing changes to %s branch", branch))
}

func (g *gitService) Pull(path string, branch string) error {
	cmd := exec.Command("git", "pull", "origin", branch)
	cmd.Dir = path
	return g.execCommand(cmd, fmt.Sprintf("pulling changes from %s branch", branch))
}

func (g *gitService) execCommand(cmd *exec.Cmd, action string) error {
	g.logger.Debug(action)
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		g.logger.Error(fmt.Sprintf("Error %s", action), err.Error())
	}
	scanner := bufio.NewScanner(stderr)
	var errMessage string
	for scanner.Scan() {
		errMessage += scanner.Text() + "\n"
	}
	err := cmd.Wait()
	if err != nil {
		g.logger.Error(action, errMessage, err.Error())
		return errors.New(fmt.Sprintf("%s: \n%s\n%s", action, errMessage, err.Error()))
	}
	return nil
}

func (g *gitService) HasChanges(path string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		g.logger.Error("Error getting git status", err.Error())
		return false, err
	}
	statusOutput := strings.TrimSpace(string(output))
	return statusOutput != "", nil
}

func configGit(logger logger.Logger) {
	cmd := exec.Command("git", "version")
	err := cmd.Run()
	if err != nil {
		logger.Error("Error getting git version", err.Error())
		panic(err)
	}
	cmd = exec.Command("git", "config", "--global", "user.email", "devportal@tempo.com.vc")
	err = cmd.Run()
	if err != nil {
		logger.Error("Error setting git user email", err.Error())
		panic(err)
	}
	cmd = exec.Command("git", "config", "--global", "user.name", "devportal")
	err = cmd.Run()
	if err != nil {
		logger.Error("Error setting git user name", err.Error())
		panic(err)
	}
	cmd = exec.Command("git", "config", "--global", "pull.rebase", "true")
	err = cmd.Run()
	if err != nil {
		logger.Error("Error setting git pull rebase", err.Error())
		panic(err)
	}
}
