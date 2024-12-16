package event

import (
	"encoding/json"
	"sphere-homework/app/model"
	"time"
)

type TransferSent struct {
	Transfer
	Status     TransferEventStatus
	SentAmount float64
}

func NewTransferSent(transfer model.Transfer) (*BaseEvent, error) {
	sent := TransferSent{
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
		SentAmount: *transfer.SentAmount,
		Status:     SentTransferEventStatus,
	}

	payload, err := json.Marshal(sent)
	if err != nil {
		return nil, err
	}

	return &BaseEvent{
		Timestamp: time.Now().UnixMilli(),
		EventType: "transfer_sent",
		Sender:    transfer.Sender,
		Payload:   payload,
	}, nil
}
