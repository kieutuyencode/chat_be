package apperror

type code string

const (
	CodeNotFound     code = "not_found"
	CodeBadRequest   code = "bad_request"
	CodeUnauthorized code = "unauthorized"
	CodeForbidden    code = "forbidden"
)
