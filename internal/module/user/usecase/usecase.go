package usecase

import (
	"context"

	"github.com/SokolovAleksandr/AvitoTestBackend/internal/model"
	"github.com/pkg/errors"
)

type Repository interface {
	AddUser(context.Context, *model.User) error
}

type BalanceUseCase interface {
	InitUserBalance(context.Context, *model.User) error
}

type useCase struct {
	rep       Repository
	balanceUC BalanceUseCase
}

func New(repository Repository, balanceUC BalanceUseCase) (*useCase, error) {
	return &useCase{
		rep:       repository,
		balanceUC: balanceUC,
	}, nil
}

func (u *useCase) AddUser(ctx context.Context, user *model.User) error {
	err := u.rep.AddUser(ctx, user)
	if err != nil {
		return errors.Wrap(err, "Repository.AddUser")
	}

	err = u.balanceUC.InitUserBalance(ctx, user)
	if err != nil {
		return errors.Wrap(err, "BalanceUseCase.SetUserBalance")
	}

	return nil
}
