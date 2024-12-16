package model

import (
	"github.com/google/uuid"
	"time"
)

type TransferStatus string
type TransferType string

const (
	UnsentTransferStatus    TransferStatus = "UNSENT"
	SentTransferStatus      TransferStatus = "SENT"
	CompletedTransferStatus TransferStatus = "COMPLETED"
	FailedTransferStatus    TransferStatus = "FAILED"
	CancelledTransferStatus TransferStatus = "CANCELLED"
)

const (
	InternalTransferType TransferType = "INTERNAL"
	ExternalTransferType TransferType = "EXTERNAL"
)

type Transfer struct {
	TransferId      uuid.UUID
	CreatedAt       time.Time
	SentAt          *time.Time
	FromAsset       string
	ToAsset         string
	RequestedAmount float64
	NetAmount       float64  // amount less fees
	SentAmount      *float64 // amount sent to the user, which is the net amount multiplied by the rate
	Fee             float64  // fee charged, in the same currency as the RequestedAmount
	Rate            float64
	Sender          string
	Recipient       string
	TransferStatus  TransferStatus
	FailureReason   *string
	TransferType    TransferType
	LockId          *uuid.UUID
}
