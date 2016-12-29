package function

import (
	"encoding/json"
	"github.com/hashnot/function/amqp"
	"github.com/hashnot/function/amqptypes"
	q "github.com/streadway/amqp"
	"log"
)

func Start(f Function) error {
	config := new(amqptypes.Configuration)
	err := UnmarshalFile("function.yaml", config)
	if err != nil {
		return err
	}

	return StartWithConfig(f, config)
}

func StartWithConfig(f Function, config *amqptypes.Configuration) error {
	if config.Dialer == nil {
		config.Dialer = &amqp.AmqpDialer{}
	}

	config.SetupOutputs()

	conn, err := config.Dial()
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	outChan, err := SetupOutput(config, ch)
	if err != nil {
		return err
	}

	return ReadInput(config, ch, f, outChan)
}

type OutChan chan<- *OutputMessage

func SetupOutput(c *amqptypes.Configuration, ch amqp.Channel) (OutChan, error) {
	outChan := make(chan *OutputMessage, 8)

	go func() {
		for d := range outChan {
			buf, err := json.Marshal(d.Payload)
			if err != nil {
				log.Print(err)
				c.Errors.Publish(ch, []byte(err.Error()))
			}

			err = c.Output.Publish(ch, buf)
			if err != nil {
				log.Print(err)
			}
		}
	}()

	return outChan, nil
}

func ReadInput(config *amqptypes.Configuration, ch amqp.Channel, f Function, outChan OutChan) error {
	msgs, err := config.Input.Consume(ch)
	if err != nil {
		return err
	}

	for d := range msgs {
		// TODO recover from panic
		log.Print("input body: ", string(d.Body))
		//spew.Fdump(os.Stderr, client)

		msg := &Message{d}
		err = f.Handle(msg, outChan)
		if err != nil {
			// TODO handle retryable errors
			err2 := d.Nack(false, false)
			if err2 != nil {
				log.Fatal(err2)
			}
			config.Errors.Publish(ch, []byte(err.Error()))
		} else {
			err2 := d.Ack(false)
			if err2 != nil {
				log.Fatal(err2)
			}
		}
	}
	return nil
}

type Message struct {
	delivery q.Delivery
}

func (m *Message) Get() *q.Delivery {
	return &m.delivery
}

func (m *Message) GetPayload(t interface{}) error {
	return json.Unmarshal(m.delivery.Body, t)
}

type InputMessage interface {
	Get() *q.Delivery
	GetPayload(interface{}) error
}

type OutputMessage struct {
	Payload interface{}
}

type Function interface {
	Handle(InputMessage, OutChan) error
}
