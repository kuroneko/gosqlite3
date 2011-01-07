package sqlite3

// #include <sqlite3.h>
// int gosqlite3_bind_text(sqlite3_stmt* s, int p, const char* q, int n) {
//     return sqlite3_bind_text(s, p, q, n, SQLITE_TRANSIENT);
// }
// int gosqlite3_bind_blob(sqlite3_stmt* s, int p, const void* q, int n) {
//     return sqlite3_bind_blob(s, p, q, n, SQLITE_TRANSIENT);
// }
import "C"

import "bytes"
import "gob"
import "unsafe"


type Column int
func (c Column) bind_blob(s *Statement, v []byte) Errno {
	return Errno(C.gosqlite3_bind_blob(s.cptr, C.int(c), unsafe.Pointer(C.CString(string(v))), C.int(len(v))))
}

func (c Column) make_buffer(s *Statement, addr interface{}) (buffer string) {
	switch addr := addr.(type) {
	case *C.uchar:
		buffer = C.GoStringN((*C.char)(unsafe.Pointer(addr)), C.int(c.ByteCount(s)))
	case unsafe.Pointer:
		buffer = C.GoStringN((*C.char)(addr), C.int(c.ByteCount(s)))
	}
	return 
}

func (c Column) Name(s *Statement) string {
	return C.GoString(C.sqlite3_column_name(s.cptr, C.int(c)))
}

func (c Column) Type(s *Statement) int {
	return int(C.sqlite3_column_type(s.cptr, C.int(c)))
}

func (c Column) ByteCount(s *Statement) int {
	return int(C.sqlite3_column_bytes(s.cptr, C.int(c)))
}

func (c Column) Value(s *Statement) (value interface{}) {
	switch c.Type(s) {
	case SQLITE_INTEGER:
		value = int64(C.sqlite3_int64(C.sqlite3_column_int64(s.cptr, C.int(c))))
	case SQLITE_FLOAT:
		value = float64(C.sqlite3_column_double(s.cptr, C.int(c)))
	case SQLITE3_TEXT:
		value = c.make_buffer(s, C.sqlite3_column_text(s.cptr, C.int(c)))
	case SQLITE_BLOB:
		buffer := c.make_buffer(s, C.sqlite3_column_blob(s.cptr, C.int(c)))
		value = gob.NewDecoder(bytes.NewBuffer([]byte(buffer)))
	case SQLITE_NULL:
		value = nil
	default:
		panic("unknown column type")
	}
	return
}

func (c Column) Bind(s *Statement, value interface{}) (e Errno) {
	switch v := value.(type) {
	case nil:
		e = Errno(C.sqlite3_bind_null(s.cptr, C.int(c)))
	case int:
		e = Errno(C.sqlite3_bind_int(s.cptr, C.int(c), C.int(v)))
	case string:
		e = Errno(C.gosqlite3_bind_text(s.cptr, C.int(c), C.CString(v), C.int(len(v))))
	case int64:
		e = Errno(C.sqlite3_bind_int64(s.cptr, C.int(c), C.sqlite3_int64(v)))
	case float32:
		e = Errno(C.sqlite3_bind_double(s.cptr, C.int(c), C.double(v)))
	case float64:
		e = Errno(C.sqlite3_bind_double(s.cptr, C.int(c), C.double(v)))
	default:
		buffer := new(bytes.Buffer)
		encoder := gob.NewEncoder(buffer)
		if err := encoder.Encode(value); err != nil {
			e = ENCODER
		} else {
			rawbuffer := string(buffer.Bytes())
			e = Errno(C.gosqlite3_bind_blob(s.cptr, C.int(c), unsafe.Pointer(C.CString(rawbuffer)), C.int(len(rawbuffer))))
		}
	}
	return
}