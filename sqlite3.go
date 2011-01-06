package sqlite3

// #include <sqlite3.h>
import "C"

func Initialize() {
	C.sqlite3_initialize()
}

func Shutdown()	{
	C.sqlite3_shutdown()
}

func Session(filename string, f func(db *Database)) {
	Initialize()
	defer Shutdown()
	if db, e := Open(filename); e == OK {
		defer db.Close()
		f(db)
	}
}

func TransientSession(f func(db *Database)) {
	Initialize()
	defer Shutdown()
	if db := TransientDatabase(); db.Open() == OK {
		defer db.Close()
		f(db)
	}
}

func LibVersion() string {
	return C.GoString(C.sqlite3_libversion())
}

type Value struct {
	cptr *C.sqlite3_value
}

type Blob struct {
	cptr *C.sqlite3_blob
}