package function

import (
	"encoding/json"
	"reflect"
	"time"
)

type Function interface {
	Handle(*Message) ([]*Message, error)
}

type Message struct {
	Body            []byte
	ContentEncoding string                 // MIME content encoding
	ContentType     string                 // MIME content type
	CorrelationId   string                 // application use - correlation identifier
	Headers         map[string]interface{} // see amqp.Table for allowed value types
	MessageId       string                 // application use - message identifier
	Timestamp       time.Time              // application use - message timestamp
	Type            string                 // application use - message type name
}

func (m *Message) DecodeBody(o interface{}) {
	if err := json.Unmarshal(m.Body, o); err != nil {
		panic(err)
	}
}

func (m *Message) EncodeBody(o interface{}) {
	buf, err := convert(o)
	if err != nil {
		panic(err)
	}

	m.Body = buf
	if m.Type == "" {
		t := reflect.TypeOf(o)
		m.Type = t.PkgPath() + "." + t.Name()
	}
}
