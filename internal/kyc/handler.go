package kyc

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/justinush/maestro-consumer/internal/vendor"
	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/run"
	"github.com/justinush/maestro/pkg/workflow"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /kyc/start", h.handleStart)
	mux.HandleFunc("GET /kyc/{runID}", h.handleGet)
	mux.HandleFunc("GET /kyc/{runID}/events", h.handleEvents)
	mux.HandleFunc("POST /kyc/{runID}/profile", h.handleProfile)
	mux.HandleFunc("POST /kyc/{runID}/document", h.handleDocument)
	mux.HandleFunc("POST /kyc/{runID}/review", h.handleReview)
	mux.HandleFunc("POST /webhooks/vendor", h.handleVendorWebhook)
	return mux
}

func (h *Handler) handleStart(w http.ResponseWriter, r *http.Request) {
	var body StartRequest
	if err := decodeJSON(r, &body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := h.svc.Start(r.Context(), body)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.Get(r.Context(), r.PathValue("runID"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleEvents(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.GetEvents(r.Context(), r.PathValue("runID"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleProfile(w http.ResponseWriter, r *http.Request) {
	var body Profile
	if err := decodeJSON(r, &body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := h.svc.SubmitProfile(r.Context(), r.PathValue("runID"), body)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleDocument(w http.ResponseWriter, r *http.Request) {
	var body Document
	if err := decodeJSON(r, &body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := h.svc.SubmitDocument(r.Context(), r.PathValue("runID"), body)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleReview(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Approved bool `json:"approved"`
	}
	if err := decodeJSON(r, &body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !body.Approved {
		http.Error(w, "demo only supports approved=true", http.StatusBadRequest)
		return
	}
	resp, err := h.svc.SubmitReview(r.Context(), r.PathValue("runID"), true)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleVendorWebhook(w http.ResponseWriter, r *http.Request) {
	var body VendorWebhookRequest
	if err := decodeJSON(r, &body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := h.svc.HandleVendorWebhook(r.Context(), body)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("trailing json")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "encode", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(data)
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, run.ErrNotFound),
		errors.Is(err, ErrNotFound),
		errors.Is(err, vendor.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, workflow.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, ErrUnknownRoute), errors.Is(err, ErrInvalid):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, ErrWrongStep):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		if _, ok := errors.AsType[*engine.InputValidationError](err); ok {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
