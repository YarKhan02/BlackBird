package postgre

import (
	"context"
	"database/sql"
	_ "embed"

	"github.com/YarKhan02/BlackBird/internal/domain/app"
	"github.com/google/uuid"
)

//go:embed sql/app_create.sql
var appCreateSQL string

//go:embed sql/app_find_by_name.sql
var appFindByNameSQL string

//go:embed sql/app_find_by_id.sql
var appFindByIDSQL string

//go:embed sql/app_find_by_client_id.sql
var appFindByClientIDSQL string

//go:embed sql/app_list.sql
var appListSQL string

type AppRepository struct {
	db *sql.DB
}

func NewAppRepository(db *sql.DB) *AppRepository {
	return &AppRepository{db: db}
}

func (r *AppRepository) Create(ctx context.Context, a *app.App) error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	a.ID = id
	err = r.db.QueryRowContext(ctx, appCreateSQL,
		a.ID,
		a.ClientID,
		a.ClientSecretHash,
		a.Name,
		a.RedirectURIs,
		a.IsActive,
	).Scan(&a.CreatedAt)
	return err
}

func (r *AppRepository) FindByName(ctx context.Context, name string) (*app.App, error) {
	row := r.db.QueryRowContext(ctx, appFindByNameSQL, name)
	return scanApp(row)
}

func (r *AppRepository) FindByID(ctx context.Context, id uuid.UUID) (*app.App, error) {
	row := r.db.QueryRowContext(ctx, appFindByIDSQL, id)
	return scanApp(row)
}

func (r *AppRepository) FindByClientID(ctx context.Context, clientID string) (*app.App, error) {
	row := r.db.QueryRowContext(ctx, appFindByClientIDSQL, clientID)
	return scanApp(row)
}

func (r *AppRepository) List(ctx context.Context) ([]*app.App, error) {
	rows, err := r.db.QueryContext(ctx, appListSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*app.App, 0)
	for rows.Next() {
		a, err := scanApp(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func scanApp(scanner interface{ Scan(dest ...any) error }) (*app.App, error) {
	var a app.App
	err := scanner.Scan(
		&a.ID,
		&a.ClientID,
		&a.ClientSecretHash,
		&a.Name,
		&a.RedirectURIs,
		&a.IsActive,
		&a.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, app.ErrAppNotFound
		}
		return nil, err
	}
	return &a, nil
}
