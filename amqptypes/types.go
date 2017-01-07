package amqptypes

import (
	"errors"
	"github.com/hashnot/function/amqp"
	q "github.com/streadway/amqp"
	"log"
	"time"
)

type Configuration struct {
	Input    *Queue
	Output   *Output
	Errors   *Output
	DialConf DialConf `yaml:"server"`
}

var defaultError = &Output{
	Key: "global.error",
	Msg: &Publishing{
		Type:        "Error",
		ContentType: "text/plain",
	},
}

func (c *Configuration) Init() {
	if c.Errors == nil {
		c.Errors = defaultError
	}
	c.DialConf.init()
}

type DialConf struct {
	Url               string
	Vhost             string
	HeartBeat         time.Duration
	ConnectionTimeout time.Duration

	Dialer amqp.Dialer
}

func (d *DialConf) init() {
	if d.Dialer == nil {
		d.Dialer = &amqp.AmqpDialer{}
	}
	if d.HeartBeat == 0 {
		d.HeartBeat = defaultHeartbeat
	}
	if d.ConnectionTimeout == 0 {
		d.ConnectionTimeout = defaultConnTimeout
	}
}

func (d DialConf) Dial() (amqp.Connection, error) {
	return d.Dialer.DialConfig(d.Url, q.Config{
		Dial:      d.dialFunc,
		Heartbeat: d.HeartBeat,
		Vhost:     d.Vhost,
	})
}

type Queue struct {
	Name      string
	Consumer  string
	AutoAck   bool `yaml:"autoAck"`
	Exclusive bool
	NoLocal   bool `yaml:"noLocal"`
	NoWait    bool `yaml:"noWait"`
	Args      q.Table
}

func (q *Queue) Consume(ch amqp.Channel) (<-chan q.Delivery, error) {
	if q.Name == "" {
		return nil, errors.New("Undefined queue")
	}

	return ch.Consume(q.Name, q.Consumer, q.AutoAck, q.Exclusive, q.NoLocal, q.NoWait, q.Args)
}

type Output struct {
	Exchange  string
	Key       string
	Mandatory bool
	Immediate bool
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

func (o *Output) Publish(ch amqp.Channel, pub *q.Publishing) error {
	log.Print("Publish to ", o.Exchange, "/", o.Key)
	return ch.Publish(o.Exchange, o.Key, o.Mandatory, o.Immediate, *pub)
}
