package sqlite

// #include <sqlite3.h>
import "C"

type Blob struct {
	cptr	*C.sqlite3_blob;
};

type Handle struct {
	chdl	*C.sqlite3;
};

type Statement struct {
	cptr	*C.sqlite3_stmt;
};

type Value struct {
	cptr	*C.sqlite3_value;
};

func Initialize() {
	C.sqlite3_initialize();
}

func Shutdown() {
	C.sqlite3_shutdown();
}

func LibVersion() (version string) {
	cver := C.sqlite3_libversion();
	version = C.GoString(cver);

	return;
}

func (s *Statement)ColumnName(column int) (name string) {
	cname := C.sqlite3_column_name(s.cptr, C.int(column));
	name = C.GoString(cname);
	return;
}

func (h *Handle) Open(filename string) (err string) {
	rv := C.sqlite3_open(C.CString(filename), &h.chdl);

	if rv != 0 {
		if nil != h.chdl {
			return C.GoString(C.sqlite3_errmsg(h.chdl));
		} else {
			return "Couldn't allocate memory for SQLite3";
		}
	}
	return "";
}

func (h *Handle) Close() {
	C.sqlite3_close(h.chdl);
	h.chdl = nil;
}