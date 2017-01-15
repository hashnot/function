package function

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/hashnot/function/amqp"
	"github.com/hashnot/function/amqptypes"
	q "github.com/streadway/amqp"
	"go/types"
	"log"
	"reflect"
	"runtime/debug"
	"strconv"
	"time"
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
	h.config.Init()

	h.stop = make(chan bool, 1)

	conn, err := h.config.DialConf.Dial()
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
		delivery: d,
	}
}

type invocation struct {
	handler  *amqpFunctionHandler
	delivery *q.Delivery
}

func (i *invocation) handle(f Function) {
	err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				msg := r.(error).Error() + "\n" + string(debug.Stack())
				log.Print("Error in function ", msg)
				err = errors.New(msg)
			}
		}()
		err = f.Handle((*Message)(i.delivery), i)
		return
	}()

	if err != nil {
		// TODO handle retryable errors
		log.Print("nack")
		if err := i.delivery.Reject(false); err != nil {
			log.Print("Error rejecting message ", err)
		}

		p := i.NewFrom(i.handler.config.Errors.Msg)
		errBody := bytes.NewBuffer([]byte(err.Error()))
		errBody.WriteString("\nIncoming message:\n")
		DumpDeliveryMeta(i.delivery, errBody)
		p.Body = errBody.Bytes()

		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Print("Error while sending error message ", r)
					log.Print(p.Body)
				}
			}()
			i.Publish(p)
			return
		}()
	} else {
		log.Print("ack")
		err = i.delivery.Ack(false)
	}
}

func (i *invocation) Publish(v interface{}) {
	var p *q.Publishing
	switch v.(type) {
	case q.Publishing:
		log.Print("Publishing value instead of pointer")
		t := v.(q.Publishing)
		p = &t
	case *q.Publishing:
		p = v.(*q.Publishing)
	default:
		p = i.New()
		i.encodeBody(v, p)
	}
	err := i.handler.config.Output.Publish(i.handler.channel, p)
	if err != nil {
		panic(err)
	}
}

func (i *invocation) New() *q.Publishing {
	return i.NewFrom(i.handler.config.Output.Msg)
}

func (i *invocation) NewFrom(t *amqptypes.PublishingTemplate) *q.Publishing {
	d := i.delivery

	if t == nil {
		log.Print("Using empty template")
		t = &amqptypes.PublishingTemplate{}
	}

	headers := make(q.Table)

	if t.Headers != nil {
		AddAll(t.Headers, headers)
	}

	var deliveryMode uint8
	if t.Persistent != nil {
		deliveryMode = t.Persistent.ToInt()
	} else {
		deliveryMode = d.DeliveryMode
	}

	var priority uint8
	if t.Priority != nil {
		priority = *t.Priority
	} else {
		priority = d.Priority
	}

	var replyTo string
	if t.ReplyTo != "" {
		replyTo = t.ReplyTo
	} else {
		replyTo = d.ReplyTo
	}

	var expiration string
	if t.Expiration != 0 {
		expiration = strconv.FormatInt(t.Expiration.Nanoseconds()/1000000, 10) //millis
	}

	return &q.Publishing{
		AppId:           t.AppId,
		ContentEncoding: t.ContentEncoding,
		ContentType:     t.ContentType,
		DeliveryMode:    deliveryMode,
		Expiration:      expiration,
		Headers:         q.Table(headers),
		Priority:        priority,
		ReplyTo:         replyTo,
		Timestamp:       time.Now(),
		Type:            t.Type,
		UserId:          t.UserId,
	}
}

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

func (i *invocation) encodeBody(o interface{}, p *q.Publishing) {
	buf, err := convert(o)
	if err != nil {
		panic(err)
	}

	p.Body = buf
	if p.Type == "" {
		p.Type = reflect.TypeOf(o).String()
	}
}
