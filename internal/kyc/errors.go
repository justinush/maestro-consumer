package kyc

import "errors"

var (
	ErrWrongStep = errors.New("wrong step")
	ErrNotFound  = errors.New("applicant: not found")
	ErrInvalid   = errors.New("invalid input")
)
