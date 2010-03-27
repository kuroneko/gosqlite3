package sqlite3

// #include <sqlite3.h>
// int gosqlite3_bind_text(sqlite3_stmt* s, int p, const char* q, int n) {
//     return sqlite3_bind_text(s, p, q, n, SQLITE_TRANSIENT);
// }
import "C"
import "unsafe"

const (
    SQLITE_OK   = 0;
    SQLITE_ROW  = 100;
    SQLITE_DONE = 101;
);

type Blob struct {
	cptr *C.sqlite3_blob;
}

type Handle struct {
	cptr *C.sqlite3;
}

type Statement struct {
	cptr *C.sqlite3_stmt;
}

type Value struct {
	cptr *C.sqlite3_value;
}

func Initialize()	{ C.sqlite3_initialize() }

func Shutdown()	{ C.sqlite3_shutdown() }

func LibVersion() (version string) {
	cver := C.sqlite3_libversion();
	version = C.GoString(cver);

	return;
}

func (s *Statement) ColumnName(column int) (name string) {
	cname := C.sqlite3_column_name(s.cptr, C.int(column));
	name = C.GoString(cname);
	return;
}

func (h *Handle) ErrMsg() (err string)	{ return C.GoString(C.sqlite3_errmsg(h.cptr)) }

func (h *Handle) Open(filename string) (err string) {
	rv := C.sqlite3_open(C.CString(filename), &h.cptr);

	if rv != 0 {
		if nil != h.cptr {
			return h.ErrMsg()
		} else {
			return "Couldn't allocate memory for SQLite3"
		}
	}
	return "";
}

func (h *Handle) Close() {
	C.sqlite3_close(h.cptr);
	h.cptr = nil;
}

func (h *Handle) LastInsertRowID() int64 {
	return int64(C.sqlite3_last_insert_rowid(h.cptr));
}

func (h *Handle) Changes() int
{
	return int(C.sqlite3_changes(h.cptr));
}

func (h *Handle) TotalChanges() int
{
	return int(C.sqlite3_total_changes(h.cptr));
}

func (h *Handle) Prepare(sql string) (s *Statement, err string) {
	s = new(Statement);

	rv := C.sqlite3_prepare(h.cptr, C.CString(sql), -1, &s.cptr, nil);
	if rv != 0 {
		return nil, h.ErrMsg()
	}
	return s, "";
}

// The sqlite3_column_type() routine returns the datatype code for the initial 
// data type of the result column. The returned value is one of SQLITE_INTEGER, 
// SQLITE_FLOAT, SQLITE_TEXT, SQLITE_BLOB, or SQLITE_NULL
func (h *Statement) ColumnType(column int) int {
	return int(C.sqlite3_column_type(h.cptr, C.int(column)));
}

func (h *Statement) ColumnText(column int) string {
	rv := C.sqlite3_column_text(h.cptr, C.int(column));
	return C.GoString((*C.char)(unsafe.Pointer(rv)));
}

func (h *Statement) ColumnInt(column int) int {
	rv := C.sqlite3_column_int(h.cptr, C.int(column));
	return int(rv);
}

func (h *Statement) ColumnInt64(column int) int64 {
	rv := C.sqlite3_column_int64(h.cptr, C.int(column))
	return int64(C.sqlite3_int64(rv))
}

func (h *Statement) ColumnDouble(column int) float64 {
	rv := C.sqlite3_column_double(h.cptr, C.int(column))
	return float64(rv)
}
// bind value for statement handler
func (h *Statement) BindInt(column int, val int) int {
	rv := C.sqlite3_bind_int(h.cptr, C.int(column), C.int(val));
	return int(rv);
}
func (h *Statement) BindText(column int, val string) int {
	rv := C.gosqlite3_bind_text(h.cptr, C.int(column), C.CString(val), C.int(len(val)));
	return int(rv);
}

func (h *Statement) BindInt64(column int, val int64) int {
	return int( C.sqlite3_bind_int64(h.cptr, C.int(column), C.sqlite3_int64(val)) )
}
func (h *Statement) BindDouble(column int, val float64) int {
	rv := C.sqlite3_bind_double(h.cptr, C.int(column), C.double(val))
	return int(rv)
}

// Return the number of columns in the result set returned by the prepared statement.
func (h *Statement) ColumnCount() int {
	return int(C.sqlite3_column_count(h.cptr));
}

// called to delete a prepared statement
func (h *Statement) Finalize() int {
	return int(C.sqlite3_finalize(h.cptr));
}

// this function must be called one or more times to evaluate a prepated statement
func (h *Statement) Step() int {
	return int(C.sqlite3_step(h.cptr));
}
