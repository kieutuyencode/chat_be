package result

type result struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Detail  any    `json:"detail,omitempty"`
}

type option func(*result)

func Success(message string, data any, opts ...option) *result {
	var res = &result{
		Status:  true,
		Message: message,
		Data:    data,
	}

	for _, opt := range opts {
		opt(res)
	}

	return res
}

func Fail(message string, data any, opts ...option) *result {
	var res = &result{
		Status:  false,
		Message: message,
		Data:    data,
	}

	for _, opt := range opts {
		opt(res)
	}

	return res
}

func WithDetail(detail any) option {
	return func(r *result) {
		r.Detail = detail
	}
}
