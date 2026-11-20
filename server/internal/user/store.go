package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// ResolveScope reads authoritative assignments for this request, so a revoked
// grant stops authorizing requests even when the access token remains valid.
func ResolveScope(ctx context.Context, q queryer, actor, org, branch int) (Scope, error) {
	var s Scope
	if actor == 0 {
		return s, nil
	}
	err := q.QueryRowContext(ctx, `SELECT a.is_administrator,a.is_sysadmin,
		COALESCE((SELECT account_type::text FROM account_restaurant WHERE account_id=a.id AND restaurant_id=$2),''),
		EXISTS(SELECT 1 FROM branch_manager WHERE account_id=a.id AND branch_id=$3),
		EXISTS(SELECT 1 FROM employee WHERE account_id=a.id AND physical_restaurant_id=$3 AND is_active)
		FROM account a WHERE a.id=$1`, actor, org, branch).Scan(&s.Administrator, &s.Sysadmin, &s.Organization, &s.Manager, &s.Employee)
	return s, dbError(err)
}

func dbError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrMissing
	}
	var pg *pq.Error
	if errors.As(err, &pg) && pg.Code == "23505" {
		return ErrConflict
	}
	return err
}

type postgres struct{ db *sql.DB }
type sqlRoles struct {
	tx  *sql.Tx
	org int
}

func (p postgres) withRoleScope(ctx context.Context, c change, fn func(roleUnit, Scope) (any, error)) (any, error) {
	if p.db == nil {
		return nil, ErrUnavailable
	}
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	u := sqlRoles{tx: tx}
	if c.slug != "" {
		// Serialize ownership, branch assignments and restaurant writes on the same row.
		if err = tx.QueryRowContext(ctx, `SELECT id FROM restaurant WHERE name_slug=$1 AND archived_at IS NULL FOR UPDATE`, c.slug).Scan(&u.org); err != nil {
			return nil, dbError(err)
		}
		if c.branch != 0 {
			var id int
			if err = tx.QueryRowContext(ctx, `SELECT id FROM physical_restaurant WHERE id=$1 AND restaurant_id=$2 AND archived_at IS NULL`, c.branch, u.org).Scan(&id); err != nil {
				return nil, dbError(err)
			}
		}
	} else if c.action != "list" {
		// Global role changes use one lock, including administrator revocations.
		if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(914028)`); err != nil {
			return nil, err
		}
	}
	scope, err := ResolveScope(ctx, tx, c.actor, u.org, c.branch)
	if err != nil {
		return nil, err
	}
	result, err := fn(u, scope)
	if err != nil {
		return nil, err
	}
	return result, tx.Commit()
}

func (u sqlRoles) list(ctx context.Context, c change) (any, error) {
	var q string
	var args []any
	switch c.role {
	case Administrator, Sysadmin:
		flag := "is_sysadmin"
		if c.role == Administrator {
			flag = "is_administrator"
		}
		q = `SELECT id,id AS account_id,created_at FROM account WHERE ` + flag + ` ORDER BY id`
	case Owner, Analyst:
		q = `SELECT m.account_id,a.email AS account_email,a.phone AS account_phone,m.account_type AS role,m.created_at FROM account_restaurant m JOIN account a ON a.id=m.account_id WHERE m.restaurant_id=$1 AND m.account_type=$2 ORDER BY m.account_id`
		args = []any{u.org, c.role}
	case Manager:
		q = `SELECT account_id,branch_id,'MANAGER' AS role FROM branch_manager WHERE branch_id=$1 ORDER BY account_id`
		args = []any{c.branch}
	case Employee:
		q = `SELECT id,physical_restaurant_id AS branch_id,account_id,name,position,is_active,created_at FROM employee WHERE physical_restaurant_id=$1 AND is_active ORDER BY id`
		args = []any{c.branch}
	}
	var result json.RawMessage
	err := u.tx.QueryRowContext(ctx, `SELECT COALESCE(json_agg(x),'[]') FROM (`+q+`) x`, args...).Scan(&result)
	return result, err
}

func (u sqlRoles) apply(ctx context.Context, c change, s Scope) (any, error) {
	if c.action == "list" {
		return u.list(ctx, c)
	}
	if c.action == "grant" {
		var eligible bool
		err := u.tx.QueryRowContext(ctx, `SELECT (email_verified AND password IS NOT NULL) OR phone_verified FROM account WHERE id=$1 FOR SHARE`, c.target).Scan(&eligible)
		if err != nil {
			return nil, dbError(err)
		}
		if !eligible {
			return nil, ErrMissing
		}
	}
	var q string
	var args []any
	grant := c.action == "grant"
	switch c.role {
	case Administrator, Sysadmin:
		flag := "is_sysadmin"
		if c.role == Administrator {
			flag = "is_administrator"
		}
		if !grant && c.role == Administrator {
			var last bool
			if err := u.tx.QueryRowContext(ctx, `SELECT is_administrator AND (SELECT count(*) FROM account WHERE is_administrator)<=1 FROM account WHERE id=$1`, c.target).Scan(&last); err != nil {
				return nil, dbError(err)
			}
			if last {
				return nil, ErrConflict
			}
		}
		q = `UPDATE account SET ` + flag + `=$2 WHERE id=$1`
		args = []any{c.target, grant}
	case Owner, Analyst:
		if grant {
			q = `INSERT INTO account_restaurant(account_id,restaurant_id,account_type) VALUES($1,$2,$3)`
		} else {
			if c.role == Owner {
				var owners int
				var target bool
				if err := u.tx.QueryRowContext(ctx, `SELECT count(*),COALESCE(bool_or(account_id=$2),false) FROM account_restaurant WHERE restaurant_id=$1 AND account_type='OWNER'`, u.org, c.target).Scan(&owners, &target); err != nil {
					return nil, err
				}
				if !target {
					return nil, ErrMissing
				}
				if owners == 1 {
					return nil, ErrConflict
				}
			}
			q = `DELETE FROM account_restaurant WHERE account_id=$1 AND restaurant_id=$2 AND account_type=$3`
		}
		args = []any{c.target, u.org, c.role}
	case Manager:
		if grant {
			q = `INSERT INTO branch_manager(account_id,branch_id) VALUES($1,$2) ON CONFLICT DO NOTHING`
		} else {
			q = `DELETE FROM branch_manager WHERE account_id=$1 AND branch_id=$2`
		}
		args = []any{c.target, c.branch}
	case Employee:
		if grant {
			q = `INSERT INTO employee(account_id,physical_restaurant_id,name,cpf,position,hired_at) SELECT id,$2,COALESCE(NULLIF(name,''),'Employee '||id),cpf,$3,current_date FROM account WHERE id=$1 ON CONFLICT (account_id,physical_restaurant_id) DO UPDATE SET is_active=true,position=EXCLUDED.position WHERE NOT employee.is_active`
			args = []any{c.target, c.branch, c.position}
		} else {
			q = `UPDATE employee SET is_active=false WHERE id=$1 AND physical_restaurant_id=$2 AND is_active`
			args = []any{c.target, c.branch}
		}
	}
	res, err := u.tx.ExecContext(ctx, q, args...)
	if err != nil {
		return nil, dbError(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		if grant && c.role == Employee {
			return nil, ErrConflict
		}
		if grant && c.role == Manager {
			return nil, nil
		}
		return nil, ErrMissing
	}
	actorRole := s.Organization
	switch {
	case s.Administrator:
		actorRole = Administrator
	case c.role == Sysadmin:
		actorRole = Sysadmin
	case actorRole == "" && s.Manager:
		actorRole = Manager
	}
	metadata, err := json.Marshal(map[string]any{"role": c.role, "position": c.position})
	if err != nil {
		return nil, err
	}
	_, err = u.tx.ExecContext(ctx, `INSERT INTO audit_entry(actor_account_id,actor_role,restaurant_id,branch_id,action,entity_type,entity_id,metadata) VALUES($1,$2,NULLIF($3,0),NULLIF($4,0),$5,'role',$6,$7)`, c.actor, actorRole, u.org, c.branch, "user.role."+c.action, fmt.Sprint(c.target), string(metadata))
	if err != nil {
		return nil, err
	}
	if !grant || c.role == Manager {
		return nil, nil
	}
	var out json.RawMessage
	switch c.role {
	case Employee:
		err = u.tx.QueryRowContext(ctx, `SELECT row_to_json(x) FROM (SELECT id,physical_restaurant_id AS branch_id,account_id,name,position,is_active,created_at FROM employee WHERE account_id=$1 AND physical_restaurant_id=$2) x`, c.target, c.branch).Scan(&out)
	case Owner, Analyst:
		err = u.tx.QueryRowContext(ctx, `SELECT row_to_json(x) FROM (SELECT m.account_id,a.email AS account_email,a.phone AS account_phone,m.account_type AS role,m.created_at FROM account_restaurant m JOIN account a ON a.id=m.account_id WHERE m.account_id=$1 AND m.restaurant_id=$2) x`, c.target, u.org).Scan(&out)
	default:
		err = u.tx.QueryRowContext(ctx, `SELECT json_build_object('id',id,'account_id',id,'created_at',created_at) FROM account WHERE id=$1`, c.target).Scan(&out)
	}
	return out, err
}
