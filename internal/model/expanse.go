package model

import "time"

type ExpanseStatus int64

const (
	StatusPending ExpanseStatus = iota
	StatusSuccess
	StatusCancel
)

func (s *ExpanseStatus) ToString() string {
	var str string
	if *s == StatusPending {
		str = "pending"
	} else if *s == StatusSuccess {
		str = "success"
	} else if *s == StatusCancel {
		str = "cancel"
	}
	return str
}

type Expanse struct {
	ID        int64
	From      *User
	To        *User
	Ts        *time.Time
	ServiceID int64
	OrderID   int64
	Cost      *Money
	Status    ExpanseStatus
}
