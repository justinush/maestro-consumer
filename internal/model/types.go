package model

import "strings"

type ApplicantRecord struct {
	ApplicantID string
	RunID       string
	Profile     Profile
	Documents   []Document
}

type Profile struct {
	FullName string `json:"fullName"`
	Email    string `json:"email"`
}

func (p Profile) Validate() error {
	if strings.TrimSpace(p.FullName) == "" {
		return ErrInvalid
	}
	if strings.TrimSpace(p.Email) == "" {
		return ErrInvalid
	}
	return nil
}

type Document struct {
	Type string `json:"documentType"`
	Ref  string `json:"documentRef"`
}

func (d Document) Validate() error {
	if strings.TrimSpace(d.Type) == "" || strings.TrimSpace(d.Ref) == "" {
		return ErrInvalid
	}
	return nil
}
