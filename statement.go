package sqlite3

// #include <sqlite3.h>
// int gosqlite3_bind_text(sqlite3_stmt* s, int p, const char* q, int n) {
//     return sqlite3_bind_text(s, p, q, n, SQLITE_TRANSIENT);
// }
// int gosqlite3_bind_blob(sqlite3_stmt* s, int p, const void* q, int n) {
//     return sqlite3_bind_text(s, p, q, n, SQLITE_TRANSIENT);
// }
import "C"
import "unsafe"

type Statement struct {
	db			*Database
	cptr		*C.sqlite3_stmt
	SQL			string
	timestamp	int64
}

func (s *Statement) ColumnName(column int) string {
	return C.GoString(C.sqlite3_column_name(s.cptr, C.int(column)))
}

func (s *Statement) ColumnType(column int) int {
	return int(C.sqlite3_column_type(s.cptr, C.int(column)))
}

func (s *Statement) Column(column int) (value interface{}) {
	//	Generalise the above so only one function is needed
	switch s.ColumnType(column) {
	case SQLITE_INTEGER:
		value = int64(C.sqlite3_int64(C.sqlite3_column_int64(s.cptr, C.int(column))))
	case SQLITE_FLOAT:
		value = float64(C.sqlite3_column_double(s.cptr, C.int(column)))
	case SQLITE3_TEXT:
		rv := C.sqlite3_column_text(s.cptr, C.int(column))
		value = C.GoString((*C.char)(unsafe.Pointer(rv)))
	case SQLITE_BLOB:
		panic("retrieving blobs is not currently supported")
	case SQLITE_NULL:
		value = nil
	default:
		panic("unknown column type")
	}
	return
}

func (s *Statement) Bind(start_column int, values... interface{}) (rv Errno, index int) {
	for i, val := range values {
		column := start_column + i
		switch val := val.(type) {
		case nil:
			rv = Errno(C.sqlite3_bind_null(s.cptr, C.int(column)))
		case int:
			rv = Errno(C.sqlite3_bind_int(s.cptr, C.int(column), C.int(val)))
		case string:
			rv = Errno(C.gosqlite3_bind_text(s.cptr, C.int(column), C.CString(val), C.int(len(val))))
		case int64:
			rv = Errno(C.sqlite3_bind_int64(s.cptr, C.int(column), C.sqlite3_int64(val)))
		case float64:
			rv = Errno(C.sqlite3_bind_double(s.cptr, C.int(column), C.double(val)))
		default:
			//	save the binary form of the value as a blob
			//	rv = Errno(C.gosqlite3_bind_blob(s.cptr, C.int(column), unsafe.Pointer(&val), C.int(unsafe.Sizeof(val))))
			rv = MISMATCH
		}
		if rv != OK {
			break
		}
	}
	return
}

func (s *Statement) Parameters() int {
	return int(C.sqlite3_bind_parameter_count(s.cptr))
}

func (s *Statement) Columns() int {
	return int(C.sqlite3_column_count(s.cptr))
}

func (s *Statement) Finalize() int {
	return int(C.sqlite3_finalize(s.cptr))
}

func (s *Statement) Step() Errno {
	return Errno(C.sqlite3_step(s.cptr))
}