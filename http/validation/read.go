package validation

type read string

const (
	ReadQuery   read = "query"
	ReadParams  read = "params"
	ReadHeaders read = "headers"
	ReadBody    read = "body"
)
