package db

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// pgForeignKeyViolation is SQLSTATE 23503, raised when a write references a row
// that is not there.
const pgForeignKeyViolation = "23503"

// IsForeignKeyViolation reports whether err is Postgres rejecting a write for
// referencing a parent row that does not exist.
//
// It exists because CDC and foreign keys disagree about ordering. Debezium makes
// no promise about the order of events across tables, so an episode can arrive
// before the anime it belongs to and Postgres will refuse it. MySQL never
// surfaced this: the constraints were declared but Vitess did not enforce them.
//
// A caller that treats this as a normal error retries forever. The parent will
// not appear by retrying the child -- it arrives on its own topic, on its own
// schedule -- so the retry occupies the consumer and the backlog stops moving.
// That is what stalled staging for hours.
//
// Discarding the event is the right response for these tables. Every one of them
// is a replica of a source of truth elsewhere; the row will be delivered again
// the next time the source updates it, and a full snapshot replays it regardless.
// Nothing is lost that cannot be recovered, whereas a stalled consumer loses
// everything queued behind it.
func IsForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgForeignKeyViolation
	}
	return false
}
