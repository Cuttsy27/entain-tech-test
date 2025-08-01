package db

import (
	"testing"

	"git.neds.sh/matty/entain/racing/proto/racing"
	"google.golang.org/protobuf/proto"
)

func TestApplyFilter(t *testing.T) {
	// Create an instance of a racesRepo to use applyFilter method
	repo := &racesRepo{}

	// Create a base query string to build upon and test against
	baseQuery := getRaceQueries()[racesList]

	// Reusable helper functions to assert the correctness of the query and arguments
	assertCorrectQuery := func(t testing.TB, gotQuery string, wantQuery string) {
		t.Helper()
		if gotQuery != wantQuery {
			t.Errorf("applyFilter() query = %q, want %q", gotQuery, wantQuery)
		}
	}
	assertCorrectArgs := func(t testing.TB, gotArgs []interface{}, wantArgs []interface{}) {
		t.Helper()
		if len(gotArgs) != len(wantArgs) {
			t.Fatalf("applyFilter() args length = %d, want %d", len(gotArgs), len(wantArgs))
		}
		for i, arg := range gotArgs {
			want := wantArgs[i]
			if arg != want {
				t.Errorf("applyFilter() arg[%d] = %v, want %v", i, arg, want)
			}
		}
	}

	t.Run("applyFilter with Visible filter set to true", func(t *testing.T) {
		// Define visible variable here so we don't pass true to the filter but want false, for example
		visible := true

		// Create a filter with Visible set to true
		filter := &racing.ListRacesRequestFilter{
			Visible: proto.Bool(visible),
		}

		// Call applyFilter with the base query and the filter
		gotQuery, gotArgs := repo.applyFilter(baseQuery, filter)

		// We know that at this point the query should only be the base query plus the WHERE clause for Visible
		wantQuery := baseQuery + " WHERE visible = ?"
		wantArgs := []interface{}{visible}

		assertCorrectQuery(t, gotQuery, wantQuery)
		assertCorrectArgs(t, gotArgs, wantArgs)
	})
	t.Run("applyFilter with Visible filter set to false", func(t *testing.T) {
		visible := false

		filter := &racing.ListRacesRequestFilter{
			Visible: proto.Bool(visible),
		}

		gotQuery, gotArgs := repo.applyFilter(baseQuery, filter)

		wantQuery := baseQuery + " WHERE visible = ?"
		wantArgs := []interface{}{visible}

		assertCorrectQuery(t, gotQuery, wantQuery)
		assertCorrectArgs(t, gotArgs, wantArgs)
	})
	t.Run("applyFilter with MeetingIds filter", func(t *testing.T) {
		// Defining a `meetingIds` variable here similar to the visible variable above would increase complexity
		// as we would need to iterate the meetingIds slice to build query and args below. Simpler test code is preferable.

		filter := &racing.ListRacesRequestFilter{
			MeetingIds: []int64{1, 2, 3},
		}

		gotQuery, gotArgs := repo.applyFilter(baseQuery, filter)

		wantQuery := baseQuery + " WHERE meeting_id IN (?,?,?)"
		wantArgs := []interface{}{int64(1), int64(2), int64(3)}

		assertCorrectQuery(t, gotQuery, wantQuery)
		assertCorrectArgs(t, gotArgs, wantArgs)
	})
	t.Run("applyFilter with both Visible and MeetingIds filters", func(t *testing.T) {
		visible := true
		filter := &racing.ListRacesRequestFilter{
			Visible:    proto.Bool(visible),
			MeetingIds: []int64{1, 2, 3},
		}

		gotQuery, gotArgs := repo.applyFilter(baseQuery, filter)

		wantQuery := baseQuery + " WHERE visible = ? AND meeting_id IN (?,?,?)"
		wantArgs := []interface{}{visible, int64(1), int64(2), int64(3)}

		assertCorrectQuery(t, gotQuery, wantQuery)
		assertCorrectArgs(t, gotArgs, wantArgs)
	})
	t.Run("applyFilter with empty filter", func(t *testing.T) {
		filter := &racing.ListRacesRequestFilter{}

		gotQuery, gotArgs := repo.applyFilter(baseQuery, filter)

		wantQuery := baseQuery
		wantArgs := []interface{}{}

		assertCorrectQuery(t, gotQuery, wantQuery)
		assertCorrectArgs(t, gotArgs, wantArgs)
	})
}
