package utils

import "testing"

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
