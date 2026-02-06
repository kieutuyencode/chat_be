package logger

type Message string

const (
	MessageRequestStarted   Message = "Request started"
	MessageRequestCompleted Message = "Request completed"
	MessageRequestFailed    Message = "Request failed"
	MessageHandlerFailed    Message = "Handler failed"
)
