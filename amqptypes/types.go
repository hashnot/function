package amqptypes

import (
	"errors"
	"github.com/hashnot/function/amqp"
	q "github.com/streadway/amqp"
	"log"
	"os"
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
	Msg: &PublishingTemplate{
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
	Msg       *PublishingTemplate `yaml:"template"`
}

type PublishingTemplate struct {
	// Application or exchange specific fields,
	// the headers exchange will inspect this field.
	Headers         map[string]interface{}
	ContentType     string        `yaml:"contentType"`     // MIME content type
	ContentEncoding string        `yaml:"contentEncoding"` // MIME content encoding
	Persistent      *DeliveryMode // Transient (false) or Persistent (true)
	Priority        *uint8        // 0 to 9
	CorrelationId   string        `yaml:"correlationId"` // correlation identifier
	ReplyTo         string        `yaml:"replyTo"`       // address to to reply to (ex: RPC)
	Expiration      time.Duration // message expiration (truncated to millis)
	Type            string        // message type name
	UserId          string        `yaml:"userId"` // creating user id - ex: "guest"
	AppId           string        `yaml:"appId"`  // creating application id
}

func (o *Output) Publish(ch amqp.Channel, pub *q.Publishing) error {
	pub.AppId = os.Args[0]
	if pub.Timestamp == (time.Time{}) {
		pub.Timestamp = time.Now()
	}

	log.Print("Publish to ", o.Exchange, "/", o.Key)
	log.Printf("%#v", pub)
	return ch.Publish(o.Exchange, o.Key, o.Mandatory, o.Immediate, *pub)
}

type DeliveryMode bool

func (m DeliveryMode) ToInt() uint8 {
	if m {
		return 2
	} else {
		return 1
	}
}
