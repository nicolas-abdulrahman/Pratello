package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	Code     string `json:"code"`
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
		mux.HandleFunc(fmt.Sprintf("POST %s/auth/register/email/request-code", p), h.requestRegistrationEmail)
		mux.HandleFunc(fmt.Sprintf("POST %s/auth/login", p), h.login)
		mux.HandleFunc(fmt.Sprintf("GET %s/auth/me", p), h.me)
		mux.HandleFunc(fmt.Sprintf("PUT %s/me/avatar", p), h.setAvatar)
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

func (h *Handler) requestRegistrationEmail(w http.ResponseWriter, r *http.Request) {
	h.requestEmail(w, r, "register-email", 0, 0)
}

func (h *Handler) requestEmail(w http.ResponseWriter, r *http.Request, purpose string, id, version int) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
		h.writeError(w, 400, "invalid request")
		return
	}
	if err := h.svc.requestEmailCode(r.Context(), purpose, id, version, req.Email); err != nil {
		h.mapServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req registerRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := decoder.Decode(&req); err != nil {
		var limit *http.MaxBytesError
		if errors.As(err, &limit) {
			h.writeError(w, 413, "request body too large")
			return
		}
		h.writeError(w, http.StatusBadRequest, "corpo da requisição inválido (JSON esperado)")
		return
	}

	input := RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Phone:    req.Phone,
		CPF:      req.CPF,
	}
	if err := h.svc.ValidateRegisterInput(input); err != nil {
		h.mapServiceError(w, err)
		return
	}
	if err := h.svc.verifyCode(r.Context(), "register-email", 0, 0, normalizeEmail(input.Email), req.Code); err != nil {
		h.mapServiceError(w, err)
		return
	}
	user, err := h.svc.Register(r.Context(), input)
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

func (h *Handler) setAvatar(w http.ResponseWriter, r *http.Request) {
	claims, err := h.authenticate(r)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	var body struct {
		URL *string `json:"avatar_url"`
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	d.DisallowUnknownFields()
	if d.Decode(&body) != nil || body.URL == nil || d.Decode(new(any)) != io.EOF {
		h.writeError(w, 400, "invalid avatar request")
		return
	}
	a, err := h.svc.SetAvatar(r.Context(), claims.AccountID, *body.URL)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	user.WriteAccount(w, a)
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
	case errors.Is(err, ErrWrongCode):
		h.writeError(w, http.StatusBadRequest, ErrWrongCode.Error())
	case errors.Is(err, ErrVerificationUnavailable):
		h.writeError(w, http.StatusServiceUnavailable, ErrVerificationUnavailable.Error())
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
	h.requestEmail(w, r, "change-email", claims.AccountID, claims.TokenVersion)
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
	var req struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		CurrentPassword string `json:"current_password"`
		Code            string `json:"code"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
		h.writeError(w, 400, "invalid request")
		return
	}
	if err := h.svc.verifyCode(r.Context(), "change-email", claims.AccountID, claims.TokenVersion, normalizeEmail(req.Email), req.Code); err != nil {
		h.mapServiceError(w, err)
		return
	}
	a, err := h.svc.SetEmail(r.Context(), claims.AccountID, req.Email, req.Password, req.CurrentPassword)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}
	writeAccount(w, a)
}

type phoneRequest struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
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
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil || req.Phone == "" {
		h.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := h.svc.verifyCode(r.Context(), "register-phone", 0, 0, strings.TrimSpace(req.Phone), req.Code); err != nil {
		h.mapServiceError(w, err)
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
	req, ok := decodePhoneRequest(w, r)
	return req.Phone, ok
}
func decodePhoneRequest(w http.ResponseWriter, r *http.Request) (phoneRequest, bool) {
	var req phoneRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil || req.Phone == "" {
		return req, false
	}
	req.Phone = strings.TrimSpace(req.Phone)
	return req, true
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
	req, ok := decodePhoneRequest(w, r)
	if !ok {
		h.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := h.svc.verifyCode(r.Context(), "login-phone", 0, 0, req.Phone, req.Code); err != nil {
		h.mapServiceError(w, err)
		return
	}
	pair, err := h.svc.LoginPhone(r.Context(), req.Phone)
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
	if !validPhone(phone) {
		h.mapServiceError(w, ErrValidation)
		return
	}
	if err := h.svc.requestVerification(r.Context(), "change-phone", claims.AccountID, claims.TokenVersion, phone); err != nil {
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
	req, ok := decodePhoneRequest(w, r)
	if !ok {
		h.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := h.svc.verifyCode(r.Context(), "change-phone", claims.AccountID, claims.TokenVersion, req.Phone, req.Code); err != nil {
		h.mapServiceError(w, err)
		return
	}
	a, err := h.svc.AddPhone(r.Context(), claims.AccountID, req.Phone)
	if err != nil {
		h.mapServiceError(w, err)
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
