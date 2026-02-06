package validation

import "github.com/go-playground/validator/v10"

type config struct {
	validate *validator.Validate
}

type option func(*config)

func WithValidate(v *validator.Validate) option {
	return func(c *config) {
		c.validate = v
	}
}
