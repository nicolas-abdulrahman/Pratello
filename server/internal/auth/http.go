package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// -------------------------------------------------------------
// Request & Response DTOs
// -------------------------------------------------------------

type registerRequest struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone,omitempty"`
}

type userResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name,omitempty"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token   string       `json:"token"`
	Account userResponse `json:"account"`
}

type accountIDRequest struct {
	AccountID int `json:"account_id"`
}

type sysadminResponse struct {
	ID        int       `json:"id"`
	AccountID int       `json:"account_id"`
	CreatedAt time.Time `json:"created_at"`
}

type managerResponse struct {
	ID             int    `json:"id"`
	RestaurantSlug string `json:"restaurant_slug"`
	AccountID      int    `json:"account_id"`
}

type ownerResponse struct {
	ID             int    `json:"id"`
	RestaurantSlug string `json:"restaurant_slug"`
	AccountID      int    `json:"account_id"`
}

type employeeRequest struct {
	Name     string `json:"name"`
	CPF      string `json:"cpf"`
	Position string `json:"position"`
}

type employeeResponse struct {
	ID                   int       `json:"id"`
	PhysicalRestaurantID int       `json:"physical_restaurant_id"`
	Name                 string    `json:"name"`
	CPF                  string    `json:"cpf"`
	Position             string    `json:"position"`
	AccountID            *int      `json:"account_id,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
}

type terminalSessionRequest struct {
	PIN      string `json:"pin"`
	DeviceID string `json:"device_id"`
}

type terminalSessionResponse struct {
	Token     string    `json:"token"`
	DeviceID  string    `json:"device_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type terminalOrderResponse struct {
	ID                   int       `json:"id"`
	PhysicalRestaurantID int       `json:"physical_restaurant_id"`
	AccountID            int       `json:"account_id"`
	OrderType            string    `json:"order_type"`
	Status               string    `json:"status"`
	Subtotal             float64   `json:"subtotal"`
	Discount             float64   `json:"discount"`
	Total                float64   `json:"total"`
	CreatedAt            time.Time `json:"created_at"`
}

// -------------------------------------------------------------
// HTTP Handler
// -------------------------------------------------------------

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

		// Superadministradores
		mux.HandleFunc(fmt.Sprintf("GET %s/sysadmins", p), h.listSysadmins)
		mux.HandleFunc(fmt.Sprintf("POST %s/sysadmins", p), h.promoteSysadmin)

		// Gerentes
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants/{slug}/managers", p), h.listManagers)
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants/{slug}/managers", p), h.assignManager)
		mux.HandleFunc(fmt.Sprintf("DELETE %s/restaurants/{slug}/managers/{account_id}", p), h.removeManager)

		// Proprietários
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants/{slug}/owners", p), h.listOwners)
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants/{slug}/owners", p), h.assignOwner)
		mux.HandleFunc(fmt.Sprintf("DELETE %s/restaurants/{slug}/owners/{account_id}", p), h.removeOwner)

		// Colaboradores (RH)
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants/{slug}/branches/{branch_id}/employees", p), h.listEmployees)
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants/{slug}/branches/{branch_id}/employees", p), h.createEmployee)
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants/{slug}/branches/{branch_id}/employees/{employee_id}/invite", p), h.inviteEmployee)

		// Terminais de Salão
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants/{slug}/branches/{branch_id}/terminal/session", p), h.openTerminalSession)
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants/{slug}/branches/{branch_id}/terminal/orders", p), h.getTerminalOrders)
	}
}

// -------------------------------------------------------------
// Handler Implementations
// -------------------------------------------------------------

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
	})
	if err != nil {
		h.mapServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(userResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt,
	})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "corpo da requisição inválido (JSON esperado)")
		return
	}

	token, user, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(loginResponse{
		Token: token,
		Account: userResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Phone:     user.Phone,
			CreatedAt: user.CreatedAt,
		},
	})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, err := h.authenticate(r)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}

	user, err := h.svc.GetAccountByID(r.Context(), claims.AccountID)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(userResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt,
	})
}

func (h *Handler) listSysadmins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	list, err := h.svc.ListSysadmins(r.Context())
	if err != nil {
		h.mapServiceError(w, err)
		return
	}

	resp := make([]sysadminResponse, 0, len(list))
	for _, sa := range list {
		resp = append(resp, sysadminResponse{
			ID:        sa.ID,
			AccountID: sa.AccountID,
			CreatedAt: sa.CreatedAt,
		})
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) promoteSysadmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req accountIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AccountID <= 0 {
		h.writeError(w, http.StatusBadRequest, "account_id é obrigatório")
		return
	}

	sa, err := h.svc.AddSysadmin(r.Context(), req.AccountID)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(sysadminResponse{
		ID:        sa.ID,
		AccountID: sa.AccountID,
		CreatedAt: sa.CreatedAt,
	})
}

func (h *Handler) listManagers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := r.PathValue("slug")

	list, err := h.svc.ListManagers(r.Context(), slug)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}

	resp := make([]managerResponse, 0, len(list))
	for _, m := range list {
		resp = append(resp, managerResponse{
			ID:             m.ID,
			RestaurantSlug: m.RestaurantSlug,
			AccountID:      m.AccountID,
		})
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) assignManager(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := r.PathValue("slug")

	var req accountIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AccountID <= 0 {
		h.writeError(w, http.StatusBadRequest, "account_id é obrigatório")
		return
	}

	m, err := h.svc.AddManager(r.Context(), slug, req.AccountID)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(managerResponse{
		ID:             m.ID,
		RestaurantSlug: m.RestaurantSlug,
		AccountID:      m.AccountID,
	})
}

func (h *Handler) removeManager(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	acctIDStr := r.PathValue("account_id")
	acctID, err := strconv.Atoi(acctIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "account_id inválido")
		return
	}

	if err := h.svc.RemoveManager(r.Context(), slug, acctID); err != nil {
		h.mapServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listOwners(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := r.PathValue("slug")

	list, err := h.svc.ListOwners(r.Context(), slug)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}

	resp := make([]ownerResponse, 0, len(list))
	for _, o := range list {
		resp = append(resp, ownerResponse{
			ID:             o.ID,
			RestaurantSlug: o.RestaurantSlug,
			AccountID:      o.AccountID,
		})
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) assignOwner(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	slug := r.PathValue("slug")

	var req accountIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AccountID <= 0 {
		h.writeError(w, http.StatusBadRequest, "account_id é obrigatório")
		return
	}

	o, err := h.svc.AddOwner(r.Context(), slug, req.AccountID)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(ownerResponse{
		ID:             o.ID,
		RestaurantSlug: o.RestaurantSlug,
		AccountID:      o.AccountID,
	})
}

func (h *Handler) removeOwner(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	acctIDStr := r.PathValue("account_id")
	acctID, err := strconv.Atoi(acctIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "account_id inválido")
		return
	}

	if err := h.svc.RemoveOwner(r.Context(), slug, acctID); err != nil {
		h.mapServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listEmployees(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	branchIDStr := r.PathValue("branch_id")
	branchID, err := strconv.Atoi(branchIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "branch_id inválido")
		return
	}

	list, err := h.svc.ListEmployees(r.Context(), branchID)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}

	resp := make([]employeeResponse, 0, len(list))
	for _, emp := range list {
		resp = append(resp, employeeResponse{
			ID:                   emp.ID,
			PhysicalRestaurantID: emp.PhysicalRestaurantID,
			Name:                 emp.Name,
			CPF:                  emp.CPF,
			Position:             emp.Position,
			AccountID:            emp.AccountID,
			CreatedAt:            emp.CreatedAt,
		})
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) createEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	branchIDStr := r.PathValue("branch_id")
	branchID, err := strconv.Atoi(branchIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "branch_id inválido")
		return
	}

	var req employeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	emp, err := h.svc.CreateEmployee(r.Context(), branchID, req.Name, req.CPF, req.Position)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(employeeResponse{
		ID:                   emp.ID,
		PhysicalRestaurantID: emp.PhysicalRestaurantID,
		Name:                 emp.Name,
		CPF:                  emp.CPF,
		Position:             emp.Position,
		AccountID:            emp.AccountID,
		CreatedAt:            emp.CreatedAt,
	})
}

func (h *Handler) inviteEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	empIDStr := r.PathValue("employee_id")
	empID, err := strconv.Atoi(empIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "employee_id inválido")
		return
	}

	var req accountIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AccountID <= 0 {
		h.writeError(w, http.StatusBadRequest, "account_id é obrigatório")
		return
	}

	emp, err := h.svc.LinkEmployeeAccount(r.Context(), empID, req.AccountID)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(employeeResponse{
		ID:                   emp.ID,
		PhysicalRestaurantID: emp.PhysicalRestaurantID,
		Name:                 emp.Name,
		CPF:                  emp.CPF,
		Position:             emp.Position,
		AccountID:            emp.AccountID,
		CreatedAt:            emp.CreatedAt,
	})
}

func (h *Handler) openTerminalSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	branchIDStr := r.PathValue("branch_id")
	branchID, err := strconv.Atoi(branchIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "branch_id inválido")
		return
	}

	var req terminalSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	ts, err := h.svc.OpenTerminalSession(r.Context(), branchID, req.PIN, req.DeviceID)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(terminalSessionResponse{
		Token:     ts.Token,
		DeviceID:  ts.DeviceID,
		ExpiresAt: ts.ExpiresAt,
	})
}

func (h *Handler) getTerminalOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	branchIDStr := r.PathValue("branch_id")
	branchID, err := strconv.Atoi(branchIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "branch_id inválido")
		return
	}

	token := extractBearerToken(r)
	deviceID := r.Header.Get("X-Device-Id")
	if token == "" || deviceID == "" {
		h.writeError(w, http.StatusUnauthorized, "token de sessão e cabeçalho X-Device-Id são obrigatórios")
		return
	}

	if _, err := h.svc.ValidateTerminalToken(r.Context(), token, deviceID); err != nil {
		h.mapServiceError(w, err)
		return
	}

	orders, err := h.svc.GetTerminalOrders(r.Context(), branchID)
	if err != nil {
		h.mapServiceError(w, err)
		return
	}

	resp := make([]terminalOrderResponse, 0, len(orders))
	for _, o := range orders {
		resp = append(resp, terminalOrderResponse{
			ID:                   o.ID,
			PhysicalRestaurantID: o.PhysicalRestaurantID,
			AccountID:            o.AccountID,
			OrderType:            o.OrderType,
			Status:               o.Status,
			Subtotal:             o.Subtotal,
			Discount:             o.Discount,
			Total:                o.Total,
			CreatedAt:            o.CreatedAt,
		})
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// -------------------------------------------------------------
// Helpers
// -------------------------------------------------------------

func (h *Handler) authenticate(r *http.Request) (*AccountClaims, error) {
	token := extractBearerToken(r)
	if token == "" {
		return nil, ErrUnauthorized
	}
	return h.svc.ParseToken(token)
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
	case errors.Is(err, ErrCannotRemoveLastOwner):
		h.writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrInvalidCredentials):
		h.writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrUnauthorized):
		h.writeError(w, http.StatusUnauthorized, "não autorizado")
	case errors.Is(err, ErrInvalidPIN):
		h.writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrTerminalLocked):
		h.writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrForbidden):
		h.writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrNotFound):
		h.writeError(w, http.StatusNotFound, "recurso não encontrado")
	case errors.Is(err, ErrConflict):
		h.writeError(w, http.StatusConflict, "este e-mail já está cadastrado")
	default:
		h.writeError(w, http.StatusInternalServerError, err.Error())
	}
}
