package utils

import (
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
)

// Returns true if a given string is present in a slice of strings, else false.
func Contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func IsProtoTimestampInPast(timestamp, currentTimestamp *timestamp.Timestamp) bool {
	if timestamp == nil || currentTimestamp == nil {
		return false
	}
	return timestamp.Seconds < currentTimestamp.Seconds
}

func GetCurrentProtoTimestamp() *timestamp.Timestamp {
	currentTime := time.Now()
	return &timestamp.Timestamp{
		Seconds: currentTime.Unix(),
		Nanos:   int32(currentTime.Nanosecond()),
	}
}

// getProtoTimestamp returns a *timestamp.Timestamp for the specified date and time in UTC.
func GetProtoTimestamp(year int, month time.Month, day, hour, min, sec int) *timestamp.Timestamp {
	t := time.Date(year, month, day, hour, min, sec, 0, time.UTC)
	return &timestamp.Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(t.Nanosecond()),
	}
}
