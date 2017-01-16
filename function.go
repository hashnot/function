package function

import (
	"encoding/json"
	"github.com/hashnot/function/amqptypes"
	"github.com/streadway/amqp"
)

type Function interface {
	Handle(*Message, Publisher) error
}

type Message amqp.Delivery

func (m *Message) DecodeBody(o interface{}) {
	if err := json.Unmarshal(m.Body, o); err != nil {
		panic(err)
	}
}

type Publisher interface {
	// Create Publishing object with fields preset from the output template
	New() *amqp.Publishing

	NewFrom(*amqptypes.PublishingTemplate) *amqp.Publishing

	// Accept Publishing or any object that can be marshalled to be used as body.
	// Should probably panic on any error, the panic will be recovered in handler function
	Publish(object interface{})

	PublishTo(*amqptypes.Output, interface{})
}

func AddAll(from, to map[string]interface{}) {
	for k, v := range from {
		to[k] = v
	}
}

type ErrorMessage struct {
	p *amqp.Publishing
}

func (e ErrorMessage) Error() string {
	return string(e.p.Body)
}

func NewErrorMessage(p *amqp.Publishing) error {
	return &ErrorMessage{p}
}
