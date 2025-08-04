package errors

const ErrInvalidEventID = EventErr("invalid Event ID")

type EventErr string

func (e EventErr) Error() string {
	return string(e)
}
