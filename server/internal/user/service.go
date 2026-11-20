package user

import (
	"context"
	"errors"
	"strings"
	"time"
)

type Role string

const (
	Administrator Role = "ADMINISTRATOR"
	Sysadmin      Role = "SYSADMIN"
	Owner         Role = "OWNER"
	Analyst       Role = "ANALYST"
	Manager       Role = "MANAGER"
	Employee      Role = "EMPLOYEE"
)

var (
	ErrInvalid     = errors.New("invalid user or role data")
	ErrForbidden   = errors.New("forbidden")
	ErrMissing     = errors.New("user or assignment not found")
	ErrConflict    = errors.New("assignment already exists or would remove the last owner")
	ErrUnavailable = errors.New("user database unavailable")
)

// Account is shared with authentication. Credential fields are never serialized.
type Account struct {
	ID            int        `json:"id"`
	Name          string     `json:"name,omitempty"`
	Email         string     `json:"email,omitempty"`
	EmailVerified bool       `json:"email_verified"`
	Password      string     `json:"-"`
	Phone         string     `json:"phone,omitempty"`
	PhoneVerified bool       `json:"phone_verified"`
	CPF           string     `json:"cpf,omitempty"`
	CPFClaimedAt  *time.Time `json:"cpf_claimed_at"`
	TokenVersion  int        `json:"-"`
	CreatedAt     time.Time  `json:"created_at"`
}

type Scope struct {
	Organization                               Role
	Administrator, Sysadmin, Manager, Employee bool
}

func (s Scope) ManagesOrganization() bool { return s.Administrator || s.Organization == Owner }
func (s Scope) ManagesBranch() bool       { return s.ManagesOrganization() || s.Manager }
func (s Scope) ReviewsBranches() bool     { return s.Administrator || s.Sysadmin }

type change struct {
	actor, branch, target int
	slug                  string
	role                  Role
	action                string
	position              string
}

func allowed(c change, s Scope) bool {
	if c.actor <= 0 {
		return false
	}
	if s.Administrator {
		return true
	}
	switch c.role {
	case Administrator:
		return false
	case Sysadmin:
		return s.Sysadmin
	case Owner, Analyst, Manager:
		return s.Organization == Owner
	case Employee:
		return s.ManagesBranch()
	default:
		return false
	}
}

type roleStore interface {
	withRoleScope(context.Context, change, func(roleUnit, Scope) (any, error)) (any, error)
}
type roleUnit interface {
	apply(context.Context, change, Scope) (any, error)
}
type service struct{ store roleStore }

func (s service) run(ctx context.Context, c change) (any, error) {
	if c.action != "list" && c.target < 1 {
		return nil, ErrInvalid
	}
	if (c.role == Manager || c.role == Employee) && c.branch < 1 {
		return nil, ErrInvalid
	}
	if c.role == Employee && c.action == "grant" {
		switch strings.ToUpper(c.position) {
		case "WAITER", "COOK", "CASHIER", "BARTENDER", "HOST", "OTHER":
			c.position = strings.ToUpper(c.position)
		default:
			return nil, ErrInvalid
		}
	}
	return s.store.withRoleScope(ctx, c, func(tx roleUnit, scope Scope) (any, error) {
		if !allowed(c, scope) {
			return nil, ErrForbidden
		}
		return tx.apply(ctx, c, scope)
	})
}
