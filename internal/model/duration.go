package model

import "time"

type Duration struct {
	Beg *time.Time
	End *time.Time
}

func GetMonth(ts *time.Time) *Duration {
	monthStart := ts.AddDate(0, 0, 1-ts.Day())
	monthStart = monthStart.Add(
		time.Hour*-time.Duration(monthStart.Hour()) +
			time.Minute*-time.Duration(monthStart.Minute()),
	)
	monthEnd := monthStart.AddDate(0, 1, 0)

	return &Duration{
		Beg: &monthStart,
		End: &monthEnd,
	}
}
