package timeutil

import (
	"fmt"
	"time"
)

var WIB = time.FixedZone("WIB", 7*60*60)

var monthNames = [...]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}

func Today(now time.Time) time.Time {
	y, m, d := now.In(WIB).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func ParseDateISO(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil || t.Format("2006-01-02") != s {
		return time.Time{}, fmt.Errorf("invalid ISO date %q", s)
	}
	return t.UTC(), nil
}

func FormatDateISO(t time.Time) string { return t.Format("2006-01-02") }

func FormatDateID(t time.Time) string {
	y, m, d := t.Date()
	return fmt.Sprintf("%d %s %d", d, monthNames[int(m)-1], y)
}

func DaysUntil(eventDate, today time.Time) int {
	e := time.Date(eventDate.Year(), eventDate.Month(), eventDate.Day(), 0, 0, 0, 0, time.UTC)
	d := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	return int(e.Sub(d).Hours() / 24)
}
