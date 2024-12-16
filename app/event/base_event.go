package event

type TransferEventStatus string

const (
	CreatedTransferEventStatus TransferEventStatus = "created"
	SentTransferEventStatus    TransferEventStatus = "sent"
	FailedTransferEventStatus  TransferEventStatus = "failed"
)

type BaseEvent struct {
	Timestamp int64
	EventType string
	Sender    string
	Payload   []byte
}
