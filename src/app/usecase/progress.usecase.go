package usecase

import (
	"github.com/zahirsis/dev-portal-backend/src/domain/entity"
	"github.com/zahirsis/dev-portal-backend/src/infrastructure/container"
)

type ProgressUseCase interface {
	Exec(ID string) ([]entity.Progress, error)
	IsFinished(ID string) bool
}

type progressUseCase struct {
	*container.Container
}

func NewProgressUseCase(c *container.Container) ProgressUseCase {
	return &progressUseCase{c}
}

func (uc *progressUseCase) Exec(ID string) ([]entity.Progress, error) {
	messages, err := uc.Repositories.ProgressRepository.GetMessages(ID)
	if err != nil {
		return nil, err
	}
	var progress []entity.Progress
	for _, message := range messages {
		progress = append(progress, message.ToStruct())
	}
	return progress, nil
}

func (uc *progressUseCase) IsFinished(ID string) bool {
	f, err := uc.Repositories.ProgressRepository.IsFinished(ID)
	uc.Logger.Debug("IsFinished", f, err)
	if err != nil {
		uc.Logger.Error("Error checking if process is finish: %s", err.Error())
		return false
	}
	return f
}
