package model

type Reserve struct {
	ID   int64
	User *User
	Size *Money
}
