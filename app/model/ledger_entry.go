package model

import "github.com/google/uuid"

type LedgerEntryType string

const (
	FeeLedgerEntryType      LedgerEntryType = "FEE"
	TransferLedgerEntryType LedgerEntryType = "TRANSFER"
)

type LedgerEntry struct {
	TransferId uuid.UUID
	Account    string
	Asset      string
	Amount     float64
	Type       LedgerEntryType
}
