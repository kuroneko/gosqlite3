package sqlite3

// #include <sqlite3.h>
import "C"
import "fmt"
import "os"
import "syscall"
import "time"

type Errno int

func (e Errno) String() (err string) {
	if err = errText[e]; err == "" {
		err = fmt.Sprintf("errno %v", int(e))
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
	ENCODER		= Errno(1000)
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
	ENCODER:	"blob encoding failed",
}

type Database struct {
	handle		*C.sqlite3
	Filename	string
	Flags		C.int
}

func TransientDatabase() (db *Database) {
	return &Database{ Filename: ":memory:" }	
}

func Open(filename string, flags... int) (db *Database, e os.Error) {
	defer func() {
		if x := recover(); x != nil {
			db.Close()
			db = nil
			e = MISUSE
		}
	}()
	db = &Database{ Filename: filename }
	if len(flags) == 0 {
		e = db.Open(C.SQLITE_OPEN_FULLMUTEX, C.SQLITE_OPEN_READWRITE, C.SQLITE_OPEN_CREATE)
	} else {
		e = db.Open(flags...)
	}
	return
}

func (db *Database) Open(flags... int) (e os.Error) {
	if C.sqlite3_threadsafe() == 0 {
		panic("sqlite library is not thread-safe")
	}
	if db.handle != nil {
		e = CANTOPEN
	} else {
		db.Flags = 0
		for _, v := range flags { db.Flags = db.Flags | C.int(v) }
		if err := Errno(C.sqlite3_open_v2(C.CString(db.Filename), &db.handle, db.Flags, nil)); err != OK {
			e = err
		} else if db.handle == nil {
			e = CANTOPEN
		}
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

func (db *Database) Error() os.Error {
	return Errno(C.sqlite3_errcode(db.handle))
}

func (db *Database) Prepare(sql string, values... interface{}) (s *Statement, e os.Error) {
	s = &Statement{ db: db, timestamp: time.Nanoseconds() }
	if rv := Errno(C.sqlite3_prepare_v2(db.handle, C.CString(sql), -1, &s.cptr, nil)); rv != OK {
		s, e = nil, rv
	} else {
		if len(values) > 0 {
			e, _ = s.BindAll(values...)
		}
	}
	return
}

func (db *Database) Execute(sql string, f... func(*Statement, ...interface{})) (c int, e os.Error) {
	var st	*Statement
	st, e = db.Prepare(sql)
	if e == nil {
		c, e = st.All(f...)
	}
	if e == OK {
		e = nil
	}
	return
}

func (db *Database) Begin() (e os.Error) {
	_, e = db.Execute("BEGIN")
	return
}

func (db *Database) Rollback() (e os.Error) {
	_, e = db.Execute("ROLLBACK")
	return
}

func (db *Database) Commit() (e os.Error) {
	_, e = db.Execute("COMMIT")
	return
}

func (db *Database) Load(source *Database, dbname string) (e os.Error) {
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

func (db *Database) Save(target *Database, dbname string) (e os.Error) {
	return target.Load(db, dbname)
}

type ProgressReport struct {
	os.Error
	PageCount		int
	Remaining		int
	Source			string
	Target			string
}

type Reporter chan *ProgressReport

type BackupParameters struct {
	Target			string
	PagesPerStep	int
	QueueLength		int
	Interval		int64
}

func (db *Database) Backup(p BackupParameters) (r Reporter, e os.Error) {
	if target, e := Open(p.Target); e == nil {
		if backup, e := NewBackup(target, "main", db, "main"); e == nil && p.PagesPerStep > 0 {
			r = make(Reporter, p.QueueLength)
			go func() {
				defer target.Close()
				defer backup.Finish()
				defer close(r)
				for {
					report := &ProgressReport{
								Source: db.Filename,
								Target: p.Target,
								Error: backup.Step(p.PagesPerStep),
								PageCount: backup.PageCount(),
								Remaining: backup.Remaining(),
								}
					r <- report
					if e, ok := report.Error.(Errno); ok && !(e == OK || e == BUSY || e == LOCKED) {
						break
					}
					if p.Interval > 0 {
						syscall.Sleep(p.Interval)
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