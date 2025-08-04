package db

import (
	"time"

	"syreclabs.com/go/faker"
)

var sportsNames = []string{
	"Soccer",
	"Basketball",
	"Baseball",
	"Tennis",
	"Cricket",
	"Rugby",
	"Golf",
	"Boxing",
	"Swimming",
	"Volleyball",
	"Hockey",
	"Table Tennis",
	"Badminton",
	"American Football",
	"Formula 1",
	"Motorsport",
	"Handball",
	"Snooker",
	"Bowling",
	"Martial Arts",
}

var eventNames = []string{
	"Championship Final",
	"Grand Prix",
	"World Cup Qualifier",
	"Friendly Match",
	"Season Opener",
	"Playoff Semifinal",
	"Exhibition Game",
	"International Test",
	"Regional Tournament",
	"Knockout Round",
	"League Match",
	"Open Invitational",
	"Charity Event",
	"All-Star Game",
	"Title Bout",
	"Gold Medal Match",
	"Quarterfinal Clash",
	"Super Series",
	"Masters Tournament",
	"Classic Derby",
}

func (r *sportsRepo) seed() error {
	statement, err := r.db.Prepare(`CREATE TABLE IF NOT EXISTS events (id INTEGER PRIMARY KEY, name TEXT, sport TEXT, visible INTEGER, advertised_start_time DATETIME)`)
	if err == nil {
		_, err = statement.Exec()
	}

	for i := 1; i <= 100; i++ {
		statement, err = r.db.Prepare(`INSERT OR IGNORE INTO events(id, name, sport, visible, advertised_start_time) VALUES (?,?,?,?,?)`)
		if err == nil {
			_, err = statement.Exec(
				i,
				eventNames[i%len(eventNames)],
				sportsNames[i%len(sportsNames)],
				faker.Number().Between(0, 1),
				faker.Time().Between(time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 2)).Format(time.RFC3339),
			)
		}
	}

	return err
}
