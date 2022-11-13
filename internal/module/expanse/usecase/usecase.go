package usecase

import (
	"context"
	"time"

	"github.com/SokolovAleksandr/AvitoTestBackend/internal/model"
	"github.com/pkg/errors"
)

var (
	moveMoneyOrderCounter int64 = 0

	ErrNotEnoughBalance = errors.New("not enough balance")
	ErrAlreadyHandled   = errors.New("expanse already handled")
)

type Repository interface {
	AddExpanse(context.Context, *model.Expanse) (*model.Expanse, error)
	GetExpanse(context.Context, *model.Expanse) (*model.Expanse, error)
	SetExpanseStatus(context.Context, *model.Expanse) error
	BuildReport(context.Context, *model.ReportParams) (*model.Report, error)
	BuildStats(context.Context, *model.StatisticsParams) (*model.Statistics, error)
}

type ReserveUseCase interface {
	AddReserve(context.Context, *model.Reserve) (*model.Reserve, error)
	ConfirmReserve(context.Context, *model.Reserve) error
	CancelReserve(context.Context, *model.Reserve) error
}

type BalanceUseCase interface {
	GetUserBalance(context.Context, *model.User) (*model.Balance, error)
	UpdateUserBalance(context.Context, *model.User, *model.Money) error
}

type useCase struct {
	rep       Repository
	balanceUC BalanceUseCase
	reserveUC ReserveUseCase
}

func New(rep Repository, balanceUC BalanceUseCase, reserveUC ReserveUseCase) (*useCase, error) {
	return &useCase{
		rep:       rep,
		balanceUC: balanceUC,
		reserveUC: reserveUC,
	}, nil
}

func (u *useCase) AddExpanse(ctx context.Context, expanse *model.Expanse) (*model.Expanse, error) {
	var err error
	if err = u.checkBalance(ctx, expanse.From, expanse.To, expanse.Cost); err != nil {
		return nil, err
	}

	var addedReserve *model.Reserve
	if expanse.From.ID != expanse.To.ID {
		newReserve := &model.Reserve{User: expanse.From, Size: expanse.Cost}
		addedReserve, err = u.reserveUC.AddReserve(ctx, newReserve)
		if err != nil {
			return nil, errors.Wrap(err, "ReserveUseCase.AddReserve")
		}
	}

	expanse.Status = model.StatusPending
	addedExpanse, err := u.rep.AddExpanse(ctx, expanse)
	if err != nil {
		if expanse.From.ID != expanse.To.ID {
			_ = u.reserveUC.CancelReserve(ctx, addedReserve)
		}
		return nil, errors.Wrap(err, "Repository.AddExpanse")
	}

	return addedExpanse, nil
}

func (u *useCase) GetExpanse(ctx context.Context, expanse *model.Expanse) (*model.Expanse, error) {
	gotExpanse, err := u.rep.GetExpanse(ctx, expanse)
	if err != nil {
		return nil, errors.Wrap(err, "Repository.GetExpanse")
	}

	return gotExpanse, nil
}

func (u *useCase) ConfirmExpanse(ctx context.Context, emptyExpanse *model.Expanse) error {
	filledExpanse, err := u.GetExpanse(ctx, emptyExpanse)
	if err != nil {
		return errors.Wrap(err, "UseCase.GetExpanse")
	}

	if filledExpanse.Status != model.StatusPending {
		return ErrAlreadyHandled
	}

	filledExpanse.Status = model.StatusSuccess
	err = u.rep.SetExpanseStatus(ctx, filledExpanse)
	if err != nil {
		return errors.Wrap(err, "Repository.SetExpanseStatus")
	}

	if filledExpanse.From.ID != filledExpanse.To.ID {
		reserve := &model.Reserve{User: filledExpanse.From, Size: filledExpanse.Cost}
		err = u.reserveUC.ConfirmReserve(ctx, reserve)
		if err != nil {
			emptyExpanse.Status = model.StatusPending
			_ = u.rep.SetExpanseStatus(ctx, filledExpanse)
			return errors.Wrap(err, "ReserveUseCase.ConfirmReserve")
		}
	}

	err = u.balanceUC.UpdateUserBalance(ctx, filledExpanse.To, filledExpanse.Cost)
	if err != nil {
		if filledExpanse.From.ID != filledExpanse.To.ID {
			_ = u.balanceUC.UpdateUserBalance(ctx, filledExpanse.From, filledExpanse.Cost)
		}
		return errors.Wrap(err, "BalanceUseCase.UpdateUserBalance")
	}

	return nil
}

func (u *useCase) CancelExpanse(ctx context.Context, emptyExpanse *model.Expanse) error {
	filledExpanse, err := u.GetExpanse(ctx, emptyExpanse)
	if err != nil {
		return errors.Wrap(err, "UseCase.GetExpanse")
	}

	if filledExpanse.Status != model.StatusPending {
		return ErrAlreadyHandled
	}

	filledExpanse.Status = model.StatusCancel
	err = u.rep.SetExpanseStatus(ctx, filledExpanse)
	if err != nil {
		return errors.Wrap(err, "Repository.SetExpanseStatus")
	}

	reserve := &model.Reserve{User: filledExpanse.From, Size: filledExpanse.Cost}
	err = u.reserveUC.CancelReserve(ctx, reserve)
	if err != nil {
		emptyExpanse.Status = model.StatusPending
		_ = u.rep.SetExpanseStatus(ctx, filledExpanse)
		return errors.Wrap(err, "ReserveUseCase.CancelReserve")
	}

	return nil
}

func (u *useCase) checkBalance(ctx context.Context, from *model.User, to *model.User, money *model.Money) error {
	fromBalance, err := u.balanceUC.GetUserBalance(ctx, from)
	if err != nil {
		return errors.Wrap(err, "Repository.GetUserBalance")
	}

	if (from.ID != to.ID && money.More(fromBalance.Money)) ||
		(from.ID == to.ID && money.Negative().More(fromBalance.Money)) {
		return ErrNotEnoughBalance
	}

	return nil
}

func (u *useCase) MoveMoney(ctx context.Context, from *model.User, to *model.User, money *model.Money) (*model.Expanse, error) {
	if err := u.checkBalance(ctx, from, to, money); err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	newExpanse := &model.Expanse{
		From:      from,
		To:        to,
		Ts:        &now,
		ServiceID: 0,
		OrderID:   moveMoneyOrderCounter,
		Status:    model.StatusPending,
		Cost:      money,
	}

	newExpanse, err := u.AddExpanse(ctx, newExpanse)
	if err != nil {
		return nil, errors.Wrap(err, "UseCase.AddExpanse")
	}

	moveMoneyOrderCounter += 1

	return newExpanse, nil
}

func (u *useCase) BuildReport(ctx context.Context, params *model.ReportParams) (*model.Report, error) {
	report, err := u.rep.BuildReport(ctx, params)
	if err != nil {
		return nil, errors.Wrap(err, "Repository.BuildReport")
	}

	return report, nil
}

func (u *useCase) BuildStats(ctx context.Context, params *model.StatisticsParams) (*model.Statistics, error) {
	stats, err := u.rep.BuildStats(ctx, params)
	if err != nil {
		return nil, errors.Wrap(err, "Repository.BuildStats")
	}

	return stats, nil
}
