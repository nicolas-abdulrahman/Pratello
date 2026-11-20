package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"pratello/server/internal/user"
	"strings"
)

type registerRequest struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone,omitempty"`
	CPF      string `json:"cpf,omitempty"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(mux *http.ServeMux) {
	prefixes := []string{"", "/api"}
	for _, p := range prefixes {
		// Autenticação & Perfil
		mux.HandleFunc(fmt.Sprintf("POST %s/auth/register", p), h.register)
		mux.HandleFunc(fmt.Sprintf("POST %s/auth/login", p), h.login)
		mux.HandleFunc(fmt.Sprintf("GET %s/auth/me", p), h.me)
		mux.HandleFunc(fmt.Sprintf("POST %s/auth/refresh", p), h.refresh)
		mux.HandleFunc(fmt.Sprintf("POST %s/auth/logout", p), h.logout)
		mux.HandleFunc(fmt.Sprintf("POST %s/auth/logout-all", p), h.logoutAll)
		mux.HandleFunc(fmt.Sprintf("POST %s/auth/login/phone/request-code", p), h.requestPhoneLogin)
		mux.HandleFunc(fmt.Sprintf("POST %s/auth/login/phone/verify", p), h.verifyPhoneLogin)
		mux.HandleFunc(fmt.Sprintf("POST %s/auth/register/phone/request-code", p), h.requestPhoneRegistration)
		mux.HandleFunc(fmt.Sprintf("POST %s/auth/register/phone/verify", p), h.registerPhone)
		mux.HandleFunc(fmt.Sprintf("POST %s/me/phone/request-code", p), h.addPhone)
		mux.HandleFunc(fmt.Sprintf("POST %s/me/phone/verify", p), h.verifyPhone)
		mux.HandleFunc(fmt.Sprintf("DELETE %s/me/phone", p), h.removePhone)
		mux.HandleFunc(fmt.Sprintf("PUT %s/me/cpf", p), h.claimCPF)
		mux.HandleFunc(fmt.Sprintf("DELETE %s/me/cpf", p), h.removeCPF)
		mux.HandleFunc(fmt.Sprintf("PUT %s/me/email", p), h.setEmail)
		mux.HandleFunc(fmt.Sprintf("DELETE %s/me/email", p), h.removeEmail)
		mux.HandleFunc(fmt.Sprintf("POST %s/me/email/verify", p), h.verifyEmail)

		// Superadministradores
	}
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "corpo da requisição inválido (JSON esperado)")
		return
	}

	user, err := h.svc.Register(r.Context(), RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Phone:    req.Phone,
		CPF:      req.CPF,
	})
	if err != nil {
		h.mapServiceError(w, err)
		return
	}

	pair, err := h.svc.IssueTokens(r.Context(), user.ID)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(pair)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "corpo da requisição inválido (JSON esperado)")
		return
	}

	_, user, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}

	pair, err := h.svc.IssueTokens(r.Context(), user.ID)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	_ = json.NewEncoder(w).Encode(pair)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil || req.RefreshToken == "" {
		h.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	pair, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pair)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	claims, err := h.authenticate(r)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	var req refreshRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil || req.RefreshToken == "" {
		h.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := h.svc.Logout(r.Context(), claims.AccountID, req.RefreshToken); err != nil {
		h.mapServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) logoutAll(w http.ResponseWriter, r *http.Request) {
	claims, err := h.authenticate(r)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	if err := h.svc.LogoutAll(r.Context(), claims.AccountID); err != nil {
		h.mapServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, err := h.authenticate(r)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}

	account, err := h.svc.Account(r.Context(), claims.AccountID)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}

	user.WriteAccount(w, account)
}

func (h *Handler) authenticate(r *http.Request) (*AccountClaims, error) {
	token := extractBearerToken(r)
	if token == "" {
		return nil, ErrUnauthorized
	}
	return h.svc.Authenticate(r.Context(), token)
}

func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

func (h *Handler) writeError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (h *Handler) mapServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		h.writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrInvalidCredentials):
		h.writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrUnauthorized):
		h.writeError(w, http.StatusUnauthorized, "não autorizado")
	case errors.Is(err, ErrForbidden):
		h.writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrNotFound):
		h.writeError(w, http.StatusNotFound, "recurso não encontrado")
	case errors.Is(err, ErrConflict):
		h.writeError(w, http.StatusConflict, err.Error())
	default:
		h.writeError(w, http.StatusInternalServerError, err.Error())
	}
}

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
	user.WriteAccount(w, a)
}
