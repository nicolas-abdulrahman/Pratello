package restaurants

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/lib/pq"
	"pratello/server/internal/user"
)

type postgres struct{ db *sql.DB }

type sqlUnit struct{ tx *sql.Tx }

func (p postgres) transact(ctx context.Context, write bool, fn func(transaction) error) error {
	if p.db == nil {
		return errUnavailable
	}
	options := &sql.TxOptions{ReadOnly: !write}
	if !write {
		options.Isolation = sql.LevelRepeatableRead
	}
	tx, err := p.db.BeginTx(ctx, options)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = fn(sqlUnit{tx}); err != nil {
		return err
	}
	return tx.Commit()
}

func databaseError(err error) error {
	var pg *pq.Error
	if errors.Is(err, sql.ErrNoRows) {
		return errMissing
	}
	if errors.As(err, &pg) && pg.Code == "23505" {
		return errConflict
	}
	return err
}

func (u sqlUnit) scope(ctx context.Context, r request, lock bool) (access, error) {
	var a access
	if r.slug == "" {
		return a, nil
	}
	q := `SELECT id FROM restaurant WHERE name_slug=$1 AND archived_at IS NULL`
	if lock {
		q += ` FOR UPDATE`
	}
	if err := u.tx.QueryRowContext(ctx, q, r.slug).Scan(&a.orgID); err != nil {
		return a, databaseError(err)
	}
	if r.actor > 0 {
		s, err := user.ResolveScope(ctx, u.tx, r.actor, a.orgID, r.branch)
		if err != nil {
			return a, err
		}
		a.orgRole, a.sysadmin, a.administrator, a.manager, a.employee = string(s.Organization), s.Sysadmin, s.Administrator, s.Manager, s.Employee
	}
	if r.branch > 0 {
		err := u.tx.QueryRowContext(ctx, `SELECT b.cnpj_verified AND b.is_active
		 FROM physical_restaurant b WHERE b.id=$1 AND b.restaurant_id=$2 AND b.archived_at IS NULL`, r.branch, a.orgID).Scan(&a.published)
		if err != nil {
			return a, databaseError(err)
		}
	}
	return a, nil
}

const orgFields = `r.name_slug AS slug, r.name, COALESCE(r.description,'') AS description, r.archived_at, r.created_at, r.updated_at`

const addressFields = `jsonb_strip_nulls(jsonb_build_object('cep',ad."CEP",'street',ad.street,'number',ad.number,'apt',ad.apt,'complement',ad.complementary,'city',ad.city,'state',ad.state))`

const branchFields = `b.id, r.name_slug AS restaurant_slug, b.name, b.cnpj, b.owner_cnpj, b.cnpj_verified, b.is_active, b.created_at, b.updated_at, ` + addressFields + ` AS address`

const branchFrom = ` FROM physical_restaurant b JOIN restaurant r ON r.id=b.restaurant_id JOIN "Address" ad ON ad.id=b.address_id`

const phoneFields = `id, physical_restaurant_id AS branch_id, phone AS phone_number, phone_type`

const tableFields = `id, physical_restaurant_id AS branch_id, number, seats, display_order, is_active`

func (u sqlUnit) json(ctx context.Context, q string, args ...any) (json.RawMessage, error) {
	var result json.RawMessage
	err := u.tx.QueryRowContext(ctx, q, args...).Scan(&result)
	return result, databaseError(err)
}

func (u sqlUnit) one(ctx context.Context, q string, args ...any) (json.RawMessage, error) {
	return u.json(ctx, `SELECT row_to_json(x) FROM (`+q+`) x`, args...)
}

func (u sqlUnit) returning(ctx context.Context, write, projection string, args ...any) (json.RawMessage, error) {
	return u.json(ctx, `WITH changed AS (`+write+` RETURNING *) SELECT row_to_json(x) FROM (SELECT `+projection+`) x`, args...)
}

func (u sqlUnit) many(ctx context.Context, q string, args ...any) (json.RawMessage, error) {
	return u.json(ctx, `SELECT COALESCE(jsonb_agg(x),'[]') FROM (`+q+`) x`, args...)
}

func (u sqlUnit) page(ctx context.Context, q string, r request, filter any, extra ...any) (any, error) {
	order := map[operation]string{listOrg: "slug", listBranch: "id"}[r.op]
	args := append([]any{filter, r.size, int64(r.page-1) * int64(r.size), r.page}, extra...)
	return u.json(ctx, `WITH rows AS (`+q+`), subset AS (SELECT * FROM rows ORDER BY `+order+` LIMIT $2 OFFSET $3)
	SELECT jsonb_build_object('data',COALESCE((SELECT jsonb_agg(subset ORDER BY `+order+`) FROM subset),'[]'),'page',$4::int,'page_size',$2::int,'total',(SELECT count(*) FROM rows))`, args...)
}

func (u sqlUnit) changed(ctx context.Context, q string, args ...any) error {
	res, err := u.tx.ExecContext(ctx, q, args...)
	if err != nil {
		return databaseError(err)
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return errMissing
	}
	return err
}

func (u sqlUnit) organization(ctx context.Context, id int) (json.RawMessage, error) {
	return u.one(ctx, `SELECT `+orgFields+` FROM restaurant r WHERE r.id=$1`, id)
}

func (u sqlUnit) branch(ctx context.Context, id int, private bool) (json.RawMessage, error) {
	if private {
		return u.one(ctx, `SELECT `+branchFields+branchFrom+` WHERE b.id=$1`, id)
	}
	return u.json(ctx, `SELECT to_jsonb(x)-'owner_cnpj' FROM (SELECT `+branchFields+branchFrom+` WHERE b.id=$1) x`, id)
}

func (u sqlUnit) saveAddress(ctx context.Context, a address) (int, error) {
	var id int
	err := u.tx.QueryRowContext(ctx, `INSERT INTO "Address"("CEP",street,number,apt,complementary,city,state) VALUES($1,$2,$3,NULLIF($4,''),NULLIF($5,''),NULLIF($6,''),NULLIF($7,'')) RETURNING id`, a.CEP, a.Street, a.Number, a.Apt, a.Complement, a.City, a.State).Scan(&id)
	return id, databaseError(err)
}

func (u sqlUnit) execute(ctx context.Context, r request, a access) (any, error) {
	switch r.op {
	case listOrg:
		return u.page(ctx, `SELECT `+orgFields+` FROM restaurant r WHERE r.archived_at IS NULL AND $1::bool`, r, true)
	case myOrg:
		return u.many(ctx, `WITH scopes AS (
		 SELECT restaurant_id,account_type::text AS role,NULL::int AS branch_id FROM account_restaurant WHERE account_id=$1
		 UNION ALL SELECT b.restaurant_id,'MANAGER',b.id FROM branch_manager m JOIN physical_restaurant b ON b.id=m.branch_id WHERE m.account_id=$1 AND b.archived_at IS NULL
		 UNION ALL SELECT b.restaurant_id,'EMPLOYEE',b.id FROM employee e JOIN physical_restaurant b ON b.id=e.physical_restaurant_id WHERE e.account_id=$1 AND e.is_active AND b.archived_at IS NULL
		), assignments AS (SELECT restaurant_id,role,COALESCE(array_agg(branch_id ORDER BY branch_id) FILTER (WHERE branch_id IS NOT NULL),'{}') AS branch_ids FROM scopes GROUP BY restaurant_id,role)
		 SELECT row_to_json(x) AS restaurant,m.role,m.branch_ids FROM assignments m JOIN LATERAL (SELECT `+orgFields+` FROM restaurant r WHERE r.id=m.restaurant_id AND r.archived_at IS NULL) x ON true ORDER BY x.slug,m.role`, r.actor)
	case getOrg:
		return u.organization(ctx, a.orgID)
	case createOrg:
		b := r.body.(*newOrg)
		var id int
		if err := u.tx.QueryRowContext(ctx, `INSERT INTO restaurant(name_slug,name,description) VALUES($1,$2,$3) RETURNING id`, b.Slug, b.Name, b.Description).Scan(&id); err != nil {
			return nil, databaseError(err)
		}
		if err := u.changed(ctx, `INSERT INTO account_restaurant(account_id,restaurant_id,account_type) VALUES($1,$2,'OWNER')`, r.actor, id); err != nil {
			return nil, err
		}
		return u.organization(ctx, id)
	case updateOrg:
		b := r.body.(*details)
		if err := u.changed(ctx, `UPDATE restaurant SET name=$2,description=$3 WHERE id=$1`, a.orgID, b.Name, b.Description); err != nil {
			return nil, err
		}
		return u.organization(ctx, a.orgID)
	case archiveOrg:
		return nil, u.changed(ctx, `UPDATE restaurant SET archived_at=now() WHERE id=$1`, a.orgID)
	case listBranch:
		return u.page(ctx, `SELECT b.id, r.name_slug AS restaurant_slug, b.name,b.cnpj,b.cnpj_verified,b.is_active,b.created_at,b.updated_at,`+addressFields+` AS address`+branchFrom+`
		 WHERE b.restaurant_id=$1 AND b.archived_at IS NULL AND ((b.is_active AND b.cnpj_verified) OR $5::bool OR EXISTS(SELECT 1 FROM branch_manager WHERE branch_id=b.id AND account_id=$6))`, r, a.orgID, a.managesOrg() || a.sysadmin, r.actor)
	case getBranch:
		return u.branch(ctx, r.branch, a.managesOrg() || a.sysadmin)
	case createBranch:
		b := r.body.(*branchInput)
		id, err := u.saveAddress(ctx, b.Address)
		if err != nil {
			return nil, err
		}
		var branch int
		if err = u.tx.QueryRowContext(ctx, `INSERT INTO physical_restaurant(restaurant_id,address_id,name,cnpj,owner_cnpj) VALUES($1,$2,$3,$4,$5) RETURNING id`, a.orgID, id, b.Name, b.CNPJ, b.OwnerCNPJ).Scan(&branch); err != nil {
			return nil, databaseError(err)
		}
		return u.branch(ctx, branch, true)
	case verifyBranch:
		if err := u.changed(ctx, `UPDATE physical_restaurant SET cnpj_verified=true WHERE id=$1 AND cnpj ~ '^[0-9]{14}$' AND owner_cnpj ~ '^[0-9]{14}$'`, r.branch); err != nil {
			return nil, err
		}
		return u.branch(ctx, r.branch, true)
	case updateBranch:
		if err := u.changed(ctx, `UPDATE physical_restaurant SET name=$2 WHERE id=$1`, r.branch, r.body.(*details).Name); err != nil {
			return nil, err
		}
		return u.branch(ctx, r.branch, a.managesOrg())
	case archiveBranch:
		return nil, u.changed(ctx, `UPDATE physical_restaurant SET archived_at=now(),is_active=false WHERE id=$1`, r.branch)
	case setAddress:
		b := r.body.(*address)
		id, err := u.saveAddress(ctx, *b)
		if err != nil {
			return nil, err
		}
		if err = u.changed(ctx, `UPDATE physical_restaurant SET address_id=$2 WHERE id=$1`, r.branch, id); err != nil {
			return nil, err
		}
		return b, nil
	case listPhones:
		return u.many(ctx, `SELECT `+phoneFields+` FROM restaurant_phone WHERE physical_restaurant_id=$1 ORDER BY id`, r.branch)
	case addPhone:
		b := r.body.(*phoneInput)
		return u.returning(ctx, `INSERT INTO restaurant_phone(physical_restaurant_id,phone,phone_type) VALUES($1,$2,$3)`, phoneFields+` FROM changed`, r.branch, b.Number, b.Type)
	case removePhone:
		return nil, u.changed(ctx, `DELETE FROM restaurant_phone WHERE physical_restaurant_id=$1 AND id=$2`, r.branch, r.target)
	case getHours:
		return u.json(ctx, `SELECT operating_hours FROM physical_restaurant WHERE id=$1`, r.branch)
	case setHours:
		data, err := json.Marshal(r.body)
		if err != nil {
			return nil, err
		}
		return dataAsJSON(data), u.changed(ctx, `UPDATE physical_restaurant SET operating_hours=$2 WHERE id=$1`, r.branch, string(data))
	case listTables:
		return u.many(ctx, `SELECT `+tableFields+` FROM dining_table WHERE physical_restaurant_id=$1 AND archived_at IS NULL ORDER BY display_order,number`, r.branch)
	case addTable, updateTable:
		b := r.body.(*tableInput)
		if r.op == addTable {
			return u.returning(ctx, `INSERT INTO dining_table(physical_restaurant_id,number,seats,display_order,is_active) VALUES($1,$2,$3,$4,COALESCE($5,true))`, tableFields+` FROM changed`, r.branch, b.Number, b.Seats, b.Order, b.Active)
		}
		return u.returning(ctx, `UPDATE dining_table SET number=$2,seats=$3,display_order=$4,is_active=COALESCE($5,is_active) WHERE physical_restaurant_id=$1 AND id=$6 AND archived_at IS NULL`, tableFields+` FROM changed`, r.branch, b.Number, b.Seats, b.Order, b.Active, r.target)
	case removeTable:
		return nil, u.changed(ctx, `UPDATE dining_table SET archived_at=now(),is_active=false WHERE physical_restaurant_id=$1 AND id=$2 AND archived_at IS NULL`, r.branch, r.target)
	}
	return nil, errInvalid
}

func dataAsJSON(data []byte) json.RawMessage { return data }

func (u sqlUnit) audit(ctx context.Context, r request, a access, result any) error {
	role := a.orgRole
	if r.op == createOrg {
		role = "OWNER"
		if err := u.tx.QueryRowContext(ctx, `SELECT id FROM restaurant WHERE name_slug=$1`, r.body.(*newOrg).Slug).Scan(&a.orgID); err != nil {
			return err
		}
	} else if a.administrator {
		role = "ADMINISTRATOR"
	} else if r.op == verifyBranch {
		role = "SYSADMIN"
	} else if !a.managesOrg() && a.manager {
		role = "MANAGER"
	}
	entityID := r.target
	if data, ok := result.(json.RawMessage); ok && (r.op == createBranch || r.op == addTable || r.op == addPhone) {
		var created struct {
			ID int `json:"id"`
		}
		if err := json.Unmarshal(data, &created); err != nil {
			return err
		}
		entityID = created.ID
		if r.op == createBranch {
			r.branch = created.ID
		}
	}
	if entityID == 0 {
		entityID = r.branch
	}
	if entityID == 0 {
		entityID = a.orgID
	}
	metadata, err := json.Marshal(r.body)
	if err != nil {
		return err
	}
	return u.changed(ctx, `INSERT INTO audit_entry(actor_account_id,actor_role,restaurant_id,branch_id,action,entity_type,entity_id,metadata) VALUES($1,$2,$3,NULLIF($4,0),$5::text,split_part($5::text,'.',1),$6::int::text,$7)`, r.actor, role, a.orgID, r.branch, r.op, entityID, string(metadata))
}
