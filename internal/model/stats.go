package model

type StatisticsParams struct {
	Dur *Duration
}

type Statistics struct {
	table map[int64]*Money
}

func NewStatistics(table map[int64]*Money) (*Statistics, error) {
	if table == nil {
		table = make(map[int64]*Money)
	}
	return &Statistics{
		table: table,
	}, nil
}

func (s *Statistics) Add(serviceID int64, money *Money) error {
	val, ok := s.table[serviceID]

	if !ok {
		val = &Money{Amount: 0}
		s.table[serviceID] = val
	}

	val.Amount += money.Amount

	return nil
}

func (s *Statistics) Keys() []int64 {
	keys := []int64{}

	for key := range s.table {
		keys = append(keys, key)
	}

	return keys
}

func (s *Statistics) Get(serviceId int64) (*Money, error) {
	val, ok := s.table[serviceId]

	if !ok {
		return &Money{Amount: 0}, nil
	}

	return val, nil
}

func (s *Statistics) Total() (*Money, error) {
	res := &Money{Amount: 0}
	for _, val := range s.table {
		res.Amount += val.Amount
	}
	return res, nil
}
