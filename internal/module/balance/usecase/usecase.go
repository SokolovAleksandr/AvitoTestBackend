package usecase

import (
	"context"

	"github.com/SokolovAleksandr/AvitoTestBackend/internal/model"
	"github.com/pkg/errors"
)

type Repository interface {
	SetUserBalance(context.Context, *model.User, *model.Balance) error
	GetUserBalance(context.Context, *model.User) (*model.Balance, error)
	UpdateUserBalance(context.Context, *model.User, *model.Money) error
}

type useCase struct {
	rep Repository
}

func New(repository Repository) (*useCase, error) {
	return &useCase{
		rep: repository,
	}, nil
}

func (u *useCase) InitUserBalance(ctx context.Context, user *model.User) error {
	balance := &model.Balance{Money: &model.Money{Amount: 0}}

	err := u.rep.SetUserBalance(ctx, user, balance)
	if err != nil {
		return errors.Wrap(err, "Repository.SetUserBalance")
	}

	return nil
}

func (u *useCase) GetUserBalance(ctx context.Context, user *model.User) (*model.Balance, error) {
	balance, err := u.rep.GetUserBalance(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "Repository.GetUserBalance")
	}

	return balance, nil
}

func (u *useCase) UpdateUserBalance(ctx context.Context, user *model.User, money *model.Money) error {
	err := u.rep.UpdateUserBalance(ctx, user, money)
	if err != nil {
		return errors.Wrap(err, "Repository.FillUserBalance")
	}

	return nil
}
