package usecase

import (
	"context"

	"github.com/SokolovAleksandr/AvitoTestBackend/internal/model"
	"github.com/pkg/errors"
)

type Repository interface {
	AddReserve(context.Context, *model.Reserve) (*model.Reserve, error)
	GetReserve(context.Context, *model.Reserve) (*model.Reserve, error)
	DeleteReserve(context.Context, *model.Reserve) error
}

type BalanceUseCase interface {
	UpdateUserBalance(context.Context, *model.User, *model.Money) error
}

type useCase struct {
	rep       Repository
	balanceUC BalanceUseCase
}

func New(rep Repository, balanceUC BalanceUseCase) (*useCase, error) {
	return &useCase{
		rep:       rep,
		balanceUC: balanceUC,
	}, nil
}

func (u *useCase) AddReserve(ctx context.Context, reserve *model.Reserve) (*model.Reserve, error) {
	err := u.balanceUC.UpdateUserBalance(ctx, reserve.User, reserve.Size.Negative())
	if err != nil {
		return nil, errors.Wrap(err, "BalanceUseCase.UpdateUserBalance")
	}

	newReserve, err := u.rep.AddReserve(ctx, reserve)
	if err != nil {
		_ = u.balanceUC.UpdateUserBalance(ctx, reserve.User, reserve.Size)
		return nil, errors.Wrap(err, "Repository.AddReserve")
	}

	return newReserve, nil
}

func (u *useCase) getReserve(ctx context.Context, emptyReserve *model.Reserve) (*model.Reserve, error) {
	filledReserve, err := u.rep.GetReserve(ctx, emptyReserve)
	if err != nil {
		return nil, errors.Wrap(err, "Repository.GetReserve")
	}

	return filledReserve, err
}

func (u *useCase) ConfirmReserve(ctx context.Context, emptyReserve *model.Reserve) error {
	filledReserve, err := u.getReserve(ctx, emptyReserve)
	if err != nil {
		return errors.Wrap(err, "UseCase.GetReserve")
	}

	err = u.rep.DeleteReserve(ctx, filledReserve)
	if err != nil {
		return errors.Wrap(err, "Repository.DeleteReserve")
	}

	return nil
}

func (u *useCase) CancelReserve(ctx context.Context, emptyReserve *model.Reserve) error {
	filledReserve, err := u.getReserve(ctx, emptyReserve)
	if err != nil {
		return errors.Wrap(err, "UseCase.GetReserve")
	}

	err = u.balanceUC.UpdateUserBalance(ctx, filledReserve.User, filledReserve.Size)
	if err != nil {
		return errors.Wrap(err, "BalanceUseCase.UpdateUserBalance")
	}

	err = u.rep.DeleteReserve(ctx, filledReserve)
	if err != nil {
		_ = u.balanceUC.UpdateUserBalance(ctx, filledReserve.User, filledReserve.Size.Negative())
		return errors.Wrap(err, "Repository.DeleteReserve")
	}

	return nil
}
