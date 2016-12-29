package function

type Closer interface {
	Close()
}

type FunctionHandler interface {
	Start(Function) Closer
}

type InputMessage interface {
	Body() []byte
	DecodeBody(interface{})
}

type Callback func(...interface{})

type Function interface {
	Handle(InputMessage, Callback) error
}
