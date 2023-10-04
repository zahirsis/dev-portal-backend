package redis

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/zahirsis/dev-portal-backend/src/domain/entity"
	"github.com/zahirsis/dev-portal-backend/src/domain/repository"
	"github.com/zahirsis/dev-portal-backend/src/pkg/logger"
)

type processRepository struct {
	logger logger.Logger
	client *redis.Client
}

func NewProcessRepository(logger logger.Logger, client *redis.Client) repository.ProcessRepository {
	return &processRepository{
		logger: logger,
		client: client,
	}
}

func (p processRepository) GetMessages(ID string) ([]entity.ProgressEntity, error) {
	ctx := context.Background()
	result, err := p.client.LRange(ctx, "process:"+ID, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	var messages []entity.ProgressEntity
	for _, row := range result {
		var message entity.Progress
		err = json.Unmarshal([]byte(row), &message)
		if err != nil {
			p.logger.Error("Error unmarshalling message: %s", err.Error())
			continue
		}
		messages = append(messages, entity.NewProgressEntity(message))
	}
	return messages, nil
}

func (p processRepository) SaveMessage(ID string, message entity.ProgressEntity) error {
	ctx := context.Background()
	progress := message.ToStruct()
	p.logger.Debug("Saving message", progress)
	jsonMessage, err := json.Marshal(progress)
	if err != nil {
		return err
	}
	err = p.client.RPush(ctx, "process:"+ID, string(jsonMessage)).Err()
	if err != nil {
		return err
	}
	return nil
}

func (p processRepository) MarkAsFinished(ID string) error {
	ctx := context.Background()
	key := "process:STATUS:" + ID
	err := p.client.Set(ctx, key, "finished", 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (p processRepository) IsFinished(ID string) (bool, error) {
	ctx := context.Background()
	key := "process:STATUS:" + ID
	status, err := p.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return status == "finished", nil
}
