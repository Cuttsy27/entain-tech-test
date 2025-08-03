package service

import (
	"sort"
	"testing"
	"time"

	"git.neds.sh/matty/entain/racing/proto/racing"
	"git.neds.sh/matty/entain/racing/utils"
	"github.com/golang/protobuf/ptypes/timestamp"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/proto"
)

// Mocking the repo for testing the service layer
type mockRacesRepo struct{}

func (m *mockRacesRepo) Init() error { return nil }

// Need to mock the return values from the List method so the repo "returns" what the service expects, so we can test the service logic
func (m *mockRacesRepo) List(filter *racing.ListRacesRequestFilter, orderBy string) ([]*racing.Race, error) {
	races := []*racing.Race{
		{Id: 1, Name: "Race 1", Visible: true, MeetingId: 1, AdvertisedStartTime: utils.GetProtoTimestamp(2024, time.December, 25, 0, 0, 0), Status: racing.Race_CLOSED},
		{Id: 2, Name: "Race 2", Visible: false, MeetingId: 2, AdvertisedStartTime: &timestamp.Timestamp{Seconds: time.Now().Add(24 * time.Hour).Unix()}, Status: racing.Race_OPEN},
	}
	filteredRaces := []*racing.Race{}
	// If the filter is set, we need to apply it
	if filter != nil {
		// Check if the filter has Visible set to true or false and there are meeting IDs
		if filter.Visible != nil && filter.MeetingIds != nil {
			for _, meetingId := range filter.MeetingIds {
				for _, race := range races {
					if race.Visible == *filter.Visible && race.MeetingId == meetingId {
						filteredRaces = append(filteredRaces, race)
					}
				}
			}
			return filteredRaces, nil
		}
		// Check if the filter has Visible set to true or false
		if filter.Visible != nil {
			for _, race := range races {
				if race.Visible == *filter.Visible {
					filteredRaces = append(filteredRaces, race)
				}
			}
			return filteredRaces, nil
		}
		// Check if the filter has MeetingIds set
		if filter.MeetingIds != nil {
			for _, meetingId := range filter.MeetingIds {
				for _, race := range races {
					if race.MeetingId == meetingId {
						filteredRaces = append(filteredRaces, race)
					}
				}
			}
			return filteredRaces, nil
		}
	}

	if orderBy != "" {
		// If orderBy is set, we can sort
		switch orderBy {
		case "advertised_start_time", "advertised_start_time asc":
			// Sort by advertised_start_time in ascending order
			sort.Slice(races, func(i, j int) bool {
				return races[i].AdvertisedStartTime.Seconds < races[j].AdvertisedStartTime.Seconds
			})
		case "advertised_start_time desc":
			// Sort by advertised_start_time in descending order
			sort.Slice(races, func(i, j int) bool {
				return races[i].AdvertisedStartTime.Seconds > races[j].AdvertisedStartTime.Seconds
			})
		}
	}
	// no filter set, return all potentially sorted races
	return races, nil
}

func TestListRacesFilter(t *testing.T) {
	repo := &mockRacesRepo{}
	service := NewRacingService(repo)
	ctx := context.Background()

	t.Run("returns all races when no filter is set", func(t *testing.T) {
		filter := &racing.ListRacesRequestFilter{}

		req := &racing.ListRacesRequest{
			Filter: filter,
		}

		resp, err := service.ListRaces(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Races) != 2 {
			t.Errorf("expected 2 races, got %d", len(resp.Races))
		}
	})
	t.Run("filters to races that have visible set to true", func(t *testing.T) {
		filter := &racing.ListRacesRequestFilter{
			Visible: proto.Bool(true),
		}

		req := &racing.ListRacesRequest{
			Filter: filter,
		}

		resp, err := service.ListRaces(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Races) != 1 {
			t.Errorf("expected 1 race, got %d", len(resp.Races))
		}
	})
	t.Run("filters to races that have visible set to false", func(t *testing.T) {
		filter := &racing.ListRacesRequestFilter{
			Visible: proto.Bool(false),
		}

		req := &racing.ListRacesRequest{
			Filter: filter,
		}

		resp, err := service.ListRaces(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Races) != 1 {
			t.Errorf("expected 1 race, got %d", len(resp.Races))
		}
	})
	t.Run("filters to races that have meeting IDs", func(t *testing.T) {
		filter := &racing.ListRacesRequestFilter{
			MeetingIds: []int64{1},
		}

		req := &racing.ListRacesRequest{
			Filter: filter,
		}

		resp, err := service.ListRaces(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Races) != 1 {
			t.Errorf("expected 1 race, got %d", len(resp.Races))
		}
		if resp.Races[0].MeetingId != 1 {
			t.Errorf("expected MeetingId 1, got %d", resp.Races[0].MeetingId)
		}
	})
	t.Run("filters to races that have visible set to true and meeting IDs", func(t *testing.T) {
		filter := &racing.ListRacesRequestFilter{
			Visible:    proto.Bool(true),
			MeetingIds: []int64{1},
		}

		req := &racing.ListRacesRequest{
			Filter: filter,
		}

		resp, err := service.ListRaces(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Races) != 1 {
			t.Errorf("expected 1 race, got %d", len(resp.Races))
		}
		if resp.Races[0].MeetingId != 1 {
			t.Errorf("expected MeetingId 1, got %d", resp.Races[0].MeetingId)
		}
	})
	t.Run("filters to races that have visible set to false and meeting IDs", func(t *testing.T) {
		filter := &racing.ListRacesRequestFilter{
			Visible:    proto.Bool(false),
			MeetingIds: []int64{2},
		}

		req := &racing.ListRacesRequest{
			Filter: filter,
		}

		resp, err := service.ListRaces(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Races) != 1 {
			t.Errorf("expected 1 race, got %d", len(resp.Races))
		}
		if resp.Races[0].MeetingId != 2 {
			t.Errorf("expected MeetingId 2, got %d", resp.Races[0].MeetingId)
		}
	})
	t.Run("filters to races that have visible set to true and non-matching meeting IDs", func(t *testing.T) {
		filter := &racing.ListRacesRequestFilter{
			Visible:    proto.Bool(true),
			MeetingIds: []int64{2},
		}

		req := &racing.ListRacesRequest{
			Filter: filter,
		}

		resp, err := service.ListRaces(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Races) != 0 {
			t.Errorf("expected 0 races, got %d", len(resp.Races))
			t.Errorf("races: %v", resp.Races)
		}
	})
}

func TestListRacesOrderBy(t *testing.T) {
	repo := &mockRacesRepo{}
	service := NewRacingService(repo)
	ctx := context.Background()

	t.Run("returns races ordered by advertised_start_time in descending order", func(t *testing.T) {
		orderBy := "advertised_start_time desc"

		req := &racing.ListRacesRequest{
			OrderBy: orderBy,
		}

		resp, err := service.ListRaces(ctx, req)

		if resp.Races[0].AdvertisedStartTime.Seconds < resp.Races[1].AdvertisedStartTime.Seconds {
			t.Errorf("expected %d to be greater than %d", resp.Races[0].AdvertisedStartTime.Seconds, resp.Races[1].AdvertisedStartTime.Seconds)
		}

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("returns races ordered by advertised_start_time in ascending order", func(t *testing.T) {
		orderBy := "advertised_start_time"
		req := &racing.ListRacesRequest{
			OrderBy: orderBy,
		}

		resp, err := service.ListRaces(ctx, req)

		if resp.Races[0].AdvertisedStartTime.Seconds > resp.Races[1].AdvertisedStartTime.Seconds {
			t.Errorf("expected %d to be less than %d", resp.Races[0].AdvertisedStartTime.Seconds, resp.Races[1].AdvertisedStartTime.Seconds)
		}

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// Does the service correctly get races with their statuses?
// This is a test to ensure that the service layer correctly retrieves the status of each race.
func TestRacesHaveStatus(t *testing.T) {
	repo := &mockRacesRepo{}
	service := NewRacingService(repo)
	ctx := context.Background()
	t.Run("returns races with correct status", func(t *testing.T) {
		req := &racing.ListRacesRequest{}
		resp, err := service.ListRaces(ctx, req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Races) != 2 {
			t.Errorf("expected 2 races, got %d", len(resp.Races))
		}

		if resp.Races[0].Status != racing.Race_CLOSED {
			t.Errorf("expected race status to be %v, got %v", racing.Race_CLOSED, resp.Races[0].Status)
		}

		if resp.Races[1].Status != racing.Race_OPEN {
			t.Errorf("expected race status to be %v, got %v", racing.Race_OPEN, resp.Races[1].Status)
		}
	})
}
