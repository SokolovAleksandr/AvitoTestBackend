package model

import "errors"

var (
	ErrIndexOutOfRange = errors.New("index out of range")
)

type ReportParams struct {
	User      *User
	Dur       *Duration
	Page      int64
	PageSize  int64
	SortField string
}

type Report struct {
	expanses []*Expanse
}

func NewReport(expanses []*Expanse) (*Report, error) {
	if expanses == nil {
		expanses = []*Expanse{}
	}
	return &Report{
		expanses: expanses,
	}, nil
}

func (r *Report) AddExpanse(expanse *Expanse) error {
	r.expanses = append(r.expanses, expanse)
	return nil
}

func (r *Report) GetExpanses() []*Expanse {
	return r.expanses
}
