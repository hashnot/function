package function

import (
	"github.com/hashnot/function/amqptypes"
	q "github.com/streadway/amqp"
	//_ "net/http"
	//_ "net/http/pprof"
	"testing"
)

/*
func init() {
	go func() {
		log.Print(http.ListenAndServe("localhost:6060", nil))
	}()
}
*/

type TestFunction struct {
	*testing.T
}

func (f *TestFunction) Handle(msg *Message, p Publisher) error {
	f.T.Log("inside")
	p.Publish("result")
	return nil
}

func TestNoOutput(t *testing.T) {

	t.Log("start")
	out := make(chan *amqptypes.TestOutput, 1)
	in := make(chan q.Delivery, 1)

	t.Log("mk cfg")
	channel := amqptypes.TestChan{
		T:      t,
		Input:  in,
		Output: out,
	}
	outExchange := "_OUT_EXCHANGE_"
	cfg := &amqptypes.Configuration{
		Input: &amqptypes.Queue{Name: "test"},
		DialConf: amqptypes.DialConf{
			Dialer: &amqptypes.TestDialer{
				Channel: &channel,
				T:       t,
			},
		},
		Output: &amqptypes.Output{Exchange: outExchange},
	}

	t.Log("start func")

	handler, err := StartWithConfig(&TestFunction{t}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer handler.Stop()

	t.Log("send")

	in <- q.Delivery{Exchange: "_IN_EXCHANGE_", Acknowledger: channel}

	t.Log("wait for result")
	result := <-out

	t.Log("result: ", result)

	if result.Exchange != outExchange {
		t.Error("Exchange ", result.Exchange)
	}

	close(in)
	close(out)
}
