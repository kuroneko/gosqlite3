package sqlite3

// #include <sqlite3.h>
// #include <stdlib.h>
// int gosqlite3_bind_text(sqlite3_stmt* s, int p, const char* q, int n) {
//     return sqlite3_bind_text(s, p, q, n, SQLITE_TRANSIENT);
// }
// int gosqlite3_bind_blob(sqlite3_stmt* s, int p, const void* q, int n) {
//     return sqlite3_bind_blob(s, p, q, n, SQLITE_TRANSIENT);
// }
import "C"
import (
	"bytes"
	"encoding/gob"
	"unsafe"
)

type QueryParameter int
func (p QueryParameter) bind_blob(s *Statement, v []byte) error {
	cs := C.CString(string(v))
	defer C.free(unsafe.Pointer(cs))
	return SQLiteError(C.gosqlite3_bind_blob(s.cptr, C.int(p), unsafe.Pointer(&cs), C.int(len(v))))
}

// Bind replaces the literals placed in the SQL statement with the actual 
// values supplied to the function.
//
// The following templates may be replaced by the values:
//   - ?
//   - ?NNN
//   - :VVV
//   - @VVV
//   - $VVV
// In the templates above, NNN represents an integer literal, VVV represents
// an alphanumeric identifier.
func (p QueryParameter) Bind(s *Statement, value interface{}) (e error) {
	switch v := value.(type) {
	case nil:
		e = SQLiteError(C.sqlite3_bind_null(s.cptr, C.int(p)))
	case int:
		e = SQLiteError(C.sqlite3_bind_int(s.cptr, C.int(p), C.int(v)))
	case string:
		e = SQLiteError(C.gosqlite3_bind_text(s.cptr, C.int(p), C.CString(v), C.int(len(v))))
	case int64:
		e = SQLiteError(C.sqlite3_bind_int64(s.cptr, C.int(p), C.sqlite3_int64(v)))
	case float32:
		e = SQLiteError(C.sqlite3_bind_double(s.cptr, C.int(p), C.double(v)))
	case float64:
		e = SQLiteError(C.sqlite3_bind_double(s.cptr, C.int(p), C.double(v)))
	default:
		buffer := new(bytes.Buffer)
		encoder := gob.NewEncoder(buffer)
		if encoder.Encode(value) != nil {
			e = ENCODER
		} else {
			rawbuffer := string(buffer.Bytes())
			cs := C.CString(rawbuffer)
			defer C.free(unsafe.Pointer(cs))
			e = SQLiteError(C.gosqlite3_bind_blob(s.cptr, C.int(p), unsafe.Pointer(cs), C.int(len(rawbuffer))))
		}
	}
	return
}
