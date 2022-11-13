package postgres

import (
	"context"
	"time"

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

func (r *postgresRepository) AddExpanse(ctx context.Context, expanse *model.Expanse) (*model.Expanse, error) {
	const (
		addReserveQuery = `
			INSERT INTO expanses (
				fromId,
				toId,
				ts,
				serviceId,
				orderId,
				cost,
				status
			)
			VALUES (
				$1,
				$2, 
				$3,
				$4,
				$5,
				$6,
				$7
			)
			RETURNING id;
		`
	)

	logger.Debug("executing postgres query...", "query", addReserveQuery, "query_args", []interface{}{expanse.From.ID, expanse.To.ID, expanse.Ts, expanse.ServiceID, expanse.OrderID, expanse.Cost.Amount, expanse.Status})

	var expanseID int64
	err := r.db.GetContext(ctx, &expanseID, addReserveQuery,
		expanse.From.ID,
		expanse.To.ID,
		*(expanse.Ts),
		expanse.ServiceID,
		expanse.OrderID,
		expanse.Cost.Amount,
		expanse.Status,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Add Expanse Query")
	}

	logger.Debug("executing postgres query finished", "query", addReserveQuery, "query_args", []interface{}{expanse.From.ID, expanse.To.ID, expanse.Ts, expanse.ServiceID, expanse.OrderID, expanse.Cost.Amount, expanse.Status}, "query_result", expanseID)

	expanse.ID = expanseID

	return expanse, nil
}

func (r *postgresRepository) GetExpanse(ctx context.Context, expanse *model.Expanse) (*model.Expanse, error) {
	const (
		getExpanseQuery = `
			SELECT id, fromId, toId, ts, serviceId, orderId, cost, status
			FROM expanses
			WHERE id = $1;
		`
	)

	logger.Debug("getting postgres query...", "query", getExpanseQuery, "query_args", []interface{}{expanse.ID})

	type exp struct {
		Id        int64
		FromId    int64
		ToId      int64
		Ts        time.Time
		ServiceId int64
		OrderId   int64
		Cost      int64
		Status    int8
	}
	var gotExpanse exp

	err := r.db.GetContext(
		ctx,
		&gotExpanse,
		getExpanseQuery,
		expanse.ID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Get Expanse Query")
	}
	logger.Debug("getting postgres query finished", "query", getExpanseQuery, "query_args", []interface{}{expanse.ID}, "query_result", []interface{}{gotExpanse.Id, gotExpanse.FromId, gotExpanse.ToId, gotExpanse.Ts, gotExpanse.ServiceId, gotExpanse.OrderId, gotExpanse.Cost, gotExpanse.Status})

	returnExpanse := &model.Expanse{
		ID:        gotExpanse.Id,
		From:      &model.User{ID: gotExpanse.FromId},
		To:        &model.User{ID: gotExpanse.ToId},
		Ts:        &gotExpanse.Ts,
		ServiceID: gotExpanse.ServiceId,
		OrderID:   gotExpanse.OrderId,
		Cost:      &model.Money{Amount: gotExpanse.Cost},
		Status:    model.ExpanseStatus(gotExpanse.Status),
	}

	return returnExpanse, nil
}

func (r *postgresRepository) SetExpanseStatus(ctx context.Context, expanse *model.Expanse) error {
	const (
		setExpanseStatusQuery = `
			UPDATE expanses
			SET status = $2
			WHERE id = $1;
		`
	)

	logger.Debug("executing postgres query...", "query", setExpanseStatusQuery, "query_args", []interface{}{expanse.ID, expanse.Status})
	_, err := r.db.ExecContext(ctx, setExpanseStatusQuery,
		expanse.ID,
		expanse.Status,
	)
	if err != nil {
		return errors.Wrap(err, "Set Expanse Status Query")
	}
	logger.Debug("executing postgres query finished", "query", setExpanseStatusQuery, "query_args", []interface{}{expanse.ID, expanse.Status})

	return nil
}

func (r *postgresRepository) BuildReport(ctx context.Context, params *model.ReportParams) (*model.Report, error) {
	const (
		selectExpansesQueryBeg = `
			SELECT 
			id, fromId, toId, ts, serviceId, orderId, cost, status 
			FROM expanses
			WHERE fromId = $1 AND (ts >= $2 AND ts <= $3) AND status = $4`
		selectExpansesQuerySortTs      = " ORDER BY ts"
		selectExpansesQuerySortService = " ORDER BY serviceId"
		selectExpansesQueryEnd         = ";"
	)

	query := selectExpansesQueryBeg
	queryArgs := []interface{}{params.User.ID, *(params.Dur.Beg), *(params.Dur.End), model.StatusSuccess}
	if params.SortField == "timestamp" {
		query += selectExpansesQuerySortTs
	} else if params.SortField == "service" {
		query += selectExpansesQuerySortService
	}
	query += selectExpansesQueryEnd

	var expanses []struct {
		Id        int64
		FromId    int64
		ToId      int64
		Ts        *time.Time
		ServiceId int64
		OrderId   int64
		Cost      int64
		Status    int8
	}
	logger.Debug("selecting postgres query...", "query", query, "query_args", queryArgs)
	err := r.db.SelectContext(ctx, &expanses, query,
		queryArgs...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Select Expanses Query Failed")
	}
	logger.Debug("selecting postgres query finished", "query", query, "query_args", "query_args", queryArgs, "query_result", expanses)

	l := int64(len(expanses))
	beg := params.PageSize * (params.Page - 1)
	if beg >= l {
		return &model.Report{}, nil
	}
	end := beg + params.PageSize
	if end >= l {
		end = l
	}

	expanses = expanses[beg:end]

	gotExpanses := make([]*model.Expanse, len(expanses))
	for i, expanse := range expanses {
		expanse := expanse
		gotExpanses[i] = &model.Expanse{
			ID:        expanse.Id,
			From:      &model.User{ID: expanse.FromId},
			To:        &model.User{ID: expanse.ToId},
			Ts:        expanse.Ts,
			ServiceID: expanse.ServiceId,
			OrderID:   expanse.OrderId,
			Cost:      &model.Money{Amount: expanse.Cost},
			Status:    model.ExpanseStatus(expanse.Status),
		}
	}

	report, err := model.NewReport(gotExpanses)
	if err != nil {
		return nil, errors.Wrap(err, "Model.NewReport")
	}

	return report, nil
}

func (r *postgresRepository) BuildStats(ctx context.Context, params *model.StatisticsParams) (*model.Statistics, error) {
	const (
		selectExpansesQuery = `
			SELECT 
			serviceId, SUM(cost) 
			FROM expanses
			WHERE (ts >= $1 AND ts <= $2) AND status = $3
			GROUP BY serviceId;
		`
	)
	queryArgs := []interface{}{*(params.Dur.Beg), *(params.Dur.End), model.StatusSuccess}

	logger.Debug("selecting postgres query...", "query", selectExpansesQuery, "query_args", queryArgs)
	type Statistics struct {
		ServiceId int64 `db:"serviceid"`
		Cost      int64 `db:"sum"`
	}
	var stats []Statistics
	err := r.db.SelectContext(
		ctx,
		&stats,
		selectExpansesQuery,
		queryArgs...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Select Expanses Query")
	}
	logger.Debug("selecting postgres query finished", "query", selectExpansesQuery, "query_args", queryArgs, "query_result", stats)

	statistics, err := model.NewStatistics(nil)
	if err != nil {
		return nil, errors.Wrap(err, "Model.NewStatistics")
	}
	for _, s := range stats {
		if err = statistics.Add(s.ServiceId, &model.Money{Amount: s.Cost}); err != nil {
			return nil, errors.Wrap(err, "Statistics.Add")
		}
	}

	return statistics, nil
}
