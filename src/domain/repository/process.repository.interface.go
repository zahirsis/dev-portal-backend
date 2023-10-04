package repository

import (
	"github.com/zahirsis/dev-portal-backend/src/domain/entity"
)

type ProcessRepository interface {
	GetMessages(ID string) ([]entity.ProgressEntity, error)
	SaveMessage(ID string, message entity.ProgressEntity) error
	MarkAsFinished(ID string) error
	IsFinished(ID string) (bool, error)
}
