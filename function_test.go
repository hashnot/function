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

func (f *TestFunction) Handle(msg InputMessage, out OutChan) error {
	f.T.Log("inside")
	out <- &OutputMessage{}
	f.T.Log("response sent")
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
		Dialer: &amqptypes.TestDialer{
			Channel: &channel,
			T:       t,
		},
		Output: &amqptypes.Output{Exchange: outExchange},
	}

	t.Log("start func")

	go func() {
		err := StartWithConfig(&TestFunction{t}, cfg)
		if err != nil {
			t.Error(err)
		}
	}()

	t.Log("send")

	in <- q.Delivery{Exchange: "_IN_EXCHANGE_", Acknowledger: channel}

	t.Log("wait for result")
	result := <-out

	t.Log(result)

	if result.Exchange != outExchange {
		t.Error("Exchange ", result.Exchange)
	}

	close(in)
}
