package function

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/hashnot/function/amqptypes"
	"github.com/hashnot/function/amqp"
	"log"
	"runtime/debug"
	"go/types"
	q "github.com/streadway/amqp"



)

type amqpFunctionHandler struct {
	config  *amqptypes.Configuration
	channel amqp.Channel

	stop chan bool
}

func Start(f Function) (*amqpFunctionHandler, error) {
	config := new(amqptypes.Configuration)
	err := UnmarshalFile("function.yaml", config)
	if err != nil {
		return nil, err
	}

	return StartWithConfig(f, config)
}

func StartWithConfig(f Function, config *amqptypes.Configuration) (*amqpFunctionHandler, error) {
	if config.Dialer == nil {
		config.Dialer = &amqp.AmqpDialer{}
	}

	handler := &amqpFunctionHandler{config: config}
	err := handler.init()
	if err != nil {
		return nil, err
	}

	return handler, handler.readInput(f)
}

func (h *amqpFunctionHandler) Stop() {
	h.stop <- false

	close(h.stop)
}

func (h *amqpFunctionHandler) init() error {
	h.config.SetupOutputs()

	h.stop = make(chan bool, 1)

	conn, err := h.config.Dial()
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	h.channel = ch

	return err
}

func (h *amqpFunctionHandler) readInput(f Function) error {
	config := h.config
	ch := h.channel
	msgs, err := config.Input.Consume(ch)
	if err != nil {
		return err
	}

	go h.readInputLoop(f, msgs)

	return nil
}

func (h *amqpFunctionHandler) readInputLoop(f Function, msgs <-chan q.Delivery) {
	for {
		select {
		case <-h.stop:
			log.Print("Received stop signal, ending loop")
			return
		case d := <-msgs:
			// TODO recover from panic
			log.Print("input body: ", string(d.Body))

			newInvocation(h, &d).handle(f)
		}
	}
}

func newInvocation(h *amqpFunctionHandler, d *q.Delivery) *invocation {
	return &invocation{
		handler:  h,
		results:  make([]*Message, 0, 1),
		errors:   make([]error, 0, 1),
		delivery: d,
	}
}

type invocation struct {
	handler  *amqpFunctionHandler
	results  []*Message
	errors   []error
	delivery *q.Delivery
}

func (i *invocation) Error() string {
	result := new(bytes.Buffer)
	for _, err := range i.errors {
		result.Write([]byte(err.Error()))
		result.WriteRune('\n')
	}
	return result.String()
}

func (i *invocation) handle(f Function) {
	err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Print(r)
				err = errors.New(r.(error).Error() + "\n" + string(debug.Stack()))
			}
		}()
		i.results, err = f.Handle(newMessage(i.delivery))
		return
	}()

	if err != nil {
		i.errors = append(i.errors, err)
	}

	if len(i.errors) == 0 {
		for _, msg := range i.results {
			output := i.handler.config.Output
			err = output.Publish(i.handler.channel, msg.toPublishing())
			if err != nil {
				i.errors = append(i.errors, err)
				break
			}
		}

	}

	if len(i.errors) == 0 {
		log.Print("ack")
		err = i.delivery.Ack(false)
	} else {
		log.Print("nack")
		err = i.handler.config.Errors.Publish(i.handler.channel, &q.Publishing{Body: []byte(i.Error())})
		if err != nil {
			log.Print(err)
		}
		// TODO handle retryable errors
		err = i.delivery.Nack(false, false)
	}

	if err != nil {
		log.Print("Problem with (n)ack ", err)
	}
}

func newMessage(d *q.Delivery) *Message {
	return &Message{
		Body:            d.Body,
		ContentEncoding: d.ContentEncoding,
		ContentType:     d.ContentType,
		CorrelationId:   d.CorrelationId,
		Headers:         d.Headers,
		MessageId:       d.MessageId,
		Timestamp:       d.Timestamp,
		Type:            d.Type,
	}
}

func (m *Message) toPublishing() *q.Publishing {
	return &q.Publishing{
		Body:            m.Body,
		ContentEncoding: m.ContentEncoding,
		ContentType:     m.ContentType,
		CorrelationId:   m.CorrelationId,
		Headers:         m.Headers,
		MessageId:       m.MessageId,
		Timestamp:       m.Timestamp,
		Type:            m.Type,
	}
}

// TODO convert directly to Publishing
func convert(object interface{}) ([]byte, error) {
	switch object.(type) {
	case types.Nil:
		return []byte{}, nil
	case []byte:
		return object.([]byte), nil
	case string:
		return []byte(object.(string)), nil
	default:
		return json.Marshal(object)
	}
}
