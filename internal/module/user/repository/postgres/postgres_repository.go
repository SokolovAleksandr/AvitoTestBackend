package postgres

import (
	"context"

	"github.com/SokolovAleksandr/AvitoTestBackend/internal/logger"
	"github.com/SokolovAleksandr/AvitoTestBackend/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type postgresRepository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) (*postgresRepository, error) {
	return &postgresRepository{db: db}, nil
}

func (r *postgresRepository) AddUser(ctx context.Context, user *model.User) error {
	const (
		addUserQuery = `
			INSERT INTO users (id)
			VALUES ($1)
			ON CONFLICT (id) 
			DO NOTHING;
		`
	)

	logger.Debug("executing postgres query...", "query", addUserQuery, "query_args", []interface{}{user.ID})
	_, err := r.db.ExecContext(ctx, addUserQuery, user.ID)
	if err != nil {
		return errors.Wrap(err, "Add User Query")
	}
	logger.Debug("executing postgres query finished", "query", addUserQuery, "query_args", []interface{}{user.ID})

	return nil
}
