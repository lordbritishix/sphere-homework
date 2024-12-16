package services

import (
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"sphere-homework/app/event"
)

type EventService struct {
	producer *kafka.Producer
}

func NewEventService(producer *kafka.Producer) EventService {
	return EventService{
		producer: producer,
	}
}

func (e *EventService) PublishEvent(event event.BaseEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}

	topic := TransferTopic
	err = e.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: value,
		Key:   []byte(event.Sender),
	}, nil)

	return err
}
