package sqlite3

// #include <sqlite3.h>
import "C"
import (
	"bytes"
	"encoding/gob"
	"unsafe"
)

const(
	INTEGER = 1
	FLOAT = 2
	TEXT = 3
	BLOB = 4
	NULL = 5
)


type ResultColumn int

func (c ResultColumn) make_buffer(s *Statement, addr interface{}) (buffer string) {
	switch addr := addr.(type) {
	case *C.uchar:
		buffer = C.GoStringN((*C.char)(unsafe.Pointer(addr)), C.int(c.ByteCount(s)))
	case unsafe.Pointer:
		buffer = C.GoStringN((*C.char)(addr), C.int(c.ByteCount(s)))
	}
	return 
}

func (c ResultColumn) Name(s *Statement) string {
	return C.GoString(C.sqlite3_column_name(s.cptr, C.int(c)))
}

func (c ResultColumn) Type(s *Statement) int {
	return int(C.sqlite3_column_type(s.cptr, C.int(c)))
}

func (c ResultColumn) ByteCount(s *Statement) int {
	return int(C.sqlite3_column_bytes(s.cptr, C.int(c)))
}

func (c ResultColumn) Value(s *Statement) (value interface{}) {
	switch c.Type(s) {
	case INTEGER:
		value = int64(C.sqlite3_int64(C.sqlite3_column_int64(s.cptr, C.int(c))))
	case FLOAT:
		value = float64(C.sqlite3_column_double(s.cptr, C.int(c)))
	case TEXT:
		value = c.make_buffer(s, C.sqlite3_column_text(s.cptr, C.int(c)))
	case BLOB:
		buffer := c.make_buffer(s, C.sqlite3_column_blob(s.cptr, C.int(c)))
		value = gob.NewDecoder(bytes.NewBuffer([]byte(buffer)))
	case NULL:
		value = nil
	default:
		panic("unknown column type")
	}
	return
}