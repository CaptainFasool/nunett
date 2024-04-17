package api

import (
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ProblemDetail struct {
	Type     string        `json:"type,omitempty" validate:"uri"`
	Status   int           `json:"status,omitempty"`
	Title    string        `json:"title,omitempty"`
	Detail   string        `json:"detail,omitempty"`
	Instance string        `json:"instance,omitempty" validate:"uri"`
	Errors   []ErrorDetail `json:"errors,omitempty"`
}

type ErrorDetail struct {
	Detail  string `json:"detail"`
	Pointer string `json:"pointer"`
}

type ProblemOption func(*ProblemDetail)

func NewProblemDetail(options ...ProblemOption) ProblemDetail {
	problem := ProblemDetail{}
	for _, option := range options {
		option(&problem)
	}
	return problem
}

func WithType(t string) ProblemOption {
	return func(p *ProblemDetail) {
		p.Type = t
	}
}

func WithStatus(s int) ProblemOption {
	return func(p *ProblemDetail) {
		p.Status = s
	}
}

func WithTitle(t string) ProblemOption {
	return func(p *ProblemDetail) {
		p.Title = t
	}
}

func WithDetail(d string) ProblemOption {
	return func(p *ProblemDetail) {
		p.Detail = d
	}
}

func WithInstance(i string) ProblemOption {
	return func(p *ProblemDetail) {
		p.Instance = i
	}
}

func WithErrors(e []ErrorDetail) ProblemOption {
	return func(p *ProblemDetail) {
		p.Errors = e
	}
}

func NewValidationProblem(e error) ProblemDetail {
	return NewProblemDetail(
		WithStatus(http.StatusBadRequest),
		WithTitle("Input Validation Error"),
		WithDetail("You request body have invalid parameters."),
		WithErrors(readableErrors(e)),
	)
}

func NewEmptyBodyProblem() ProblemDetail {
	return NewProblemDetail(
		WithStatus(http.StatusBadRequest),
		WithTitle("Empty Request Body"),
		WithDetail("Your request did not include a body."),
	)
}

// TODO: Update readableErrors to accept a ut.Translator parameter for human-readable messages
// (github.com/go-playground/universal-translator)
func readableErrors(err error) []ErrorDetail {
	var errors []ErrorDetail
	errs, ok := err.(validator.ValidationErrors)
	if ok {
		for _, e := range errs {
			var detail, pointer string
			switch e.Tag() {
			case "required":
				detail = "is required"
			case "exists":
				detail = "is missing"
			default:
				detail = "is invalid"
			}
			pointer = "#/" + strings.ToLower(e.Field())
			errors = append(errors, ErrorDetail{Detail: detail, Pointer: pointer})
		}
	}
	return errors
}

func exists(fl validator.FieldLevel) bool {
	return !fl.Field().IsNil()
}
