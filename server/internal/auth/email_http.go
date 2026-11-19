package auth

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) setEmail(w http.ResponseWriter, r *http.Request) {
	claims, err := h.authenticate(r)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	var req struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		CurrentPassword string `json:"current_password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	a, err := h.svc.SetEmail(r.Context(), claims.AccountID, req.Email, req.Password, req.CurrentPassword)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	writeAccount(w, a)
}
func (h *Handler) removeEmail(w http.ResponseWriter, r *http.Request) {
	claims, err := h.authenticate(r)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	if err = h.svc.RemoveEmail(r.Context(), claims.AccountID); err != nil {
		h.mapServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) verifyEmail(w http.ResponseWriter, r *http.Request) {
	claims, err := h.authenticate(r)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	a, err := h.svc.Account(r.Context(), claims.AccountID)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	if !a.EmailVerified {
		h.mapServiceError(w, ErrValidation)
		return
	}
	writeAccount(w, a)
}
