package sqlite3

// #include <sqlite3.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unsafe"
)


type DBFlag	int

func (d DBFlag) String() string {
	flags := []string{}
	for i := O_READONLY; i < d; i <<= 1 {
		if s, ok := flagText[d & i]; ok {
			flags = append(flags, s)
		}
	}
	return strings.Join(flags, "|")
}

const (
	O_READONLY DBFlag =			0x00000001
	O_READWRITE DBFlag =		0x00000002
	O_CREATE DBFlag =			0x00000004
	O_DELETEONCLOSE DBFlag =	0x00000008
	O_EXCLUSIVE DBFlag =		0x00000010
	O_AUTOPROXY DBFlag =		0x00000020
	O_URI DBFlag =				0x00000040
	O_MAIN_DB DBFlag =			0x00000100
	O_TEMP_DB DBFlag =			0x00000200
	O_TRANSIENT_DB DBFlag =		0x00000400
	O_MAIN_JOURNAL DBFlag =		0x00000800
	O_TEMP_JOURNAL DBFlag =		0x00001000
	O_SUBJOURNAL DBFlag =		0x00002000
	O_MASTER_JOURNAL DBFlag =	0x00004000
	O_NOMUTEX DBFlag =			0x00008000
	O_FULLMUTEX DBFlag =		0x00010000
	O_SHAREDCACHE DBFlag =		0x00020000
	O_PRIVATECACHE DBFlag =		0x00040000
	O_WAL DBFlag =				0x00080000
)

var flagText = map[DBFlag]string{
	O_READONLY:			"O_READONLY",
	O_READWRITE:		"O_READWRITE",
	O_CREATE:			"O_CREATE",
	O_DELETEONCLOSE:	"O_DELETEONCLOSE",
	O_EXCLUSIVE:		"O_EXCLUSIVE",
	O_AUTOPROXY:		"O_AUTOPROXY",
	O_URI:				"O_URI",
	O_MAIN_DB:			"O_MAIN_DB",
	O_TEMP_DB:			"O_TEMP_DB",
	O_TRANSIENT_DB:		"O_TRANSIENT_DB",
	O_MAIN_JOURNAL:		"O_MAIN_JOURNAL",
	O_TEMP_JOURNAL:		"O_TEMP_JOURNAL",
	O_SUBJOURNAL:		"O_SUBJOURNAL",
	O_MASTER_JOURNAL:	"O_MASTER_JOURNAL",
	O_NOMUTEX:			"O_NOMUTEX",
	O_FULLMUTEX:		"O_FULLMUTEX",
	O_SHAREDCACHE:		"O_SHAREDCACHE",
	O_PRIVATECACHE:		"O_PRIVATECACHE",
	O_WAL:				"O_WAL",
}

// Database implements high level view of the underlying database.
type Database struct {
	handle     *C.sqlite3
	Filename   string
	DBFlag
	Savepoints []interface{}
}

// TransientDatabase returns a handle to an in-memory database.
func TransientDatabase() (db *Database) {
	return &Database{Filename: ":memory:"}
}

// Open returns a handle to the sqlite3 database specified by filename.
func Open(filename string, flags ...DBFlag) (db *Database, e error) {
	defer func() {
		if x := recover(); x != nil {
			db.Close()
			db = nil
			e = MISUSE
		}
	}()
	db = &Database{Filename: filename}
	if len(flags) == 0 {
		e = db.Open(O_FULLMUTEX, O_READWRITE, O_CREATE)
	} else {
		e = db.Open(flags...)
	}
	return
}

// Open initializes and opens the database.
func (db *Database) Open(flags ...DBFlag) (e error) {
	if C.sqlite3_threadsafe() == 0 {
		panic("sqlite library is not thread-safe")
	}
	if db.handle != nil {
		e = CANTOPEN
	} else {
		db.DBFlag = 0
		for _, v := range flags {
			db.DBFlag |= v
		}

		cs := C.CString(db.Filename)
		defer C.free(unsafe.Pointer(cs))
		e = SQLiteError(C.sqlite3_open_v2(cs, &db.handle, C.int(db.DBFlag), nil))

		if e == nil && db.handle == nil {
			e = CANTOPEN
		}
	}
	return
}

// Close is used to close the database.
func (db *Database) Close() {
	C.sqlite3_close(db.handle)
	db.handle = nil
}

// LastInsertRowID returns the id of the most recently successful INSERT.
//
// Each entry in an SQLite table has a unique 64-bit signed integer key 
// called the "rowid". The rowid is always available as an undeclared column 
// named ROWID, OID, or _ROWID_ as long as those names are not also used by 
// explicitly declared columns. If the table has a column of type 
// INTEGER PRIMARY KEY then that column is another alias for the rowid.
//
// This routine returns the rowid of the most recently successful INSERT into 
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

// Error returns the numeric result code for the most recently failed database
// call.
func (db *Database) Error() error {
	return SQLiteError(C.sqlite3_errcode(db.handle))
}

// Prepare compiles the SQL query into a byte-code program and binds the 
// supplied values.
func (db *Database) Prepare(sql string, values ...interface{}) (s *Statement, e error) {
	s = &Statement{db: db, timestamp: time.Now().UnixNano()}
	cs := C.CString(sql)
	defer C.free(unsafe.Pointer(cs))
	if e = SQLiteError(C.sqlite3_prepare_v2(db.handle, cs, -1, &s.cptr, nil)); e != nil {
		s = nil
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
	if st, e = db.Prepare(sql); e == nil {
		c, e = st.All(f...)
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
func (db *Database) MergeSteps(id interface{}) (e error) {
	if st, err := db.Prepare("RELEASE SAVEPOINT ?", savepointID(id)); err == nil {
		_, e = st.All()
	} else {
		e = err
	}
	return
}

// Release rolls back all transactions to the specified SAVEPOINT (Mark).
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

// Backup creates a copy (backup) of the current database to the target file 
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
					if e := report.Error; !(e == nil || e == BUSY || e == LOCKED) {
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
