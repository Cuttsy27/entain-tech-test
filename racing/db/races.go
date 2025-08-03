package db

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"

	"git.neds.sh/matty/entain/racing/proto/racing"

	"git.neds.sh/matty/entain/racing/utils"
)

var advertised_start_time = "advertised_start_time"
var validOrderByFields = []string{advertised_start_time}

// RacesRepo provides repository access to races.
type RacesRepo interface {
	// Init will initialise our races repository.
	Init() error

	// List will return a list of races.
	List(filter *racing.ListRacesRequestFilter, orderBy string) ([]*racing.Race, error)
}

type racesRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewRacesRepo creates a new races repository.
func NewRacesRepo(db *sql.DB) RacesRepo {
	return &racesRepo{db: db}
}

// Init prepares the race repository dummy data.
func (r *racesRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy races.
		err = r.seed()
	})

	return err
}

func (r *racesRepo) List(filter *racing.ListRacesRequestFilter, orderBy string) ([]*racing.Race, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getRaceQueries()[racesList]

	query, args = r.applyFilter(query, filter)

	query = r.applyOrderBy(query, orderBy)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	// Get the races as variable so they can be modified before returning
	races, err := r.scanRaces(rows)

	// Return early if there was an error
	if err != nil {
		return nil, err
	}

	// Add status to each race derived from their AdvertisedStartTime
	r.addStatusToRaces(races)

	return races, nil
}

func (r *racesRepo) applyFilter(query string, filter *racing.ListRacesRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}

	// Check if "Visible" is set as a filter option, and if so, add it to the query clauses.
	if filter.Visible != nil {
		clauses = append(clauses, "visible = ?")
		args = append(args, *filter.Visible)
	}

	if len(filter.MeetingIds) > 0 {
		clauses = append(clauses, "meeting_id IN ("+strings.Repeat("?,", len(filter.MeetingIds)-1)+"?)")

		for _, meetingID := range filter.MeetingIds {
			args = append(args, meetingID)
		}
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	return query, args
}

func (repo *racesRepo) applyOrderBy(query string, orderBy string) string {
	// Using a dict to set default order_by fields that can be overwritten by the user.
	orderByDict := map[string]string{
		advertised_start_time: "asc",
	}

	orderBy = strings.TrimSpace(orderBy)
	// split string by commas to allow multiple order by fields
	orderByFields := strings.Split(orderBy, ",")

	if len(orderByFields) > 0 && orderBy != "" {
		// If there are order_by fields provided by the user, loop and check for valid fields before updating the orderByDict
		for _, field := range orderByFields {
			field = strings.TrimSpace(field)

			// Extract the field name (before any direction like ASC or DESC)
			fieldName := strings.Fields(field)[0]

			// Check if the field is valid
			isValidField := utils.Contains(validOrderByFields, fieldName)
			if !isValidField {
				continue // Skip invalid fields
			}

			fieldDirection := "asc"

			if len(strings.Fields(field)) > 1 {
				fieldDirection = strings.ToLower(strings.Fields(field)[1])
				orderByDict[fieldName] = fieldDirection
			}
		}
	}

	// We have a default orderBy, so we always append to the query
	query += " ORDER BY "

	queriesToAdd := []string{}

	// Iterate the map to apply each field and its direction
	for orderBy, direction := range orderByDict {
		queriesToAdd = append(queriesToAdd, orderBy+" "+direction)
	}

	query += strings.Join(queriesToAdd, ", ")

	// Return the potentially modified query
	// If no valid fields were found, the query remains unchanged.
	return query
}

func (repo *racesRepo) addStatusToRaces(races []*racing.Race) {
	for _, race := range races {
		repo.addStatusToRace(race)
	}
}

func (repo *racesRepo) addStatusToRace(race *racing.Race) {
	if utils.IsProtoTimestampInPast(race.AdvertisedStartTime, utils.GetCurrentProtoTimestamp()) {
		race.Status = racing.Race_CLOSED
	} else {
		race.Status = racing.Race_OPEN
	}
}

func (m *racesRepo) scanRaces(
	rows *sql.Rows,
) ([]*racing.Race, error) {
	var races []*racing.Race

	for rows.Next() {
		var race racing.Race
		var advertisedStart time.Time

		if err := rows.Scan(&race.Id, &race.MeetingId, &race.Name, &race.Number, &race.Visible, &advertisedStart); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		ts, err := ptypes.TimestampProto(advertisedStart)
		if err != nil {
			return nil, err
		}

		race.AdvertisedStartTime = ts

		races = append(races, &race)
	}

	return races, nil
}
