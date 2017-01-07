package amqptypes

import (
	"github.com/hashnot/function/amqp"
	q "github.com/streadway/amqp"
	"testing"
)

type TestDialer struct {
	*testing.T
	Channel *TestChan
}

func (d *TestDialer) DialConfig(string, q.Config) (amqp.Connection, error) {
	d.T.Log("Dial")
	return &TestConn{
		TestDialer: d,
	}, nil
}

type TestBaseStruct struct {
}

func (*TestBaseStruct) Close() error {
	return nil
}

type TestConn struct {
	TestBaseStruct
	*TestDialer
}

func (c *TestConn) Channel() (amqp.Channel, error) {
	c.T.Log("Channel()")
	return c.TestDialer.Channel, nil
}

type TestChan struct {
	TestBaseStruct
	*testing.T
	Output chan *TestOutput
	Input  chan q.Delivery
}

type TestOutput struct {
	Exchange  string
	Key       string
	Mandatory bool
	Immediate bool
	Msg       q.Publishing
}

func (c *TestChan) Publish(exchange, key string, mandatory, immediate bool, msg q.Publishing) error {
	c.T.Log("publish to ", c.Output)
	c.Output <- &TestOutput{
		Exchange:  exchange,
		Key:       key,
		Mandatory: mandatory,
		Immediate: immediate,
		Msg:       msg,
	}
	return nil
}

func (c *TestChan) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args q.Table) (<-chan q.Delivery, error) {
	c.T.Log("[test] Consume ", queue)
	return c.Input, nil
}

func (c TestChan) Ack(tag uint64, multiple bool) error {
	return nil
}
func (c TestChan) Nack(tag uint64, multiple bool, requeue bool) error {
	return nil
}
func (c TestChan) Reject(tag uint64, requeue bool) error {
	return nil
}
