// Package sqlite3 implements a Go interface to the SQLite database.
package sqlite3

// #cgo LDFLAGS: -lsqlite3
// #include <sqlite3.h>
import "C"

// Initialize starts the SQLite3 engine.
func Initialize() {
	C.sqlite3_initialize()
}

// Shutdown stops the SQLite3 engine.
func Shutdown()	{
	C.sqlite3_shutdown()
}

// Session initializes a database and calls `f` to access it.
func Session(filename string, f func(db *Database)) {
	Initialize()
	defer Shutdown()
	if db, e := Open(filename); e == nil {
		defer db.Close()
		f(db)
	}
}

// TransientDatabase initializes a in-memory database and calls `f` to 
// access it.
func TransientSession(f func(db *Database)) {
	Initialize()
	defer Shutdown()
	if db := TransientDatabase(); db.Open() == nil {
		defer db.Close()
		f(db)
	}
}

// LibVersion returns the version of the SQLite3 engine.
func LibVersion() string {
	return C.GoString(C.sqlite3_libversion())
}

// Value represents any SQLite3 value.
type Value struct {
	cptr *C.sqlite3_value
}

// Blob represents the SQLite3 blob type.
type Blob struct {
	cptr *C.sqlite3_blob
}
