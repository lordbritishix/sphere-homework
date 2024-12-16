package event

import (
	"encoding/json"
	"github.com/google/uuid"
	"sphere-homework/app/dto"
	"time"
)

type TransferCreated struct {
	Transfer
	Status TransferEventStatus
}

func NewTransferCreated(request dto.TransferRequest, fee float64, rate float64, transferId uuid.UUID) (*BaseEvent, error) {
	created := TransferCreated{
		Transfer: Transfer{
			TransferId: transferId,
			FromAsset:  request.FromAsset,
			ToAsset:    request.ToAsset,
			Sender:     request.Sender,
			Recipient:  request.Recipient,
			Amount:     request.Amount,
			Fee:        fee,
			Rate:       rate,
		},
		Status: CreatedTransferEventStatus,
	}

	payload, err := json.Marshal(created)
	if err != nil {
		return nil, err
	}

	return &BaseEvent{
		Timestamp: time.Now().UnixMilli(),
		EventType: "transfer_created",
		Sender:    request.Sender,
		Payload:   payload,
	}, nil
}
