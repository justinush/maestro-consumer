package kyc

import (
	"errors"

	"github.com/justinush/maestro-consumer/internal/model"
)

var (
	ErrWrongStep    = errors.New("wrong step")
	ErrUnknownRoute = errors.New("unknown route")

	ErrNotFound = model.ErrNotFound
	ErrInvalid  = model.ErrInvalid
)
