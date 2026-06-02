package kyc

import (
	"strings"

	"github.com/justinush/maestro-consumer/internal/model"
)

type ApplicantRecord = model.ApplicantRecord
type Profile = model.Profile
type Document = model.Document

// StartRequest selects which workflow to start (host routing).
type StartRequest struct {
	Entity string `json:"entity"`
	Flow   string `json:"flow"`
}

func (r StartRequest) Validate() error {
	if strings.TrimSpace(r.Entity) == "" || strings.TrimSpace(r.Flow) == "" {
		return ErrInvalid
	}
	return nil
}
