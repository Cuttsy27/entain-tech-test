package errors

const ErrInvalidRaceID = RacingErr("invalid Race ID")

type RacingErr string

func (e RacingErr) Error() string {
	return string(e)
}
