package db

import (
	"testing"
	"time"

	"git.neds.sh/matty/entain/sports/proto/sports"
	"git.neds.sh/matty/entain/sports/utils"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// 100% coverage for applyFilter method
func Events_TestApplyFilter(t *testing.T) {
	// Create an instance of a sportsRepo to use applyFilter method
	repo := &sportsRepo{}

	// Create a base query string to build upon and test against
	baseQuery := getEventQueries()[eventsList]

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
		filter := "visible = true"

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

		filter := "visible = false"

		gotQuery, gotArgs := repo.applyFilter(baseQuery, filter)

		wantQuery := baseQuery + " WHERE visible = ?"
		wantArgs := []interface{}{visible}

		assertCorrectQuery(t, gotQuery, wantQuery)
		assertCorrectArgs(t, gotArgs, wantArgs)
	})
	t.Run("applyFilter with sport filter", func(t *testing.T) {
		// Defining a `sport` variable here similar to the visible variable above would increase complexity
		// as we would need to iterate the sport slice to build query and args below. Simpler test code is preferable.

		filter := "sport IN (\"Soccer\", \"Basketball\")"

		gotQuery, gotArgs := repo.applyFilter(baseQuery, filter)

		wantQuery := baseQuery + " WHERE sport IN (?,?)"
		wantArgs := []interface{}{"Soccer", "Basketball"}

		assertCorrectQuery(t, gotQuery, wantQuery)
		assertCorrectArgs(t, gotArgs, wantArgs)
	})
	t.Run("applyFilter with both Visible and Sport filters", func(t *testing.T) {
		visible := true
		filter := "visible = true AND sport IN (\"Soccer\", \"Basketball\")"

		gotQuery, gotArgs := repo.applyFilter(baseQuery, filter)

		wantQuery := baseQuery + " WHERE visible = ? AND sport IN (?,?)"
		wantArgs := []interface{}{visible, "Soccer", "Basketball"}

		assertCorrectQuery(t, gotQuery, wantQuery)
		assertCorrectArgs(t, gotArgs, wantArgs)
	})
	t.Run("applyFilter with empty filter", func(t *testing.T) {
		filter := ""

		gotQuery, gotArgs := repo.applyFilter(baseQuery, filter)

		wantQuery := baseQuery
		wantArgs := []interface{}{}

		assertCorrectQuery(t, gotQuery, wantQuery)
		assertCorrectArgs(t, gotArgs, wantArgs)
	})
	t.Run("applyFilter with nil filter", func(t *testing.T) {
		gotQuery, gotArgs := repo.applyFilter(baseQuery, "")

		wantQuery := baseQuery
		wantArgs := []interface{}{}

		assertCorrectQuery(t, gotQuery, wantQuery)
		assertCorrectArgs(t, gotArgs, wantArgs)
	})
}

// 100% coverage for applyOrderBy method
func TestApplyOrderBy(t *testing.T) {
	assertOrderBy := func(t testing.TB, got, want string) {
		t.Helper()

		if got != want {
			t.Errorf("applyOrderBy() = %q, want %q", got, want)
		}
	}
	t.Run("applyOrderBy with empty orderBy", func(t *testing.T) {
		repo := &sportsRepo{}

		baseQuery := getEventQueries()[eventsList]

		got := repo.applyOrderBy(baseQuery, "")

		want := baseQuery + " ORDER BY advertised_start_time asc"

		assertOrderBy(t, got, want)

	})
	t.Run("applyOrderBy with valid orderBy", func(t *testing.T) {
		repo := &sportsRepo{}

		baseQuery := getEventQueries()[eventsList]
		orderBy := "advertised_start_time desc"

		got := repo.applyOrderBy(baseQuery, orderBy)

		want := baseQuery + " ORDER BY advertised_start_time desc"

		assertOrderBy(t, got, want)
	})
	t.Run("applyOrderBy with invalid orderBy", func(t *testing.T) {
		repo := &sportsRepo{}

		baseQuery := getEventQueries()[eventsList]
		orderBy := "invalid_field"

		got := repo.applyOrderBy(baseQuery, orderBy)

		// Invalid orderBy should not change the query
		want := baseQuery + " ORDER BY advertised_start_time asc"

		assertOrderBy(t, got, want)
	})
}

// Test addStatusToEvent method to ensure it correctly sets the status of a single event
func TestAddStatusToEvent(t *testing.T) {
	repo := &sportsRepo{}

	t.Run("adds CLOSED status when advertised_start_time is in the past", func(t *testing.T) {
		event := &sports.Event{
			Id:                  1,
			Name:                "Championship Final",
			Visible:             true,
			Sport:               "Soccer",
			AdvertisedStartTime: utils.GetProtoTimestamp(2024, time.December, 25, 0, 0, 0),
		}
		repo.addStatusToEvent(event)
		if event.Status != sports.Event_CLOSED {
			t.Errorf("expected event status to be CLOSED, got %v", event.Status)
		}
	})
	t.Run("adds OPEN status when advertised_start_time is in the future", func(t *testing.T) {
		now := time.Now().Unix()
		event := &sports.Event{
			Id:                  2,
			Name:                "Grand Prix",
			Visible:             true,
			Sport:               "Basketball",
			AdvertisedStartTime: &timestamp.Timestamp{Seconds: now + 100000},
		}
		repo.addStatusToEvent(event)
		if event.Status != sports.Event_OPEN {
			t.Errorf("expected event status to be OPEN, got %v", event.Status)
		}
	})
}

// Test addStatusToEvents method to ensure it correctly sets the status for multiple events
func TestAddStatusToEvents(t *testing.T) {
	repo := &sportsRepo{}

	t.Run("adds status to all events", func(t *testing.T) {
		now := time.Now().Unix()
		events := []*sports.Event{
			{Id: 1, Name: "Championship Final", Visible: true, Sport: "Soccer", AdvertisedStartTime: utils.GetProtoTimestamp(2024, time.December, 25, 0, 0, 0)},
			{Id: 2, Name: "Grand Prix", Visible: true, Sport: "Basketball", AdvertisedStartTime: &timestamp.Timestamp{Seconds: now + 100000}},
		}
		repo.addStatusToEvents(events)
		// The first event in the slice is specifically set to be in the past, so it should be CLOSED
		if events[0].Status != sports.Event_CLOSED {
			t.Errorf("expected event status to be %v, got %v", sports.Event_CLOSED, events[0].Status)
		}
		// The second event in the slice is specifically set to be in the future, so it should be OPEN
		if events[1].Status != sports.Event_OPEN {
			t.Errorf("expected event status to be %v, got %v", sports.Event_OPEN, events[1].Status)
		}
	})
}
