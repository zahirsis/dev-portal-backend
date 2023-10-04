package usecase

import (
	"github.com/zahirsis/dev-portal-backend/config"
)

type CiCdDataDto struct {
	RepositoryBaseUrl string `json:"repositoryBaseUrl"`
}

type GetCiCdDataUseCase interface {
	Exec() (CiCdDataDto, error)
}

type getCiCdDataUseCase struct {
	cfg *config.Config
}

func NewGetCiCdDataUseCase(c *config.Config) GetCiCdDataUseCase {
	return &getCiCdDataUseCase{c}
}

func (uc *getCiCdDataUseCase) Exec() (CiCdDataDto, error) {
	return CiCdDataDto{
		RepositoryBaseUrl: uc.cfg.GitConfig.GetRepositoryUrl(""),
	}, nil
}
