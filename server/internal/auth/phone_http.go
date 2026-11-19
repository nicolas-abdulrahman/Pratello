package auth

import (
	"encoding/json"
	"net/http"
)

type phoneRequest struct {
	Phone string `json:"phone"`
}

func (h *Handler) requestPhoneRegistration(w http.ResponseWriter, r *http.Request) {
	phone, ok := decodePhone(w, r)
	if !ok {
		h.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := h.svc.RequestPhoneRegistration(r.Context(), phone); err != nil {
		h.mapServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) registerPhone(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone string `json:"phone"`
		Name  string `json:"name"`
		CPF   string `json:"cpf"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil || req.Phone == "" {
		h.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	pair, err := h.svc.RegisterPhone(r.Context(), req.Phone, req.Name, req.CPF)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(pair)
}

func decodePhone(w http.ResponseWriter, r *http.Request) (string, bool) {
	var req phoneRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil || req.Phone == "" {
		return "", false
	}
	return req.Phone, true
}

func (h *Handler) requestPhoneLogin(w http.ResponseWriter, r *http.Request) {
	phone, ok := decodePhone(w, r)
	if !ok {
		h.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := h.svc.RequestPhoneLogin(r.Context(), phone); err != nil {
		h.mapServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) verifyPhoneLogin(w http.ResponseWriter, r *http.Request) {
	phone, ok := decodePhone(w, r)
	if !ok {
		h.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	pair, err := h.svc.LoginPhone(r.Context(), phone)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pair)
}

func (h *Handler) addPhone(w http.ResponseWriter, r *http.Request) {
	claims, err := h.authenticate(r)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	phone, ok := decodePhone(w, r)
	if !ok {
		h.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if _, err := h.svc.AddPhone(r.Context(), claims.AccountID, phone); err != nil {
		h.mapServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) verifyPhone(w http.ResponseWriter, r *http.Request) {
	claims, err := h.authenticate(r)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	phone, ok := decodePhone(w, r)
	if !ok {
		h.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	a, err := h.svc.Account(r.Context(), claims.AccountID)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	if a.Phone != phone || !a.PhoneVerified {
		h.mapServiceError(w, ErrValidation)
		return
	}
	writeAccount(w, a)
}

func (h *Handler) removePhone(w http.ResponseWriter, r *http.Request) {
	claims, err := h.authenticate(r)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	if err := h.svc.RemovePhone(r.Context(), claims.AccountID); err != nil {
		h.mapServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeAccount(w http.ResponseWriter, a *Account) {
	var email, phone, name, cpf any
	if a.Email != "" {
		email = a.Email
	}
	if a.Phone != "" {
		phone = a.Phone
	}
	if a.Name != "" {
		name = a.Name
	}
	if a.CPF != "" {
		cpf = a.CPF
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"id": a.ID, "email": email, "email_verified": a.EmailVerified, "phone": phone, "phone_verified": a.PhoneVerified, "name": name, "cpf": cpf, "cpf_claimed_at": a.CPFClaimedAt, "is_shadow": a.Email == "" && a.Phone == "" && a.Password == "", "created_at": a.CreatedAt})
}
