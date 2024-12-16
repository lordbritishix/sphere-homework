package dto

import "github.com/google/uuid"

type TransferRequest struct {
	FromAsset string  `json:"from_asset"`
	ToAsset   string  `json:"to_asset"`
	Amount    float64 `json:"amount"`
	Sender    string  `json:"sender"`
	Recipient string  `json:"recipient"`
}

type TransferResponse struct {
	TransferId uuid.UUID `json:"transfer_id"`
}
