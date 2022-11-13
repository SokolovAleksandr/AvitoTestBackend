package model

type Balance struct {
	Money *Money
}

func (b *Balance) More(other *Balance) bool {
	return b.Money.More(other.Money)
}
