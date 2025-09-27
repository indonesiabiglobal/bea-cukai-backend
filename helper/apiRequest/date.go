package apiRequest

import (
	"time"

	"github.com/gin-gonic/gin"
)

func parseDate(value string, loc *time.Location) (time.Time, error) {
	// accept RFC3339 date or YYYY-MM-DD
	if value == "" {
		return time.Time{}, nil
	}
	if t, err := time.ParseInLocation(time.RFC3339, value, loc); err == nil {
		return t, nil
	}
	return time.ParseInLocation("2006-01-02", value, loc)
}

func GetRange(ctx *gin.Context) (time.Time, time.Time, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	fromStr := ctx.Query("from")
	toStr := ctx.Query("to")

	// defaults: last 30 days
	now := time.Now().In(loc)
	from := now.AddDate(0, 0, -30)
	to := now

	if t, err := parseDate(fromStr, loc); err == nil && !t.IsZero() {
		from = t
	} else if fromStr != "" && err != nil {
		return time.Time{}, time.Time{}, err
	}
	if t, err := parseDate(toStr, loc); err == nil && !t.IsZero() {
		to = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	} else if toStr != "" && err != nil {
		return time.Time{}, time.Time{}, err
	}
	return from, to, nil
}