package sqlite3

// #include <sqlite3.h>
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
	if e := Errno(C.gosqlite3_bind_blob(s.cptr, C.int(p), unsafe.Pointer(C.CString(string(v))), C.int(len(v)))); e != OK {
		return e
	}
	return nil
}

// Bind replaces the literals placed in the SQL statement with the actual 
// values supplied to the function.
//
// The following templates may be replaced by the parameters:
//   - ?
//   - ?NNN
//   - :VVV
//   - @VVV
//   - $VVV
// In the templates above, NNN represents an integer literal, VVV represents
// an alphanumeric identifier.
func (p QueryParameter) Bind(s *Statement, value interface{}) (e error) {
	var rv	Errno
	switch v := value.(type) {
	case nil:
		rv = Errno(C.sqlite3_bind_null(s.cptr, C.int(p)))
	case int:
		rv = Errno(C.sqlite3_bind_int(s.cptr, C.int(p), C.int(v)))
	case string:
		rv = Errno(C.gosqlite3_bind_text(s.cptr, C.int(p), C.CString(v), C.int(len(v))))
	case int64:
		rv = Errno(C.sqlite3_bind_int64(s.cptr, C.int(p), C.sqlite3_int64(v)))
	case float32:
		rv = Errno(C.sqlite3_bind_double(s.cptr, C.int(p), C.double(v)))
	case float64:
		rv = Errno(C.sqlite3_bind_double(s.cptr, C.int(p), C.double(v)))
	default:
		buffer := new(bytes.Buffer)
		encoder := gob.NewEncoder(buffer)
		if encoder.Encode(value) != nil {
			rv = ENCODER
		} else {
			rawbuffer := string(buffer.Bytes())
			rv = Errno(C.gosqlite3_bind_blob(s.cptr, C.int(p), unsafe.Pointer(C.CString(rawbuffer)), C.int(len(rawbuffer))))
		}
	}
	if rv != OK {
		e = rv
	}
	return
}
