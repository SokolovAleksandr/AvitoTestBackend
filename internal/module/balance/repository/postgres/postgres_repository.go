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

func (r *postgresRepository) SetUserBalance(ctx context.Context, user *model.User, balance *model.Balance) error {
	const (
		setUserBalanceQuery = `
			INSERT INTO balances(userId, value)
			VALUES ($1, $2)
			ON CONFLICT (userId) 
			DO NOTHING;
		`
	)

	logger.Debug("executing postgres query...", "query", setUserBalanceQuery, "query_args", []interface{}{user.ID, balance.Money.Amount})
	_, err := r.db.ExecContext(ctx, setUserBalanceQuery, user.ID, balance.Money.Amount)
	if err != nil {
		return errors.Wrap(err, "Set User Balance Query")
	}
	logger.Debug("executing postgres query finished", "query", setUserBalanceQuery, "query_args", []interface{}{user.ID, balance.Money.Amount})

	return nil
}

func (r *postgresRepository) GetUserBalance(ctx context.Context, user *model.User) (*model.Balance, error) {
	const (
		getUserBalanceQuery = `
			SELECT value
			FROM balances
			WHERE userId = $1;
		`
	)

	logger.Debug("getting postgres query...", "query", getUserBalanceQuery, "query_args", []interface{}{user.ID})
	var balanceValue int64
	err := r.db.GetContext(
		ctx,
		&balanceValue,
		getUserBalanceQuery,
		user.ID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Get User Balance Query")
	}
	logger.Debug("getting postgres query finished", "query", getUserBalanceQuery, "query_args", []interface{}{user.ID}, "query_result", balanceValue)

	return &model.Balance{Money: &model.Money{Amount: balanceValue}}, nil
}

func (r *postgresRepository) UpdateUserBalance(ctx context.Context, user *model.User, money *model.Money) error {
	const (
		updateUserBalanceQuery = `
			UPDATE balances
			SET value = value + $2
			WHERE userId = $1;
		`
	)

	logger.Debug("executing postgres query...", "query", updateUserBalanceQuery, "query_args", []interface{}{user.ID, money.Amount})
	_, err := r.db.ExecContext(
		ctx,
		updateUserBalanceQuery,
		user.ID,
		money.Amount,
	)
	if err != nil {
		return errors.Wrap(err, "Update User Balance Query")
	}
	logger.Debug("executing postgres query finished", "query", updateUserBalanceQuery, "query_args", []interface{}{user.ID, money.Amount})

	return nil
}
