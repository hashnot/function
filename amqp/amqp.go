package amqp

import (
	q "github.com/streadway/amqp"
)

type Closer interface {
	Close() error
}

// Dialer

type Dialer interface {
	DialConfig(url string, config q.Config) (Connection, error)
}

type AmqpDialer struct{}

func (d *AmqpDialer) DialConfig(addr string, config q.Config) (Connection, error) {
	c, err := q.DialConfig(addr, config)
	return &amqpConnection{c}, err
}

// Connection

type Connection interface {
	Closer
	Channel() (Channel, error)
}

type amqpConnection struct {
	c *q.Connection
}

func (c *amqpConnection) Close() error {
	return c.c.Close()
}

func (c *amqpConnection) Channel() (Channel, error) {
	ch, err := c.c.Channel()
	return &amqpChannel{ch}, err
}

// Channel

type Channel interface {
	Closer
	Publish(exchange, key string, mandatory, immediate bool, msg q.Publishing) error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args q.Table) (<-chan q.Delivery, error)
}

type amqpChannel struct {
	c *q.Channel
}

func (c *amqpChannel) Close() error {
	return c.c.Close()
}

func (c *amqpChannel) Publish(exchange, key string, mandatory, immediate bool, msg q.Publishing) error {
	return c.c.Publish(exchange, key, mandatory, immediate, msg)
}

func (c *amqpChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args q.Table) (<-chan q.Delivery, error) {
	return c.c.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
}
