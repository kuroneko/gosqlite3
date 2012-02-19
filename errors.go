package sqlite3

import "C"
import "fmt"

type Errno int

func (e Errno) Error() (err string) {
	if err = errText[e]; err == "" {
		err = fmt.Sprintf("errno %v", int(e))
	}
	return
}

const (
	OK = Errno(iota)
	ERROR
	INTERNAL
	PERM
	ABORT
	BUSY
	LOCKED
	NOMEM
	READONLY
	INTERRUPT
	IOERR
	CORRUPT
	NOTFOUND
	FULL
	CANTOPEN
	PROTOCOL
	EMPTY
	SCHEMA
	TOOBIG
	CONSTRAINT
	MISMATCH
	MISUSE
	NOLFS
	AUTH
	FORMAT
	RANGE
	NOTDB
	ROW       = Errno(100)
	DONE      = Errno(101)
	ENCODER   = Errno(1000)
	SAVEPOINT = Errno(1001)
)

var errText = map[Errno]string{
	ERROR:      "SQL error or missing database",
	INTERNAL:   "Internal logic error in SQLite",
	PERM:       "Access permission denied",
	ABORT:      "Callback routine requested an abort",
	BUSY:       "The database file is locked",
	LOCKED:     "A table in the database is locked",
	NOMEM:      "A malloc() failed",
	READONLY:   "Attempt to write a readonly database",
	INTERRUPT:  "Operation terminated by sqlite3_interrupt()",
	IOERR:      "Some kind of disk I/O error occurred",
	CORRUPT:    "The database disk image is malformed",
	NOTFOUND:   "NOT USED. Table or record not found",
	FULL:       "Insertion failed because database is full",
	CANTOPEN:   "Unable to open the database file",
	PROTOCOL:   "NOT USED. Database lock protocol error",
	EMPTY:      "Database is empty",
	SCHEMA:     "The database schema changed",
	TOOBIG:     "String or BLOB exceeds size limit",
	CONSTRAINT: "Abort due to constraint violation",
	MISMATCH:   "Data type mismatch",
	MISUSE:     "Library used incorrectly",
	NOLFS:      "Uses OS features not supported on host",
	AUTH:       "Authorization denied",
	FORMAT:     "Auxiliary database format error",
	RANGE:      "2nd parameter to sqlite3_bind out of range",
	NOTDB:      "File opened that is not a database file",
	ROW:        "sqlite3_step() has another row ready",
	DONE:       "sqlite3_step() has finished executing",
	ENCODER:    "blob encoding failed",
	SAVEPOINT:  "invalid or unknown savepoint identifier",
}

func SQLiteError(code C.int) (e error) {
	if e = Errno(code); e == OK {
		e = nil
	}
	return
}