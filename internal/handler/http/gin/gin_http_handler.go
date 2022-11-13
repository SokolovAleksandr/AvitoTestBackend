package handler

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	_ "github.com/SokolovAleksandr/AvitoTestBackend/docs"
	"github.com/SokolovAleksandr/AvitoTestBackend/internal/logger"
	"github.com/SokolovAleksandr/AvitoTestBackend/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	swagger_files "github.com/swaggo/files"
	gin_swagger "github.com/swaggo/gin-swagger"
)

var (
	ErrInvalidArgument       = errors.New("invalid path argument")
	ErrInvalidBody           = errors.New("invalid body parameters")
	ErrMoneyShouldBePositive = errors.New("money amount should be positive")
)

type UseCase interface {
	AddUser(context.Context, *model.User) error
	GetUserBalance(context.Context, *model.User) (*model.Balance, error)
	MoveMoney(context.Context, *model.User, *model.User, *model.Money) (*model.Expanse, error)

	AddExpanse(context.Context, *model.Expanse) (*model.Expanse, error)
	GetExpanse(context.Context, *model.Expanse) (*model.Expanse, error)
	ConfirmExpanse(context.Context, *model.Expanse) error
	CancelExpanse(context.Context, *model.Expanse) error

	BuildReport(context.Context, *model.ReportParams) (*model.Report, error)
	BuildStats(context.Context, *model.StatisticsParams) (*model.Statistics, error)
}

type Handler struct {
	uc     UseCase
	router *gin.Engine
}

// @title           Gin Expanses Service
// @version         1.0
// @description     A balances management service API in Go using Gin framework.

// @contact.name   Sokolov Aleksandr
// @contact.email  sokolov.alex5@yandex.ru

// @host      localhost:8080
// @BasePath  /
// @schemes http
func New(mode string, uc UseCase) (*Handler, error) {
	h := &Handler{uc: uc}

	gin.SetMode(mode)

	engine := gin.New()

	engine.Use(LoggerMiddleware())
	engine.Use(gin.Recovery())

	engine.GET("/swagger/*any", gin_swagger.WrapHandler(swagger_files.Handler))

	engine.GET("/file/:file_name", h.DownloadFile)

	engine.GET("/balance/:id", h.GetUserBalance)
	engine.PUT("/balance/:id", h.UpdateUserBalance)
	engine.POST("/balance/:id/move", h.MoveMoney)
	engine.GET("/balance/:id/report", h.BuildReport)

	engine.POST("/expanse", h.AddExpanse)
	engine.GET("/expanse", h.BuildStats)
	engine.GET("/expanse/:id", h.GetExpanse)
	engine.POST("/expanse/:id/confirm", h.ConfirmExpanse)
	engine.POST("/expanse/:id/cancel", h.CancelExpanse)

	return &Handler{
		router: engine,
		uc:     uc,
	}, nil
}

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("incoming request handling...", "path", c.FullPath(), "params", c.Params)
		c.Next()
		logger.Info("incoming request handling finished", "path", c.FullPath(), "params", c.Params)
	}
}

func (h *Handler) Serve(port int) error {
	addr := fmt.Sprintf(":%d", port)
	logger.Info("serving...")
	return h.router.Run(addr)
}

// DownloadFile godoc
// @Summary Download file
// @Description Download built report
// @Param file_name path string true "File name to download"
// @Success 200 {binary} File
// @Failure 404 {string} NotFoundError
// @Router /file/{file_name} [get]
func (h *Handler) DownloadFile(c *gin.Context) {
	fileName := c.Param("file_name")
	c.FileAttachment("./file/"+fileName, fileName)
}

// BuildReport godoc
// @Summary Generate report
// @Description Generate report with user expanses
// @Accept json
// @Param id path int true "User ID"
// @Param beg body string false "Report begin datetime"
// @Param end body string false "Report end datetime"
// @Produce json
// @Success 200 {object} ResponseExpanses
// @Failure 400 {string} InvalidParams
// @Failure 500 {string} BuildReportError
// @Router /balance/{id}/report [post]
func (h *Handler) BuildReport(c *gin.Context) {
	userIDString := c.Param("id")
	userID, err := strconv.ParseInt(userIDString, 10, 64)
	if err != nil || userID <= 0 {
		sendResponse(c, http.StatusBadRequest, ErrInvalidArgument)
		return
	}

	reportReq := RequestReport{}
	if err := c.ShouldBindJSON(&reportReq); err != nil {
		sendResponse(c, http.StatusBadRequest, ErrInvalidBody)
		return
	}

	params := reportReq.toModel()
	params.User = &model.User{ID: userID}

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*2)
	defer cancel()
	report, err := h.uc.BuildReport(ctx, params)
	if err != nil {
		sendResponse(c, http.StatusInternalServerError, err)
		return
	}

	sendResponse(c, http.StatusOK, toResponseExpanses(report))
}

type RequestReport struct {
	Beg  *time.Time `json:"beg" binding:"omitempty" time_format:"2006-01-02T15:04:05Z07:00"`
	End  *time.Time `json:"end" binding:"omitempty" time_format:"2006-01-02T15:04:05Z07:00"`
	Page int64      `json:"page" binding:"omitempty,gte=1"`
	Sort string     `json:"sort" binding:"omitempty"`
}

func (r *RequestReport) toModel() *model.ReportParams {
	if r.Page == 0 {
		r.Page = 1
	}

	now := time.Now().Local().UTC()
	currMonth := model.GetMonth(&now)

	if r.End == nil {
		r.End = currMonth.End
	}

	if r.Beg == nil {
		r.Beg = currMonth.Beg
	}

	return &model.ReportParams{
		Dur:       &model.Duration{Beg: r.Beg, End: r.End},
		Page:      r.Page,
		PageSize:  2,
		SortField: r.Sort,
	}
}

type ResponseExpanses struct {
	Expanses []*ResponseExpanse `json:"expanses"`
}

func toResponseExpanses(report *model.Report) *ResponseExpanses {
	expanses := report.GetExpanses()
	respExpanses := make([]*ResponseExpanse, len(expanses))
	for i, expanse := range expanses {
		expanse := expanse
		respExpanses[i] = toResponseExpanse(expanse)
	}
	return &ResponseExpanses{Expanses: respExpanses}
}

// GetUserBalance godoc
// @Summary Get user balance
// @Description Return user balance
// @Accept json
// @Param id path int true "User ID"
// @Produce json
// @Success 200 {object} ResponseBalance
// @Failure 400 {string} InvalidParams
// @Failure 500 {string} GetUserBalanceError
// @Router /balance/{id} [get]
func (h *Handler) GetUserBalance(c *gin.Context) {
	userIDString := c.Param("id")
	userID, err := strconv.ParseInt(userIDString, 10, 64)
	if err != nil || userID <= 0 {
		sendResponse(c, http.StatusBadRequest, ErrInvalidArgument)
		return
	}
	user := &model.User{ID: userID}

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*2)
	defer cancel()
	balance, err := h.uc.GetUserBalance(ctx, user)
	if err != nil {
		sendResponse(c, http.StatusInternalServerError, err)
		return
	}

	sendResponse(c, http.StatusOK, *toResponseBalance(balance))
}

type ResponseBalance struct {
	Balance string `json:"balance"`
}

func toResponseBalance(balance *model.Balance) *ResponseBalance {
	return &ResponseBalance{Balance: balance.Money.ToString()}
}

// UpdateUserBalance godoc
// @Summary Change user balance
// @Description Add money to user balance
// @Accept json
// @Param id path int true "User ID"
// @Param money body int true "Adding money to balance"
// @Produce json
// @Success 202 {string} SuccessMsg
// @Failure 400 {string} InvalidParams
// @Failure 500 {string} MoveMoneyError
// @Router /balance/{id} [put]
func (h *Handler) UpdateUserBalance(c *gin.Context) {
	userIDString := c.Param("id")
	userID, err := strconv.ParseInt(userIDString, 10, 64)
	if err != nil || userID <= 0 {
		sendResponse(c, http.StatusBadRequest, ErrInvalidArgument)
		return
	}
	user := &model.User{ID: userID}

	updateBalanceReq := RequestUpdateBalance{}
	if err := c.ShouldBindJSON(&updateBalanceReq); err != nil {
		sendResponse(c, http.StatusBadRequest, ErrInvalidBody)
		return
	}
	money, err := updateBalanceReq.toModel()
	if err != nil {
		sendResponse(c, http.StatusBadRequest, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*2)
	defer cancel()
	err = h.uc.AddUser(ctx, user)
	if err != nil {
		sendResponse(c, http.StatusInternalServerError, err)
		return
	}

	expanse, err := h.uc.MoveMoney(ctx, user, user, money)
	if err != nil {
		sendResponse(c, http.StatusInternalServerError, err)
		return
	}
	err = h.uc.ConfirmExpanse(ctx, expanse)
	if err != nil {
		sendResponse(c, http.StatusInternalServerError, err)
		return
	}

	sendResponse(c, http.StatusAccepted, fmt.Sprintf("update user: %d balance with money: %s", user.ID, money.ToString()))
}

type RequestUpdateBalance struct {
	MoneyAmount string `json:"money" binding:"required"`
}

func (req *RequestUpdateBalance) toModel() (*model.Money, error) {
	money, err := model.MoneyFromString(req.MoneyAmount)
	if err != nil {
		return nil, err
	}

	return money, nil
}

// MoveMoney godoc
// @Summary Move money
// @Description Move money from user balance to other user balance
// @Accept json
// @Param id path int true "User ID to move money from"
// @Param to body int true "User ID to move money to"
// @Param money body int true "Moving money"
// @Produce json
// @Success 202 {object} ResponseExpanse
// @Failure 400 {string} InvalidParams
// @Failure 500 {string} MoveMoneyError
// @Router /balance/{id}/move [post]
func (h *Handler) MoveMoney(c *gin.Context) {
	fromUserIDString := c.Param("id")
	fromUserID, err := strconv.ParseInt(fromUserIDString, 10, 64)
	if err != nil || fromUserID <= 0 {
		sendResponse(c, http.StatusBadRequest, ErrInvalidArgument)
		return
	}
	from := &model.User{ID: fromUserID}

	moveReq := RequestMove{}
	if err := c.ShouldBindJSON(&moveReq); err != nil {
		sendResponse(c, http.StatusBadRequest, ErrInvalidBody)
		return
	}
	to, money, err := moveReq.toModel()
	if err != nil {
		sendResponse(c, http.StatusBadRequest, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*2)
	defer cancel()
	newExpanse, err := h.uc.MoveMoney(ctx, from, to, money)
	if err != nil {
		sendResponse(c, http.StatusInternalServerError, err)
		return
	}

	sendResponse(c, http.StatusAccepted, *toResponseExpanse(newExpanse))
}

type RequestMove struct {
	ToID  int64  `json:"to" binding:"required,gte=1"`
	Money string `json:"money" binding:"required"`
}

func (r *RequestMove) toModel() (*model.User, *model.Money, error) {
	user := &model.User{ID: r.ToID}
	money, err := model.MoneyFromString(r.Money)
	if err != nil {
		return user, nil, err
	}
	if money.Amount <= 0 {
		return user, nil, ErrMoneyShouldBePositive
	}
	return user, money, nil

}

type ResponseExpanse struct {
	ID        int64  `json:"id"`
	FromID    int64  `json:"from"`
	ToID      int64  `json:"to"`
	Ts        string `json:"ts"`
	ServiceID int64  `json:"service_id"`
	OrderID   int64  `json:"order_id"`
	Cost      string `json:"cost"`
	Status    string `json:"status"`
}

func toResponseExpanse(expanse *model.Expanse) *ResponseExpanse {
	return &ResponseExpanse{
		ID:        expanse.ID,
		FromID:    expanse.From.ID,
		ToID:      expanse.To.ID,
		Ts:        expanse.Ts.Format(time.RFC3339),
		ServiceID: expanse.ServiceID,
		OrderID:   expanse.OrderID,
		Cost:      expanse.Cost.ToString(),
		Status:    expanse.Status.ToString(),
	}
}

// AddExpanse godoc
// @Summary Add expanse
// @Description Add new expanse
// @Accept json
// @Param from body int true "User ID to move money from"
// @Param to body int true "User ID to move money to"
// @Param ts body string false "Expanse timestamp"
// @Param service_id body int true "Expanse service ID"
// @Param order_id body int true "Expanse order ID"
// @Param cost body int true "Expanse cost"
// @Produce json
// @Success 201 {object} ResponseExpanse
// @Failure 400 {string} InvalidParams
// @Failure 500 {string} AddExpanseError
// @Router /expanse [post]
func (h *Handler) AddExpanse(c *gin.Context) {
	addExpanseReq := RequestAddExpanseRequest{}
	if err := c.ShouldBindJSON(&addExpanseReq); err != nil {
		sendResponse(c, http.StatusBadRequest, ErrInvalidBody)
		return
	}

	expanse, err := addExpanseReq.toModel()
	if err != nil {
		sendResponse(c, http.StatusBadRequest, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*2)
	defer cancel()
	expanse, err = h.uc.AddExpanse(ctx, expanse)
	if err != nil {
		sendResponse(c, http.StatusInternalServerError, err)
		return
	}

	sendResponse(c, http.StatusCreated, *toResponseExpanse(expanse))
}

type RequestAddExpanseRequest struct {
	FromID    int64      `json:"from" binding:"required,gte=1"`
	ToID      int64      `json:"to" binding:"required,gte=1"`
	Ts        *time.Time `json:"ts" binding:"omitempty" time_format:"2006-01-02T15:04:05Z07:00"`
	ServiceID int64      `json:"service_id" binding:"required,gte=1"`
	OrderID   int64      `json:"order_id" binding:"required,gte=1"`
	Cost      string     `json:"cost" binding:"required"`
}

func (r *RequestAddExpanseRequest) toModel() (*model.Expanse, error) {
	money, err := model.MoneyFromString(r.Cost)
	if err != nil {
		return nil, err
	}
	if money.Amount < 0 {
		return nil, ErrMoneyShouldBePositive
	}
	if r.Ts == nil {
		now := time.Now().Local().UTC()
		r.Ts = &now
	}

	return &model.Expanse{
		From:      &model.User{ID: r.FromID},
		To:        &model.User{ID: r.ToID},
		Ts:        r.Ts,
		ServiceID: r.ServiceID,
		OrderID:   r.OrderID,
		Cost:      money,
	}, nil
}

// BuildStats godoc
// @Summary Build statistics
// @Description Build statistics of user expanses by services
// @Accept json
// @Param date body string true "Year and month to build statistics for"
// @Produce json
// @Success 201 {string} StatsURL
// @Failure 400 {string} InvalidParams
// @Failure 500 {string} BuildStatsError
// @Router /expanse [get]
func (h *Handler) BuildStats(c *gin.Context) {
	statsReq := RequestStats{}
	if err := c.ShouldBindJSON(&statsReq); err != nil {
		sendResponse(c, http.StatusBadRequest, ErrInvalidBody)
		return
	}

	params := statsReq.toModel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*2)
	defer cancel()
	stats, err := h.uc.BuildStats(ctx, params)
	if err != nil {
		sendResponse(c, http.StatusInternalServerError, err)
		return
	}

	path, err := StatsToCSV(&statsReq, stats)
	if err != nil {
		sendResponse(c, http.StatusInternalServerError, err)
		return
	}
	sendResponse(c, http.StatusCreated, path)
}

type RequestStats struct {
	Date *time.Time `json:"date" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
}

func (r *RequestStats) toModel() *model.StatisticsParams {
	month := model.GetMonth(r.Date)
	return &model.StatisticsParams{
		Dur: month,
	}
}

func StatsToCSV(statsReq *RequestStats, stats *model.Statistics) (string, error) {
	fileName := statsReq.Date.Format("2006-01") + ".csv"
	fileLocation := "./file"

	err := os.MkdirAll(fileLocation, os.ModePerm)
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(fileLocation, fileName)

	_, err = os.Stat(filePath)
	if err == nil {
		err = os.Remove(filePath)
		if err != nil {
			return "", err
		}
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	if err := w.Write([]string{"service_id", "total_amount"}); err != nil {
		return "", err
	}
	for _, key := range stats.Keys() {
		value, _ := stats.Get(key)
		if err := w.Write([]string{strconv.FormatInt(key, 10), value.ToString()}); err != nil {
			return "", err
		}
	}

	return filePath, nil
}

// GetExpanse godoc
// @Summary Get expanse
// @Description Return expanse info by expanse id
// @Accept json
// @Param id path int true "Expanse ID"
// @Produce json
// @Success 200 {object} ResponseExpanse
// @Failure 400 {string} InvalidParams
// @Failure 500 {string} GetExpanseError
// @Router /expanse/{id} [get]
func (h *Handler) GetExpanse(c *gin.Context) {
	expanseIDString := c.Param("id")
	expanseID, err := strconv.ParseInt(expanseIDString, 10, 64)
	if err != nil || expanseID <= 0 {
		sendResponse(c, http.StatusBadRequest, ErrInvalidArgument)
		return
	}
	expanse := &model.Expanse{ID: expanseID}

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*2)
	defer cancel()
	expanse, err = h.uc.GetExpanse(ctx, expanse)
	if err != nil {
		sendResponse(c, http.StatusInternalServerError, err)
		return
	}

	sendResponse(c, http.StatusOK, *toResponseExpanse(expanse))
}

// ConfirmExpanse godoc
// @Summary Confirm expanse
// @Description Confirm expanse
// @Accept json
// @Param id path int true "Expanse ID"
// @Produce json
// @Success 200 {string} Confirmed
// @Failure 400 {string} InvalidParams
// @Failure 500 {string} ConfirmExpanseError
// @Router /expanse/{id}/confirm [post]
func (h *Handler) ConfirmExpanse(c *gin.Context) {
	expanseIDString := c.Param("id")
	expanseID, err := strconv.ParseInt(expanseIDString, 10, 64)
	if err != nil || expanseID <= 0 {
		sendResponse(c, http.StatusBadRequest, ErrInvalidArgument)
		return
	}
	expanse := &model.Expanse{ID: expanseID}

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*2)
	defer cancel()
	if err = h.uc.ConfirmExpanse(ctx, expanse); err != nil {
		sendResponse(c, http.StatusInternalServerError, err)
		return
	}

	sendResponse(c, http.StatusOK, fmt.Sprintf("expanse: %d confirmed", expanse.ID))
}

// CancelExpanse godoc
// @Summary Cancel expanse
// @Description Cancel expanse
// @Accept json
// @Param id path int true "Expanse ID"
// @Produce json
// @Success 200 {string} Canceled
// @Failure 400 {string} InvalidParams
// @Failure 500 {string} CancelExpanseError
// @Router /expanse/{id}/cancel [post]
func (h *Handler) CancelExpanse(c *gin.Context) {
	expanseIDString := c.Param("id")
	expanseID, err := strconv.ParseInt(expanseIDString, 10, 64)
	if err != nil || expanseID <= 0 {
		sendResponse(c, http.StatusBadRequest, ErrInvalidArgument)
		return
	}
	expanse := &model.Expanse{ID: expanseID}

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*2)
	defer cancel()
	if err = h.uc.CancelExpanse(ctx, expanse); err != nil {
		sendResponse(c, http.StatusInternalServerError, err)
		return
	}

	sendResponse(c, http.StatusOK, fmt.Sprintf("expanse: %d canceled", expanse.ID))
}

func sendResponse(c *gin.Context, httpCode int, obj interface{}) {
	var data gin.H
	switch obj := obj.(type) {
	case error:
		data = gin.H{"error": obj.Error()}
	default:
		data = gin.H{"data": obj}
	}
	c.IndentedJSON(httpCode, data)
}
