package utils

import (
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
)

// 100% coverage for Contains function by checking both existing and non-existing items in the slice.
func TestContains(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		item  string
		want  bool
	}{
		{"the given item exists in the slice", []string{"apple", "banana", "cherry"}, "banana", true},
		{"the given item does not exist in the slice", []string{"apple", "banana", "cherry"}, "grape", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Contains(tt.slice, tt.item)
			if got != tt.want {
				t.Errorf("Contains(%v, %q) = %v; want %v",
					tt.slice, tt.item, got, tt.want)
			}
		})
	}
}

func TestIsTimestampInPast(t *testing.T) {
	now := time.Now().Unix()
	tests := []struct {
		name      string
		timestamp timestamp.Timestamp
		current   *timestamp.Timestamp
		want      bool
	}{
		{"timestamp is in the past", timestamp.Timestamp{Seconds: 1000}, &timestamp.Timestamp{Seconds: now}, true},
		{"timestamp is in the future", timestamp.Timestamp{Seconds: now + 100000}, &timestamp.Timestamp{Seconds: now}, false},
		{"timestamp is now", timestamp.Timestamp{Seconds: now}, &timestamp.Timestamp{Seconds: now}, false},
		{"timestamp is nil", timestamp.Timestamp{}, nil, false},
		{"current timestamp is nil", timestamp.Timestamp{Seconds: now}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsProtoTimestampInPast(&tt.timestamp, tt.current)
			if got != tt.want {
				t.Errorf("IsTimestampInPast(%v, %v) = %v; want %v",
					&tt.timestamp, tt.current, got, tt.want)
			}
		})
	}
}

func TestGetCurrentTimestamp(t *testing.T) {
	got := GetCurrentProtoTimestamp()
	if got == nil {
		t.Errorf("GetCurrentTimestamp() = nil; want non-nil")
		return
	}
	if got.Seconds != time.Now().Unix() {
		t.Errorf("GetCurrentTimestamp() = %v; want %v",
			got.Seconds, time.Now().Unix())
	}
}

func TestGetTimestamp(t *testing.T) {
	tests := []struct {
		name  string
		year  int
		month time.Month
		day   int
		hour  int
		min   int
		sec   int
	}{
		{"valid timestamp", 2024, time.December, 25, 0, 0, 0},
		{"another valid timestamp", 2023, time.January, 1, 12, 30, 45},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetProtoTimestamp(tt.year, tt.month, tt.day, tt.hour, tt.min, tt.sec)
			if got == nil {
				t.Errorf("GetTimestamp() = nil; want non-nil")
				return
			}
			expectedTime := time.Date(tt.year, tt.month, tt.day, tt.hour, tt.min, tt.sec, 0, time.UTC)
			if got.Seconds != expectedTime.Unix() || got.Nanos != int32(expectedTime.Nanosecond()) {
				t.Errorf("GetTimestamp() = %v; want %v",
					got, expectedTime)
			}
		})
	}
}
