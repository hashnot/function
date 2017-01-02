package amqptypes

import (
	"net"
	"time"
)

const defaultHeartbeat = 10 * time.Second

func (d *DialConf) dialFunc(network, addr string) (net.Conn, error) {
	conn, err := net.DialTimeout(network, addr, d.ConnectionTimeout)
	if err != nil {
		return nil, err
	}

	// Heartbeating hasn't started yet, don't stall forever on a dead server.
	if err := conn.SetReadDeadline(time.Now().Add(d.ConnectionTimeout)); err != nil {
		return nil, err
	}
	return conn, nil
}
