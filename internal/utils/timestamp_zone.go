package utils

import (
	"fmt"
	"time"
)

type TimestampAndZone struct {
	Timestamp int64
	Timezone  string
}

func GetTimestampAndZone() TimestampAndZone {
	now := time.Now()
	timestamp := now.Unix()
	_, offset := now.Zone()

	sign := "+"
	if offset < 0 {
		sign = "-"
		offset = -offset
	}
	hours := offset / 3600
	minutes := (offset % 3600) / 60
	timezone := fmt.Sprintf("%s%02d%02d", sign, hours, minutes)

	return TimestampAndZone{
		Timestamp: timestamp,
		Timezone:  timezone,
	}
}
