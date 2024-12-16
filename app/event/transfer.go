package event

import "github.com/google/uuid"

type Transfer struct {
	TransferId uuid.UUID `json:"transfer_id"`
	FromAsset  string    `json:"from_asset"`
	ToAsset    string    `json:"to_asset"`
	Sender     string    `json:"sender"`
	Recipient  string    `json:"recipient"`
	Amount     float64   `json:"amount"`
	Fee        float64   `json:"fee"`
	Rate       float64   `json:"rate"`
}
