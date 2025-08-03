package db

import (
	"database/sql"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"

	"git.neds.sh/matty/entain/sports/proto/sports"

	"git.neds.sh/matty/entain/sports/utils"
)

var advertised_start_time = "advertised_start_time"
var sport_filter = "sport"
var visible_filter = "visible"

// The properties a user is allowed to filter events by
var validFilterProperties = map[string]bool{sport_filter: true, visible_filter: true}

// Add more in the future as we want to order by other properties
var validOrderByProperties = map[string]bool{advertised_start_time: true}

var (
	// Currently only matches `=` operators, e.g. `sport = "Soccer"`
	// Could add other operators in the future to enhance
	eqPattern = regexp.MustCompile(`(?i)^(\w+)\s*=\s*("[^"]+"|\w+|true|false|\d+)$`)
	inPattern = regexp.MustCompile(`(?i)^(\w+)\s+IN\s+\(\s*("[^"]+"|\w+)(\s*,\s*("[^"]+"|\w+))*\s*\)$`)
)

// SportsRepo provides repository access to sports.
type SportsRepo interface {
	// Init will initialise our sports repository.
	Init() error

	// List will return a list of sports.
	ListEvents(filter string, orderBy string) ([]*sports.Event, error)

	// Get will return a specific Event by its ID.
	GetEvent(id int64) (*sports.Event, error)
}

type sportsRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewSportsRepo creates a new sports repository.
func NewSportsRepo(db *sql.DB) SportsRepo {
	return &sportsRepo{db: db}
}

// Init prepares the sport repository dummy data.
func (r *sportsRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy sports.
		err = r.seed()
	})

	return err
}

func (r *sportsRepo) ListEvents(filter string, orderBy string) ([]*sports.Event, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getEventQueries()[eventsList]

	query, args = r.applyFilter(query, filter)

	query = r.applyOrderBy(query, orderBy)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	// Get the events as variable so they can be modified before returning
	events, err := r.scanEvents(rows)

	// Return early if there was an error
	if err != nil {
		return nil, err
	}

	// Add status to each event derived from their AdvertisedStartTime
	r.addStatusToEvents(events)

	return events, nil
}

func (r *sportsRepo) GetEvent(id int64) (*sports.Event, error) {
	var (
		query string
		args  []interface{}
	)

	query = getEventQueries()[eventsGet]

	args = append(args, id)

	row := r.db.QueryRow(query, args...)

	event, err := r.scanEvent(row)

	if err != nil {
		return nil, err
	}

	r.addStatusToEvent(event)

	return event, nil

}

func (r *sportsRepo) applyFilter(query string, filter string) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == "" {
		return query, args
	}

	// Find valid filter properties in the provided filter string.
	// the filter string may be formatted like: sport IN ("Soccer", "Basketball") AND visible = true
	filter = strings.TrimSpace(filter)
	filterConditions := strings.Split(filter, " AND ")
	for _, condition := range filterConditions {
		condition = strings.TrimSpace(condition)
		if eqPattern.MatchString(condition) {
			matches := eqPattern.FindStringSubmatch(condition)
			field := matches[1]
			value := matches[2]

			if !validFilterProperties[field] {
				continue
			}

			value = strings.Trim(value, `"`)
			clauses = append(clauses, field+" = ?")

			// Add other bool fields as needed
			if field == visible_filter {
				if value == "true" {
					args = append(args, true)
				} else if value == "false" {
					args = append(args, false)
				} else {
					// If the value is not a valid boolean, we can skip this condition or return an error.
					continue
				}
			} else {
				args = append(args, value)
			}

		} else if inPattern.MatchString(condition) {
			matches := inPattern.FindStringSubmatch(condition)
			field := matches[1]
			valuesRaw := matches[2] + matches[3] // matches[2] is first value, matches[3] is rest

			// Validate field
			if !validFilterProperties[field] {
				continue // or return error
			}

			// Split values by comma, trim spaces and quotes
			values := []string{}
			for _, v := range strings.Split(valuesRaw, ",") {
				v = strings.TrimSpace(v)
				v = strings.Trim(v, `"`)
				values = append(values, v)
			}

			// Build IN clause with correct number of placeholders
			placeholders := strings.Repeat("?,", len(values)-1) + "?"
			clauses = append(clauses, field+" IN ("+placeholders+")")
			for _, v := range values {
				args = append(args, v)
			}
		} else {
			// Invalid format, skip or return error
			continue
		}
	}

	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	return query, args
}

func (repo *sportsRepo) applyOrderBy(query string, orderBy string) string {
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
			isValidField := validOrderByProperties[fieldName]
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

func (repo *sportsRepo) addStatusToEvents(events []*sports.Event) {
	for _, event := range events {
		repo.addStatusToEvent(event)
	}
}

func (repo *sportsRepo) addStatusToEvent(event *sports.Event) {
	if utils.IsProtoTimestampInPast(event.AdvertisedStartTime, utils.GetCurrentProtoTimestamp()) {
		event.Status = sports.Event_CLOSED
	} else {
		event.Status = sports.Event_OPEN
	}
}

func (m *sportsRepo) scanEvents(
	rows *sql.Rows,
) ([]*sports.Event, error) {
	var events []*sports.Event

	for rows.Next() {
		var event sports.Event
		var advertisedStartTime time.Time

		if err := rows.Scan(&event.Id, &event.Name, &event.Sport, &event.Visible, &advertisedStartTime); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		ts, err := ptypes.TimestampProto(advertisedStartTime)
		if err != nil {
			return nil, err
		}

		event.AdvertisedStartTime = ts

		events = append(events, &event)
	}

	return events, nil
}

func (m *sportsRepo) scanEvent(
	row *sql.Row,
) (*sports.Event, error) {
	var event sports.Event
	var advertisedStartTime time.Time

	if err := row.Scan(&event.Id, &event.Name, &event.Sport, &event.Visible, &advertisedStartTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	ts, err := ptypes.TimestampProto(advertisedStartTime)
	if err != nil {
		return nil, err
	}

	event.AdvertisedStartTime = ts

	return &event, nil
}
