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

func (r *postgresRepository) AddReserve(ctx context.Context, reserve *model.Reserve) (*model.Reserve, error) {
	const (
		addReserveQuery = `
			INSERT INTO reserves (
				userId, 
				size
			)
			VALUES (
				$1,
				$2
			)
			RETURNING id;
		`
	)

	logger.Debug("executing postgres query...", "query", addReserveQuery, "query_args", []interface{}{reserve.User.ID, reserve.Size.Amount})
	var reserveID int64
	err := r.db.GetContext(ctx, &reserveID, addReserveQuery,
		reserve.User.ID,
		reserve.Size.Amount,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Add Reserve Query")
	}

	logger.Debug("executing postgres query finished", "query", addReserveQuery, "query_args", []interface{}{reserve.User.ID, reserve.Size.Amount}, "query_result", reserveID)

	reserve.ID = reserveID

	return reserve, nil
}

func (r *postgresRepository) GetReserve(ctx context.Context, reserve *model.Reserve) (*model.Reserve, error) {
	const (
		getReserveQuery = `
			SELECT id, userId, size
			FROM reserves
			WHERE id = $1 OR (userId = $2 AND size = $3);
		`
	)

	logger.Debug("getting postgres query...", "query", getReserveQuery, "query_args", []interface{}{reserve.ID, reserve.User.ID, reserve.Size.Amount})

	type res struct {
		Id     int64
		UserId int64
		Size   int64
	}
	var gotReserve res

	err := r.db.GetContext(
		ctx,
		&gotReserve,
		getReserveQuery,
		reserve.ID,
		reserve.User.ID,
		reserve.Size.Amount,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Get Reserve Query")
	}
	logger.Debug("getting postgres query finished", "query", getReserveQuery, "query_args", []interface{}{reserve.ID, reserve.User.ID, reserve.Size.Amount}, "query_result", []interface{}{gotReserve.Id, gotReserve.UserId, gotReserve.Size})

	returnReserve := &model.Reserve{
		ID:   gotReserve.Id,
		User: &model.User{ID: gotReserve.UserId},
		Size: &model.Money{Amount: gotReserve.Size},
	}

	return returnReserve, nil
}

func (r *postgresRepository) DeleteReserve(ctx context.Context, reserve *model.Reserve) error {
	const (
		deleteReserveQuery = `
			DELETE FROM reserves
			WHERE id = $1 OR (userId = $2 AND size = $3);
		`
	)

	logger.Debug("executing postgres query...", "query", deleteReserveQuery, "query_args", []interface{}{reserve.ID, reserve.User.ID, reserve.Size.Amount})
	_, err := r.db.ExecContext(ctx, deleteReserveQuery,
		reserve.ID,
		reserve.User.ID,
		reserve.Size.Amount,
	)
	if err != nil {
		return errors.Wrap(err, "Delete Reserve Query")
	}
	logger.Debug("executing postgres query finished", "query", deleteReserveQuery, "query_args", []interface{}{reserve.ID, reserve.User.ID, reserve.Size.Amount})

	return nil
}
