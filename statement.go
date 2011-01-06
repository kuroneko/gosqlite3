package sqlite3

// #include <sqlite3.h>
import "C"

type Statement struct {
	db			*Database
	cptr		*C.sqlite3_stmt
	SQL			string
	timestamp	int64
}

func (s *Statement) ColumnName(column int) string {
	return Column(column).Name(s)
}

func (s *Statement) ColumnType(column int) int {
	return Column(column).Type(s)
}

func (s *Statement) Column(column int) (value interface{}) {
	return Column(column).Value(s)
}

func (s *Statement) Bind(start_column int, values... interface{}) (rv Errno, index int) {
	column := Column(0)
	for i, v := range values {
		column++
		if rv = column.Bind(s, v); rv != OK {
			index = i
			return
		}
	}
	return
}

func (s *Statement) SQLSource() string {
	return C.GoString(C.sqlite3_sql(s.cptr))
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
