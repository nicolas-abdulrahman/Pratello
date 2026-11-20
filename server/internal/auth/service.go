package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/mail"
	"os"
	"pratello/server/internal/common"
	"strings"
	"time"
)

var (
	ErrInvalidCredentials      = errors.New("e-mail ou senha inválidos")
	ErrNotFound                = errors.New("recurso não encontrado")
	ErrConflict                = errors.New("conflito de dados (registro já existente)")
	ErrValidation              = errors.New("erro de validação dos dados")
	ErrUnauthorized            = errors.New("não autorizado")
	ErrForbidden               = errors.New("acesso proibido para este nível de privilégio")
	ErrWrongCode               = errors.New("código incorreto ou expirado")
	ErrVerificationUnavailable = errors.New("verificação temporariamente indisponível")
)

type authStore interface {
	CreateAccount(ctx context.Context, name, email, passwordHash, phone string) (*Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)
	GetAccountByID(ctx context.Context, id int) (*Account, error)
}

type AccountClaims struct {
	AccountID    int    `json:"account_id"`
	Email        string `json:"email,omitempty"`
	TokenVersion int    `json:"token_version"`
	Type         string `json:"typ"`
	jwt.RegisteredClaims
}

type User struct {
	ID        int
	Name      string
	Email     string
	Phone     string
	CreatedAt time.Time
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
	Phone    string
	CPF      string
}

type Service struct {
	store        authStore
	jwtSecret    []byte
	hashCost     int
	verification verificationStore
}

func (s *Service) Identify(ctx context.Context, token string) (int, error) {
	claims, err := s.Authenticate(ctx, token)
	if errors.Is(err, ErrUnauthorized) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return claims.AccountID, nil
}

const fixedCode = "123456"
const verificationTTL = 5 * time.Minute

type verificationStore interface {
	put(context.Context, string, string, time.Duration) error
	consume(context.Context, string, string) (bool, error)
}

func challengeKey(purpose string, id, version int, contact string) string {
	return "pratello:auth:code:" + purpose + ":" + tokenHash(fmt.Sprintf("%d:%d:%s", id, version, contact))
}

func (s *Service) requestVerification(ctx context.Context, purpose string, id, version int, contact string) error {
	if s.verification == nil {
		return ErrVerificationUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := s.verification.put(ctx, challengeKey(purpose, id, version, contact), fixedCode, verificationTTL); err != nil {
		return fmt.Errorf("%w: %v", ErrVerificationUnavailable, err)
	}
	return nil
}

func (s *Service) verifyCode(ctx context.Context, purpose string, id, version int, contact, code string) error {
	if s.verification == nil {
		return ErrVerificationUnavailable
	}
	if len(code) != 6 {
		return ErrWrongCode
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	ok, err := s.verification.consume(ctx, challengeKey(purpose, id, version, contact), code)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrVerificationUnavailable, err)
	}
	if !ok {
		return ErrWrongCode
	}
	return nil
}

func normalizeEmail(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

func (s *Service) requestEmailCode(ctx context.Context, purpose string, id, version int, value string) error {
	email := normalizeEmail(value)
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || len(email) > 255 {
		return ErrValidation
	}
	return s.requestVerification(ctx, purpose, id, version, email)
}

func NewService(store authStore) *Service {
	secret := []byte(os.Getenv("JWT_SECRET"))
	if len(secret) == 0 {
		secret = make([]byte, 32)
		rand.Read(secret)
	}
	return &Service{
		store:     store,
		jwtSecret: secret,
		hashCost:  bcrypt.DefaultCost,
	}
}

// SetHashCost allows tests to use bcrypt.MinCost for maximum execution speed.
func (s *Service) SetHashCost(cost int) {
	s.hashCost = cost
}

func (s *Service) ValidateRegisterInput(input RegisterInput) error {
	if input.CPF != "" && !validCPF(input.CPF) {
		return errors.Join(ErrValidation, errors.New("CPF inválido"))
	}
	email := strings.TrimSpace(input.Email)
	if email == "" {
		return errors.Join(ErrValidation, errors.New("e-mail é obrigatório"))
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.Join(ErrValidation, errors.New("formato de e-mail inválido"))
	}
	if len(input.Password) < 8 {
		return errors.Join(ErrValidation, errors.New("a senha deve ter no mínimo 8 caracteres"))
	}
	if input.Phone != "" && !validPhone(strings.TrimSpace(input.Phone)) {
		return errors.Join(ErrValidation, errors.New("telefone inválido"))
	}
	return nil
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*User, error) {
	if err := s.ValidateRegisterInput(input); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), s.hashCost)
	if err != nil {
		return nil, fmt.Errorf("falha ao criptografar senha: %w", err)
	}
	var acct *Account
	if input.CPF != "" {
		store, ok := s.store.(cpfStore)
		if !ok {
			return nil, ErrUnauthorized
		}
		acct, err = store.RegisterWithCPF(ctx, strings.TrimSpace(input.Name), strings.ToLower(strings.TrimSpace(input.Email)), string(hash), strings.TrimSpace(input.Phone), input.CPF, false)
	} else {
		acct, err = s.store.CreateAccount(ctx, strings.TrimSpace(input.Name), input.Email, string(hash), strings.TrimSpace(input.Phone))
	}
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        acct.ID,
		Name:      acct.Name,
		Email:     acct.Email,
		Phone:     acct.Phone,
		CreatedAt: acct.CreatedAt,
	}, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, *User, error) {
	email = strings.TrimSpace(email)
	if email == "" || password == "" {
		return "", nil, ErrInvalidCredentials
	}

	acct, err := s.store.GetAccountByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, fmt.Errorf("erro no banco ao buscar conta: %w", err)
	}

	if !acct.EmailVerified || bcrypt.CompareHashAndPassword([]byte(acct.Password), []byte(password)) != nil {
		return "", nil, ErrInvalidCredentials
	}

	token, err := s.GenerateToken(acct.ID, acct.Email)
	if err != nil {
		return "", nil, fmt.Errorf("falha ao gerar token de acesso: %w", err)
	}

	user := &User{
		ID:        acct.ID,
		Name:      acct.Name,
		Email:     acct.Email,
		Phone:     acct.Phone,
		CreatedAt: acct.CreatedAt,
	}

	return token, user, nil
}

func (s *Service) GenerateToken(accountID int, email string) (string, error) {
	a, err := s.store.GetAccountByID(context.Background(), accountID)
	if err != nil {
		return "", err
	}
	return s.signAccess(a)
}

func (s *Service) signAccess(a *Account) (string, error) {
	claims := AccountClaims{
		AccountID:    a.ID,
		Email:        a.Email,
		TokenVersion: a.TokenVersion,
		Type:         "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *Service) ParseToken(tokenStr string) (*AccountClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccountClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("método de assinatura inesperado: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	}, jwt.WithExpirationRequired())
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AccountClaims); ok && token.Valid && claims.Type == "access" && claims.AccountID > 0 {
		return claims, nil
	}
	return nil, ErrUnauthorized
}

func (s *Service) GetAccountByID(ctx context.Context, id int) (*User, error) {
	acct, err := s.store.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:        acct.ID,
		Name:      acct.Name,
		Email:     acct.Email,
		Phone:     acct.Phone,
		CreatedAt: acct.CreatedAt,
	}, nil
}

func (s *Service) Account(ctx context.Context, id int) (*Account, error) {
	return s.store.GetAccountByID(ctx, id)
}

func (s *Service) SetAvatar(ctx context.Context, id int, avatar string) (*Account, error) {
	if avatar != "" {
		var err error
		avatar, err = common.ImageURL(avatar)
		if err != nil {
			return nil, ErrValidation
		}
	}
	store, ok := s.store.(interface {
		SetAvatar(context.Context, int, string) error
	})
	if !ok {
		return nil, errors.New("profile store unavailable")
	}
	if err := store.SetAvatar(ctx, id, avatar); err != nil {
		return nil, err
	}
	return s.Account(ctx, id)
}

func validCPF(cpf string) bool {
	if len(cpf) != 11 {
		return false
	}
	for _, c := range cpf {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func (s *Service) ClaimCPF(ctx context.Context, id int, cpf string) (*Account, error) {
	if !validCPF(cpf) {
		return nil, ErrValidation
	}
	store, ok := s.store.(cpfStore)
	if !ok {
		return nil, ErrUnauthorized
	}
	if err := store.ClaimCPF(ctx, id, cpf); err != nil {
		return nil, err
	}
	return s.store.GetAccountByID(ctx, id)
}

func (s *Service) RemoveCPF(ctx context.Context, id int) error {
	store, ok := s.store.(cpfStore)
	if !ok {
		return ErrUnauthorized
	}
	return store.RemoveCPF(ctx, id)
}

type emailStore interface {
	SetVerifiedEmail(context.Context, int, string, string, string) error
	RemoveEmail(context.Context, int) error
}

func (s *Service) SetEmail(ctx context.Context, id int, email, password, currentPassword string) (*Account, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || len(password) < 8 {
		return nil, ErrValidation
	}
	a, err := s.store.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if a.Email != "" && bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(currentPassword)) != nil {
		return nil, ErrInvalidCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.hashCost)
	if err != nil {
		return nil, err
	}
	store, ok := s.store.(emailStore)
	if !ok {
		return nil, ErrUnauthorized
	}
	if other, err := s.store.GetAccountByEmail(ctx, email); err == nil && other.ID != id {
		return nil, ErrConflict
	} else if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	if err = store.SetVerifiedEmail(ctx, id, email, string(hash), a.Password); err != nil {
		return nil, err
	}
	return s.store.GetAccountByID(ctx, id)
}

func (s *Service) RemoveEmail(ctx context.Context, id int) error {
	store, ok := s.store.(emailStore)
	if !ok {
		return ErrUnauthorized
	}
	return store.RemoveEmail(ctx, id)
}

func validPhone(phone string) bool {
	if len(phone) < 8 || len(phone) > 20 {
		return false
	}
	for i, c := range phone {
		if c == '+' && i == 0 {
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// RequestPhoneLogin always succeeds for unknown or unverified phone numbers.
func (s *Service) RequestPhoneLogin(ctx context.Context, phone string) error {
	phone = strings.TrimSpace(phone)
	if !validPhone(phone) {
		return ErrValidation
	}
	store, ok := s.store.(phoneStore)
	if !ok {
		return ErrUnauthorized
	}
	a, err := store.GetAccountByPhone(ctx, phone)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if !a.PhoneVerified || a.Phone != phone {
		return nil
	}
	return s.requestVerification(ctx, "login-phone", 0, 0, phone)
}

func (s *Service) RequestPhoneRegistration(ctx context.Context, phone string) error {
	phone = strings.TrimSpace(phone)
	if !validPhone(phone) {
		return ErrValidation
	}
	store, ok := s.store.(phoneStore)
	if !ok {
		return ErrUnauthorized
	}
	_, err := store.GetAccountByPhone(ctx, phone)
	if err == nil {
		return nil
	}
	if !errors.Is(err, ErrNotFound) {
		return err
	}
	return s.requestVerification(ctx, "register-phone", 0, 0, phone)
}

func (s *Service) RegisterPhone(ctx context.Context, phone, name, cpf string) (*TokenPair, error) {
	phone = strings.TrimSpace(phone)
	if !validPhone(phone) || (cpf != "" && !validCPF(cpf)) {
		return nil, ErrValidation
	}
	store, ok := s.store.(phoneStore)
	if !ok {
		return nil, ErrUnauthorized
	}
	var a *Account
	var err error
	if cpf != "" {
		cpfStore, ok := s.store.(cpfStore)
		if !ok {
			return nil, ErrUnauthorized
		}
		a, err = cpfStore.RegisterWithCPF(ctx, strings.TrimSpace(name), "", "", phone, cpf, true)
	} else {
		a, err = store.CreatePhoneAccount(ctx, phone, strings.TrimSpace(name))
	}
	if err != nil {
		return nil, err
	}
	return s.IssueTokens(ctx, a.ID)
}

func (s *Service) LoginPhone(ctx context.Context, phone string) (*TokenPair, error) {
	phone = strings.TrimSpace(phone)
	if !validPhone(phone) {
		return nil, ErrInvalidCredentials
	}
	store, ok := s.store.(phoneStore)
	if !ok {
		return nil, ErrUnauthorized
	}
	a, err := store.GetAccountByPhone(ctx, phone)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if !a.PhoneVerified || a.Phone != phone {
		return nil, ErrInvalidCredentials
	}
	return s.IssueTokens(ctx, a.ID)
}

func (s *Service) AddPhone(ctx context.Context, id int, phone string) (*Account, error) {
	phone = strings.TrimSpace(phone)
	if !validPhone(phone) {
		return nil, ErrValidation
	}
	store, ok := s.store.(phoneStore)
	if !ok {
		return nil, ErrUnauthorized
	}
	existing, err := store.GetAccountByPhone(ctx, phone)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	if err == nil && existing.ID != id {
		return nil, ErrConflict
	}
	if err = store.SetVerifiedPhone(ctx, id, phone); err != nil {
		return nil, err
	}
	return s.store.GetAccountByID(ctx, id)
}

func (s *Service) RemovePhone(ctx context.Context, id int) error {
	store, ok := s.store.(phoneStore)
	if !ok {
		return ErrUnauthorized
	}
	return store.RemovePhone(ctx, id)
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func tokenHash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (s *Service) IssueTokens(ctx context.Context, id int) (*TokenPair, error) {
	return s.issueTokens(ctx, id, nil)
}

func (s *Service) issueTokens(ctx context.Context, id int, expectedVersion *int) (*TokenPair, error) {
	store, ok := s.store.(sessionStore)
	if !ok {
		return nil, ErrUnauthorized
	}
	a, err := s.store.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if expectedVersion != nil && a.TokenVersion != *expectedVersion {
		return nil, ErrUnauthorized
	}
	access, err := s.signAccess(a)
	if err != nil {
		return nil, err
	}
	var nonce [32]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return nil, err
	}
	now := time.Now()
	expiry := now.Add(30 * 24 * time.Hour)
	claims := AccountClaims{AccountID: a.ID, TokenVersion: a.TokenVersion, Type: "refresh", RegisteredClaims: jwt.RegisteredClaims{ID: hex.EncodeToString(nonce[:]), IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(expiry)}}
	refresh, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}
	if err = store.SaveRefresh(ctx, id, a.TokenVersion, tokenHash(refresh), expiry); err != nil {
		return nil, err
	}
	return &TokenPair{access, refresh, "Bearer", 900}, nil
}

func (s *Service) Refresh(ctx context.Context, raw string) (*TokenPair, error) {
	store, ok := s.store.(sessionStore)
	if !ok {
		return nil, ErrUnauthorized
	}
	token, err := jwt.ParseWithClaims(raw, &AccountClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrUnauthorized
		}
		return s.jwtSecret, nil
	}, jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return nil, ErrUnauthorized
	}
	claims, ok := token.Claims.(*AccountClaims)
	if !ok || claims.Type != "refresh" || claims.AccountID <= 0 {
		return nil, ErrUnauthorized
	}
	id, err := store.ConsumeRefresh(ctx, tokenHash(raw), claims.TokenVersion)
	if err != nil {
		return nil, err
	}
	if id != claims.AccountID {
		return nil, ErrUnauthorized
	}
	return s.issueTokens(ctx, id, &claims.TokenVersion)
}

func (s *Service) Authenticate(ctx context.Context, token string) (*AccountClaims, error) {
	claims, err := s.ParseToken(token)
	if err != nil {
		return nil, ErrUnauthorized
	}
	a, err := s.store.GetAccountByID(ctx, claims.AccountID)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrUnauthorized
		}
		return nil, err
	}
	if a.TokenVersion != claims.TokenVersion || (a.Email == "" && a.Phone == "" && a.Password == "") {
		return nil, ErrUnauthorized
	}
	return claims, nil
}

func (s *Service) Logout(ctx context.Context, id int, raw string) error {
	store, ok := s.store.(sessionStore)
	if !ok {
		return ErrUnauthorized
	}
	return store.RevokeRefresh(ctx, tokenHash(raw), id)
}

func (s *Service) LogoutAll(ctx context.Context, id int) error {
	store, ok := s.store.(sessionStore)
	if !ok {
		return ErrUnauthorized
	}
	return store.RevokeAll(ctx, id)
}
