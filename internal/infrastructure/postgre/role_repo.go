package postgre

import (
	"context"
	"database/sql"
	_ "embed"

	"github.com/YarKhan02/BlackBird/internal/domain/role"
	"github.com/google/uuid"
)

type RoleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

//go:embed sql/role_list_global.sql
var roleListGlobalSQL string

//go:embed sql/role_id_by_name.sql
var roleIDByNameSQL string

//go:embed sql/role_add_global.sql
var roleAddGlobalSQL string

//go:embed sql/role_remove_global.sql
var roleRemoveGlobalSQL string

//go:embed sql/role_get_user_global.sql
var roleGetUserGlobalSQL string

//go:embed sql/role_add_user_app.sql
var roleAddUserAppSQL string

//go:embed sql/role_remove_user_app.sql
var roleRemoveUserAppSQL string

//go:embed sql/role_get_all_user_app.sql
var roleGetAllUserAppSQL string

//go:embed sql/role_exists.sql
var roleExistsSQL string

func (r *RoleRepository) ListGlobalRoles(ctx context.Context) ([]role.GlobalRole, error) {
	rows, err := r.db.QueryContext(ctx, roleListGlobalSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []role.GlobalRole
	for rows.Next() {
		var item role.GlobalRole
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *RoleRepository) GetGlobalRoleIDByName(ctx context.Context, roleName string) (uuid.UUID, error) {
	var id uuid.UUID

	err := r.db.QueryRowContext(ctx, roleIDByNameSQL, roleName).Scan(&id)

	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (r *RoleRepository) AddGlobalRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	res, err := r.db.ExecContext(ctx, roleAddGlobalSQL, userID, roleName)
	if err != nil {
		return err
	}
	return ensureRoleExists(ctx, r.db, roleName, res)
}

func (r *RoleRepository) RemoveGlobalRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	res, err := r.db.ExecContext(ctx, roleRemoveGlobalSQL, userID, roleName)
	if err != nil {
		return err
	}
	return ensureRoleExists(ctx, r.db, roleName, res)
}

func (r *RoleRepository) GetUserGlobalRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, roleGetUserGlobalSQL, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		roles = append(roles, name)
	}
	return roles, rows.Err()
}

func (r *RoleRepository) AddUserAppRole(ctx context.Context, userID uuid.UUID, appID uuid.UUID, roleName string) error {
	_, err := r.db.ExecContext(ctx, roleAddUserAppSQL, userID, appID, roleName)
	return err
}

func (r *RoleRepository) RemoveUserAppRole(ctx context.Context, userID uuid.UUID, appID uuid.UUID, roleName string) error {
	_, err := r.db.ExecContext(ctx, roleRemoveUserAppSQL, userID, appID, roleName)
	return err
}

func (r *RoleRepository) GetAllUserAppRoles(ctx context.Context, userID uuid.UUID) (map[string][]string, error) {
	rows, err := r.db.QueryContext(ctx, roleGetAllUserAppSQL, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string][]string)
	for rows.Next() {
		var clientID string
		var roleName string
		if err := rows.Scan(&clientID, &roleName); err != nil {
			return nil, err
		}
		out[clientID] = append(out[clientID], roleName)
	}
	return out, rows.Err()
}

func ensureRoleExists(ctx context.Context, db *sql.DB, roleName string, res sql.Result) error {
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows > 0 {
		return nil
	}

	var exists bool
	err = db.QueryRowContext(ctx, roleExistsSQL, roleName).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return role.ErrRoleNotFound
	}
	return nil
}
