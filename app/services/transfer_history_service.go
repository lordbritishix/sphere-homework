package services

import (
	"context"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
	eventModel "sphere-homework/app/event"
	"sphere-homework/app/repository"
)

type TransferHistoryService struct {
	consumer   *kafka.Consumer
	logger     *zap.Logger
	ctx        context.Context
	repository *repository.TransferHistoryRepository
}

func NewTransferHistoryService(ctx context.Context, consumer *kafka.Consumer, logger *zap.Logger, repository *repository.TransferHistoryRepository) *TransferHistoryService {
	return &TransferHistoryService{
		consumer:   consumer,
		logger:     logger,
		ctx:        ctx,
		repository: repository,
	}
}

func (t *TransferHistoryService) Init() error {
	err := t.consumer.Subscribe(TransferTopic, nil)
	if err != nil {
		return err
	}

	// this go-routine listens to kafka for all transfer events - and writes it to the transfer_history table
	go func() {
		t.logger.Info("Starting transfer history service consumer")
		for {
			msg, err := t.consumer.ReadMessage(-1)
			if err != nil {
				t.logger.Error("Error reading message from consumer", zap.Error(err))
				continue
			}

			if msg == nil {
				continue
			}

			err = t.handleMessage(msg)
			if err != nil {
				t.logger.Error("Unable to handle the event", zap.Error(err))
				continue
			}
		}
	}()

	return nil
}

func (t *TransferHistoryService) handleMessage(msg *kafka.Message) error {
	event := eventModel.BaseEvent{}

	err := json.Unmarshal(msg.Value, &event)
	if err != nil {
		return err
	}

	err = t.repository.InsertTransferHistory(event)

	return err
}
