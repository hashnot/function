package function

type InputMessage interface {
	Body() []byte
	DecodeBody(interface{})
}

type Callback func(...interface{})

type Function interface {
	Handle(InputMessage, Callback) error
}
