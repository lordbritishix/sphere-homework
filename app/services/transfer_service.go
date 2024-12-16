package services

import (
	"context"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"sphere-homework/app/config"
	eventModel "sphere-homework/app/event"
	"sphere-homework/app/model"
	"sphere-homework/app/repository"
	"time"
)

// TransferDelay provides a map of artificial transfer delay (sec) by destination asset
var TransferDelay = map[string]time.Duration{
	"USD": time.Duration(3) * time.Second,
	"EUR": time.Duration(2) * time.Second,
	"JPY": time.Duration(3) * time.Second,
	"GBP": time.Duration(2) * time.Second,
	"AUD": time.Duration(3) * time.Second,
}

// TransferService is responsible for:
// 1. Listening to the event bus for transfer created events
// 2. Creating a transfer order on the outbox table
// 3. Fulfilling the orders placed on the outbox table
type TransferService struct {
	consumer           *kafka.Consumer
	logger             *zap.Logger
	ctx                context.Context
	transferRepository *repository.TransferRepository
	ledgerRepository   *repository.LedgerRepository
	config             config.Config
	eventService       *EventService
}

func NewTransferService(consumer *kafka.Consumer, logger *zap.Logger, transferRepository *repository.TransferRepository,
	ledgerRepository *repository.LedgerRepository, eventService *EventService, ctx context.Context, config config.Config) *TransferService {

	return &TransferService{
		consumer:           consumer,
		logger:             logger,
		transferRepository: transferRepository,
		ledgerRepository:   ledgerRepository,
		ctx:                ctx,
		config:             config,
		eventService:       eventService,
	}
}

func (t *TransferService) Init() error {
	err := t.consumer.Subscribe(TransferTopic, nil)
	if err != nil {
		return err
	}

	// this go-routine listens to kafka for transfer created events - and writes it to the outbox
	go func() {
		t.logger.Info("Starting transfer service consumer")
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

	// this go-routine polls the outbox for unsent transfers, and sends them -
	// to do - maybe schedule this as a cron in a more reliable task scheduler like asynq
	go func() {
		t.logger.Info("Starting transfer outbox processor")

		ticker := time.NewTicker(time.Duration(t.config.TransferOutboxPollFrequencySec) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-t.ctx.Done():
				t.logger.Info("Shutting down transfer outbox processor")
				return
			case <-ticker.C:
				transfers, err := t.transferRepository.GetUnsentTransfers(250)

				if err != nil {
					t.logger.Error("Unable to get unsent transfers", zap.Error(err))
					continue
				}

				// TODO: this transfer queue processor can probably be parallelized by some sort of partition - like destination asset
				for _, transfer := range transfers {
					err = t.processTransfer(transfer)
					if err != nil {
						t.logger.Error("Unable to process transfer", zap.Any("transfer", transfer), zap.Error(err))
						continue
					}
				}
			}
		}
	}()

	return nil
}

func applyArtificialDelay(asset string) {
	delay, ok := TransferDelay[asset]
	if !ok {
		delay = time.Duration(0) * time.Second
	}

	time.Sleep(delay)
}

func (t *TransferService) processTransfer(transfer model.Transfer) error {
	logger := t.logger.With(
		zap.String("id", transfer.TransferId.String()),
		zap.String("from_asset", transfer.FromAsset),
		zap.String("to_asset", transfer.ToAsset))

	logger.Info("Processing transfer", zap.Any("transfer", transfer))

	applyArtificialDelay(transfer.ToAsset)

	// Lock transfer so it cannot be picked up by another transfer processor in case we have multiple instances of it running
	lockedTransfer, err := t.transferRepository.LockTransfer(transfer.TransferId)
	if err != nil {
		logger.Error("Unable to lock transfer", zap.Error(err))
		return err
	}

	defer func() {
		var event *eventModel.BaseEvent
		var errPub error

		if err != nil {
			lockedTransfer.TransferStatus = model.FailedTransferStatus
			reason := err.Error()
			lockedTransfer.FailureReason = &reason

			logger.Info("Failed to process transfer", zap.Any("transfer", lockedTransfer))

			event, errPub = eventModel.NewTransferFailed(*lockedTransfer)
			if errPub != nil {
				logger.Error("Unable to create transfer failed event", zap.Error(err))
			}
		} else {
			now := time.Now().UTC()
			lockedTransfer.TransferStatus = model.SentTransferStatus
			lockedTransfer.SentAt = &now

			logger.Info("Successfully processed transfer", zap.Any("transfer", lockedTransfer))

			event, errPub = eventModel.NewTransferSent(*lockedTransfer)
			if errPub != nil {
				logger.Error("Unable to create transfer sent event", zap.Error(err))
			}
		}

		err = t.eventService.PublishEvent(*event)
		if err != nil {
			logger.Error("Unable to publish transfer sent event", zap.Error(err))
		}

		_, err = t.transferRepository.UnlockAndUpdateTransfer(*lockedTransfer)
		if err != nil {
			logger.Error("Unable to unlock transfer", zap.Error(err))
			return
		}

	}()

	err = t.ledgerRepository.InsertNewEntryIfNotExists(lockedTransfer.FromAsset, lockedTransfer.Sender)
	if err != nil {
		return err
	}

	err = t.ledgerRepository.InsertNewEntryIfNotExists(lockedTransfer.ToAsset, lockedTransfer.Recipient)
	if err != nil {
		return err
	}

	err = t.ledgerRepository.Transfer(lockedTransfer)
	if err != nil {
		logger.Error("Unable to perform transfer to ledger", zap.Error(err))
		return err
	}

	return nil
}

func (t *TransferService) handleMessage(msg *kafka.Message) error {
	event := eventModel.BaseEvent{}

	err := json.Unmarshal(msg.Value, &event)
	if err != nil {
		return err
	}

	// ignore non transfer_created events
	if event.EventType != "transfer_created" {
		return nil
	}

	transferCreatedEvent := eventModel.TransferCreated{}

	err = json.Unmarshal(event.Payload, &transferCreatedEvent)
	if err != nil {
		return err
	}

	t.logger.Info("Received TransferCreated event",
		zap.String("key", string(msg.Key)),
		zap.String("message", string(msg.Value)),
		zap.Int32("partition", msg.TopicPartition.Partition),
		zap.Any("transfer", transferCreatedEvent),
	)

	var transferType model.TransferType

	// transfers like a re-balance is an internal transfer
	if transferCreatedEvent.Sender == repository.SystemAccount && transferCreatedEvent.Recipient == repository.SystemAccount {
		transferType = model.InternalTransferType
	} else {
		transferType = model.ExternalTransferType
	}

	fee := transferCreatedEvent.Fee * transferCreatedEvent.Amount

	err = t.transferRepository.InsertOutgoingTransfer(model.Transfer{
		TransferId:      uuid.New(),
		CreatedAt:       time.Now().UTC(),
		FromAsset:       transferCreatedEvent.FromAsset,
		ToAsset:         transferCreatedEvent.ToAsset,
		RequestedAmount: transferCreatedEvent.Amount,
		Fee:             fee,
		Rate:            transferCreatedEvent.Rate,
		Sender:          transferCreatedEvent.Sender,
		Recipient:       transferCreatedEvent.Recipient,
		TransferType:    transferType,
	})

	if err != nil {
		t.logger.Error("Unable to insert outgoing transfer", zap.Error(err))
		return err
	}

	return nil
}
