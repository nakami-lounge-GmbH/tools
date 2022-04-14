package helpers

import "time"

func Time0(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
func Time24(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.UTC)
}

func Time4(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 4, 0, 0, 0, time.UTC)
}
