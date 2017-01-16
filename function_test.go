package function

import (
	"github.com/hashnot/function/amqptypes"
	q "github.com/streadway/amqp"
	//_ "net/http"
	//_ "net/http/pprof"
	at "github.com/hashnot/function/amqptypes/amqptest"
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
	out := make(chan *at.TestOutput, 1)
	in := make(chan q.Delivery, 1)

	t.Log("mk cfg")
	channel := at.TestChan{
		T:      t,
		Input:  in,
		Output: out,
	}
	outExchange := "_OUT_EXCHANGE_"
	cfg := &amqptypes.Configuration{
		Input: &amqptypes.Queue{Name: "test"},
		DialConf: amqptypes.DialConf{
			Dialer: &at.TestDialer{
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

	if result.Exchange != outExchange {
		t.Error("Exchange ", result.Exchange)
	}

	close(in)
	close(out)
}

type errorMsgHandler struct{}

func (*errorMsgHandler) Handle(m *Message, p Publisher) error {
	return NewErrorMessage(&q.Publishing{Body: []byte("Error Message")})
}

func TestErrorMessage(t *testing.T) {
	out := make(chan *at.TestOutput, 1)
	ch := &at.TestChan{
		T:      t,
		Output: out,
	}
	i := &invocation{
		handler: &amqpFunctionHandler{
			config: &amqptypes.Configuration{
				Errors: &amqptypes.Output{},
			},
			channel: ch,
		},
		delivery: &q.Delivery{Acknowledger: ch},
	}
	i.handle(&errorMsgHandler{})

	result := <-out
	if string(result.Msg.Body) != "Error Message" {
		t.Fail()
	}
}
