package function

import (
	"bytes"
	"encoding/json"
	"github.com/hashnot/function/amqptypes"
	"github.com/streadway/amqp"
	"log"
)

type amqpFunctionHandler struct {
	config  *amqptypes.Configuration
	channel *amqp.Channel

	stop    chan bool
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
	handler := &amqpFunctionHandler{config:config}
	err := handler.init()
	if err != nil {
		return nil, err
	}

	return handler, handler.readInput(f)
}

func (h*amqpFunctionHandler)Stop() {
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

func (h*amqpFunctionHandler) readInputLoop(f Function, msgs <-chan amqp.Delivery) {
	for {
		select {
		case <-h.stop:
			log.Print("Received stop signal, ending loop")
			return
		case d := <-msgs:
		// TODO recover from panic
			log.Print("input body: ", string(d.Body))

			msg := &Message{d}
			i := &invocation{
				handler: h,
				results: make([][]byte, 0, 1),
				errors:  make([]error, 0, 1),
			}
			i.handle(f, msg)
		}
	}
}

type invocation struct {
	handler *amqpFunctionHandler
	results [][]byte
	errors  []error
}

func (i *invocation) Error() string {
	result := new(bytes.Buffer)
	for _, err := range i.errors {
		result.Write([]byte(err.Error()))
		result.WriteRune('\n')
	}
	return result.String()
}

func (i *invocation) handle(f Function, msg *Message) {
	err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Print(r)
				err = r.(error)
			}
		}()
		err = f.Handle(msg, i.functionResultCallback)
		return
	}()

	if err != nil {
		i.errors = append(i.errors, err)
	}

	if len(i.errors) == 0 {
		for _, buf := range i.results {
			output := i.handler.config.Output
			err = output.Publish(i.handler.channel, buf)
			if err != nil {
				i.errors = append(i.errors, err)
				break
			}
		}

	}

	if len(i.errors) == 0 {
		log.Print("ack")
		err = msg.delivery.Ack(false)
	} else {
		log.Print("nack")
		err = i.handler.config.Errors.Publish(i.handler.channel, []byte(i.Error()))
		if err != nil {
			log.Print(err)
		}
		// TODO handle retryable errors
		err = msg.delivery.Nack(false, false)
	}

	if err != nil {
		log.Print("Problem with (n)ack ", err)
	}

}

func (i *invocation) functionResultCallback(data ...interface{}) {
	for _, o := range data {
		buf, err := convert(o)

		if err == nil {
			i.results = append(i.results, buf)
		} else {
			i.errors = append(i.errors, err)
		}
	}
}

// TODO convert directly to Publishing
func convert(object interface{}) ([]byte, error) {
	return json.Marshal(object)
}

type Message struct {
	delivery amqp.Delivery
}

func (m *Message) Body() []byte {
	return m.delivery.Body
}

func (m *Message) DecodeBody(t interface{}) {
	if err := json.Unmarshal(m.delivery.Body, t); err != nil {
		panic(err)
	}
}
