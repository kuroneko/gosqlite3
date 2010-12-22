package sqlite3

// #include <sqlite3.h>
import "C"
import "fmt"
import "os"
import "time"

const(
	SQLITE_INTEGER = 1
	SQLITE_FLOAT = 2
	SQLITE3_TEXT = 3
	SQLITE_BLOB = 4
	SQLITE_NULL = 5
)

type Errno int

func (e Errno) String() (err string) {
	if err = errText[e]; err == "" {
		err = fmt.Sprintf("errno %d", e)
	}
	return 
}

const(
	OK			= Errno(iota)
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
	ROW			= Errno(100)
	DONE		= Errno(101)
)

var errText = map[Errno]string {
	ERROR:		"SQL error or missing database",
	INTERNAL:	"Internal logic error in SQLite",
	PERM:		"Access permission denied",
	ABORT:		"Callback routine requested an abort",
	BUSY:		"The database file is locked",
	LOCKED:		"A table in the database is locked",
	NOMEM:		"A malloc() failed",
	READONLY:	"Attempt to write a readonly database",
	INTERRUPT:	"Operation terminated by sqlite3_interrupt()",
	IOERR:		"Some kind of disk I/O error occurred",
	CORRUPT:	"The database disk image is malformed",
	NOTFOUND:	"NOT USED. Table or record not found",
	FULL:		"Insertion failed because database is full",
	CANTOPEN:	"Unable to open the database file",
	PROTOCOL:	"NOT USED. Database lock protocol error",
	EMPTY:		"Database is empty",
	SCHEMA:		"The database schema changed",
	TOOBIG:		"String or BLOB exceeds size limit",
	CONSTRAINT:	"Abort due to constraint violation",
	MISMATCH:	"Data type mismatch",
	MISUSE:		"Library used incorrectly",
	NOLFS:		"Uses OS features not supported on host",
	AUTH:		"Authorization denied",
	FORMAT:		"Auxiliary database format error",
	RANGE:		"2nd parameter to sqlite3_bind out of range",
	NOTDB:		"File opened that is not a database file",
	ROW:		"sqlite3_step() has another row ready",
	DONE:		"sqlite3_step() has finished executing",
}

type Database struct {
	handle		*C.sqlite3
	Filename	string
	Flags		C.int
}

func Open(filename string, flags... int) (db *Database, e os.Error) {
	defer func() {
		if x := recover(); x != nil {
			db.Close()
			db = nil
			e = MISUSE
		}
	}()
	db = new(Database)
	db.Filename = filename
	switch len(flags) {
	case 0:		db.Flags = C.SQLITE_OPEN_FULLMUTEX | C.SQLITE_OPEN_READWRITE | C.SQLITE_OPEN_CREATE
	default:	for _, v := range flags { db.Flags = db.Flags | C.int(v) }
	}
	e = db.Open()
	return
}

func (db *Database) Open() (e os.Error) {
	if C.sqlite3_threadsafe() == 0 {
		panic("sqlite library is not thread-safe")
	} else if rv := C.sqlite3_open_v2(C.CString(db.Filename), &db.handle, db.Flags, nil); rv != 0 {
		e = Errno(rv)
	} else if &db.handle == nil {
		panic("sqlite failed to return a database")
	} else {
		e = OK
	}
	return
}

func (db *Database) Close() {
	C.sqlite3_close(db.handle)
	db.handle = nil
}

func (db *Database) LastInsertRowID() int64 {
	return int64(C.sqlite3_last_insert_rowid(db.handle))
}

func (db *Database) Changes() int {
	return int(C.sqlite3_changes(db.handle))
}

func (db *Database) TotalChanges() int {
	return int(C.sqlite3_total_changes(db.handle))
}

func (db *Database) Error() (e string) {
	return C.GoString(C.sqlite3_errmsg(db.handle))
}

func (db *Database) Prepare(sql string) (s *Statement, e os.Error) {
	s = &Statement{ db: db, SQL: sql, timestamp: time.Nanoseconds() }
	if e = Errno(C.sqlite3_prepare_v2(db.handle, C.CString(sql), -1, &s.cptr, nil)); e != OK {
		s = nil
	}
	return
}