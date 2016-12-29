package function

import (
	"github.com/streadway/amqp"
)

type OutChan chan<- *OutputMessage

type InputMessage interface {
	Get() *amqp.Delivery
	GetPayload(interface{}) error
}

type OutputMessage struct {
	Payload interface{}
}

type Function interface {
	Handle(InputMessage, OutChan) error
}
