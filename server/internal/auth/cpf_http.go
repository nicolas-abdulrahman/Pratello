package auth

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) claimCPF(w http.ResponseWriter, r *http.Request) {
	claims, err := h.authenticate(r)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	var req struct {
		CPF string `json:"cpf"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	a, err := h.svc.ClaimCPF(r.Context(), claims.AccountID, req.CPF)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	writeAccount(w, a)
}

func (h *Handler) removeCPF(w http.ResponseWriter, r *http.Request) {
	claims, err := h.authenticate(r)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	if err = h.svc.RemoveCPF(r.Context(), claims.AccountID); err != nil {
		h.mapServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
