package function

import (
	"fmt"
	"github.com/streadway/amqp"
	"io"
	"time"
)

func DumpProperties(d *amqp.Delivery) map[string]interface{} {
	m := make(map[string]interface{})

	addStrIfNotEmpty("appId", d.AppId, m)
	addStrIfNotEmpty("consumerTag", d.ConsumerTag, m)
	addStrIfNotEmpty("contentEncoding", d.ContentEncoding, m)
	addStrIfNotEmpty("contentType", d.ContentType, m)
	addStrIfNotEmpty("correlationId", d.CorrelationId, m)
	if d.DeliveryMode != 0 {
		m["deliveryMode"] = d.DeliveryMode
	}
	if d.DeliveryTag != 0 {
		m["deliveryTag"] = d.DeliveryTag
	}
	if d.MessageCount != 0 {
		m["messageCount"] = d.MessageCount
	}
	addStrIfNotEmpty("exchange", d.Exchange, m)
	addStrIfNotEmpty("expiration", d.Expiration, m)
	addStrIfNotEmpty("messageId", d.MessageId, m)
	if d.Priority != 0 {
		m["priority"] = d.Priority
	}
	if d.Redelivered {
		m["redelivered"] = d.Redelivered
	}
	addStrIfNotEmpty("replyTo", d.ReplyTo, m)
	addStrIfNotEmpty("routingKey", d.RoutingKey, m)
	if (d.Timestamp != time.Time{}) {
		m["timestamp"] = d.Timestamp
	}
	addStrIfNotEmpty("type", d.Type, m)
	addStrIfNotEmpty("userId", d.UserId, m)

	return m
}

func DumpPublishingProperties(p *amqp.Publishing) map[string]interface{} {
	m := make(map[string]interface{})

	addStrIfNotEmpty("appId", p.AppId, m)
	addStrIfNotEmpty("contentEncoding", p.ContentEncoding, m)
	addStrIfNotEmpty("contentType", p.ContentType, m)
	addStrIfNotEmpty("correlationId", p.CorrelationId, m)
	if p.DeliveryMode != 0 {
		m["deliveryMode"] = p.DeliveryMode
	}
	addStrIfNotEmpty("expiration", p.Expiration, m)
	addStrIfNotEmpty("messageId", p.MessageId, m)
	if p.Priority != 0 {
		m["priority"] = p.Priority
	}

	addStrIfNotEmpty("replyTo", p.ReplyTo, m)
	if (p.Timestamp != time.Time{}) {
		m["timestamp"] = p.Timestamp
	}
	addStrIfNotEmpty("type", p.Type, m)
	addStrIfNotEmpty("userId", p.UserId, m)

	return m
}

func addStrIfNotEmpty(key, value string, dst map[string]interface{}) {
	if value != "" {
		dst[key] = value
	}
}

func DumpDeliveryMeta(d *amqp.Delivery, w io.Writer) {
	props := DumpProperties(d)
	if len(props) != 0 {
		fmt.Fprintln(w, "Properties:")
		printMap(props, w)
	}
	if len(d.Headers) != 0 {
		fmt.Fprintln(w, "Headers:")
		printMap(d.Headers, w)
	}
}

func DumpPublishingMeta(d *amqp.Publishing, w io.Writer) {
	props := DumpPublishingProperties(d)
	if len(props) != 0 {
		fmt.Fprintln(w, "Properties:")
		printMap(props, w)
	}
	if len(d.Headers) != 0 {
		fmt.Fprintln(w, "Headers:")
		printMap(d.Headers, w)
	}
}

func printMap(m map[string]interface{}, w io.Writer) error {
	for k, v := range m {
		_, err := fmt.Fprintf(w, "%v = %v\n", k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
