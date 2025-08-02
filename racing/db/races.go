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

	return r.scanRaces(rows)
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
	if orderBy == "" {
		return query
	}
	// Validate orderBy against a list of valid fields to prevent SQL injection.
	validFields := []string{"advertised_start_time"}

	orderBy = strings.TrimSpace(orderBy)
	// split string by commas to allow multiple order by fields
	orderByFields := strings.Split(orderBy, ",")

	queryOrderBy := []string{}

	for _, field := range orderByFields {
		field = strings.TrimSpace(field)

		// Extract the field name (before any direction like ASC or DESC)
		fieldName := strings.Fields(field)[0]

		// Check if the field is valid
		isValidField := utils.Contains(validFields, fieldName)
		if !isValidField {
			continue // Skip invalid fields
		}
		queryOrderBy = append(queryOrderBy, field)
	}

	// If queryOrderBy has fields it means they are valid and can be added to the query. Only add " ORDER BY " if there are valid fields.
	if len(queryOrderBy) > 0 {
		query += " ORDER BY " + strings.Join(queryOrderBy, ", ")
	}

	// Return the potentially modified query
	// If no valid fields were found, the query remains unchanged.
	return query
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
