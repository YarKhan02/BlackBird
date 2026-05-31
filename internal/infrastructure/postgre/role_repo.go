package postgre

import (
	"context"
	"database/sql"

	"github.com/YarKhan02/BlackBird/internal/domain/role"
	"github.com/google/uuid"
)

type RoleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) ListGlobalRoles(ctx context.Context) ([]role.GlobalRole, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name FROM global_roles ORDER BY id`)
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

func (r *RoleRepository) AddGlobalRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO user_global_roles (user_id, role_id)
		SELECT $1, id FROM global_roles WHERE name = $2
		ON CONFLICT DO NOTHING
	`, userID, roleName)
	if err != nil {
		return err
	}
	return ensureRoleExists(ctx, r.db, roleName, res)
}

func (r *RoleRepository) RemoveGlobalRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM user_global_roles
		WHERE user_id = $1 AND role_id = (SELECT id FROM global_roles WHERE name = $2)
	`, userID, roleName)
	if err != nil {
		return err
	}
	return ensureRoleExists(ctx, r.db, roleName, res)
}

func (r *RoleRepository) GetUserGlobalRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT gr.name
		FROM user_global_roles ugr
		JOIN global_roles gr ON gr.id = ugr.role_id
		WHERE ugr.user_id = $1
		ORDER BY gr.id
	`, userID)
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
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_app_roles (user_id, app_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`, userID, appID, roleName)
	return err
}

func (r *RoleRepository) RemoveUserAppRole(ctx context.Context, userID uuid.UUID, appID uuid.UUID, roleName string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM user_app_roles
		WHERE user_id = $1 AND app_id = $2 AND role = $3
	`, userID, appID, roleName)
	return err
}

func (r *RoleRepository) GetAllUserAppRoles(ctx context.Context, userID uuid.UUID) (map[string][]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.client_id, uar.role
		FROM user_app_roles uar
		JOIN registered_apps a ON a.id = uar.app_id
		WHERE uar.user_id = $1
		ORDER BY a.client_id
	`, userID)
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
	err = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM global_roles WHERE name = $1)`, roleName).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return role.ErrRoleNotFound
	}
	return nil
}
