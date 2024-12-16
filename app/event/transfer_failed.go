package event

import (
	"encoding/json"
	"sphere-homework/app/model"
	"time"
)

type TransferFailed struct {
	Transfer
	Status        TransferEventStatus
	FailureReason string
}

func NewTransferFailed(transfer model.Transfer) (*BaseEvent, error) {
	sent := TransferFailed{
		Transfer: Transfer{
			TransferId: transfer.TransferId,
			FromAsset:  transfer.FromAsset,
			ToAsset:    transfer.ToAsset,
			Sender:     transfer.Sender,
			Recipient:  transfer.Recipient,
			Amount:     transfer.RequestedAmount,
			Fee:        transfer.Fee,
			Rate:       transfer.Rate,
		},
		Status:        FailedTransferEventStatus,
		FailureReason: *transfer.FailureReason,
	}

	payload, err := json.Marshal(sent)
	if err != nil {
		return nil, err
	}

	return &BaseEvent{
		Timestamp: time.Now().UnixMilli(),
		EventType: "transfer_failed",
		Sender:    transfer.Sender,
		Payload:   payload,
	}, nil
}
