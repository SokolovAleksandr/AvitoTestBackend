package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const Pattern = `(-?\d+(.\d{2})?)`

var (
	ErrParseAmount          = errors.New("parse amount string failed")
	ErrIncorrectDecimalPart = errors.New("incorrect decimal part")
)

type Money struct {
	Amount int64
}

func (m *Money) Negative() *Money {
	return &Money{Amount: -m.Amount}
}

func (m *Money) More(other *Money) bool {
	return m.Amount > other.Amount
}

func MoneyFromString(moneyString string) (*Money, error) {
	isNeg := strings.HasPrefix(moneyString, "-")
	if isNeg {
		moneyString = moneyString[1:]
	}

	amountInt := int64(0)

	parts := strings.Split(moneyString, ".")

	amountPart, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, ErrParseAmount
	}
	amountInt += int64(amountPart) * 100

	if len(parts) > 1 {
		if len(parts[1]) != 2 {
			return nil, ErrIncorrectDecimalPart
		}
		amountPart, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil, ErrParseAmount
		}
		amountInt += int64(amountPart)
	}

	if isNeg {
		amountInt *= -1
	}

	return &Money{Amount: int64(amountInt)}, nil
}

func (m *Money) ToString() string {
	whole, part := m.Amount/100, m.Amount%100
	if m.Amount < 0 {
		part *= -1
	}
	return fmt.Sprintf("%d.%02d", whole, part)
}
