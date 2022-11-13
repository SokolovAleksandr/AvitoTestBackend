package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"

	handler "github.com/SokolovAleksandr/AvitoTestBackend/internal/handler/http/gin"
	"github.com/SokolovAleksandr/AvitoTestBackend/internal/logger"
	"github.com/SokolovAleksandr/AvitoTestBackend/internal/model"
	balance_rep "github.com/SokolovAleksandr/AvitoTestBackend/internal/module/balance/repository/postgres"
	balance_uc "github.com/SokolovAleksandr/AvitoTestBackend/internal/module/balance/usecase"
	expanse_rep "github.com/SokolovAleksandr/AvitoTestBackend/internal/module/expanse/repository/postgres"
	expanse_uc "github.com/SokolovAleksandr/AvitoTestBackend/internal/module/expanse/usecase"
	reserve_rep "github.com/SokolovAleksandr/AvitoTestBackend/internal/module/reserve/repository/postgres"
	reserve_uc "github.com/SokolovAleksandr/AvitoTestBackend/internal/module/reserve/usecase"
	user_rep "github.com/SokolovAleksandr/AvitoTestBackend/internal/module/user/repository/postgres"
	user_uc "github.com/SokolovAleksandr/AvitoTestBackend/internal/module/user/usecase"
)

type BalanceUseCase interface {
	InitUserBalance(context.Context, *model.User) error
	GetUserBalance(context.Context, *model.User) (*model.Balance, error)
	UpdateUserBalance(context.Context, *model.User, *model.Money) error
}

type UserUseCase interface {
	AddUser(context.Context, *model.User) error
}

type ReserveUseCase interface {
	AddReserve(context.Context, *model.Reserve) (*model.Reserve, error)
	ConfirmReserve(context.Context, *model.Reserve) error
	CancelReserve(context.Context, *model.Reserve) error
}

type ExpanseUseCase interface {
	AddExpanse(context.Context, *model.Expanse) (*model.Expanse, error)
	GetExpanse(context.Context, *model.Expanse) (*model.Expanse, error)
	ConfirmExpanse(context.Context, *model.Expanse) error
	CancelExpanse(context.Context, *model.Expanse) error
	MoveMoney(context.Context, *model.User, *model.User, *model.Money) (*model.Expanse, error)
	BuildReport(context.Context, *model.ReportParams) (*model.Report, error)
	BuildStats(context.Context, *model.StatisticsParams) (*model.Statistics, error)
}

type UseCase interface {
	BalanceUseCase
	UserUseCase
	ReserveUseCase
	ExpanseUseCase
}

type useCase struct {
	BalanceUseCase
	UserUseCase
	ReserveUseCase
	ExpanseUseCase
}

type Handler interface {
	Serve(port int) error
}

type App struct {
	port   int
	handle Handler
}

type RepositoryParams interface {
	GetHost() string
	GetPort() int
	GetUser() string
	GetPassword() string
	GetDBName() string
}

func New(port int, repParams RepositoryParams) (*App, error) {
	logger.Info("initing app...")
	defer logger.Info("initing app finished")

	logger.Info("initing DB...")
	db, err := initDB(repParams)
	if err != nil {
		return nil, errors.Wrap(err, "initDB")
	}
	logger.Info("initing DB finished")

	logger.Info("initing UseCase...")
	uc, err := initUC(db)
	if err != nil {
		return nil, errors.Wrap(err, "initUC")
	}
	logger.Info("initing UseCase finished")

	logger.Info("initing handler...")
	handle, err := initHandler(uc)
	if err != nil {
		return nil, errors.Wrap(err, "initHandler")
	}
	logger.Info("initing handler finished")

	return &App{
		port:   port,
		handle: handle,
	}, nil
}

func initDB(repParams RepositoryParams) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		repParams.GetHost(),
		repParams.GetPort(),
		repParams.GetUser(),
		repParams.GetPassword(),
		repParams.GetDBName(),
	)

	logger.Debug("opening connection to postgres...")
	postgresDB, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	logger.Debug("opening connection to postgres finished")

	return postgresDB, nil
}

func initUC(db *sqlx.DB) (UseCase, error) {
	balanceUC, err := initBalanceUC(db)
	if err != nil {
		return nil, errors.Wrap(err, "initBalanceUC")
	}

	userUC, err := initUserUC(db, balanceUC)
	if err != nil {
		return nil, errors.Wrap(err, "initUserUC")
	}

	reserveUC, err := initReserveUC(db, balanceUC)
	if err != nil {
		return nil, errors.Wrap(err, "initReserveUC")
	}

	expanseUC, err := initExpanseUC(db, balanceUC, reserveUC)
	if err != nil {
		return nil, errors.Wrap(err, "initExpanseUC")
	}

	return &useCase{
		balanceUC,
		userUC,
		reserveUC,
		expanseUC,
	}, nil
}

func initBalanceUC(db *sqlx.DB) (BalanceUseCase, error) {
	logger.Debug("initing balance Repository...")
	balanceRep, err := balance_rep.New(db)
	if err != nil {
		return nil, errors.Wrap(err, "balance_rep.New")
	}
	logger.Debug("initing balance Repository finished")

	logger.Debug("initing balance UseCase...")
	balanceUC, err := balance_uc.New(balanceRep)
	if err != nil {
		return nil, errors.Wrap(err, "balance_uc.New")
	}
	logger.Debug("initing balance UseCase finished")
	return balanceUC, nil
}

func initUserUC(db *sqlx.DB, balanceUC BalanceUseCase) (UserUseCase, error) {
	logger.Debug("initing user Repository...")
	userRep, err := user_rep.New(db)
	if err != nil {
		return nil, errors.Wrap(err, "user_rep.New")
	}
	logger.Debug("initing user Repository finished")

	logger.Debug("initing user UseCase...")
	userUC, err := user_uc.New(userRep, balanceUC)
	if err != nil {
		return nil, errors.Wrap(err, "user_uc.New")
	}
	logger.Debug("initing user UseCase finished")
	return userUC, nil
}

func initReserveUC(db *sqlx.DB, balanceUC BalanceUseCase) (ReserveUseCase, error) {
	logger.Debug("initing reserve Repository...")
	reserveRep, err := reserve_rep.New(db)
	if err != nil {
		return nil, errors.Wrap(err, "reserve_rep.New")
	}
	logger.Debug("initing reserve Repository finished")

	logger.Debug("initing reserve UseCase...")
	reserveUC, err := reserve_uc.New(reserveRep, balanceUC)
	if err != nil {
		return nil, errors.Wrap(err, "reserve_uc.New")
	}
	logger.Debug("initing reserve UseCase finished")
	return reserveUC, nil
}

func initExpanseUC(db *sqlx.DB, balanceUC BalanceUseCase, reserveUC ReserveUseCase) (ExpanseUseCase, error) {
	logger.Debug("initing expanse Repository...")
	expanseRep, err := expanse_rep.New(db)
	if err != nil {
		return nil, errors.Wrap(err, "expanse_rep.New")
	}
	logger.Debug("initing expanse Repository finished")

	logger.Debug("initing expanse UseCase...")
	expanseUC, err := expanse_uc.New(expanseRep, balanceUC, reserveUC)
	if err != nil {
		return nil, errors.Wrap(err, "expanse_uc.New")
	}
	logger.Debug("initing expanse UseCase finished")
	return expanseUC, nil
}

func initHandler(uc UseCase) (Handler, error) {
	handle, err := handler.New("debug", uc)
	if err != nil {
		return nil, errors.Wrap(err, "handler.New")
	}
	return handle, nil
}

func (a *App) Run() error {
	logger.Info("running app...")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go func(ctx context.Context, a *App) {
		logger.Info("starting http server", "port", a.port)

		if err := a.handle.Serve(a.port); err != nil {
			logger.Error("serving http server failed", "error", err.Error())
		}
	}(ctx, a)

	<-ctx.Done()
	return nil
}
