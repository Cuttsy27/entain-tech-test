package service

import (
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	sportsErrors "git.neds.sh/matty/entain/sports/errors"
	"git.neds.sh/matty/entain/sports/proto/sports"
	"git.neds.sh/matty/entain/sports/utils"
	"github.com/golang/protobuf/ptypes/timestamp"
	"golang.org/x/net/context"
)

// Mocking the repo for testing the service layer
type mockSportsRepo struct {
	shouldError bool
}

func (m *mockSportsRepo) Init() error { return nil }

// Need to mock the return values from the List method so the repo "returns" what the service expects, so we can test the service logic
func (m *mockSportsRepo) ListEvents(filter string, orderBy string) ([]*sports.Event, error) {
	if m.shouldError {
		return nil, errors.New("[List] mock error")
	}
	events := []*sports.Event{
		{Id: 1, Name: "Championship Final", Sport: "Soccer", Visible: true, AdvertisedStartTime: utils.GetProtoTimestamp(2024, time.December, 25, 0, 0, 0), Status: sports.Event_CLOSED},
		{Id: 2, Name: "Grand Prix", Sport: "Basketball", Visible: false, AdvertisedStartTime: &timestamp.Timestamp{Seconds: time.Now().Add(24 * time.Hour).Unix()}, Status: sports.Event_OPEN},
	}
	filteredEvents := events
	// If the filter is set, we need to apply it
	if filter != "" {
		conditions := strings.Split(filter, " AND ")
		for _, cond := range conditions {
			cond = strings.TrimSpace(cond)
			if strings.Contains(cond, "visible = true") {
				tmp := []*sports.Event{}
				for _, event := range filteredEvents {
					if event.Visible {
						tmp = append(tmp, event)
					}
				}
				filteredEvents = tmp
			} else if strings.Contains(cond, "visible = false") {
				tmp := []*sports.Event{}
				for _, event := range filteredEvents {
					if !event.Visible {
						tmp = append(tmp, event)
					}
				}
				filteredEvents = tmp
			}
			// Check for sport filtering
			if strings.Contains(cond, "sport in") || strings.Contains(cond, "sport IN") {
				start := strings.Index(cond, "(")
				end := strings.Index(cond, ")")
				if start != -1 && end != -1 && end > start {
					sportsList := cond[start+1 : end]
					sportsList = strings.ReplaceAll(sportsList, `"`, "")
					sportsList = strings.ReplaceAll(sportsList, "'", "")
					sportsNames := strings.Split(sportsList, ",")
					for i := range sportsNames {
						sportsNames[i] = strings.TrimSpace(sportsNames[i])
					}
					tmp := []*sports.Event{}
					for _, event := range filteredEvents {
						for _, sport := range sportsNames {
							if event.Sport == sport {
								tmp = append(tmp, event)
							}
						}
					}
					filteredEvents = tmp
				}
			}
		}
	}

	if orderBy != "" {
		// If orderBy is set, we can sort
		switch orderBy {
		case "advertised_start_time", "advertised_start_time asc":
			// Sort by advertised_start_time in ascending order
			sort.Slice(events, func(i, j int) bool {
				return events[i].AdvertisedStartTime.Seconds < events[j].AdvertisedStartTime.Seconds
			})
		case "advertised_start_time desc":
			// Sort by advertised_start_time in descending order
			sort.Slice(events, func(i, j int) bool {
				return events[i].AdvertisedStartTime.Seconds > events[j].AdvertisedStartTime.Seconds
			})
		}
	}
	// no filter set, return all potentially sorted events
	return filteredEvents, nil
}
func (m *mockSportsRepo) GetEvent(id int64) (*sports.Event, error) {
	if m.shouldError {
		return nil, errors.New("[Get] mock error")
	}
	return &sports.Event{
		Id:                  1,
		Name:                "Championship Final",
		Sport:               "Soccer",
		Visible:             true,
		AdvertisedStartTime: utils.GetProtoTimestamp(2024, time.December, 25, 0, 0, 0),
		Status:              sports.Event_CLOSED,
	}, nil
}

func TestListEventsFilter(t *testing.T) {
	repo := &mockSportsRepo{}
	service := NewSportsService(repo)
	ctx := context.Background()

	t.Run("returns all events when no filter is set", func(t *testing.T) {
		filter := ""

		req := &sports.ListEventsRequest{
			Filter: filter,
		}

		resp, err := service.ListEvents(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Events) != 2 {
			t.Errorf("expected 2 events, got %d", len(resp.Events))
		}
	})
	t.Run("filters to events that have visible set to true", func(t *testing.T) {
		filter := "visible = true"

		req := &sports.ListEventsRequest{
			Filter: filter,
		}

		resp, err := service.ListEvents(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Events) != 1 {
			t.Errorf("expected 1 event, got %d", len(resp.Events))
		}
	})
	t.Run("filters to events that have visible set to false", func(t *testing.T) {
		filter := "visible = false"

		req := &sports.ListEventsRequest{
			Filter: filter,
		}

		resp, err := service.ListEvents(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Events) != 1 {
			t.Errorf("expected 1 event, got %d", len(resp.Events))
		}
	})
	t.Run("filters to events that have sports", func(t *testing.T) {
		filter := "sport in (\"Soccer\")"

		req := &sports.ListEventsRequest{
			Filter: filter,
		}

		resp, err := service.ListEvents(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Events) != 1 {
			t.Errorf("expected 1 event, got %d", len(resp.Events))
		}
		if resp.Events[0].Sport != "Soccer" {
			t.Errorf("expected Sport 'Soccer', got %s", resp.Events[0].Sport)
		}
	})
	t.Run("filters to events that have visible set to true and sports", func(t *testing.T) {
		filter := "visible = true AND sport in (\"Soccer\")"

		req := &sports.ListEventsRequest{
			Filter: filter,
		}

		resp, err := service.ListEvents(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Events) != 1 {
			t.Errorf("expected 1 event, got %d", len(resp.Events))
		}
		if resp.Events[0].Sport != "Soccer" {
			t.Errorf("expected Sport 'Soccer', got %s", resp.Events[0].Sport)
		}
	})
	t.Run("filters to events that have visible set to false and sports", func(t *testing.T) {
		filter := "sport in (\"Soccer\", \"Basketball\") AND visible = false"

		req := &sports.ListEventsRequest{
			Filter: filter,
		}

		resp, err := service.ListEvents(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Events) != 1 {
			t.Errorf("expected 1 event, got %d", len(resp.Events))
		}

		if resp.Events[0].Sport != "Basketball" {
			t.Errorf("expected Sport 'Basketball', got %s", resp.Events[0].Sport)
		}
	})
	t.Run("filters to events that have visible set to true and non-matching sports", func(t *testing.T) {
		filter := "sport in (\"Basketball\") AND visible = true"

		req := &sports.ListEventsRequest{
			Filter: filter,
		}

		resp, err := service.ListEvents(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Events) != 0 {
			t.Errorf("expected 0 events, got %d", len(resp.Events))
			t.Errorf("events: %v", resp.Events)
		}
	})
}

func TestListEventsOrderBy(t *testing.T) {
	repo := &mockSportsRepo{}
	service := NewSportsService(repo)
	ctx := context.Background()

	t.Run("returns events ordered by advertised_start_time in descending order", func(t *testing.T) {
		orderBy := "advertised_start_time desc"

		req := &sports.ListEventsRequest{
			OrderBy: orderBy,
		}

		resp, err := service.ListEvents(ctx, req)

		if resp.Events[0].AdvertisedStartTime.Seconds < resp.Events[1].AdvertisedStartTime.Seconds {
			t.Errorf("expected %d to be greater than %d", resp.Events[0].AdvertisedStartTime.Seconds, resp.Events[1].AdvertisedStartTime.Seconds)
		}

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("returns events ordered by advertised_start_time in ascending order", func(t *testing.T) {
		orderBy := "advertised_start_time"
		req := &sports.ListEventsRequest{
			OrderBy: orderBy,
		}

		resp, err := service.ListEvents(ctx, req)

		if resp.Events[0].AdvertisedStartTime.Seconds > resp.Events[1].AdvertisedStartTime.Seconds {
			t.Errorf("expected %d to be less than %d", resp.Events[0].AdvertisedStartTime.Seconds, resp.Events[1].AdvertisedStartTime.Seconds)
		}

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// Does the service correctly get events with their statuses?
// This is a test to ensure that the service layer correctly retrieves the status of each event.
func TestEventsHaveStatus(t *testing.T) {
	repo := &mockSportsRepo{}
	service := NewSportsService(repo)
	ctx := context.Background()
	t.Run("returns events with correct status", func(t *testing.T) {
		req := &sports.ListEventsRequest{}
		resp, err := service.ListEvents(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Events) != 2 {
			t.Errorf("expected 2 events, got %d", len(resp.Events))
		}

		if resp.Events[0].Status != sports.Event_CLOSED {
			t.Errorf("expected event status to be %v, got %v", sports.Event_CLOSED, resp.Events[0].Status)
		}

		if resp.Events[1].Status != sports.Event_OPEN {
			t.Errorf("expected event status to be %v, got %v", sports.Event_OPEN, resp.Events[1].Status)
		}
	})
}

func TestListEvents(t *testing.T) {
	repo := &mockSportsRepo{}
	service := NewSportsService(repo)
	ctx := context.Background()

	t.Run("returns all events", func(t *testing.T) {
		req := &sports.ListEventsRequest{}
		resp, err := service.ListEvents(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Events) != 2 {
			t.Errorf("expected 2 events, got %d", len(resp.Events))
		}
	})
	t.Run("returns an error", func(t *testing.T) {
		repo.shouldError = true                     // Simulate an error in the repo
		defer func() { repo.shouldError = false }() // Reset after test
		req := &sports.ListEventsRequest{}
		resp, err := service.ListEvents(ctx, req)

		if err == nil {
			t.Fatalf("expected error, got nil")
		}

		if resp != nil {
			t.Errorf("expected nil response, got %v", resp)
		}
	})
}

func TestGetEvent(t *testing.T) {
	repo := &mockSportsRepo{}
	service := NewSportsService(repo)
	ctx := context.Background()

	t.Run("returns event by ID", func(t *testing.T) {
		req := &sports.GetEventRequest{
			Id: 1,
		}

		event, err := service.GetEvent(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if event == nil {
			t.Errorf("expected event, got nil")
		}

		if event.Id != 1 {
			t.Errorf("expected event ID 1, got %d", event.Id)
		}
	})

	t.Run("returns an error for non-existent event", func(t *testing.T) {
		repo.shouldError = true                     // Simulate an error in the repo
		defer func() { repo.shouldError = false }() // Reset after test
		req := &sports.GetEventRequest{
			Id: 999,
		}

		event, err := service.GetEvent(ctx, req)

		if err == nil {
			t.Fatalf("expected error, got nil")
		}

		if event != nil {
			t.Errorf("expected nil event, got %v", event)
		}
	})

	t.Run("returns ErrInvalidEventID error for invalid event ID", func(t *testing.T) {
		req := &sports.GetEventRequest{
			Id: 0,
		}

		event, err := service.GetEvent(ctx, req)

		if err == nil {
			t.Fatalf("expected error, got nil")
		}

		if event != nil {
			t.Errorf("expected nil event, got %v", event)
		}

		if err != sportsErrors.ErrInvalidEventID {
			t.Errorf("expected error %v, got %v", sportsErrors.ErrInvalidEventID, err)
		}
	})
}
