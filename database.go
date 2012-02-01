package sqlite3

// #include <sqlite3.h>
import "C"
import (
	"fmt"
	"strconv"
	"time"
)

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

type Database struct {
	handle     *C.sqlite3
	Filename   string
	Flags      C.int
	Savepoints []interface{}
}

// TransientDatabase returns a handle to an in-memory database.
func TransientDatabase() (db *Database) {
	return &Database{Filename: ":memory:"}
}

// Open returns a handle to the sqlite3 database specified by filename.
func Open(filename string, flags ...int) (db *Database, e error) {
	defer func() {
		if x := recover(); x != nil {
			db.Close()
			db = nil
			e = MISUSE
		}
	}()
	db = &Database{Filename: filename}
	if len(flags) == 0 {
		e = db.Open(C.SQLITE_OPEN_FULLMUTEX, C.SQLITE_OPEN_READWRITE, C.SQLITE_OPEN_CREATE)
	} else {
		e = db.Open(flags...)
	}
	return
}

// Open initializes and opens the database.
func (db *Database) Open(flags ...int) (e error) {
	if C.sqlite3_threadsafe() == 0 {
		panic("sqlite library is not thread-safe")
	}
	if db.handle != nil {
		e = CANTOPEN
	} else {
		db.Flags = 0
		for _, v := range flags {
			db.Flags = db.Flags | C.int(v)
		}
		if err := Errno(C.sqlite3_open_v2(C.CString(db.Filename), &db.handle, db.Flags, nil)); err != OK {
			e = err
		} else if db.handle == nil {
			e = CANTOPEN
		}
	}
	return
}

// Close shuts down the database engine for this database.
func (db *Database) Close() {
	C.sqlite3_close(db.handle)
	db.handle = nil
}

// LastInsertRowID returns the id of the most recent successful INSERT.
//
// Each entry in an SQLite table has a unique 64-bit signed integer key 
// called the "rowid". The rowid is always available as an undeclared column 
// named ROWID, OID, or _ROWID_ as long as those names are not also used by 
// explicitly declared columns. If the table has a column of type 
// INTEGER PRIMARY KEY then that column is another alias for the rowid.
//
// This routine returns the rowid of the most recent successful INSERT into 
// the database from the database connection in the first argument. As of 
// SQLite version 3.7.7, this routines records the last insert rowid of both 
// ordinary tables and virtual tables. If no successful INSERTs have ever 
// occurred on that database connection, zero is returned.
func (db *Database) LastInsertRowID() int64 {
	return int64(C.sqlite3_last_insert_rowid(db.handle))
}

// Changes returns the number of database rows that were changed or inserted 
// or deleted by the most recently completed SQL statement.
func (db *Database) Changes() int {
	return int(C.sqlite3_changes(db.handle))
}

// TotalChanges retruns the number of row changes. 
//
// This function returns the number of row changes caused by INSERT, UPDATE 
// or DELETE statements since the database connection was opened. The count 
// returned by TotalChanges includes all changes from all trigger contexts 
// and changes made by foreign key actions. However, the count does not 
// include changes used to implement REPLACE constraints, do rollbacks or 
// ABORT processing, or DROP TABLE processing. The count does not include 
// rows of views that fire an INSTEAD OF trigger, though if the INSTEAD OF 
// trigger makes changes of its own, those changes are counted. The 
// TotalChanges function counts the changes as soon as the statement that 
// makes them is completed.
func (db *Database) TotalChanges() int {
	return int(C.sqlite3_total_changes(db.handle))
}

// Error returns the numeric result code for the most recent failed database
// call.
func (db *Database) Error() error {
	return Errno(C.sqlite3_errcode(db.handle))
}

// Prepare compiles the SQL query into a byte-code program and binds the 
// supplied values.
func (db *Database) Prepare(sql string, values ...interface{}) (s *Statement, e error) {
	s = &Statement{db: db, timestamp: time.Now().UnixNano()}
	if rv := Errno(C.sqlite3_prepare_v2(db.handle, C.CString(sql), -1, &s.cptr, nil)); rv != OK {
		s, e = nil, rv
	} else {
		if len(values) > 0 {
			e, _ = s.BindAll(values...)
		}
	}
	return
}

// Execute runs the SQL statement. 
func (db *Database) Execute(sql string, f ...func(*Statement, ...interface{})) (c int, e error) {
	var st *Statement
	st, e = db.Prepare(sql)
	if e == nil {
		c, e = st.All(f...)
	}
	if e == OK {
		e = nil
	}
	return
}

// Begin initializes a SQL Transaction block.
func (db *Database) Begin() (e error) {
	_, e = db.Execute("BEGIN")
	return
}

// Rollback reverts the changes since the most recent Begin() call.
func (db *Database) Rollback() (e error) {
	_, e = db.Execute("ROLLBACK")
	return
}

// Commit ends the current transaction and makes all changes performed in the
// transaction permanent.
func (db *Database) Commit() (e error) {
	_, e = db.Execute("COMMIT")
	return
}

func savepointID(id interface{}) (s string) {
	switch id := id.(type) {
	case string:
		s = id
	case []byte:
		s = string(id)
	case fmt.Stringer:
		s = id.String()
	case int:
		s = strconv.Itoa(id)
	case uint:
		s = strconv.FormatUint(uint64(id), 10)
	default:
		panic(SAVEPOINT)
	}
	return
}

// Mark creates a SAVEPOINT.
//
// A SAVEPOINT is a method of creating transactions, similar to BEGIN and
// COMMIT, except that Mark and MergeSteps are named and may be nested.
func (db *Database) Mark(id interface{}) (e error) {
	if st, err := db.Prepare("SAVEPOINT ?", savepointID(id)); err == nil {
		_, e = st.All()
	} else {
		e = err
	}
	return
}

// MergeSteps can be seen as the equivalent of COMMIT for a Mark command.
//
//   More specificly ...
//   - Some people view RELEASE as the equivalent of COMMIT for a SAVEPOINT.
//     This is an acceptable point of view as long as one remembers that the
//     changes committed by an inner transaction might later be undone by a
//     rollback in an outer transaction.
//   - Another view of RELEASE is that it merges a named transaction into its
//     parent transaction, so that the named transaction and its parent 
//     become the same transaction. After RELEASE, the named transaction and 
//     its parent will commit or rollback together, whatever their fate may 
//     be.
//   - One can also think of savepoints as "marks" in the transaction 
//     timeline. In this view, the SAVEPOINT command creates a new mark, the 
//     ROLLBACK TO command rewinds the timeline back to a point just after 
//     the named mark, and the RELEASE command erases marks from the timeline
//     without actually making any changes to the database.
func (db *Database) MergeSteps(id interface{}) (e error) {
	if st, err := db.Prepare("RELEASE SAVEPOINT ?", savepointID(id)); err == nil {
		_, e = st.All()
	} else {
		e = err
	}
	return
}

// Release rolls back all transactions to the specified SAVEPOINT (Mark).
func (db *Database) Release(id interface{}) (e error) {
	if st, err := db.Prepare("ROLLBACK TRANSACTION TO SAVEPOINT ?", savepointID(id)); err == nil {
		_, e = st.All()
	} else {
		e = err
	}
	return
}

// SavePoints returns the currently active SAVEPOINTs created using Mark.
func (db *Database) SavePoints() (s []interface{}) {
	s = make([]interface{}, len(db.Savepoints))
	copy(s, db.Savepoints)
	return
}

// Load creates a backup of the source database and loads that.
func (db *Database) Load(source *Database, dbname string) (e error) {
	if dbname == "" {
		dbname = "main"
	}
	if backup, rv := NewBackup(db, dbname, source, dbname); rv == nil {
		e = backup.Full()
	} else {
		e = rv
	}
	return
}

// Save stores the content of the database in the target database.
func (db *Database) Save(target *Database, dbname string) (e error) {
	return target.Load(db, dbname)
}

type Reporter chan *ProgressReport

type BackupParameters struct {
	Target       string
	PagesPerStep int
	QueueLength  int
	Interval     time.Duration
}

// Backup creates a copy (backup) of the current database to the Target file 
// specified in BackupParameters.
func (db *Database) Backup(p BackupParameters) (r Reporter, e error) {
	if target, e := Open(p.Target); e == nil {
		if backup, e := NewBackup(target, "main", db, "main"); e == nil && p.PagesPerStep > 0 {
			r = make(Reporter, p.QueueLength)
			go func() {
				defer target.Close()
				defer backup.Finish()
				defer close(r)
				for {
					report := &ProgressReport{
						Source:    db.Filename,
						Target:    p.Target,
						Error:     backup.Step(p.PagesPerStep),
						Total:     backup.PageCount(),
						Remaining: backup.Remaining(),
					}
					r <- report
					if e, ok := report.Error.(Errno); ok && !(e == OK || e == BUSY || e == LOCKED) {
						break
					}
					if p.Interval > 0 {
						time.Sleep(p.Interval)
					}
				}
			}()
		} else {
			target.Close()
			e = target.Error()
		}
	}
	return
}
