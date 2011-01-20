package sqlite3

// #include <sqlite3.h>
import "C"
import "os"

type Statement struct {
	db			*Database
	cptr		*C.sqlite3_stmt
	timestamp	int64
}

func (s *Statement) Parameters() int {
	return int(C.sqlite3_bind_parameter_count(s.cptr))
}

func (s *Statement) Columns() int {
	return int(C.sqlite3_column_count(s.cptr))
}

func (s *Statement) ColumnName(column int) string {
	return ResultColumn(column).Name(s)
}

func (s *Statement) ColumnType(column int) int {
	return ResultColumn(column).Type(s)
}

func (s *Statement) Column(column int) (value interface{}) {
	return ResultColumn(column).Value(s)
}

func (s *Statement) Row() (values []interface{}) {
	for i := 0; i < s.Columns(); i++ {
		values = append(values, s.Column(i))
	}
	return
}

func (s *Statement) Bind(start_column int, values... interface{}) (e os.Error, index int) {
	column := QueryParameter(start_column)
	for i, v := range values {
		column++
		if e = column.Bind(s, v); e != nil {
			index = i
			return
		}
	}
	return
}

func (s *Statement) BindAll(values... interface{}) (e os.Error, index int) {
	return s.Bind(0, values...)
}

func (s *Statement) SQLSource() (sql string) {
	if s.cptr != nil {
		sql = C.GoString(C.sqlite3_sql(s.cptr))
	}
	return
}

func (s *Statement) Finalize() os.Error {
	if e := Errno(C.sqlite3_finalize(s.cptr)); e != OK {
		return e
	}
	return nil
}

func (s *Statement) Step(f func(*Statement, ...interface{})) (e os.Error) {
	r := Errno(C.sqlite3_step(s.cptr))
	switch r {
	case DONE:
		e = nil
	case ROW:
		if f != nil {
			defer func() {
				switch x := recover().(type) {
				case nil:		e = ROW
				case os.Error:	e = x
				default:		e = MISUSE
				}
			}()
			f(s, s.Row()...)
		}
	default:
		e = r
	}
	return
}

func (s *Statement) All(f func(*Statement, ...interface{})) (c int, e os.Error) {
	for {
		if e = s.Step(f); e != nil {
			if r, ok := e.(Errno); ok {
				switch r {
				case ROW:
					c++
					continue
				default:
					e = r
					break
				}
			}
		} else {
			break
		}
	}
	if e == nil {
		s.Finalize()
	}
	return
}

func (s *Statement) Reset() os.Error {
	if e := Errno(C.sqlite3_reset(s.cptr)); e != OK {
		return e
	}
	return nil
}

func (s *Statement) ClearBindings() os.Error {
	if e := Errno(C.sqlite3_clear_bindings(s.cptr)); e != OK {
		return e
	}
	return nil
}