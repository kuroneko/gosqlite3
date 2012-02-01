package sqlite3

// #include <sqlite3.h>
import "C"

// Statement represents a "SQL prepared Statement" also known as "compiled SQL statement".
type Statement struct {
	db			*Database
	cptr		*C.sqlite3_stmt
	timestamp	int64
}

// Parameters returns the number of SQL parameters.
//
// This routine actually returns the index of the largest (rightmost) 
// parameter. For all forms except ?NNN, this will correspond to the number 
// of unique parameters. If parameters of the ?NNN form are used, there may 
// be gaps in the list.
func (s *Statement) Parameters() int {
	return int(C.sqlite3_bind_parameter_count(s.cptr))
}

// Columns returns the number of columns in the result set of the 
// prepared statement.
func (s *Statement) Columns() int {
	return int(C.sqlite3_column_count(s.cptr))
}

// ColumnName returns the name of the column.
func (s *Statement) ColumnName(column int) string {
	return ResultColumn(column).Name(s)
}

// ColumnType returns the type of the column.
func (s *Statement) ColumnType(column int) int {
	return ResultColumn(column).Type(s)
}

// Column returns the value of the column.
func (s *Statement) Column(column int) (value interface{}) {
	return ResultColumn(column).Value(s)
}

// Row returns all values of the row.
func (s *Statement) Row() (values []interface{}) {
	for i := 0; i < s.Columns(); i++ {
		values = append(values, s.Column(i))
	}
	return
}

// Bind replaces the SQL parameters with actual values.
func (s *Statement) Bind(start_column int, values... interface{}) (e error, index int) {
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

// BindAll replaces all SQL parameters with their actual values.
func (s *Statement) BindAll(values... interface{}) (e error, index int) {
	return s.Bind(0, values...)
}

// SQLSource can be used to retrieve a saved copy of the original SQL text 
// used to create the prepared statement -- if that statement was compiled 
// using `Prepare`.
func (s *Statement) SQLSource() (sql string) {
	if s.cptr != nil {
		sql = C.GoString(C.sqlite3_sql(s.cptr))
	}
	return
}

// Finalize is used to delete a prepared statement in the SQLite engine.
func (s *Statement) Finalize() error {
	if e := Errno(C.sqlite3_finalize(s.cptr)); e != OK {
		return e
	}
	return nil
}

// Step must be called one or more times to evaluate the statement after the 
// prepared statement has been prepared.
func (s *Statement) Step(f... func(*Statement, ...interface{})) (e error) {
	r := Errno(C.sqlite3_step(s.cptr))
	switch r {
	case DONE:
		e = nil
	case ROW:
		if f != nil {
			defer func() {
				switch x := recover().(type) {
				case nil:		e = ROW
				case error:		e = x
				default:		e = MISUSE
				}
			}()
			r := s.Row()
			for _, fn := range f {
				fn(s, r...)
			}
		}
	default:
		e = r
	}
	return
}

// All can be used to return all rows of a prepared statement after the 
// statement has been prepared.
func (s *Statement) All(f... func(*Statement, ...interface{})) (c int, e error) {
	for {
		if e = s.Step(f...); e != nil {
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

// Reset may be used to reset the statement to its initial state, ready to
// be re-executed.
//
// Any SQL statement variables that had values bound to them retain 
// their values. Use `ClearBindings` to reset the bindings.
func (s *Statement) Reset() error {
	if e := Errno(C.sqlite3_reset(s.cptr)); e != OK {
		return e
	}
	return nil
}

// ClearBindings is used to reset all parameters to NULL.
func (s *Statement) ClearBindings() error {
	if e := Errno(C.sqlite3_clear_bindings(s.cptr)); e != OK {
		return e
	}
	return nil
}
