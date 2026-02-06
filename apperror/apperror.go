package apperror

import "fmt"

type AppError struct {
	Code    code
	Message string
	Data    any
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s | cause: %s", e.Code, e.Message, e.Err.Error())
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code code, message string, data any, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Data:    data,
		Err:     err,
	}
}

func NotFound(message string, data any, err error) *AppError {
	return New(CodeNotFound, message, data, err)
}

func BadRequest(message string, data any, err error) *AppError {
	return New(CodeBadRequest, message, data, err)
}

func Unauthorized(message string, data any, err error) *AppError {
	return New(CodeUnauthorized, message, data, err)
}

func Forbidden(message string, data any, err error) *AppError {
	return New(CodeForbidden, message, data, err)
}
