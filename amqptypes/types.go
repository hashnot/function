package amqptypes

import (
	"errors"
	"github.com/streadway/amqp"
	"time"
)

type Configuration struct {
	Url     string `yaml:"url"`
	Input   *Queue `yaml:"input"`
	*Output `yaml:"output"`
	Errors  *Output `yaml:"errors"`
}

var defaultError = &Output{
	Key: "global.error",
	Msg: &Publishing{
		Type:        "Error",
		ContentType: "text/plain",
	},
}

func (c *Configuration) SetupOutputs() {
	if c.Errors == nil {
		c.Errors = defaultError
	}
	if c.Output == nil {
		c.Output = defaultError
	}
}

func (c *Configuration) Dial() (*amqp.Connection, error) {
	return amqp.Dial(c.Url)
}

type Queue struct {
	Name      string     `yaml:"name"`
	Consumer  string     `yaml:"consumer"`
	AutoAck   bool       `yaml:"autoAck"`
	Exclusive bool       `yaml:"exclusive"`
	NoLocal   bool       `yaml:"noLocal"`
	NoWait    bool       `yaml:"noWait"`
	Args      amqp.Table `yaml:"args"`
}

func (q *Queue) Consume(ch *amqp.Channel) (<-chan amqp.Delivery, error) {
	if q.Name == "" {
		return nil, errors.New("Undefined queue")
	}

	return ch.Consume(q.Name, q.Consumer, q.AutoAck, q.Exclusive, q.NoLocal, q.NoWait, q.Args)
}

type Output struct {
	Exchange  string      `yaml:"exchange"`
	Key       string      `yaml:"key"`
	Mandatory bool        `yaml:"mandatory"`
	Immediate bool        `yaml:"immediate"`
	Msg       *Publishing `yaml:"publishing"`
}

type Publishing struct {
	// Application or exchange specific fields,
	// the headers exchange will inspect this field.
	//Headers         amqp.Table

	// Properties
	ContentType     string // MIME content type
	ContentEncoding string // MIME content encoding
	DeliveryMode    uint8  // Transient (0 or 1) or Persistent (2)
	//Priority        uint8     // 0 to 9
	CorrelationId string // correlation identifier
	//ReplyTo         string    // address to to reply to (ex: RPC)
	//Expiration      string    // message expiration spec
	//MessageId       string    // message identifier
	//Timestamp       time.Time // message timestamp
	Type string // message type name
	//UserId          string    // creating user id - ex: "guest"
	//AppId           string    // creating application id

	// The application specific payload of the message
	//Body            []byte `yaml:",-"`
}

func (o *Output) Publish(ch *amqp.Channel, body []byte) error {
	pub := &amqp.Publishing{
		Body:      body,
		Timestamp: time.Now(),
	}

	p := o.Msg
	if p != nil {
		pub.ContentType = p.ContentType
		pub.ContentEncoding = p.ContentEncoding
		pub.DeliveryMode = p.DeliveryMode
		pub.Type = p.Type
		pub.Body = body
	}
	return ch.Publish(o.Exchange, o.Key, o.Mandatory, o.Immediate, *pub)
}
